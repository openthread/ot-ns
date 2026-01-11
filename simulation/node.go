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
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/pkg/errors"

	"github.com/openthread/ot-ns/dispatcher"
	"github.com/openthread/ot-ns/event"
	"github.com/openthread/ot-ns/logger"
	. "github.com/openthread/ot-ns/types"
)

const (
	DefaultCommandTimeout = time.Second * 10
	NodeExitTimeout       = time.Second * 3
	SendCoapResourceName  = "t"
	SendMcastPrefix       = "ff13::deed"
	SendUdpPort           = 10000
)

type Node struct {
	S      *Simulation
	Id     int
	DNode  *dispatcher.Node
	Logger *logger.NodeLogger

	nameStr       string
	cfg           *NodeConfig
	cmd           *exec.Cmd
	isExiting     bool
	cmdErr        error // store the last CLI command error; nil if none.
	version       string
	threadVersion uint16
	isSendStarted bool
	sendGroupIds  map[int]struct{}

	pendingLines  chan string       // OT node CLI output lines, pending processing.
	pendingEvents chan *event.Event // OT-RFSIM incoming events for sim-node, pending processing.
	pipeIn        io.WriteCloser
	pipeOut       io.ReadCloser
	pipeErr       io.ReadCloser
	uartLine      bytes.Buffer // builds a line based on received UART characters/string-parts.
	uartType      NodeUartType
	uartHasEcho   bool
}

// newNode creates a new simulation node. If unsuccessful, it returns an error != nil and the Node object created
// so far which may be nil if the node object wasn't created yet.
func newNode(s *Simulation, nodeid NodeId, cfg *NodeConfig, dnode *dispatcher.Node) (*Node, error) {
	var err error

	if !cfg.Restore {
		flashFile := fmt.Sprintf("%s/%d_%d.flash", s.cfg.OutputDir, s.cfg.Id, nodeid)
		if err = os.RemoveAll(flashFile); err != nil {
			logger.Errorf("Remove flash file %s failed: %+v", flashFile, err)
			return nil, err
		}
		if cfg.IsRcp {
			eui64 := GetDefaultRcpIeeeEui64(nodeid)
			flashFile = fmt.Sprintf("%s/%d_%x.data", s.cfg.OutputDir, s.cfg.Id, eui64)
			if err = os.RemoveAll(flashFile); err != nil {
				logger.Errorf("Remove flash file %s failed: %+v", flashFile, err)
				return nil, err
			}
		}
	}

	var args []string
	var exePath string

	if !cfg.IsRcp {
		exePath = cfg.ExecutablePath
		args = append(args, strconv.Itoa(nodeid))
		args = append(args, s.d.GetUnixSocketName())
		if cfg.RandomSeed != 0 {
			args = append(args, fmt.Sprintf("%d", cfg.RandomSeed))
		}
	} else {
		// The executable and args formed here are for the Posix host process that will fork an RCP.
		exePath = cfg.HostExePath
		// Flag -v to send OT log messages also to stderr (and not only syslog)
		// Flag -d 5 to enable all levels of log messages to be captured in the node's log file.
		args = append(args, "-d", "5")
		// Provide the args: node-id, socket name and random seed, through the
		// SPINEL URL's forkpty-arg query parameter, that can be repeated.
		spinelUrl := fmt.Sprintf("spinel+hdlc+forkpty://%s?forkpty-arg=%d&forkpty-arg=%s",
			cfg.ExecutablePath, nodeid, s.d.GetUnixSocketName())
		if cfg.RandomSeed != 0 {
			spinelUrl += fmt.Sprintf("&forkpty-arg=%d", cfg.RandomSeed)
		}
		args = append(args, spinelUrl)
	}
	cmd := exec.CommandContext(context.Background(), exePath, args...)

	logger.Tracef("Starting node %d: %s %v", nodeid, exePath, args)

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
		uartType:      nodeUartTypeUndefined,
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
		go node.lineReaderStdOut(node.pipeOut) // reader for Posix host CLI output and logging
	} else {
		node.uartType = nodeUartTypeVirtualTime
		// for a regular CLI node (not RCP), stdout is not used: UART output is sent via events.
	}
	go node.lineReaderStdErr(node.pipeErr) // reader for OT node process errors/failures written to stderr

	return node, err
}

func (node *Node) String() string {
	if len(node.nameStr) == 0 {
		node.nameStr = GetNodeName(node.Id)
	}
	return node.nameStr
}

func (node *Node) error(err error) {
	if err != nil {
		node.Logger.Error(err)
		node.cmdErr = err
	}
}

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

func (node *Node) signalExit() error {
	if node.cmd.Process == nil || node.cmd.ProcessState != nil {
		return nil
	}
	node.Logger.Tracef("Sending SIGTERM to node process PID %d", node.cmd.Process.Pid)
	return node.cmd.Process.Signal(syscall.SIGTERM)
}

func (node *Node) exit() error {
	node.isExiting = true
	_ = node.signalExit()

	// Pipes are closed to allow cmd.Wait() to be successful and not hang.
	if node.pipeIn != nil {
		_ = node.pipeIn.Close()
	}
	if node.pipeErr != nil {
		_ = node.pipeErr.Close()
	}
	if node.pipeOut != nil {
		_ = node.pipeOut.Close()
	}

	var err error = nil
	if node.cmd.Process != nil {
		processDone := make(chan bool)
		node.Logger.Tracef("Waiting for process PID %d to exit ...", node.cmd.Process.Pid)
		deadline := time.After(NodeExitTimeout)
		go func() {
			select {
			case processDone <- true:
				break
			case <-deadline:
				node.Logger.Warn("Node did not exit in time, sending SIGKILL.")
				_ = node.cmd.Process.Kill()
				processDone <- true
			}
		}()
		err = node.cmd.Wait() // wait for process end
		node.Logger.Tracef("Node process exited. Wait().err=%v", err)
		<-processDone // signal above kill-goroutine to end
	}
	node.Logger.Debugf("Node exited.")

	// finalize the log file/printing.
	node.DisplayPendingLogEntries()
	node.DisplayPendingLines()
	node.Logger.Close()

	// typical errors can be nil, or "signal: killed" if SIGKILL was necessary, or "signal: broken pipe", or
	// any of the "exit: ..." errors listed in code of node.cmd.Wait(). Most errors aren't critical: they
	// will still stop the node.
	return err
}

// inputCommand is a helper method to send a CLI command to the node's UART
func (node *Node) inputCommand(cmd string) error {
	var err error
	node.cmdErr = nil // reset last command error
	cmdBytes := []byte(cmd + "\n")

	switch node.uartType {
	case nodeUartTypeRealTime:
		_, err = node.pipeIn.Write(cmdBytes)
		node.S.Dispatcher().NotifyCommand(node.Id, false)
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
		_, err = node.expectLine(cmd, DefaultCommandTimeout)
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
	timeout := DefaultCommandTimeout

	err := node.inputCommand(cmd)
	if err != nil {
		node.error(err)
		return []string{}
	}

	if node.uartHasEcho {
		_, err = node.expectLine(cmd, timeout)
		if err != nil {
			node.error(err)
			return []string{}
		}
	}

	var output []string
	output, err = node.expectLine(doneOrErrorRegexp, timeout)
	if err != nil {
		node.error(err)
		return []string{}
	}

	var result string
	logger.AssertTrue(len(output) >= 1) // regexp matched - there's a Done or Error line in last output line.
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
		err := fmt.Errorf("%v - expected 1 line, but received %d: %#v", node, len(output), output)
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

func (node *Node) CommandExpectUint64(cmd string) uint64 {
	s := node.CommandExpectString(cmd)
	var iv uint64
	var err error

	if strings.HasPrefix(s, "0x") || strings.HasPrefix(s, "0X") {
		iv, err = strconv.ParseUint(s[2:], 16, 64)
	} else {
		iv, err = strconv.ParseUint(s, 10, 64)
	}

	if err != nil {
		node.Logger.Errorf("parsing unexpected Uint64 number: '%#v'", s)
		return 0
	}
	return iv
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

func (node *Node) SetChannel(ch ChannelId) {
	node.Command(fmt.Sprintf("channel %d", ch))
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

func (node *Node) GetRxSensitivity() int8 {
	return int8(node.getOrSetRfSimParam(false, ParamRxSensitivity, RssiInvalid))
}

func (node *Node) SetRxSensitivity(rxSens int8) {
	logger.AssertTrue(rxSens >= RssiMin && rxSens <= RssiMax)
	node.getOrSetRfSimParam(true, ParamRxSensitivity, RfSimParamValue(rxSens))
}

func (node *Node) GetCslAccuracy() uint8 {
	return uint8(node.getOrSetRfSimParam(false, ParamCslAccuracy, 255))
}

func (node *Node) SetCslAccuracy(accPpm uint8) {
	node.getOrSetRfSimParam(true, ParamCslAccuracy, RfSimParamValue(accPpm))
}

func (node *Node) GetCslUncertainty() uint8 {
	return uint8(node.getOrSetRfSimParam(false, ParamCslUncertainty, 255))
}

func (node *Node) SetCslUncertainty(unc10us uint8) {
	node.getOrSetRfSimParam(true, ParamCslUncertainty, RfSimParamValue(unc10us))
}

func (node *Node) getOrSetRfSimParam(isSet bool, param RfSimParam, value RfSimParamValue) RfSimParamValue {
	node.cmdErr = nil
	err := node.DNode.SendRfSimEvent(isSet, param, value)
	if err == nil {
		// wait for response event
		for {
			evt, err := node.expectEvent(event.EventTypeRadioRfSimParamRsp, DefaultCommandTimeout)
			if err != nil {
				node.error(err)
				return value
			}
			if evt.NodeId == node.Id && evt.RfSimParamData.Param == param {
				return RfSimParamValue(evt.RfSimParamData.Value)
			}
		}
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

func (node *Node) GetChildList() (childlist []int) {
	s := node.CommandExpectString("child list")
	ss := strings.Split(s, " ")

	for _, ids := range ss {
		id, err := strconv.Atoi(ids)
		if err != nil {
			node.Logger.Errorf("unexpected child list: '%#v'", s)
		}
		childlist = append(childlist, id)
	}
	return
}

func (node *Node) GetChildTimeout() int {
	return node.CommandExpectInt("childtimeout")
}

func (node *Node) SetChildTimeout(timeout int) {
	node.Command(fmt.Sprintf("childtimeout %d", timeout))
}

func (node *Node) GetContextReuseDelay() int {
	return node.CommandExpectInt("contextreusedelay")
}

func (node *Node) SetContextReuseDelay(delay int) {
	node.Command(fmt.Sprintf("contextreusedelay %d", delay))
}

func (node *Node) GetNetworkName() string {
	return node.CommandExpectString("networkname")
}

func (node *Node) SetNetworkName(name string) {
	node.Command(fmt.Sprintf("networkname %s", name))
}

func (node *Node) GetEui64() string {
	return node.CommandExpectString("eui64")
}

func (node *Node) SetEui64(eui64 string) {
	node.Command(fmt.Sprintf("eui64 %s", eui64))
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

func (node *Node) SetExtAddr(extaddr uint64) {
	node.Command(fmt.Sprintf("extaddr %016x", extaddr))
}

func (node *Node) GetExtPanid() string {
	return node.CommandExpectString("extpanid")
}

func (node *Node) SetExtPanid(extpanid string) {
	node.Command(fmt.Sprintf("extpanid %s", extpanid))
}

func (node *Node) GetIfconfig() string {
	return node.CommandExpectString("ifconfig")
}

func (node *Node) IfconfigUp() {
	node.Command("ifconfig up")
}

func (node *Node) IfconfigDown() {
	node.Command("ifconfig down")
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

func (node *Node) GetIpMaddr() []string {
	addrs := node.Command("ipmaddr")
	return addrs
}

func (node *Node) GetIpMaddrPromiscuous() bool {
	return node.CommandExpectEnabledOrDisabled("ipmaddr promiscuous")
}

func (node *Node) IpMaddrPromiscuousEnable() {
	node.Command("ipmaddr promiscuous enable")
}

func (node *Node) IpMaddrPromiscuousDisable() {
	node.Command("ipmaddr promiscuous disable")
}

func (node *Node) GetPromiscuous() bool {
	return node.CommandExpectEnabledOrDisabled("promiscuous")
}

func (node *Node) PromiscuousEnable() {
	node.Command("promiscuous enable")
}

func (node *Node) PromiscuousDisable() {
	node.Command("promiscuous disable")
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

func (node *Node) GetJoinerPort() int {
	return node.CommandExpectInt("joinerport")
}

func (node *Node) SetJoinerPort(port int) {
	node.Command(fmt.Sprintf("joinerport %d", port))
}

func (node *Node) GetKeySequenceCounter() int {
	return node.CommandExpectInt("keysequence counter")
}

func (node *Node) SetKeySequenceCounter(counter int) {
	node.Command(fmt.Sprintf("keysequence counter %d", counter))
}

func (node *Node) GetKeySequenceGuardTime() int {
	return node.CommandExpectInt("keysequence guardtime")
}

func (node *Node) SetKeySequenceGuardTime(guardtime int) {
	node.Command(fmt.Sprintf("keysequence guardtime %d", guardtime))
}

type LeaderData struct {
	PartitionID       int
	Weighting         int
	DataVersion       int
	StableDataVersion int
	LeaderRouterID    int
}

func (node *Node) GetLeaderData() (leaderData LeaderData) {
	var err error
	output := node.Command("leaderdata")
	for _, line := range output {
		if strings.HasPrefix(line, "Partition ID:") {
			leaderData.PartitionID, err = strconv.Atoi(line[14:])
			node.Logger.Error(err)
		}

		if strings.HasPrefix(line, "Weighting:") {
			leaderData.Weighting, err = strconv.Atoi(line[11:])
			node.Logger.Error(err)
		}

		if strings.HasPrefix(line, "Data Version:") {
			leaderData.DataVersion, err = strconv.Atoi(line[14:])
			node.Logger.Error(err)
		}

		if strings.HasPrefix(line, "Stable Data Version:") {
			leaderData.StableDataVersion, err = strconv.Atoi(line[21:])
			node.Logger.Error(err)
		}

		if strings.HasPrefix(line, "Leader Router ID:") {
			leaderData.LeaderRouterID, err = strconv.Atoi(line[18:])
			node.Logger.Error(err)
		}
	}
	return
}

func (node *Node) GetLeaderWeight() int {
	return node.CommandExpectInt("leaderweight")
}

func (node *Node) SetLeaderWeight(weight int) {
	node.Command(fmt.Sprintf("leaderweight %d", weight))
}

func (node *Node) Ping(addr string, payloadSize int, count int, interval float64, hopLimit int) {
	cmd := fmt.Sprintf("ping async %s %d %d %f %d", addr, payloadSize, count, interval, hopLimit)
	node.Command(cmd)
}

func (node *Node) GetNetworkKey() string {
	return node.CommandExpectString("networkkey")
}

func (node *Node) SetNetworkKey(key string) {
	node.Command(fmt.Sprintf("networkkey %s", key))
}

func (node *Node) GetMode() string {
	// TODO: return Mode type rather than just string
	return node.CommandExpectString("mode")
}

func (node *Node) SetMode(mode string) {
	node.Command(fmt.Sprintf("mode %s", mode))
}

func (node *Node) GetPanid() uint16 {
	return uint16(node.CommandExpectInt("panid"))
}

func (node *Node) SetPanid(panid uint16) {
	node.Command(fmt.Sprintf("panid 0x%x", panid))
}

func (node *Node) GetRloc16() uint16 {
	return uint16(node.CommandExpectHex("rloc16"))
}

func (node *Node) GetRouterSelectionJitter() int {
	return node.CommandExpectInt("routerselectionjitter")
}

func (node *Node) SetRouterSelectionJitter(timeout int) {
	node.Command(fmt.Sprintf("routerselectionjitter %d", timeout))
}

func (node *Node) GetRouterUpgradeThreshold() int {
	return node.CommandExpectInt("routerupgradethreshold")
}

func (node *Node) SetRouterUpgradeThreshold(timeout int) {
	node.Command(fmt.Sprintf("routerupgradethreshold %d", timeout))
}

func (node *Node) GetRouterDowngradeThreshold() int {
	return node.CommandExpectInt("routerdowngradethreshold")
}

func (node *Node) SetRouterDowngradeThreshold(timeout int) {
	node.Command(fmt.Sprintf("routerdowngradethreshold %d", timeout))
}

func (node *Node) GetState() string {
	return node.CommandExpectString("state")
}

func (node *Node) ThreadStart() error {
	return node.CommandChecked("thread start")
}

func (node *Node) ThreadStop() error {
	return node.CommandChecked("thread stop")
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

func (node *Node) GetUptimeMs() uint64 {
	return node.CommandExpectUint64("uptime ms")
}

func (node *Node) GetExecutablePath() string {
	return node.cfg.ExecutablePath
}

func (node *Node) GetExecutableName() string {
	return filepath.Base(node.cfg.ExecutablePath)
}

func (node *Node) GetSingleton() bool {
	s := node.CommandExpectString("singleton")
	if s == "true" {
		return true
	} else if s == "false" {
		return false
	} else if len(s) == 0 {
		node.Logger.Errorf("GetSingleton(): no data received")
		return false
	} else {
		node.Logger.Errorf("expected true/false, but read: '%#v'", s)
		return false
	}
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
		idx := bytes.IndexByte(buf, '\n')
		if idx == -1 {
			break // any remaining data stays in node.uartLine for next time.
		}
		if bytes.HasPrefix(buf, OtCliPrompt) {
			node.uartLine.Next(OtCliPromptLen) // consume the OT prompt, if any
			idx -= OtCliPromptLen
		}
		lineBytes := node.uartLine.Next(idx + 1)
		lineStr := bytes.TrimRight(lineBytes, "\r\n")
		node.pendingLines <- string(lineStr)
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

func (node *Node) onProcessFailure() {
	node.Logger.Errorf("Node process exited after an error occurred.")
	node.S.Dispatcher().NotifyNodeProcessFailure(node.Id)
	node.S.PostAsync(func() {
		_, nodeExists := node.S.nodes[node.Id]
		if node.S.ctx.Err() == nil && nodeExists {
			logger.Warnf("Deleting node %v due to process failure.", node.Id)
			_ = node.S.DeleteNode(node.Id)
		}
	})
}

// lineReaderStdErr is a goroutine to read stderr lines from an OT node process and log any lines
// as node errors.
func (node *Node) lineReaderStdErr(reader io.Reader) {
	scanner := bufio.NewScanner(reader)
	scanner.Split(bufio.ScanLines)

	isError := false
	for scanner.Scan() {
		line := scanner.Text()
		err := fmt.Errorf("StdErr: %s", line)
		node.Logger.Error(err)
		isError = true
	}

	// when the stderr of the node closes and any error was raised, inform simulation of node's failure.
	if isError && !node.isExiting {
		node.onProcessFailure()
	}
}

// lineReaderStdOut is a goroutine to read stdout lines from a Posix host CLI process and turn these into
// one of 1) UART-write event, 2) Log-write event, or 3) Status-push event + log-write event, depending
// on line format.
func (node *Node) lineReaderStdOut(reader io.Reader) {
	logger.AssertTrue(node.cfg.IsRcp)
	scanner := bufio.NewScanner(reader)
	scanner.Split(bufio.ScanLines)

	for scanner.Scan() {
		var evType event.EventType
		var data []byte

		line := scanner.Text()
		isOtLogLine := false

		if isStatusPush, status := logger.ParseOtnsStatusPush(line); isStatusPush {
			evType = event.EventTypeStatusPush
			data = []byte(status)

			ev := &event.Event{
				Delay:  0,
				Type:   evType,
				Data:   data,
				NodeId: node.Id,
			}
			node.S.Dispatcher().PostEventAsync(ev)
		}

		if isOtLogLine, _ = logger.ParseOtLogLine(line); isOtLogLine {
			evType = event.EventTypeLogWrite
			data = []byte(line)
		} else {
			evType = event.EventTypeUartWrite
			data = []byte(line + "\n")
		}

		ev := &event.Event{
			Delay:  0,
			Type:   evType,
			Data:   data,
			NodeId: node.Id,
		}
		node.S.Dispatcher().PostEventAsync(ev)
	}
}

// expectEvent waits for an event of the specified type to arrive from the OT node, for at most the
// timeout duration.
func (node *Node) expectEvent(evtType event.EventType, timeout time.Duration) (*event.Event, error) {
	done := node.S.ctx.Done()
	deadline := time.After(timeout)

	for {
		select {
		case <-done:
			return nil, CommandInterruptedError
		case evt := <-node.pendingEvents:
			node.Logger.Tracef("expectEvent() received: %v", evt)
			if evt.Type == evtType {
				return evt, nil
			} else {
				node.Logger.Errorf("expectEvent() received unexpected event type %d", evt.Type)
			}
		case <-deadline:
			err := fmt.Errorf("expectEvent timeout: expected event type %d", evtType)
			return nil, err
		default:
			node.S.waitForSimulation(true)
		}
	}
}

// readLine attempts to read a line from the node. Returns false when the node is asleep and
// therefore cannot return any more lines at the moment.
func (node *Node) readLine() (string, bool) {
	node.S.waitForSimulation(false)

	select {
	case line := <-node.pendingLines:
		node.Logger.Tracef("UART: %s", line)
		return line, true
	default:
		return "", false
	}
}

// expectLine reads potentially multiple lines from the node until it finds a matching line, or timeout
// occurs. The matching line is returned as the final line in the output string array. The expected line
// can be defined using various data types as detailed in isLineMatch().
func (node *Node) expectLine(line interface{}, timeout time.Duration) ([]string, error) {
	var outputLines []string
	done := node.S.ctx.Done()
	deadline := time.After(timeout)

	for {
		select {
		case <-done:
			return []string{}, CommandInterruptedError
		case <-deadline:
			//_ = pprof.Lookup("goroutine").WriteTo(os.Stdout, 2) // @DEBUG: useful log info when node stuck
			err := fmt.Errorf("expectLine timeout: expected '%v'", line)
			return outputLines, err
		case readLine := <-node.pendingLines:
			node.Logger.Tracef("UART: %s", readLine)
			outputLines = append(outputLines, readLine)
			if node.isLineMatch(readLine, line) { // found a matching line
				return outputLines, nil
			}
		default:
			node.S.waitForSimulation(true)
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

func (node *Node) DumpStat() string {
	return fmt.Sprintf("extaddr %016x, addr %04x, state %-6s", node.GetExtAddr(), node.GetRloc16(), node.GetState())
}

// setupCli sets up and tests CLI communication (via PTY, UART, virtual-UART, or other) before
// first-time CLI command usage.
func (node *Node) setupCli() error {
	var testCmdOutput string
	testCmd := "ifconfig"
	expectedTestCmdOutput := "down"

	// sending initial '\n' is required for MacOS terminal setup. It will trigger the CLI to write
	// the prompt '> ' as output without newline. The prompt gets filtered out by node.processUartData.
	// In Linux, sending the '\n' will lock up the ot-cli node (TODO look into why)
	if runtime.GOOS == "darwin" {
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

	if testCmdOutput != expectedTestCmdOutput {
		return fmt.Errorf("node did not provide expected output for '%s' ('%s')", testCmd, expectedTestCmdOutput)
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
	var prefix string

	useNodePrefix := node.S.cmdRunner.GetNodeContext() != node.Id
	if useNodePrefix {
		prefix = node.String()
	}

loop:
	for {
		select {
		case line := <-node.pendingLines:
			if useNodePrefix {
				logger.Println(prefix + line)
			} else {
				logger.Println(line)
			}
		default:
			break loop
		}
	}
}

func (node *Node) DisplayPendingLogEntries() {
	node.Logger.DisplayPendingLogEntries(node.S.Dispatcher().CurTime)
}
