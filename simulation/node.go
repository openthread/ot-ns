// Copyright (c) 2020-2026, The OTNS Authors.
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

package simulation

import (
	"bufio"
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"io"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"slices"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/pkg/errors"

	"github.com/openthread/ot-ns/dispatcher"
	"github.com/openthread/ot-ns/event"
	"github.com/openthread/ot-ns/logger"
	. "github.com/openthread/ot-ns/types"
)

const (
	NodeExitTimeout      = time.Second * 3
	SendCoapResourceName = "t"
	SendMcastPrefix      = "ff13::deed"
	SendUdpPort          = 10000
)

type Node struct {
	S      *Simulation
	Id     int
	DNode  *dispatcher.Node
	Logger *logger.NodeLogger

	cfg           *NodeConfig
	cmd           *exec.Cmd
	cmdErr        error // store the last CLI command error; nil if none.
	version       string
	threadVersion uint16
	isSendStarted bool
	isExiting     bool
	sendGroupIds  map[int]struct{}

	pendingLines     chan string       // OT node CLI output lines, pending processing.
	pendingEvents    chan *event.Event // OT node emitted events to be processed.
	pipeIn           io.WriteCloser
	pipeOut          io.ReadCloser
	pipeErr          io.ReadCloser
	pipesDone        sync.WaitGroup // tracks active pipeOut/pipeErr reader goroutines
	stdoutReaderDone sync.WaitGroup // signals when lineReaderStdOut has posted its last event
	uartLine         bytes.Buffer   // builds a line based on received UART characters/string-parts.
	uartType         NodeUartType
	uartHasEcho      bool
}

// newNode creates a new simulation node. If unsuccessful, it returns an error != nil and the Node object created
// so far (if node return != nil), or nil if node object wasn't created yet.
func newNode(s *Simulation, nodeid NodeId, cfg *NodeConfig, dnode *dispatcher.Node) (*Node, error) {
	var err error
	var args []string
	var exePath string

	// check executables and construct process args
	if cfg.IsRcp {
		// First check if the to-be-forked RCP executable can be found.
		if !isFile(cfg.ExecutablePath) {
			return nil, fmt.Errorf("target RCP file '%s' not found", cfg.ExecutablePath)
		}
		if !isFileExecutable(cfg.ExecutablePath) {
			return nil, fmt.Errorf("target RCP file '%s' is not executable", cfg.ExecutablePath)
		}
		if cfg.RandomSeed != 0 {
			return nil, fmt.Errorf("random seed != 0 not supported for RCP (got %d)", cfg.RandomSeed)
		}
		// The executable and args formed here are for the Posix host process that will fork an RCP.
		exePath = cfg.HostExePath

		// Flag -d 5 to enable all levels of log messages to be captured in the node's log file.
		// Flag -v to also send log messages to stderr, so OTNS can capture them.
		args = append(args, "-d", "5", "-v")
		// Provide the args: node-id, socket name and random seed, through the
		// SPINEL URL's forkpty-arg query parameter, that can be repeated.
		// TODO: change to url.URL url.Values query builder, but only after ot-cli accepts percent-encoded URLs.
		spinelUrl := fmt.Sprintf("spinel+hdlc+forkpty://%s?forkpty-arg=%d&forkpty-arg=%s",
			cfg.ExecutablePath, nodeid, s.d.GetUnixSocketName())
		args = append(args, spinelUrl)
	} else if cfg.Type == MATTER {
		args = append(args, fmt.Sprintf("--thread-args=%d", nodeid))
		args = append(args, fmt.Sprintf("--thread-args=%s", s.d.GetUnixSocketName()))
		if cfg.RandomSeed != 0 {
			args = append(args, fmt.Sprintf("--thread-args=%d", cfg.RandomSeed))
		}
		exePath = cfg.ExecutablePath
	} else {
		exePath = cfg.ExecutablePath
		args = append(args, strconv.Itoa(nodeid))
		args = append(args, s.d.GetUnixSocketName())
		if cfg.RandomSeed != 0 {
			args = append(args, fmt.Sprintf("%d", cfg.RandomSeed))
		}
	}

	// erase prior NVM state (if desired)
	if !cfg.Restore && !cfg.IsExternal {
		flashFile := fmt.Sprintf("%s/%d_%d.flash", s.cfg.OutputDir, s.cfg.Id, nodeid)
		if err = os.RemoveAll(flashFile); err != nil {
			err = fmt.Errorf("remove OT flash file %s failed: %w", flashFile, err)
			return nil, err
		}
		if cfg.IsRcp {
			eui64 := GetDefaultRcpIeeeEui64(nodeid)
			settingsFile := fmt.Sprintf("%s/%d_%x.data", s.cfg.OutputDir, s.cfg.Id, eui64)
			if err = os.RemoveAll(settingsFile); err != nil {
				err = fmt.Errorf("remove OT settings file %s failed: %w", settingsFile, err)
				return nil, err
			}
		}
		// Note: OTBR files are owned by root - not touched here.
	}

	cmd := exec.CommandContext(context.Background(), exePath, args...)
	cmd.Env = append(os.Environ(), fmt.Sprintf("%s=%d", OtSimulationIdEnv, s.cfg.Id))

	node := &Node{
		S:             s,
		Id:            nodeid,
		Logger:        logger.GetNodeLogger(s.cfg.OutputDir, s.cfg.Id, cfg),
		DNode:         dnode,
		cfg:           cfg,
		cmd:           cmd,
		isExiting:     false,
		pendingLines:  make(chan string, 10000),
		pendingEvents: make(chan *event.Event, 100),
		uartType:      nodeUartTypeVirtualTime,
		uartLine:      bytes.Buffer{},
		uartHasEcho:   false,
		version:       "",
		sendGroupIds:  make(map[int]struct{}),
	}

	node.Logger.SetFileLevel(s.cfg.LogFileLevel)
	node.Logger.Debugf("Node config: type=%s IsMtd=%t IsRcp=%t IsRouter=%t IsBR=%t RxOffWhenIdle=%t", cfg.Type, cfg.IsMtd,
		cfg.IsRcp, cfg.IsRouter, cfg.IsBorderRouter, cfg.RxOffWhenIdle)
	node.Logger.Debugf("  exe cmd : %v", cmd)
	node.Logger.Debugf("  position: (%d,%d,%d)", cfg.X, cfg.Y, cfg.Z)

	if !cfg.IsExternal {
		if node.pipeIn, err = cmd.StdinPipe(); err != nil {
			return node, err
		}

		if node.pipeOut, err = cmd.StdoutPipe(); err != nil {
			return node, err
		}

		if node.pipeErr, err = cmd.StderrPipe(); err != nil {
			return node, err
		}

		if err = cmd.Start(); err != nil {
			return node, err
		}

		if cfg.IsRcp {
			node.uartType = nodeUartTypeRealTime
			node.stdoutReaderDone.Add(1)
			node.pipesDone.Add(1)
			go func() {
				defer node.stdoutReaderDone.Done()
				defer node.pipesDone.Done()
				node.lineReaderStdOut(node.pipeOut)
			}()
		}

		node.pipesDone.Add(1)
		go func() {
			defer node.pipesDone.Done()
			node.lineReaderStdErr(node.pipeErr)
		}()
	}

	return node, err
}

func (node *Node) String() string {
	return GetNodeName(node.Id)
}

func (node *Node) error(err error) {
	if err != nil {
		node.Logger.Error(err)
		node.cmdErr = err
	}
}

// runScript runs a node script on this node, consisting of a series of node CLI commands.
// Returns immediately with error value in case any CLI command fails.
func (node *Node) runScript(cfg []string) error {
	logger.AssertNotNil(cfg)
	if len(cfg) == 0 {
		return nil
	}

	for _, cmd := range cfg {
		cmd = strings.TrimSpace(cmd)
		if len(cmd) == 0 || strings.HasPrefix(cmd, "#") {
			continue // skip empty lines and comments
		}
		if node.CommandResult() != nil {
			return node.CommandResult()
		}
		if node.S.ctx.Err() != nil {
			return nil
		}
		node.Command(cmd)
	}
	return node.CommandResult()
}

func (node *Node) sendSigTerm() error {
	if node.cmd.Process == nil || node.cmd.ProcessState != nil {
		return nil // if not started or already exited
	}
	node.Logger.Tracef("Sending SIGTERM to node process PID %d", node.cmd.Process.Pid)
	return node.cmd.Process.Signal(syscall.SIGTERM)
}

// initiateExit initiates the node exit process, to gracefully shut down the node.
// Must be followed by finalizeExit() at a later point in time.
func (node *Node) initiateExit() {
	if node.isExiting {
		return
	}
	node.isExiting = true

	attemptSendExit := false
	if node.cfg.IsExternal {
		// The node process is external, so we cannot control it here. Disconnecting the DNode socket
		// is used instead. It triggers the Unix socket server GoRoutine to emit a final exit event.
		node.DNode.DisconnectSocket()
	} else if node.cfg.IsRcp {
		attemptSendExit = true
	} else {
		if node.DNode.IsConnected() {
			attemptSendExit = true
		} else {
			node.S.Dispatcher().NotifyCommand(node.Id) // set alive: expecting NodeExit event as final one in queue.
			_ = node.sendSigTerm()
		}
	}

	if attemptSendExit {
		err := node.inputCommand("exit")
		if err != nil { // Only send SIGTERM to the process if the 'exit' CLI command failed.
			_ = node.sendSigTerm()
		}
	}
}

// finalizeExit completes the node exit process, ensuring all resources are released and any
// remaining (exit) log messages are captured in the node log.
func (node *Node) finalizeExit() {
	var err error

	if !node.isExiting {
		node.initiateExit()
	}

	// Close stdin only — no more input to the process, without interrupting the stdout/stderr readers.
	if node.pipeIn != nil {
		_ = node.pipeIn.Close()
	}

	if node.cmd.Process != nil {
		node.Logger.Tracef("Waiting for process PID %d to exit ...", node.cmd.Process.Pid)
		waitDone := make(chan error, 1)
		go func() {
			// cmd.Wait() reaps the process and closes pipeOut/pipeErr (per StdoutPipe/StderrPipe docs),
			// delivering EOF to the reader goroutines so they drain cleanly.
			waitDone <- node.cmd.Wait()
		}()
		select {
		case err = <-waitDone:
			// continue
		case <-time.After(NodeExitTimeout):
			node.Logger.Warn("Node did not exit in time, sending SIGKILL.")
			_ = node.cmd.Process.Kill()
			err = <-waitDone
		}
		// typical err values: nil, "signal: killed" (SIGKILL), "signal: broken pipe", or "exit: ..."
		node.Logger.Tracef("Node process exited. Wait().err=%v", err)
		node.Logger.Debugf("Node process exited.")
	}

	// pipeOut/pipeErr are now closed by cmd.Wait(); wait for the reader goroutines to finish
	// reading any remaining lines, and putting these into eventChan, before finalizing.
	node.pipesDone.Wait()

	// post a task to get remaining (log/UART) items from eventChan and display final node's output
	node.S.PostAsync(func() {
		node.S.waitForSimulation()
		node.DisplayPendingLogEntries()
		node.DisplayPendingLines()
		node.Logger.Close()
	})
}

// inputCommand is a helper method to send a CLI command to the node's UART
func (node *Node) inputCommand(cmd string) error {
	var err error
	node.cmdErr = nil // reset last command error
	cmdBytes := []byte(cmd + "\n")

	switch node.uartType {
	case nodeUartTypeRealTime:
		if _, err = node.pipeIn.Write(cmdBytes); err == nil {
			node.S.Dispatcher().NotifyCommand(node.Id)
		}
	case nodeUartTypeVirtualTime:
		err = node.DNode.SendToVirtualUART(cmdBytes)
	default:
		err = fmt.Errorf("invalid node.uartType: %d", node.uartType)
	}
	return err
}

// CommandNoDone executes the command without necessarily expecting 'Done' (e.g. it's a background command).
// It just reads output lines until the node is asleep. If nevertheless 'Done' is received,
// it returns the flag hasDoneOutput = true.
func (node *Node) CommandNoDone(cmd string) ([]string, bool) {
	hasDoneOutput := false

	err := node.inputCommand(cmd)
	if err != nil {
		node.error(err)
		return []string{}, false
	}

	if node.uartHasEcho {
		_, err = node.expectLine(cmd)
		if err != nil {
			node.error(err)
			return []string{}, false
		}
	}

	output := []string{}
	for {
		line, ok := node.readLine()
		if !ok {
			break
		}
		lineTrimmed := strings.TrimSpace(line)
		if strings.HasPrefix(line, "Error") {
			node.error(errors.New(line))
		} else if lineTrimmed == "Done" {
			// potential background commands that return a 'Done' CLI output at the end are flagged here.
			hasDoneOutput = true
		} else {
			output = append(output, line)
			if len(lineTrimmed) > 0 {
				hasDoneOutput = false
			}
		}
	}
	return output, hasDoneOutput
}

// Command executes the command and waits for 'Done', or an 'Error', at end of output.
func (node *Node) Command(cmd string) []string {
	err := node.inputCommand(cmd)
	if err != nil {
		node.error(err)
		return []string{}
	}

	if node.uartHasEcho {
		_, err = node.expectLine(cmd)
		if err != nil {
			node.error(err)
			return []string{}
		}
	}

	var output []string
	output, err = node.expectLine(doneOrErrorRegexp)
	if err != nil {
		node.error(err)
		return []string{}
	}

	var result string
	logger.AssertTrue(len(output) >= 1) // regexp matched - there's a Done or Error line as the last output line.
	output, result = output[:len(output)-1], output[len(output)-1]
	result = strings.TrimSpace(result)
	if result != "Done" {
		node.error(errors.New(result))
	}
	return output
}

// CommandResult gets the last result of any Command...() call, either nil or an Error.
func (node *Node) CommandResult() error {
	return node.cmdErr
}

// CommandChecked executes a command, does not provide any output lines, but returns error resulting from the cmd.
func (node *Node) CommandChecked(cmd string) error {
	node.Command(cmd)
	return node.CommandResult()
}

func (node *Node) CommandExpectString(cmd string) string {
	output := node.Command(cmd)
	if len(output) != 1 {
		err := fmt.Errorf("expected 1 line, but received %d: %#v", len(output), output)
		node.Logger.Error(err)
		return ""
	}
	return output[0]
}

func (node *Node) CommandExpectInt(cmd string) int {
	s := node.CommandExpectString(cmd)
	var iv int64
	var err error

	if strings.HasPrefix(s, "0x") || strings.HasPrefix(s, "0X") {
		iv, err = strconv.ParseInt(s[2:], 16, 0)
	} else {
		iv, err = strconv.ParseInt(s, 10, 0)
	}

	if err != nil {
		node.Logger.Errorf("parsing unexpected Int number: '%#v'", s)
		return 0
	}
	return int(iv)
}

func (node *Node) CommandExpectHex(cmd string) int {
	s := node.CommandExpectString(cmd)
	var iv int64
	var err error

	iv, err = strconv.ParseInt(s[2:], 16, 0)

	if err != nil {
		node.Logger.Errorf("hex parsing unexpected number: '%#v'", s)
		return 0
	}
	return int(iv)
}

func (node *Node) CommandExpectEnabledOrDisabled(cmd string) bool {
	output := node.CommandExpectString(cmd)
	if output == "Enabled" {
		return true
	} else if output == "Disabled" {
		return false
	} else if len(output) == 0 {
		node.Logger.Errorf("CommandExpectEnabledOrDisabled() did not get data from node")
		return false
	} else {
		node.Logger.Errorf("expect Enabled/Disabled, but read: '%#v'", output)
	}
	return false
}

func (node *Node) GetRfSimParam(param RfSimParam) RfSimParamValue {
	switch param {
	case ParamRxSensitivity,
		ParamCslUncertainty,
		ParamTxInterferer,
		ParamClockDrift,
		ParamCslAccuracy,
		ParamPhyBitrate:
		return node.getOrSetRfSimParam(false, param, 0)
	case ParamCcaThreshold:
		return node.GetCcaThreshold()
	default:
		node.error(fmt.Errorf("unknown RfSim parameter: %d", param))
		return 0
	}
}

func (node *Node) SetRfSimParam(param RfSimParam, value RfSimParamValue) {
	switch param {
	case ParamRxSensitivity:
		if value < RssiMin || value > RssiMax {
			node.error(fmt.Errorf("parameter out of range %d - %d", RssiMin, RssiMax))
			return
		}
		node.getOrSetRfSimParam(true, param, value)
	case ParamCslAccuracy,
		ParamCslUncertainty,
		ParamTxInterferer:
		if value < 0 || value > 255 {
			node.error(fmt.Errorf("parameter out of range 0-255"))
			return
		}
		node.getOrSetRfSimParam(true, param, value)
	case ParamCcaThreshold:
		node.SetCcaThreshold(value)
	case ParamClockDrift:
		if value < math.MinInt16 || value > math.MaxInt16 {
			node.error(fmt.Errorf("parameter out of range %d - %d", math.MinInt16, math.MaxInt16))
			return
		}
		node.getOrSetRfSimParam(true, param, value)
	case ParamPhyBitrate:
		if value < 1 || value > RfSimValueMax {
			node.error(fmt.Errorf("parameter out of range 1 - %d", RfSimValueMax))
			return
		}
		node.getOrSetRfSimParam(true, param, value)
	default:
		node.error(fmt.Errorf("unknown RfSim parameter: %d", param))
	}
}

func (node *Node) getOrSetRfSimParam(isSet bool, param RfSimParam, value RfSimParamValue) RfSimParamValue {
	node.cmdErr = nil
	err := node.DNode.SendRfSimEvent(isSet, param, value)
	if err == nil { // wait for response event
		evt, err := node.expectEvent(event.EventTypeRadioRfSimParamRsp)
		if err != nil {
			node.error(err)
			return value
		}
		if evt.RfSimParamData.Param == param {
			return RfSimParamValue(evt.RfSimParamData.Value)
		}
		node.error(fmt.Errorf("RfSimEvent with unexpected param %d received - expected %d", evt.RfSimParamData.Param, param))
	}
	node.error(err)
	return value
}

func (node *Node) GetCcaThreshold() RfSimParamValue {
	s := node.CommandExpectString("ccathreshold")
	idx := strings.Index(s, " dBm")
	iv, err := strconv.ParseInt(s[0:idx], 10, 0)
	if err != nil {
		node.error(err)
		return RssiInvalid
	}
	return RfSimParamValue(iv)
}

func (node *Node) SetCcaThreshold(thresh RfSimParamValue) {
	if thresh >= RssiMin && thresh <= RssiMax {
		node.Command(fmt.Sprintf("ccathreshold %d", thresh))
	} else {
		node.error(fmt.Errorf("parameter out of range 0-255"))
	}
}

func (node *Node) GetChannel() int {
	return node.CommandExpectInt("channel")
}

func (node *Node) GetEui64() string {
	return node.CommandExpectString("eui64")
}

func (node *Node) GetExtAddr() uint64 {
	s := node.CommandExpectString("extaddr")
	v, err := strconv.ParseUint(s, 16, 64)
	if err != nil {
		if len(s) > 0 {
			node.Logger.Errorf("GetExtAddr() unknown address format: %s", s)
		} else {
			node.Logger.Errorf("GetExtAddr() address not received")
		}
		return InvalidExtAddr
	}
	return v
}

func (node *Node) GetIpAddr() []string {
	addrs := node.Command("ipaddr")
	return addrs
}

func (node *Node) GetIpAddrLinkLocal() []string {
	addrs := node.Command("ipaddr linklocal")
	return addrs
}

func (node *Node) GetIpAddrMleid() []string {
	addrs := node.Command("ipaddr mleid")
	return addrs
}

func (node *Node) GetIpAddrRloc() []string {
	addrs := node.Command("ipaddr rloc")
	return addrs
}

func (node *Node) GetIpAddrSlaac() []string {
	addrs := node.Command("ipaddr -v")
	slaacAddrs := make([]string, 0)
	for _, addr := range addrs {
		idx := strings.Index(addr, " origin:slaac ")
		if idx > 0 {
			slaacAddrs = append(slaacAddrs, addr[0:idx])
		}
	}
	return slaacAddrs
}

func (node *Node) GetRouterEligible() bool {
	return node.CommandExpectEnabledOrDisabled("routereligible")
}

func (node *Node) RouterEligibleEnable() {
	node.Command("routereligible enable")
}

func (node *Node) RouterEligibleDisable() {
	node.Command("routereligible disable")
}

type LeaderData struct {
	PartitionID       int
	Weighting         int
	DataVersion       int
	StableDataVersion int
	LeaderRouterID    int
}

func (node *Node) Ping(addr string, payloadSize int, count int, interval float64, hopLimit int) {
	cmd := fmt.Sprintf("ping async %s %d %d %f %d", addr, payloadSize, count, interval, hopLimit)
	node.Command(cmd)
}

func (node *Node) GetNetworkKey() string {
	return node.CommandExpectString("networkkey")
}

func (node *Node) GetMode() string {
	// TODO: return Mode type rather than just string
	return node.CommandExpectString("mode")
}

func (node *Node) SetMode(mode string) {
	node.Command(fmt.Sprintf("mode %s", mode))
}

func (node *Node) GetType() string {
	return node.cfg.Type
}

func (node *Node) GetPanid() uint16 {
	return uint16(node.CommandExpectInt("panid"))
}

func (node *Node) GetRloc16() uint16 {
	return uint16(node.CommandExpectHex("rloc16"))
}

func (node *Node) GetState() string {
	return node.CommandExpectString("state")
}

// SendInit inits the node to participate in OTNS 'send' command sending or receiving.
func (node *Node) SendInit() error {
	if node.isSendStarted {
		return nil
	}
	if err := node.CommandChecked("udp open"); err != nil {
		return err
	}
	if err := node.UdpBindAny(SendUdpPort); err != nil {
		return err
	}
	if err := node.CommandChecked("coap start"); err != nil {
		return err
	}
	if err := node.CommandChecked(fmt.Sprintf("coap resource %s", SendCoapResourceName)); err != nil {
		return err
	}
	node.isSendStarted = true
	return nil
}

// SendReset resets the node after, or before, using a series of OTNS 'send' commands.
func (node *Node) SendReset() error {
	if !node.isSendStarted {
		return nil
	}
	if err := node.CommandChecked("coap stop"); err != nil {
		return err
	}
	if err := node.CommandChecked("udp close"); err != nil {
		return err
	}
	// clear any mcast group memberships
	for gid := range node.sendGroupIds {
		if err := node.SendGroupMembership(gid, false); err != nil {
			return err
		}
	}
	node.isSendStarted = false

	return nil
}

// SendGroupMembership modifies multicast group membership for the groups used by the OTNS 'send' command.
func (node *Node) SendGroupMembership(groupId int, isMember bool) error {
	addOrDel := "del"
	if isMember {
		addOrDel = "add"
	}
	addr := fmt.Sprintf("%s:%x", SendMcastPrefix, groupId)
	cmd := fmt.Sprintf("ipmaddr %s %s", addOrDel, addr)
	err := node.CommandChecked(cmd)
	if err != nil {
		if !isMember && strings.HasPrefix(err.Error(), "Error 23") {
			return nil // OT err address not found - no problem: already a non-member.
		}
		return err
	}
	if isMember {
		node.sendGroupIds[groupId] = struct{}{}
	} else {
		delete(node.sendGroupIds, groupId)
	}
	return err
}

func (node *Node) UdpSend(addr string, port int, data []byte) error {
	cmd := fmt.Sprintf("udp send %s %d -x %s", addr, port, hex.EncodeToString(data))
	return node.CommandChecked(cmd)
}

func (node *Node) UdpSendTestData(addr string, port int, dataSize int) error {
	cmd := fmt.Sprintf("udp send %s %d -s %d", addr, port, dataSize)
	return node.CommandChecked(cmd)
}

func (node *Node) UdpBindAny(port int) error {
	cmd := fmt.Sprintf("udp bind :: %d", port)
	return node.CommandChecked(cmd)
}

func (node *Node) CoapPostTestData(addr string, uri string, confirmable bool, dataSize int) error {
	payloadStr := randomString(dataSize)
	conNonStr := "non"
	if confirmable {
		conNonStr = "con"
	}
	cmd := fmt.Sprintf("coap post %s %s %s %s", addr, uri, conNonStr, payloadStr)
	return node.CommandChecked(cmd)
}

// GetThreadVersion gets the Thread version integer of the OpenThread node.
func (node *Node) GetThreadVersion() uint16 {
	if node.threadVersion == 0 { // lazy init
		node.threadVersion = uint16(node.CommandExpectInt("thread version"))
	}
	return node.threadVersion
}

// GetVersion gets the version string of the OpenThread node.
func (node *Node) GetVersion() string {
	if node.version == "" { // lazy init
		node.version = node.CommandExpectString("version")
	}
	return node.version
}

func (node *Node) GetExecutablePath() string {
	return node.cfg.ExecutablePath
}

func (node *Node) GetExecutableName() string {
	return filepath.Base(node.cfg.ExecutablePath)
}

func (node *Node) GetCounters(counterType string, keyPrefix string) NodeCounters {
	lines := node.Command("counters " + counterType)
	res := make(NodeCounters)
	for _, line := range lines {
		kv := strings.Split(line, ": ")
		if len(kv) != 2 {
			node.Logger.Errorf("GetCounters(): unexpected data '%v'", line)
			return nil
		}
		val, err := strconv.ParseUint(kv[1], 10, 64)
		if err != nil {
			node.Logger.Errorf("GetCounters(): unexpected value string '%v' (not int)", kv[1])
			return nil
		}
		key := strings.ReplaceAll(kv[0], " ", "")
		res[keyPrefix+key] = val
	}
	return res
}

// processUartData is called by the Simulation to deliver new UART data (from the OT node) to the sim-node.
func (node *Node) processUartData(data []byte) {
	node.uartLine.Write(data)

	// find completed UART line(s) and push into the node.pendingLines queue.
	for {
		buf := node.uartLine.Bytes()
		if bytes.HasPrefix(buf, OtCliPrompt) {
			node.uartLine.Next(OtCliPromptLen) // consume the OT prompt(s), if any
			continue
		}

		idx := bytes.IndexByte(buf, '\n')
		if idx == -1 {
			break // any remaining data stays in node.uartLine for next time.
		}
		lineBytes := node.uartLine.Next(idx + 1)
		lineStr := string(bytes.TrimRight(lineBytes, "\r\n"))

		select {
		case node.pendingLines <- lineStr:
			// ok, written (highly likely case)
		default: // in case pendingLines is full, empty it first by doing an early display.
			node.DisplayPendingLines()
			node.pendingLines <- lineStr
		}
	}
}

func (node *Node) onStart() {
	if node.Logger.IsLevelVisible(logger.InfoLevel) {
		node.Logger.Infof("started, panid=0x%04x, chan=%d, eui64=%#v, extaddr=%#v, state=%s, key=%#v, mode=%v",
			node.GetPanid(), node.GetChannel(), node.GetEui64(), node.GetExtAddr(), node.GetState(),
			node.GetNetworkKey(), node.GetMode())
		node.Logger.Infof("         version=%s", node.GetVersion())
	}
	if node.cfg.Type == WIFI {
		node.SetRfSimParam(ParamTxInterferer, defaultWiFiTxInterfererPercentage)
	}
}

// lineReaderStdErr reads the StdErr of any OT nodes and turns each line into a log event.
// For RCP+Posix, OTNS status push lines will be detected, since these are routed as log entries.
func (node *Node) lineReaderStdErr(reader io.Reader) {
	syslogPrefix := ""
	scanner := bufio.NewScanner(reader)
	scanner.Split(bufio.ScanLines)

	for scanner.Scan() {
		line := scanner.Text()

		if node.cfg.IsRcp {
			// ot-cli RCPs have a syslog prefix that is removed here.
			if syslogPrefix == "" {
				if prefix := logger.ParseSyslogPrefix(line); prefix != "" {
					syslogPrefix = prefix // prefix remains identical for node's lifetime
				}
			}
			if syslogPrefix != "" && strings.HasPrefix(line, syslogPrefix) {
				line = line[len(syslogPrefix):]
			}
			if isStatusPush, status := logger.ParseOtnsStatusPush(line); isStatusPush {
				ev := &event.Event{
					Delay:  0,
					Type:   event.EventTypeStatusPush,
					Data:   []byte(status),
					NodeId: node.Id,
				}
				node.S.Dispatcher().PostEventAsync(ev)
			}
		}

		ev := &event.Event{
			Delay:  0,
			Type:   event.EventTypeLogWriteHost,
			Data:   []byte(line),
			NodeId: node.Id,
		}
		node.S.Dispatcher().PostEventAsync(ev)
	}

	// For RCP nodes, wait for the stdout reader goroutine to post all its events before signaling
	// disconnect. Since eventChan is FIFO, this guarantees the dispatcher gets all stdout log
	// events (e.g. RCP exit messages) in the event queue before EventTypeUartDisconnected is posted.
	node.stdoutReaderDone.Wait()

	// send an event to the queue to unblock waitForSimulation() in case it is waiting for new
	// events from this node. The event handler will set the node to sleeping/disconnected, thereby unblocking
	// any ongoing wait for the node.
	node.S.Dispatcher().PostEventAsync(&event.Event{
		Delay:  0,
		Type:   event.EventTypeUartDisconnected,
		NodeId: node.Id,
	})
}

// lineReaderStdOut is a goroutine to read stdout lines from a Posix host CLI process (i.e. for RCP nodes
// only) and turn these into one UART-write events. Any log lines are filtered out: these are already
// handled by lineReaderStdErr, which is a faster path (pipe that is not kernel-buffered).
func (node *Node) lineReaderStdOut(reader io.Reader) {
	logger.AssertTrue(node.cfg.IsRcp)
	scanner := bufio.NewScanner(reader)
	scanner.Split(bufio.ScanLines)

	for scanner.Scan() {
		line := scanner.Text()

		if isOtLogLine, _ := logger.ParseOtLogLine(line); !isOtLogLine {
			ev := &event.Event{
				Delay:  0,
				Type:   event.EventTypeUartWrite,
				NodeId: node.Id,
				Data:   []byte(line + "\n"),
			}
			node.S.Dispatcher().PostEventAsync(ev)
		}
	}
}

// expectEvent waits for an event of the specified type to arrive from the OT node.
func (node *Node) expectEvent(evtType event.EventType) (*event.Event, error) {
	done := node.S.ctx.Done()

	node.S.waitForSimulation()

	for {
		select {
		case evt := <-node.pendingEvents:
			if evt.Type == evtType {
				node.Logger.Tracef("expectEvent() received: %v", evt)
				return evt, nil
			}
			node.Logger.Warnf("expectEvent() received unexpected event, discarding: %v", evt)
			continue
		case <-done:
			return nil, CommandInterruptedError
		default:
			return nil, fmt.Errorf("expectEvent: expected event type %d not received", evtType)
		}
	}
}

// readLine attempts to read a line from the node. Returns false when there is no more line to return.
func (node *Node) readLine() (string, bool) {
	select {
	case line, ok := <-node.pendingLines:
		if !ok {
			return "", false
		}
		node.Logger.Tracef("readLine() UART: %s", line)
		return line, true
	default:
		break
	}

	node.S.waitForSimulation()

	select {
	case line, ok := <-node.pendingLines:
		if !ok {
			return "", false
		}
		node.Logger.Tracef("readLine() UART: %s", line)
		return line, true
	default:
		return "", false
	}
}

// expectLine reads potentially multiple lines from the node until it finds a matching line, or until there are
// no more lines to read. The matching line is returned as the final line in the output string array. The expected
// line can be defined using various data types as detailed in isLineMatch().
func (node *Node) expectLine(line interface{}) ([]string, error) {
	output := []string{}
	done := node.S.ctx.Done()
	eventReceiver := node.S.Dispatcher().EventReceiver()
	deadline := time.After(dispatcher.DefaultReadTimeout)

	for {
		select {
		case readLine, ok := <-node.pendingLines:
			if !ok {
				return output, NodeExitedError
			}
			node.Logger.Tracef("expectLine() UART: %s", readLine)
			output = append(output, readLine)
			if node.isLineMatch(readLine, line) { // found a matching line
				return output, nil
			}
		case evt := <-eventReceiver:
			node.S.Dispatcher().HandleEvent(evt)
		case <-deadline:
			return output, fmt.Errorf("expectLine timeout: did not receive expected '%v'", line)
		case <-done:
			return []string{}, CommandInterruptedError
		}
	}
}

func (node *Node) isLineMatch(line string, _expectedLine interface{}) bool {
	switch expectedLine := _expectedLine.(type) {
	case string:
		return line == expectedLine
	case *regexp.Regexp:
		return expectedLine.MatchString(line)
	case []string:
		for _, s := range expectedLine {
			if s == line {
				return true
			}
		}
	default:
		node.Logger.Panicf("unknown data type %v, expected string, Regexp or []string", expectedLine)
	}
	return false
}

// setupCli sets up and tests CLI communication (via PTY, UART, virtual-UART, or other) before
// first-time CLI command usage.
func (node *Node) setupCli() error {
	var testCmdOutput string
	var expectedTestCmdOutput []string

	testCmd := "ifconfig"
	expectedTestCmdOutput = []string{"down"}

	// Sending initial '\n' is required for MacOS terminal setup for the real-time UART. It will trigger
	// the CLI to write the prompt '> ' as output without a newline character. The prompt gets filtered out by
	// node.processUartData. In Linux, we don't send '\n' here because it would otherwise cause CLI lock-up.
	if node.uartType == nodeUartTypeRealTime && runtime.GOOS == "darwin" {
		_, err := node.pipeIn.Write([]byte("\n"))
		if err != nil {
			return fmt.Errorf("internal error on node.pipeIn.Write(): %w", err)
		}
	}

	// use a test command to check terminal's default echo behavior - and record this.
	outputLines := node.Command(testCmd)
	if node.cmdErr != nil {
		return node.cmdErr
	}

	switch len(outputLines) {
	case 0:
		return fmt.Errorf("node did not provide any output for '%s'", testCmd)
	case 1:
		testCmdOutput = outputLines[0]
	case 2:
		logger.AssertEqual(testCmd, outputLines[0])
		node.uartHasEcho = true
		testCmdOutput = outputLines[1]
	default:
		return fmt.Errorf("received unexpected (longer) node output: %v", outputLines)
	}

	if !slices.Contains(expectedTestCmdOutput, testCmdOutput) {
		return fmt.Errorf("node did not provide expected output for '%s' (%v), but: '%s'", testCmd, expectedTestCmdOutput, testCmdOutput)
	}
	node.Logger.Debugf("setupCli done: uartHasEcho=%t", node.uartHasEcho)

	return nil
}

func (node *Node) setupMode() {
	if node.cfg.IsRouter {
		// routers should be full functional and rx always on
		logger.AssertFalse(node.cfg.IsMtd)
		logger.AssertFalse(node.cfg.RxOffWhenIdle)
	}

	// only MED can use RxOffWhenIdle
	logger.AssertTrue(!node.cfg.RxOffWhenIdle || node.cfg.IsMtd)

	// Externally launched node does not receive 'mode' configuration
	if node.cfg.IsExternal {
		return
	}

	mode := ""
	if !node.cfg.RxOffWhenIdle {
		mode += "r"
	}
	if !node.cfg.IsMtd {
		mode += "d"
	}
	mode += "n"

	node.SetMode(mode)

	if !node.cfg.IsRouter && !node.cfg.IsMtd {
		node.RouterEligibleDisable()
	}
}

func (node *Node) DisplayPendingLines() {
	prefix := ""
loop:
	for {
		select {
		case line, ok := <-node.pendingLines:
			if !ok {
				break loop
			}
			if len(prefix) == 0 && node.S.cmdRunner.GetNodeContext() != node.Id {
				prefix = node.String() // lazy init of node-specific prefix
			}
			if node.isExiting && line == "exit" {
				continue // suppress the echo of an 'exit' CLI command
			}
			logger.Println(prefix + line)
		default:
			break loop
		}
	}
}

func (node *Node) DisplayPendingLogEntries() {
	node.Logger.DisplayPendingLogEntries(node.S.Dispatcher().CurTime)
}
