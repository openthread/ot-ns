// Copyright (c) 2020-2024, The OTNS Authors.
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
	"context"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/pkg/errors"

	"github.com/openthread/ot-ns/dispatcher"
	"github.com/openthread/ot-ns/event"
	"github.com/openthread/ot-ns/logger"
	"github.com/openthread/ot-ns/otoutfilter"
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

	cfg           *NodeConfig
	cmd           *exec.Cmd
	cmdErr        error // store the last CLI command error; nil if none.
	version       string
	threadVersion uint16
	isSendStarted bool
	sendGroupIds  map[int]struct{}

	pendingLines  chan string       // OT node CLI output lines, pending processing.
	pendingEvents chan *event.Event // OT node emitted events to be processed.
	pipeIn        io.WriteCloser
	pipeOut       io.ReadCloser
	pipeErr       io.ReadCloser
	uartReader    chan []byte
	uartType      NodeUartType
}

// newNode creates a new simulation node. If unsuccessful, it returns an error != nil and the Node object created
// so far (if node return != nil), or nil if node object wasn't created yet.
func newNode(s *Simulation, nodeid NodeId, cfg *NodeConfig, dnode *dispatcher.Node) (*Node, error) {
	var err error

	if !cfg.Restore {
		flashFile := fmt.Sprintf("%s/%d_%d.flash", s.cfg.OutputDir, s.cfg.Id, nodeid)
		if err = os.RemoveAll(flashFile); err != nil {
			logger.Errorf("Remove flash file %s failed: %+v", flashFile, err)
			return nil, err
		}
	}

	var cmd *exec.Cmd
	if cfg.RandomSeed == 0 {
		cmd = exec.CommandContext(context.Background(), cfg.ExecutablePath, strconv.Itoa(nodeid), s.d.GetUnixSocketName())
	} else {
		seedParam := fmt.Sprintf("%d", cfg.RandomSeed)
		cmd = exec.CommandContext(context.Background(), cfg.ExecutablePath, strconv.Itoa(nodeid), s.d.GetUnixSocketName(), seedParam)
	}

	node := &Node{
		S:             s,
		Id:            nodeid,
		Logger:        logger.GetNodeLogger(s.cfg.OutputDir, s.cfg.Id, cfg),
		DNode:         dnode,
		cfg:           cfg,
		cmd:           cmd,
		pendingLines:  make(chan string, 10000),
		pendingEvents: make(chan *event.Event, 100),
		uartType:      nodeUartTypeUndefined,
		uartReader:    make(chan []byte, 10000),
		version:       "",
		sendGroupIds:  make(map[int]struct{}),
	}

	node.Logger.SetFileLevel(s.cfg.LogFileLevel)
	node.Logger.Debugf("Node config: type=%s IsMtd=%t IsRouter=%t IsBR=%t RxOffWhenIdle=%t", cfg.Type, cfg.IsMtd,
		cfg.IsRouter, cfg.IsBorderRouter, cfg.RxOffWhenIdle)
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

	go node.lineReaderStdErr(node.pipeErr) // reads StdErr output from OT node exe and acts on failures

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
	if node.cmd.Process == nil {
		return nil
	}
	node.Logger.Tracef("Sending SIGTERM to node process PID %d", node.cmd.Process.Pid)
	return node.cmd.Process.Signal(syscall.SIGTERM)
}

func (node *Node) exit() error {
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
		timeout := time.After(NodeExitTimeout)
		go func() {
			select {
			case processDone <- true:
				break
			case <-timeout:
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

func (node *Node) assurePrompt() {
	err := node.inputCommand("")
	if err == nil {
		if _, err := node.expectLine("", time.Second); err == nil {
			return
		}
	}

	err = node.inputCommand("")
	if err == nil {
		if _, err := node.expectLine("", time.Second); err == nil {
			return
		}
	}

	err = node.inputCommand("")
	if err == nil {
		_, err := node.expectLine("", DefaultCommandTimeout)
		if err != nil {
			node.Logger.Error(err)
		}
	}
}

func (node *Node) inputCommand(cmd string) error {
	logger.AssertTrue(node.uartType != nodeUartTypeUndefined)
	var err error
	node.cmdErr = nil // reset last command error

	if node.uartType == nodeUartTypeRealTime {
		_, err = node.pipeIn.Write([]byte(cmd + "\n"))
		node.S.Dispatcher().NotifyCommand(node.Id)
	} else {
		err = node.DNode.SendToUART([]byte(cmd + "\n"))
	}
	return err
}

// CommandNoDone executes the command without expecting 'Done' (e.g. it's a background command).
// It just reads output lines until the node is asleep.
func (node *Node) CommandNoDone(cmd string) []string {
	err := node.inputCommand(cmd)
	if err != nil {
		node.error(err)
		return []string{}
	}

	_, err = node.expectLine(cmd, DefaultCommandTimeout)
	if err != nil {
		node.error(err)
		return []string{}
	}

	var output []string
	for {
		line, ok := node.readLine()
		if !ok {
			break
		}
		if strings.HasPrefix(line, "Error") {
			node.error(fmt.Errorf(line))
		} else {
			output = append(output, line)
		}
	}
	return output
}

// Command executes the command and waits for 'Done', or an 'Error', at end of output.
func (node *Node) Command(cmd string) []string {
	timeout := DefaultCommandTimeout

	err := node.inputCommand(cmd)
	if err != nil {
		node.error(err)
		return []string{}
	}

	_, err = node.expectLine(cmd, timeout)
	if err != nil {
		node.error(err)
		return []string{}
	}

	var output []string
	output, err = node.expectLine(doneOrErrorRegexp, timeout)
	if err != nil {
		node.error(err)
		return []string{}
	}

	var result string
	logger.AssertTrue(len(output) >= 1) // there's always a Done or Error line in output.
	output, result = output[:len(output)-1], output[len(output)-1]
	if result != "Done" {
		node.error(fmt.Errorf(result))
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
		ParamCslAccuracy:
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
			node.error(fmt.Errorf("parameter out of range %d to %d", RssiMin, RssiMax))
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
		if value < -127 || value > 127 {
			node.error(fmt.Errorf("parameter out of range -127 - +127"))
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

func (node *Node) FactoryReset() {
	node.Logger.Warn("node factoryreset")
	err := node.inputCommand("factoryreset")
	if err != nil {
		node.error(err)
		return
	}
	node.assurePrompt()
	node.Logger.Info("factoryreset complete")
}

func (node *Node) Reset() {
	node.Logger.Warn("node reset")
	err := node.inputCommand("reset")
	if err != nil {
		node.error(err)
		return
	}
	node.assurePrompt()
	node.Logger.Warn("reset complete")
}

func (node *Node) Ping(addr string, payloadSize int, count int, interval int, hopLimit int) {
	cmd := fmt.Sprintf("ping async %s %d %d %d %d", addr, payloadSize, count, interval, hopLimit)
	err := node.inputCommand(cmd)
	if err != nil {
		node.error(err)
		return
	}
	_, err = node.expectLine(cmd, DefaultCommandTimeout)
	node.Logger.Error(err)
	node.assurePrompt()
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
		val, err := strconv.Atoi(kv[1])
		if err != nil {
			node.Logger.Errorf("GetCounters(): unexpected value string '%v' (not int)", kv[1])
			return nil
		}
		res[keyPrefix+kv[0]] = val
	}
	return res
}

func (node *Node) processUartData() {
	var deadline <-chan time.Time
	done := node.S.ctx.Done()

loop:
	for {
		select {
		case <-done:
			break loop
		case data := <-node.uartReader:
			line := string(data)
			if line == "> " { // filter out the prompt.
				continue
			}
			idxNewLine := strings.IndexByte(line, '\n')
			lineTrim := strings.TrimSpace(line)
			isLogLine, otLevelChar := otoutfilter.DetectLogLine(line)
			if isLogLine {
				lev := logger.ParseOtLevelChar(otLevelChar)
				node.Logger.Log(lev, lineTrim)
			} else if idxNewLine == -1 { // if no newline, get more items until a line can be formed.
				deadline = time.After(dispatcher.DefaultReadTimeout)

			loop2:
				for {
					select {
					case nextData := <-node.uartReader:
						nextPart := string(nextData)
						idxNewLinePart := strings.IndexByte(nextPart, '\n')
						line += nextPart
						if idxNewLinePart >= 0 {
							node.pendingLines <- strings.TrimSpace(line)
							break loop2
						}
					case <-done:
						break loop
					case <-deadline:
						line = strings.TrimSpace(line)
						node.Logger.Panicf("processUart deadline: line=%s", line)
						break loop2
					}
				}
			} else {
				node.pendingLines <- lineTrim
			}

		default:
			break loop
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

func (node *Node) lineReaderStdErr(reader io.Reader) {
	scanner := bufio.NewScanner(bufio.NewReader(reader)) // no filter applied.
	scanner.Split(bufio.ScanLines)

	var errProc error = nil
	for scanner.Scan() {
		line := scanner.Text()
		stderrLine := fmt.Sprintf("StdErr: %s", line)

		// mark the first error output line of the node
		if errProc == nil {
			errProc = errors.New(stderrLine)
		}
		node.Logger.Error(errProc)
	}

	// when the stderr of the node closes and any error was raised, inform simulation of node's failure.
	if errProc != nil {
		node.onProcessFailure()
	}
}

func (node *Node) expectEvent(evtType event.EventType, timeout time.Duration) (*event.Event, error) {
	done := node.S.ctx.Done()
	deadline := time.After(timeout)

	for {
		select {
		case <-done:
			return nil, CommandInterruptedError
		case <-deadline:
			err := fmt.Errorf("expectEvent timeout: expected type %d", evtType)
			return nil, err
		case evt := <-node.pendingEvents:
			node.Logger.Tracef("expectEvent() received: %v", evt)
			return evt, nil
		default:
			if !node.S.Dispatcher().IsAlive(node.Id) {
				time.Sleep(time.Millisecond * 10) // in case of node connection error, this becomes busy-loop.
			}
			node.S.Dispatcher().RecvEvents() // keep virtual-UART events coming.
		}
	}
}

// readLine attempts to read a line from the node. Returns false when the node is asleep and
// therefore cannot return any more lines at the moment.
func (node *Node) readLine() (string, bool) {
	done := node.S.ctx.Done()
	pendingLinesReceived := false

	for {
		select {
		case <-done:
			return "", false
		case readLine := <-node.pendingLines:
			if len(readLine) > 0 {
				node.Logger.Tracef("UART: %s", readLine)
			}
			return readLine, true
		default:
			node.S.Dispatcher().RecvEvents() // keep virtual-UART events coming.
			node.processUartData()           // forge data from UART-events into lines (fills up node.pendingLines)
			if pendingLinesReceived && !node.S.Dispatcher().IsAlive(node.Id) {
				return "", false
			}
			pendingLinesReceived = true
		}
	}
}

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
			err := fmt.Errorf("expectLine timeout: expected %v", line)
			return outputLines, err
		case readLine := <-node.pendingLines:
			if len(readLine) > 0 {
				node.Logger.Tracef("UART: %s", readLine)
			}

			outputLines = append(outputLines, readLine)
			if node.isLineMatch(readLine, line) { // found the exact line
				node.S.Dispatcher().RecvEvents() // this blocks until node is asleep.
				return outputLines, nil
			}
		default:
			if !node.S.Dispatcher().IsAlive(node.Id) {
				time.Sleep(time.Millisecond * 10) // in case of node connection error, this becomes busy-loop.
			}
			node.S.Dispatcher().RecvEvents() // keep virtual-UART events coming.
			node.processUartData()           // forge data from UART-events into lines (fills up node.pendingLines)
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
	prefix := ""
loop:
	for {
		select {
		case line := <-node.pendingLines:
			if len(prefix) == 0 && node.S.cmdRunner.GetNodeContext() != node.Id {
				prefix = node.String() // lazy init of node-specific prefix
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
