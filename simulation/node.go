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

package simulation

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/openthread/ot-ns/dispatcher"
	"github.com/openthread/ot-ns/otoutfilter"
	. "github.com/openthread/ot-ns/types"
	"github.com/pkg/errors"
	"github.com/simonlingoogle/go-simplelogger"
)

const (
	DefaultCommandTimeout = time.Second * 10
)

var (
	DoneOrErrorRegexp = regexp.MustCompile(`(Done|Error \d+: .*)`)
)

type NodeUartType int

const (
	NodeUartTypeUndefined   NodeUartType = iota
	NodeUartTypeRealTime    NodeUartType = iota
	NodeUartTypeVirtualTime NodeUartType = iota
)

type Node struct {
	S   *Simulation
	Id  int
	cfg *NodeConfig

	cmd     *exec.Cmd
	logFile *os.File
	err     error

	pendingLines      chan string
	pipeIn            io.WriteCloser
	pipeOut           io.Reader
	pipeErr           io.ReadCloser
	virtualUartReader *io.PipeReader
	virtualUartPipe   *io.PipeWriter
	uartType          NodeUartType
}

func newNode(s *Simulation, id NodeId, cfg *NodeConfig) (*Node, error) {
	var err error
	logFileName := fmt.Sprintf("tmp/%d_%d.log", s.cfg.Id, id)
	if !cfg.Restore {
		flashFile := fmt.Sprintf("tmp/%d_%d.flash", s.cfg.Id, id)
		if err = os.RemoveAll(flashFile); err != nil {
			simplelogger.Errorf("Remove flash file %s failed: %+v", flashFile, err)
			return nil, err
		}
		if err = os.RemoveAll(logFileName); err != nil {
			simplelogger.Errorf("Remove node log file %s failed: %+v", logFileName, err)
			return nil, err
		}
	}

	// open log file for node's OT output
	logFile, err := os.OpenFile(logFileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0664)
	if err != nil {
		simplelogger.Errorf("Opening node log file %s failed: %+v", logFileName, err)
		return nil, err
	}

	// determine path of the OT CLI app and start OT node process.
	otCliPath := s.cfg.OtCliFtdPath
	if cfg.ExecutablePath != "" {
		otCliPath = cfg.ExecutablePath
	} else if cfg.IsBorderRouter {
		otCliPath = s.cfg.OtBrPath
	} else if cfg.IsMtd {
		otCliPath = s.cfg.OtCliMtdPath
	}
	simplelogger.Debugf("node exe path: %s", otCliPath)
	cmd := exec.CommandContext(context.Background(), otCliPath, strconv.Itoa(id), s.d.GetUnixSocketName())

	node := &Node{
		S:            s,
		Id:           id,
		cfg:          cfg,
		cmd:          cmd,
		pendingLines: make(chan string, 10000),
		uartType:     NodeUartTypeUndefined,
		logFile:      logFile,
		err:          nil,
	}

	node.virtualUartReader, node.virtualUartPipe = io.Pipe()

	if node.pipeIn, err = cmd.StdinPipe(); err != nil {
		return nil, err
	}

	if node.pipeOut, err = cmd.StdoutPipe(); err != nil {
		return nil, err
	}

	if node.pipeErr, err = cmd.StderrPipe(); err != nil {
		return nil, err
	}

	if err = cmd.Start(); err != nil {
		return nil, err
	}

	go node.lineReader(node.pipeOut, NodeUartTypeRealTime)
	go node.lineReader(node.virtualUartReader, NodeUartTypeVirtualTime)
	go node.lineReaderStdErr(node.pipeErr)

	return node, nil
}

func (node *Node) String() string {
	return fmt.Sprintf("Node<%d>", node.Id)
}

func (node *Node) SetupNetworkParameters(cfg []string) {
	for _, cmd := range cfg {
		node.Command(cmd, DefaultCommandTimeout)
	}
}

func (node *Node) Start() {
	simplelogger.Infof("%v - started, panid=0x%04x, chan=%d, eui64=%#v, extaddr=%#v, state=%s, key=%#v, mode=%v", node,
		node.GetPanid(), node.GetChannel(), node.GetEui64(), node.GetExtAddr(), node.GetState(),
		node.GetNetworkKey(), node.GetMode())
}

func (node *Node) IsFED() bool {
	return !node.cfg.IsMtd
}

func (node *Node) Stop() {
	node.ThreadStop()
	node.IfconfigDown()
	simplelogger.Debugf("%v - stopped, state = %s", node, node.GetState())
}

func (node *Node) Exit() error {
	_ = node.cmd.Process.Signal(syscall.SIGTERM)

	err := node.cmd.Wait()
	node.S.Dispatcher().RecvEvents() // ensure to receive any remaining events of exited node.

	// no more events or lineReader lines should come, so we can close the log file and virtual-UART.
	_ = node.virtualUartReader.Close()
	if node.logFile != nil {
		_ = node.logFile.Close()
	}

	return err
}

func (node *Node) AssurePrompt() {
	node.inputCommand("")
	if found, _ := node.TryExpectLine("", time.Second); found {
		return
	}

	node.inputCommand("")
	if found, _ := node.TryExpectLine("", time.Second); found {
		return
	}

	node.inputCommand("")
	node.expectLine("", DefaultCommandTimeout)
}

func (node *Node) inputCommand(cmd string) {
	simplelogger.AssertTrue(node.uartType != NodeUartTypeUndefined)

	if node.uartType == NodeUartTypeRealTime {
		_, _ = node.pipeIn.Write([]byte(cmd + "\n"))
		node.S.Dispatcher().NotifyCommand(node.Id)
	} else {
		node.S.Dispatcher().SendToUART(node.Id, []byte(cmd+"\n"))
	}
}

func (node *Node) CommandExpectNone(cmd string, timeout time.Duration) {
	node.inputCommand(cmd)
	node.expectLine(cmd, timeout)
}

func (node *Node) Command(cmd string, timeout time.Duration) []string {
	node.inputCommand(cmd)
	node.expectLine(cmd, timeout)
	output := node.expectLine(DoneOrErrorRegexp, timeout)

	var result string
	output, result = output[:len(output)-1], output[len(output)-1]
	if result != "Done" {
		simplelogger.Panicf("%v - Unexpected cmd result: %s", node, result)
	}
	return output
}

func (node *Node) CommandExpectString(cmd string, timeout time.Duration) string {
	output := node.Command(cmd, timeout)
	if len(output) != 1 {
		simplelogger.Panicf("expected 1 line, but received %d: %#v", len(output), output)
	}

	return output[0]
}

func (node *Node) CommandExpectInt(cmd string, timeout time.Duration) int {
	s := node.CommandExpectString(cmd, timeout)
	var iv int64
	var err error

	if strings.HasPrefix(s, "0x") || strings.HasPrefix(s, "0X") {
		iv, err = strconv.ParseInt(s[2:], 16, 0)
	} else {
		iv, err = strconv.ParseInt(s, 10, 0)
	}

	if err != nil {
		simplelogger.Panicf("unexpected number: %#v", s)
	}
	return int(iv)
}

func (node *Node) CommandExpectHex(cmd string, timeout time.Duration) int {
	s := node.CommandExpectString(cmd, timeout)
	var iv int64
	var err error

	iv, err = strconv.ParseInt(s[2:], 16, 0)

	if err != nil {
		simplelogger.Panicf("unexpected number: %#v", s)
	}
	return int(iv)
}

func (node *Node) SetChannel(ch int) {
	simplelogger.AssertTrue(11 <= ch && ch <= 26)
	node.Command(fmt.Sprintf("channel %d", ch), DefaultCommandTimeout)
}

func (node *Node) GetChannel() int {
	return node.CommandExpectInt("channel", DefaultCommandTimeout)
}

func (node *Node) GetChildList() (childlist []int) {
	s := node.CommandExpectString("child list", DefaultCommandTimeout)
	ss := strings.Split(s, " ")

	for _, ids := range ss {
		id, err := strconv.Atoi(ids)
		if err != nil {
			simplelogger.Panicf("unpexpected child list: %#v", s)
		}
		childlist = append(childlist, id)
	}
	return
}

func (node *Node) GetChildTable() {
	// todo: not implemented yet
}

func (node *Node) GetChildTimeout() int {
	return node.CommandExpectInt("childtimeout", DefaultCommandTimeout)
}

func (node *Node) SetChildTimeout(timeout int) {
	node.Command(fmt.Sprintf("childtimeout %d", timeout), DefaultCommandTimeout)
}

func (node *Node) GetContextReuseDelay() int {
	return node.CommandExpectInt("contextreusedelay", DefaultCommandTimeout)
}

func (node *Node) SetContextReuseDelay(delay int) {
	node.Command(fmt.Sprintf("contextreusedelay %d", delay), DefaultCommandTimeout)
}

func (node *Node) GetNetworkName() string {
	return node.CommandExpectString("networkname", DefaultCommandTimeout)
}

func (node *Node) SetNetworkName(name string) {
	node.Command(fmt.Sprintf("networkname %s", name), DefaultCommandTimeout)
}

func (node *Node) GetEui64() string {
	return node.CommandExpectString("eui64", DefaultCommandTimeout)
}

func (node *Node) SetEui64(eui64 string) {
	node.Command(fmt.Sprintf("eui64 %s", eui64), DefaultCommandTimeout)
}

func (node *Node) GetExtAddr() uint64 {
	s := node.CommandExpectString("extaddr", DefaultCommandTimeout)
	v, err := strconv.ParseUint(s, 16, 64)
	simplelogger.PanicIfError(err)
	return v
}

func (node *Node) SetExtAddr(extaddr uint64) {
	node.Command(fmt.Sprintf("extaddr %016x", extaddr), DefaultCommandTimeout)
}

func (node *Node) GetExtPanid() string {
	return node.CommandExpectString("extpanid", DefaultCommandTimeout)
}

func (node *Node) SetExtPanid(extpanid string) {
	node.Command(fmt.Sprintf("extpanid %s", extpanid), DefaultCommandTimeout)
}

func (node *Node) GetIfconfig() string {
	return node.CommandExpectString("ifconfig", DefaultCommandTimeout)
}

func (node *Node) IfconfigUp() {
	node.Command("ifconfig up", DefaultCommandTimeout)
}

func (node *Node) IfconfigDown() {
	node.Command("ifconfig down", DefaultCommandTimeout)
}

func (node *Node) GetIpAddr() []string {
	// todo: parse IPv6 addresses
	addrs := node.Command("ipaddr", DefaultCommandTimeout)
	return addrs
}

func (node *Node) GetIpAddrLinkLocal() []string {
	// todo: parse IPv6 addresses
	addrs := node.Command("ipaddr linklocal", DefaultCommandTimeout)
	return addrs
}

func (node *Node) GetIpAddrMleid() []string {
	// todo: parse IPv6 addresses
	addrs := node.Command("ipaddr mleid", DefaultCommandTimeout)
	return addrs
}

func (node *Node) GetIpAddrRloc() []string {
	addrs := node.Command("ipaddr rloc", DefaultCommandTimeout)
	return addrs
}

func (node *Node) GetIpMaddr() []string {
	// todo: parse IPv6 addresses
	addrs := node.Command("ipmaddr", DefaultCommandTimeout)
	return addrs
}

func (node *Node) GetIpMaddrPromiscuous() bool {
	return node.CommandExpectEnabledOrDisabled("ipmaddr promiscuous", DefaultCommandTimeout)
}

func (node *Node) IpMaddrPromiscuousEnable() {
	node.Command("ipmaddr promiscuous enable", DefaultCommandTimeout)
}

func (node *Node) IpMaddrPromiscuousDisable() {
	node.Command("ipmaddr promiscuous disable", DefaultCommandTimeout)
}

func (node *Node) GetPromiscuous() bool {
	return node.CommandExpectEnabledOrDisabled("promiscuous", DefaultCommandTimeout)
}

func (node *Node) PromiscuousEnable() {
	node.Command("promiscuous enable", DefaultCommandTimeout)
}

func (node *Node) PromiscuousDisable() {
	node.Command("promiscuous disable", DefaultCommandTimeout)
}

func (node *Node) GetRouterEligible() bool {
	return node.CommandExpectEnabledOrDisabled("routereligible", DefaultCommandTimeout)
}

func (node *Node) RouterEligibleEnable() {
	node.Command("routereligible enable", DefaultCommandTimeout)
}

func (node *Node) RouterEligibleDisable() {
	node.Command("routereligible disable", DefaultCommandTimeout)
}

func (node *Node) GetJoinerPort() int {
	return node.CommandExpectInt("joinerport", DefaultCommandTimeout)
}

func (node *Node) SetJoinerPort(port int) {
	node.Command(fmt.Sprintf("joinerport %d", port), DefaultCommandTimeout)
}

func (node *Node) GetKeySequenceCounter() int {
	return node.CommandExpectInt("keysequence counter", DefaultCommandTimeout)
}

func (node *Node) SetKeySequenceCounter(counter int) {
	node.Command(fmt.Sprintf("keysequence counter %d", counter), DefaultCommandTimeout)
}

func (node *Node) GetKeySequenceGuardTime() int {
	return node.CommandExpectInt("keysequence guardtime", DefaultCommandTimeout)
}

func (node *Node) SetKeySequenceGuardTime(guardtime int) {
	node.Command(fmt.Sprintf("keysequence guardtime %d", guardtime), DefaultCommandTimeout)
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
	output := node.Command("leaderdata", DefaultCommandTimeout)
	for _, line := range output {
		if strings.HasPrefix(line, "Partition ID:") {
			leaderData.PartitionID, err = strconv.Atoi(line[14:])
			simplelogger.PanicIfError(err)
		}

		if strings.HasPrefix(line, "Weighting:") {
			leaderData.Weighting, err = strconv.Atoi(line[11:])
			simplelogger.PanicIfError(err)
		}

		if strings.HasPrefix(line, "Data Version:") {
			leaderData.DataVersion, err = strconv.Atoi(line[14:])
			simplelogger.PanicIfError(err)
		}

		if strings.HasPrefix(line, "Stable Data Version:") {
			leaderData.StableDataVersion, err = strconv.Atoi(line[21:])
			simplelogger.PanicIfError(err)
		}

		if strings.HasPrefix(line, "Leader Router ID:") {
			leaderData.LeaderRouterID, err = strconv.Atoi(line[18:])
			simplelogger.PanicIfError(err)
		}
	}
	return
}

func (node *Node) GetLeaderPartitionId() int {
	return node.CommandExpectInt("leaderpartitionid", DefaultCommandTimeout)
}

func (node *Node) SetLeaderPartitionId(partitionid int) {
	node.Command(fmt.Sprintf("leaderpartitionid 0x%x", partitionid), DefaultCommandTimeout)
}

func (node *Node) GetLeaderWeight() int {
	return node.CommandExpectInt("leaderweight", DefaultCommandTimeout)
}

func (node *Node) SetLeaderWeight(weight int) {
	node.Command(fmt.Sprintf("leaderweight 0x%x", weight), DefaultCommandTimeout)
}

func (node *Node) FactoryReset() {
	simplelogger.Warnf("%v - factoryreset", node)
	node.inputCommand("factoryreset")
	node.AssurePrompt()
	simplelogger.Debugf("%v - ready", node)
}

func (node *Node) Reset() {
	simplelogger.Warnf("%v - reset", node)
	node.inputCommand("reset")
	node.AssurePrompt()
	simplelogger.Debugf("%v - ready", node)
}

func (node *Node) GetNetworkKey() string {
	return node.CommandExpectString("networkkey", DefaultCommandTimeout)
}

func (node *Node) SetNetworkKey(key string) {
	node.Command(fmt.Sprintf("networkkey %s", key), DefaultCommandTimeout)
}

func (node *Node) GetMode() string {
	// todo: return Mode type rather than just string
	return node.CommandExpectString("mode", DefaultCommandTimeout)
}

func (node *Node) SetMode(mode string) {
	node.Command(fmt.Sprintf("mode %s", mode), DefaultCommandTimeout)
}

func (node *Node) GetPanid() uint16 {
	// todo: return Mode type rather than just string
	return uint16(node.CommandExpectInt("panid", DefaultCommandTimeout))
}

func (node *Node) SetPanid(panid uint16) {
	node.Command(fmt.Sprintf("panid 0x%x", panid), DefaultCommandTimeout)
}

func (node *Node) GetRloc16() uint16 {
	return uint16(node.CommandExpectHex("rloc16", DefaultCommandTimeout))
}

func (node *Node) GetRouterSelectionJitter() int {
	return node.CommandExpectInt("routerselectionjitter", DefaultCommandTimeout)
}

func (node *Node) SetRouterSelectionJitter(timeout int) {
	node.Command(fmt.Sprintf("routerselectionjitter %d", timeout), DefaultCommandTimeout)
}

func (node *Node) GetRouterUpgradeThreshold() int {
	return node.CommandExpectInt("routerupgradethreshold", DefaultCommandTimeout)
}

func (node *Node) SetRouterUpgradeThreshold(timeout int) {
	node.Command(fmt.Sprintf("routerupgradethreshold %d", timeout), DefaultCommandTimeout)
}

func (node *Node) GetRouterDowngradeThreshold() int {
	return node.CommandExpectInt("routerdowngradethreshold", DefaultCommandTimeout)
}

func (node *Node) SetRouterDowngradeThreshold(timeout int) {
	node.Command(fmt.Sprintf("routerdowngradethreshold %d", timeout), DefaultCommandTimeout)
}

func (node *Node) GetState() string {
	return node.CommandExpectString("state", DefaultCommandTimeout)
}

func (node *Node) ThreadStart() {
	node.Command("thread start", DefaultCommandTimeout)
}

func (node *Node) ThreadStop() {
	node.Command("thread stop", DefaultCommandTimeout)
}

func (node *Node) GetVersion() string {
	return node.CommandExpectString("version", DefaultCommandTimeout)
}

func (node *Node) GetSingleton() bool {
	s := node.CommandExpectString("singleton", DefaultCommandTimeout)
	if s == "true" {
		return true
	} else if s == "false" {
		return false
	} else {
		simplelogger.Panicf("expect true/false, but read: %#v", s)
		return false
	}
}

func (node *Node) lineReader(reader io.Reader, uartType NodeUartType) {
	// close the line channel after line reader routine exit.
	// Below filter takes out any OT node log-message lines and sends these to the handler.
	scanner := bufio.NewScanner(otoutfilter.NewOTOutFilter(bufio.NewReader(reader), "", node.handlerLogMsg))
	scanner.Split(bufio.ScanLines)

	for scanner.Scan() {
		line := scanner.Text()

		// append line to node-specific log file.
		node.writeToLogFile(line)

		if node.uartType == NodeUartTypeUndefined {
			simplelogger.Debugf("%v's UART type is %v", node, uartType)
			node.uartType = uartType
		}

		select {
		case node.pendingLines <- line:
			break
		default:
			// if we failed to append line, panic - should not happen normally. If so, needs fix.
			simplelogger.Panicf("%v - node.pendingLines exceeded length %v", node, len(node.pendingLines))
			break
		}
	}
}

func (node *Node) lineReaderStdErr(reader io.Reader) {
	scanner := bufio.NewScanner(bufio.NewReader(reader)) // no filter applied.
	scanner.Split(bufio.ScanLines)

	for scanner.Scan() {
		line := scanner.Text()

		// mark the first error in the node
		if node.err == nil {
			node.err = errors.New(line)
		}

		// append it to node-specific log file.
		node.writeToLogFile(line)

		// send it to watch output.
		node.S.PostAsync(false, func() {
			node.S.Dispatcher().WatchMessage(node.Id, dispatcher.WatchCritLevel, fmt.Sprintf("%v StdErr: %v", node, line))
		})
	}

	// when the stderr of the node closes and any error was raised, inform simulation of node's failure.
	if node.err != nil {
		node.S.OnNodeProcessFailure(node)
	}
}

// handles an incoming log message from the OT-Node and its detected loglevel (otLevel). It is written
// to the node's log file. Display of the log message is done by the Dispatcher.
func (node *Node) handlerLogMsg(otLevel string, msg string) {
	// append msg to node-specific log file.
	node.writeToLogFile(msg)

	// create a node-specific log message that may be used by the Dispatcher's Watch function.
	lev := dispatcher.ParseWatchLogLevel(otLevel)
	node.S.PostAsync(false, func() {
		node.S.Dispatcher().WatchMessage(node.Id, lev, msg)
	})
}

func (node *Node) TryExpectLine(line interface{}, timeout time.Duration) (bool, []string) {
	var outputLines []string

	deadline := time.After(timeout)

	for {
		select {
		case <-deadline:
			return false, outputLines
		case readLine, ok := <-node.pendingLines:
			simplelogger.AssertTrue(ok, "%s EOF: node.pendingLines channel was closed", node)

			if !strings.HasPrefix(readLine, "|") && !strings.HasPrefix(readLine, "+") {
				simplelogger.Debugf("%v %s", node, readLine)
			}

			outputLines = append(outputLines, readLine)
			if node.isLineMatch(readLine, line) {
				// found the exact line
				return true, outputLines
			} else {
				// TODO hack: output scan result here, should have better implementation
				//| J | Network Name     | Extended PAN     | PAN  | MAC Address      | Ch | dBm | LQI |
				if strings.HasPrefix(readLine, "|") || strings.HasPrefix(readLine, "+") {
					fmt.Printf("%s\n", readLine)
				}
			}
		default:
			node.S.Dispatcher().RecvEvents() // keep virtual-UART events coming.
		}
	}
}

func (node *Node) expectLine(line interface{}, timeout time.Duration) []string {
	found, output := node.TryExpectLine(line, timeout)
	if !found {
		node.err = errors.Errorf("expect line timeout: %#v", line)
		return []string{}
	}

	return output
}

func (node *Node) CommandExpectEnabledOrDisabled(cmd string, timeout time.Duration) bool {
	output := node.CommandExpectString(cmd, timeout)
	if output == "Enabled" {
		return true
	} else if output == "Disabled" {
		return false
	} else {
		simplelogger.Panicf("expect Enabled/Disabled, but read: %#v", output)
	}
	return false
}

func (node *Node) Ping(addr string, payloadSize int, count int, interval int, hopLimit int) {
	cmd := fmt.Sprintf("ping async %s %d %d %d %d", addr, payloadSize, count, interval, hopLimit)
	node.inputCommand(cmd)
	node.expectLine(cmd, DefaultCommandTimeout)
	node.AssurePrompt()
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
		simplelogger.Panic("unknown expected string")
	}
	return false
}

func (node *Node) DumpStat() string {
	return fmt.Sprintf("extaddr %016x, addr %04x, state %-6s", node.GetExtAddr(), node.GetRloc16(), node.GetState())
}

func (node *Node) setupMode() {
	if node.cfg.IsRouter {
		// routers should be full functional and rx always on
		simplelogger.AssertFalse(node.cfg.IsMtd)
		simplelogger.AssertFalse(node.cfg.RxOffWhenIdle)
	}

	// only MED can use RxOffWhenIdle
	simplelogger.AssertTrue(!node.cfg.RxOffWhenIdle || node.cfg.IsMtd)

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

func (node *Node) onUartWrite(data []byte) {
	_, _ = node.virtualUartPipe.Write(data)
}

func (node *Node) writeToLogFile(line string) {
	if node.logFile != nil {
		_, err := node.logFile.WriteString(line + "\n")
		if err != nil {
			simplelogger.Error("Couldn't write to log file of %v, closing it (%s)", node, node.logFile)
			_ = node.logFile.Close()
			node.logFile = nil
		}
	}
}
