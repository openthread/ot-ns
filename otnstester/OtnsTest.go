// Copyright (c) 2020-2023, The OTNS Authors.
// All rights reserved.
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions are met:
// 1. Redistributions of source code must retain the above copyright
//    notice, this list of conditions and the following disclaimer.
// 2. Redistributions in binary form must reproduce the above copyright
//    notice, this list of conditions and the following disclaimer in the
//    documentation and/or other materials provided with the distribution.
// 3. Neither the name of the copyright holder nor the
//    names of its contributors may be used to endorse or promote products
//    derived from this software without specific prior written permission.
//
// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
// AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
// IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE
// ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE
// LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR
// CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF
// SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS
// INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN
// CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE)
// ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE
// POSSIBILITY OF SUCH DAMAGE.

package otnstester

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"

	"google.golang.org/grpc/credentials/insecure"

	visualize_grpc_pb "github.com/openthread/ot-ns/visualize/grpc/pb"
	"google.golang.org/grpc"

	"github.com/chzyer/readline"
	"github.com/openthread/ot-ns/cli/runcli"
	"github.com/openthread/ot-ns/dispatcher"
	"github.com/openthread/ot-ns/otns_main"
	"github.com/openthread/ot-ns/progctx"
	. "github.com/openthread/ot-ns/types"
	"github.com/openthread/ot-ns/visualize"
	"github.com/simonlingoogle/go-simplelogger"
	"github.com/stretchr/testify/assert"
)

var (
	stdinPipeFile               = "stdin.namedpipe"
	stdoutPipeFile              = "stdout.namedpipe"
	otnsTestSingleton *OtnsTest = nil
)

type OtnsTest struct {
	*testing.T

	stdin                  *os.File
	stdout                 *os.File
	otnsDone               chan struct{}
	pendingOutput          chan string
	pendingVisualizeEvents chan *visualize_grpc_pb.VisualizeEvent
	stdinCloser            *readline.CancelableStdin
	ctx                    *progctx.ProgCtx
	grpcClient             visualize_grpc_pb.VisualizeGrpcServiceClient
	visualizeStream        visualize_grpc_pb.VisualizeGrpcService_VisualizeClient
}

func (ot *OtnsTest) Go(duration time.Duration) {
	seconds := float64(duration) / float64(time.Second)
	ot.sendCommand(fmt.Sprintf("go %f", seconds))
	ot.expectDone()
}

func (ot *OtnsTest) Join() {
	<-ot.otnsDone
}

func (ot *OtnsTest) AddNode(role string, x int, y int) NodeId {
	cmd := fmt.Sprintf("add %s x %d y %d", role, x, y)
	ot.sendCommand(cmd)
	return ot.expectCommandResultInt()
}

// AddNode with radio-range (rr)
func (ot *OtnsTest) AddNodeRr(role string, x int, y int, rr int) NodeId {
	cmd := fmt.Sprintf("add %s x %d y %d rr %d", role, x, y, rr)
	ot.sendCommand(cmd)
	return ot.expectCommandResultInt()
}

func (ot *OtnsTest) sendCommand(cmd string) {
	simplelogger.Infof("> %s", cmd)
	_, err := ot.stdin.WriteString(cmd + "\n")
	simplelogger.PanicIfError(err)
}

func (ot *OtnsTest) sendCommandf(format string, args ...interface{}) {
	cmd := fmt.Sprintf(format, args...)
	ot.sendCommand(cmd)
}

func (ot *OtnsTest) stdoutReadRoutine() {
	simplelogger.Infof("OTNS Stdout reader started.")

	scanner := bufio.NewScanner(ot.stdout)
	scanner.Split(bufio.ScanLines)

	for scanner.Scan() {
		simplelogger.Infof("read stdout: %#v", scanner.Text())
		ot.pendingOutput <- scanner.Text()
	}
}

func (ot *OtnsTest) expectDone() {
	ot.expectCommandResultLines()
}

func (ot *OtnsTest) expectCommandResultLines() (output []string) {
loop:
	for {
		line := <-ot.pendingOutput

		if line == "Done" {
			break loop
		} else if strings.HasPrefix(line, "Error") {
			simplelogger.Panicf("%s", line)
		} else {
			output = append(output, line)
		}
	}

	simplelogger.Infof("expectCommandResultLines: %#v", output)
	return
}

func (ot *OtnsTest) expectCommandResultInt() int {
	lines := ot.expectCommandResultLines()
	simplelogger.AssertTrue(len(lines) == 1)

	v, err := strconv.Atoi(lines[0])
	simplelogger.PanicIfError(err)

	return v
}

func (ot *OtnsTest) Shutdown() {
	ot.ctx.Cancel(nil)
	ot.Join()
}

func (ot *OtnsTest) SetSpeed(speed int) {
	ot.sendCommandf("speed %d", speed)
	ot.expectDone()
}

func (ot *OtnsTest) Start(testFunc string) {
	ot.Reset()
	simplelogger.Infof("Go test Start(): %v", testFunc)
}

func (ot *OtnsTest) Reset() {
	ot.SetSpeed(dispatcher.MaxSimulateSpeed)
	ot.SetPacketLossRatio(0)
	ot.RemoveAllNodes()
	ot.Go(time.Second)
}

func (ot *OtnsTest) SetPacketLossRatio(ratio float32) {
	ot.sendCommandf("plr %f", ratio)
	ot.expectDone()
}

func (ot *OtnsTest) RemoveAllNodes() {
	nodes := ot.ListNodes()
	simplelogger.Infof("Remove all nodes: %+v", nodes)
	for nodeid := range nodes {
		ot.DeleteNode(nodeid)
	}
}

type NodeInfo struct {
	ExtAddr uint64
	Rloc16  uint16
	X, Y    int
	Failed  bool
}

func (ot *OtnsTest) ListNodes() map[NodeId]*NodeInfo {
	ot.sendCommand("nodes")
	lines := ot.expectCommandResultLines()
	nodes := map[NodeId]*NodeInfo{}

	for _, line := range lines {
		var err error
		var id NodeId
		var extaddr uint64
		var rloc16 uint16
		var x, y int
		failed := false

		for _, sec := range strings.Split(line, "\t") {
			kv := strings.Split(sec, "=")

			switch kv[0] {
			case "id":
				id, err = strconv.Atoi(kv[1])
				ot.ExpectNoError(err)
			case "extaddr":
				extaddr, err = strconv.ParseUint(kv[1], 16, 64)
				ot.ExpectNoError(err)
			case "rloc16":
				var value uint64
				value, err = strconv.ParseUint(kv[1], 16, 16)
				ot.ExpectNoError(err)
				rloc16 = uint16(value)
			case "x":
				x, err = strconv.Atoi(kv[1])
				ot.ExpectNoError(err)
			case "y":
				y, err = strconv.Atoi(kv[1])
				ot.ExpectNoError(err)
			case "failed":
				ot.ExpectTrue(kv[1] == "false" || kv[1] == "true")
				failed = kv[1] == "failed"
			}
		}
		nodes[id] = &NodeInfo{
			ExtAddr: extaddr,
			Rloc16:  rloc16,
			X:       x,
			Y:       y,
			Failed:  failed,
		}
	}

	return nodes
}

func (ot *OtnsTest) ExpectNoError(err error) {
	if err != nil {
		ot.Shutdown()
	}
	assert.Nil(ot, err, "unexpected error")
	if err != nil {
		ot.FailNow()
	}
}

func (ot *OtnsTest) ExpectTrue(value bool, msgAndArgs ...interface{}) {
	if !value {
		ot.Shutdown()
	}
	assert.True(ot, value, msgAndArgs...)

	if !value {
		ot.FailNow()
	}
}

func (ot *OtnsTest) DeleteNode(ids ...NodeId) {
	if len(ids) == 0 {
		return
	}

	cmd := "del"
	for _, id := range ids {
		cmd = cmd + fmt.Sprintf(" %d", id)
	}
	ot.executeCommand(cmd)
}

func (ot *OtnsTest) executeCommand(cmd string) []string {
	ot.sendCommand(cmd)
	return ot.expectCommandResultLines()
}

func (ot *OtnsTest) GetNodeState(id NodeId) string {
	lines := ot.executeCommandNodeContext(id, "state")
	ot.ExpectTrue(len(lines) == 1)
	return lines[0]
}

func (ot *OtnsTest) executeCommandNodeContext(id NodeId, cmd string) []string {
	return ot.executeCommand(fmt.Sprintf("node %d \"%s\"", id, cmd))
}

func (ot *OtnsTest) Command(cmd string) []string {
	ot.sendCommand(cmd)
	return ot.expectCommandResultLines()
}

func (ot *OtnsTest) Commandf(format string, args ...interface{}) []string {
	ot.sendCommandf(format, args...)
	return ot.expectCommandResultLines()
}

func (ot *OtnsTest) visualizeStreamReadRoutine() {
	if ot.visualizeStream == nil {
		simplelogger.Errorf("No ot.visualizeStream was created yet (due to timeout)")
		return
	}
	vctx := ot.visualizeStream.Context()

	for vctx.Err() == nil {
		evt, err := ot.visualizeStream.Recv()

		ot.ExpectTrue(err == nil || ot.visualizeStream.Context().Err() != nil)
		if err == nil {
			simplelogger.Warnf("Visualize: %+v", evt)
			ot.pendingVisualizeEvents <- evt
		}
	}
}

func (ot *OtnsTest) ExpectVisualizeEvent(match func(evt *visualize_grpc_pb.VisualizeEvent) bool) {
	deadline := time.After(time.Second * 10)
	for {
		select {
		case evt := <-ot.pendingVisualizeEvents:
			if match(evt) {
				return
			}
		case <-deadline:
			ot.ExpectTrue(false, "ExpectVisualizeEvent timeout")
		}
	}
}

func (ot *OtnsTest) ExpectVisualizeAddNode(nodeid NodeId, x int, y int, radioRange int) {
	ot.ExpectVisualizeEvent(func(evt *visualize_grpc_pb.VisualizeEvent) bool {
		addNode := evt.GetAddNode()
		if addNode == nil {
			return false
		}

		return addNode.NodeId == int32(nodeid) && addNode.X == int32(x) && addNode.Y == int32(y) && addNode.RadioRange == int32(radioRange)
	})
}

func Instance(t *testing.T) *OtnsTest {
	if otnsTestSingleton == nil {
		otnsTestSingleton = NewOtnsTest(t)
	}
	return otnsTestSingleton
}

func NewOtnsTest(t *testing.T) *OtnsTest {
	// ensure test is run from the repo base directory.
	_, filename, _, _ := runtime.Caller(0)
	dir := path.Join(path.Dir(filename), "..")
	err := os.Chdir(dir)
	if err != nil {
		panic(err)
	}

	ot := &OtnsTest{
		T:                      t,
		otnsDone:               make(chan struct{}),
		pendingOutput:          make(chan string, 1000),
		pendingVisualizeEvents: make(chan *visualize_grpc_pb.VisualizeEvent, 1000),
	}

	os.Args = append(os.Args, "-log", "debug", "-web=false", "-autogo=false", "-watch", "info")

	_ = os.Remove(stdinPipeFile)
	_ = os.Remove(stdoutPipeFile)

	err = syscall.Mkfifo(stdinPipeFile, 0644)
	simplelogger.PanicIfError(err)

	ot.stdin, err = os.OpenFile(stdinPipeFile, os.O_RDWR, os.ModeNamedPipe)
	simplelogger.PanicIfError(err)
	ot.stdinCloser = readline.NewCancelableStdin(ot.stdin)

	err = syscall.Mkfifo(stdoutPipeFile, 0644)
	simplelogger.PanicIfError(err)

	ot.stdout, err = os.OpenFile(stdoutPipeFile, os.O_RDWR, os.ModeNamedPipe)
	simplelogger.PanicIfError(err)

	ot.ctx = progctx.New(context.Background())

	go func() {
		defer func() {
			simplelogger.Infof("OTNS exited.")
			close(ot.otnsDone)
		}()

		otns_main.Main(ot.ctx, func(ctx *progctx.ProgCtx, args *otns_main.MainArgs) visualize.Visualizer {
			return nil
		}, &runcli.CliOptions{
			EchoInput: false,
			Stdin:     ot.stdin,
			Stdout:    ot.stdout,
		})
	}()

	grpcConn, err := grpc.Dial("localhost:8999", grpc.WithTransportCredentials(insecure.NewCredentials()))
	ot.ExpectNoError(err)

	grpcClient := visualize_grpc_pb.NewVisualizeGrpcServiceClient(grpcConn)
	ot.grpcClient = grpcClient

	deadline := time.Now().Add(time.Second * 10)
	for time.Now().Before(deadline) {
		visualizeStream, err := grpcClient.Visualize(ot.ctx, &visualize_grpc_pb.VisualizeRequest{})
		if err != nil {
			continue
		}

		ot.visualizeStream = visualizeStream
		break
	}

	go ot.stdoutReadRoutine()
	go ot.visualizeStreamReadRoutine()
	return ot
}
