package otnstester

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"

	visualize_grpc_pb "github.com/openthread/ot-ns/visualize/grpc/pb"
	"google.golang.org/grpc"

	"github.com/openthread/ot-ns/dispatcher"
	"github.com/stretchr/testify/assert"

	"github.com/chzyer/readline"
	"github.com/openthread/ot-ns/cli/runcli"
	"github.com/openthread/ot-ns/otns_main"
	"github.com/openthread/ot-ns/progctx"
	. "github.com/openthread/ot-ns/types"
	"github.com/openthread/ot-ns/visualize"
	"github.com/simonlingoogle/go-simplelogger"
)

var (
	stdinPipeFile  = "stdin.namedpipe"
	stdoutPipeFile = "stdout.namedpipe"
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

func (ot *OtnsTest) AddNode(role string) NodeId {
	cmd := fmt.Sprintf("add %s", role)
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

func (ot *OtnsTest) shutdown() {

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

func (ot *OtnsTest) Reset() {
	ot.SetSpeed(dispatcher.MaxSimulateSpeed)
	ot.SetPacketLossRatio(0)
	ot.RemoveAllNodes()
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
	assert.Nil(ot, err, "unexpected error")
	if err != nil {
		ot.FailNow()
	}
}

func (ot *OtnsTest) ExpectTrue(value bool, msgAndArgs ...interface{}) {
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

func NewOtnsTest(t *testing.T) *OtnsTest {
	ot := &OtnsTest{
		T:                      t,
		otnsDone:               make(chan struct{}),
		pendingOutput:          make(chan string, 1000),
		pendingVisualizeEvents: make(chan *visualize_grpc_pb.VisualizeEvent, 1000),
	}

	os.Args = append(os.Args, "-log", "debug", "-web=false", "-autogo=false")

	_ = os.Remove(stdinPipeFile)
	_ = os.Remove(stdoutPipeFile)

	err := syscall.Mkfifo(stdinPipeFile, 0644)
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
			ot.shutdown()
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

	grpcConn, err := grpc.Dial("localhost:8999", grpc.WithInsecure())
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
