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
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/pkg/errors"
	"github.com/simonlingoogle/go-simplelogger"

	"github.com/openthread/ot-ns/dispatcher"
	"github.com/openthread/ot-ns/otoutfilter"
	. "github.com/openthread/ot-ns/types"
)

const (
	DefaultCommandTimeout = time.Second * 10
	NodeExitTimeout       = time.Second * 3
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
	S       *Simulation
	Id      int
	cfg     *NodeConfig
	cmd     *exec.Cmd
	logFile *os.File
	cmdErr  error // store the last CLI command error; nil if none.

	pendingLines chan string   // OT node CLI output lines, pending processing.
	logEntries   chan LogEntry // OT node log entries, pending display on OTNS CLI or other viewers.
	pipeIn       io.WriteCloser
	pipeOut      io.ReadCloser
	pipeErr      io.ReadCloser
	uartReader   chan []byte
	uartType     NodeUartType
}

func newNode(s *Simulation, nodeid NodeId, cfg *NodeConfig) (*Node, error) {
	var err error
	logFileName := fmt.Sprintf("tmp/%d_%d.log", s.cfg.Id, nodeid)
	var logFile *os.File

	if !cfg.Restore {
		flashFile := fmt.Sprintf("tmp/%d_%d.flash", s.cfg.Id, nodeid)
		if err = os.RemoveAll(flashFile); err != nil {
			simplelogger.Errorf("Remove flash file %s failed: %+v", flashFile, err)
			return nil, err
		}
		if cfg.NodeLogFile {
			if err = os.RemoveAll(logFileName); err != nil {
				simplelogger.Errorf("Remove node log file %s failed: %+v", logFileName, err)
				return nil, err
			}
		}
	}

	simplelogger.Debugf("node exe path: %s", cfg.ExecutablePath)
	cmd := exec.CommandContext(context.Background(), cfg.ExecutablePath, strconv.Itoa(nodeid), s.d.GetUnixSocketName())

	node := &Node{
		S:            s,
		Id:           nodeid,
		cfg:          cfg,
		cmd:          cmd,
		pendingLines: make(chan string, 10000),
		logEntries:   make(chan LogEntry, 10000),
		uartType:     NodeUartTypeUndefined,
		uartReader:   make(chan []byte, 10000),
		logFile:      nil,
		cmdErr:       nil,
	}

	if node.pipeIn, err = cmd.StdinPipe(); err != nil {
		return nil, err
	}

	if node.pipeOut, err = cmd.StdoutPipe(); err != nil {
		return nil, err
	}

	if node.pipeErr, err = cmd.StderrPipe(); err != nil {
		return nil, err
	}

	// open log file for node's OT output
	if cfg.NodeLogFile {
		logFile, err = os.OpenFile(logFileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0664)
		if err != nil {
			simplelogger.Errorf("Opening node log file %s failed: %+v", logFileName, err)
			return nil, err
		}
		node.logFile = logFile
		header := fmt.Sprintf("# OpenThread node log for %s Created %s\n", GetNodeName(nodeid),
			time.Now().Format(time.RFC3339)) +
			fmt.Sprintf("# Executable: %s\n", cfg.ExecutablePath) +
			"# SimTimeUs NodeTime     Lev LogModule       Message"
		_ = node.writeToLogFile(header)
		simplelogger.Debugf("Node log file '%s' opened.", logFileName)
	}

	if err = cmd.Start(); err != nil {
		if logFile != nil {
			_ = logFile.Close()
		}
		return nil, err
	}

	go node.lineReaderStdErr(node.pipeErr) // reads StdErr output from OT node exe and acts on failures

	return node, nil
}

func (node *Node) String() string {
	return GetNodeName(node.Id)
}

func (node *Node) runInitScript(cfg []string) error {
	simplelogger.AssertNotNil(cfg)
	for _, cmd := range cfg {
		if node.CommandResult() != nil {
			return node.CommandResult()
		}
		if node.S.ctx.Err() != nil {
			return nil
		}
		node.Command(cmd, DefaultCommandTimeout)
	}
	return node.CommandResult()
}

func (node *Node) onStart() {
	node.log(WatchInfoLevel, fmt.Sprintf("started, panid=0x%04x, chan=%d, eui64=%#v, extaddr=%#v, state=%s, key=%#v, mode=%v",
		node.GetPanid(), node.GetChannel(), node.GetEui64(), node.GetExtAddr(), node.GetState(),
		node.GetNetworkKey(), node.GetMode()))
}

func (node *Node) IsFED() bool {
	return !node.cfg.IsMtd
}

func (node *Node) SignalExit() error {
	return node.cmd.Process.Signal(syscall.SIGTERM)
}

func (node *Node) Exit() error {
	_ = node.cmd.Process.Signal(syscall.SIGTERM)

	// Pipes are closed to allow cmd.Wait() to be successful and not hang.
	_ = node.pipeIn.Close()
	_ = node.pipeErr.Close()
	_ = node.pipeOut.Close()

	processDone := make(chan bool)
	isKilled := false
	simplelogger.Debugf("%s Waiting for process to exit ...", node.String())
	timeout := time.After(NodeExitTimeout)
	go func() {
		select {
		case processDone <- true:
			break
		case <-timeout:
			simplelogger.Warnf("%s did not exit in time, sending SIGKILL.", node.String())
			isKilled = true
			_ = node.cmd.Process.Kill()
			processDone <- true
		}
	}()
	err := node.cmd.Wait() // wait for process end
	simplelogger.Debugf("%s process exited. Wait().err=%v", node.String(), err)
	<-processDone

	if isKilled { // when killed, a "signal: killed" error is always given by Wait(). Suppress this.
		return nil
	}

	return err
}

func (node *Node) AssurePrompt() {
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
			node.logError(err)
		}
	}
}

func (node *Node) inputCommand(cmd string) error {
	simplelogger.AssertTrue(node.uartType != NodeUartTypeUndefined)
	var err error
	node.cmdErr = nil

	if node.uartType == NodeUartTypeRealTime {
		_, err = node.pipeIn.Write([]byte(cmd + "\n"))
		node.S.Dispatcher().NotifyCommand(node.Id)
	} else {
		err = node.S.Dispatcher().SendToUART(node.Id, []byte(cmd+"\n"))
	}
	return err
}

func (node *Node) CommandExpectNone(cmd string, timeout time.Duration) {
	err := node.inputCommand(cmd)
	if err != nil {
		node.logError(err)
		node.cmdErr = err
	} else {
		_, err = node.expectLine(cmd, timeout)
		if err != nil {
			node.logError(err)
			node.cmdErr = err
		}
	}
}

func (node *Node) Command(cmd string, timeout time.Duration) []string {
	err2 := node.inputCommand(cmd)
	if err2 == nil {
		_, err1 := node.expectLine(cmd, timeout)
		if err1 != nil {
			node.logError(err1)
			node.cmdErr = err1
			return []string{}
		}
	} else {
		node.logError(err2)
		node.cmdErr = err2
		return []string{}
	}

	output, err := node.expectLine(DoneOrErrorRegexp, timeout)
	if err != nil {
		node.logError(err)
		node.cmdErr = err
		return []string{}
	}
	if len(output) == 0 {
		err = fmt.Errorf("Command() response empty for cmd '%s'", cmd)
		node.logError(err)
		node.cmdErr = err
		return []string{}
	}

	var result string
	output, result = output[:len(output)-1], output[len(output)-1]
	if result != "Done" {
		err = fmt.Errorf("unexpected result for cmd '%s': %s", cmd, result)
		node.logError(err)
		node.cmdErr = fmt.Errorf(result)
	}
	return output
}

// CommandResult gets the last result of any Command...() call, either nil or an Error.
func (node *Node) CommandResult() error {
	return node.cmdErr
}

func (node *Node) CommandExpectString(cmd string, timeout time.Duration) string {
	output := node.Command(cmd, timeout)
	if len(output) != 1 {
		err := fmt.Errorf("%v - expected 1 line, but received %d: %#v", node, len(output), output)
		node.logError(err)
		return ""
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
		node.logError(fmt.Errorf("parsing unexpected Int number: '%#v'", s))
		return 0
	}
	return int(iv)
}

func (node *Node) CommandExpectHex(cmd string, timeout time.Duration) int {
	s := node.CommandExpectString(cmd, timeout)
	var iv int64
	var err error

	iv, err = strconv.ParseInt(s[2:], 16, 0)

	if err != nil {
		node.logError(fmt.Errorf("hex parsing unexpected number: '%#v'", s))
		return 0
	}
	return int(iv)
}

func (node *Node) SetChannel(ch int) {
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
			node.logError(fmt.Errorf("unexpected child list: '%#v'", s))
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
	if err != nil {
		if len(s) > 0 {
			node.logError(fmt.Errorf("GetExtAddr() unknown address format: %s", s))
		} else {
			node.logError(fmt.Errorf("GetExtAddr() address not received"))
		}
		return InvalidExtAddr
	}
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
			node.logError(err)
		}

		if strings.HasPrefix(line, "Weighting:") {
			leaderData.Weighting, err = strconv.Atoi(line[11:])
			node.logError(err)
		}

		if strings.HasPrefix(line, "Data Version:") {
			leaderData.DataVersion, err = strconv.Atoi(line[14:])
			node.logError(err)
		}

		if strings.HasPrefix(line, "Stable Data Version:") {
			leaderData.StableDataVersion, err = strconv.Atoi(line[21:])
			node.logError(err)
		}

		if strings.HasPrefix(line, "Leader Router ID:") {
			leaderData.LeaderRouterID, err = strconv.Atoi(line[18:])
			node.logError(err)
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
	node.log(WatchWarnLevel, "node factoryreset")
	err := node.inputCommand("factoryreset")
	if err != nil {
		node.logError(err)
		node.cmdErr = err
		return
	}
	node.AssurePrompt()
	node.log(WatchInfoLevel, "factoryreset complete")
}

func (node *Node) Reset() {
	node.log(WatchWarnLevel, "node reset")
	err := node.inputCommand("reset")
	if err != nil {
		node.logError(err)
		node.cmdErr = err
		return
	}
	node.AssurePrompt()
	node.log(WatchInfoLevel, "reset complete")
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

// GetVersion gets the version string of the OpenThread node.
func (node *Node) GetVersion() string {
	return node.CommandExpectString("version", DefaultCommandTimeout)
}

func (node *Node) GetExecutablePath() string {
	return node.cfg.ExecutablePath
}

func (node *Node) GetExecutableName() string {
	return filepath.Base(node.cfg.ExecutablePath)
}

func (node *Node) GetSingleton() bool {
	s := node.CommandExpectString("singleton", DefaultCommandTimeout)
	if s == "true" {
		return true
	} else if s == "false" {
		return false
	} else if len(s) == 0 {
		node.logError(fmt.Errorf("GetSingleton(): no data received"))
		return false
	} else {
		node.logError(fmt.Errorf("expected true/false, but read: '%#v'", s))
		return false
	}
}

func (node *Node) processUartData() {
	var deadline <-chan time.Time
loop:
	for {
		select {
		case data := <-node.uartReader:
			line := string(data)
			if line == "> " { // filter out the prompt.
				continue
			}
			idxNewLine := strings.IndexByte(line, '\n')
			lineTrim := strings.TrimSpace(line)
			isLogLine, otLevelChar := otoutfilter.DetectLogLine(line)
			if isLogLine {
				lev := ParseWatchLogLevel(otLevelChar)
				node.logEntries <- LogEntry{
					NodeId:  node.Id,
					Level:   lev,
					Msg:     lineTrim,
					IsWatch: true,
				}
			} else if idxNewLine == -1 { // if no newline, get more items until a line can be formed.
				deadline = time.After(dispatcher.DefaultReadTimeout)
				done := node.S.ctx.Done()
				isExiting := false
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
						if !isExiting {
							isExiting = true
							deadline = time.After(time.Millisecond * 200)
						}
						time.Sleep(time.Millisecond * 10)
					case <-deadline:
						line = strings.TrimSpace(line)
						if !isExiting {
							simplelogger.Panicf("%s processUart deadline: line=%s", node, line)
						}
						node.pendingLines <- line
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

func (node *Node) onProcessFailure(err error) {
	node.log(WatchCritLevel, "Node process failed.")
	node.S.PostAsync(false, func() {
		if node.S.ctx.Err() == nil {
			simplelogger.Warnf("Deleting node %v due to process failure.", node.Id)
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
		node.log(WatchCritLevel, stderrLine)
	}

	// when the stderr of the node closes and any error was raised, inform simulation of node's failure.
	if errProc != nil {
		node.onProcessFailure(errProc)
	}
}

func (node *Node) expectLine(line interface{}, timeout time.Duration) ([]string, error) {
	var outputLines []string

	deadline := time.After(timeout)
	for {
		select {
		case <-deadline:
			//_ = pprof.Lookup("goroutine").WriteTo(os.Stdout, 2) // @DEBUG: useful log info when node stuck
			outputLines = append(outputLines, "Done")
			err := fmt.Errorf("expectLine timeout: expected %v", line)
			return outputLines, err
		case readLine := <-node.pendingLines:
			if len(readLine) > 0 {
				node.log(WatchTraceLevel, "UART: "+readLine)
			}

			outputLines = append(outputLines, readLine)
			if node.isLineMatch(readLine, line) {
				// found the exact line
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

func (node *Node) CommandExpectEnabledOrDisabled(cmd string, timeout time.Duration) bool {
	output := node.CommandExpectString(cmd, timeout)
	if output == "Enabled" {
		return true
	} else if output == "Disabled" {
		return false
	} else if len(output) == 0 {
		node.logError(fmt.Errorf("CommandExpectEnabledOrDisabled() did not get data from node"))
		return false
	} else {
		node.logError(fmt.Errorf("expect Enabled/Disabled, but read: '%#v'", output))
	}
	return false
}

func (node *Node) Ping(addr string, payloadSize int, count int, interval int, hopLimit int) {
	cmd := fmt.Sprintf("ping async %s %d %d %d %d", addr, payloadSize, count, interval, hopLimit)
	err := node.inputCommand(cmd)
	if err != nil {
		node.logError(err)
		node.cmdErr = err
		return
	}
	_, err = node.expectLine(cmd, DefaultCommandTimeout)
	node.logError(err)
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
		simplelogger.Panicf("unknown data type %v, expected string, Regexp or []string", expectedLine)
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

func (node *Node) log(level WatchLogLevel, msg string) {
	node.logEntries <- LogEntry{
		NodeId: node.Id,
		Level:  level,
		Msg:    msg,
	}
}

func (node *Node) logError(err error) {
	if err == nil {
		return
	}
	node.logEntries <- LogEntry{
		NodeId: node.Id,
		Level:  WatchCritLevel,
		Msg:    err.Error(),
	}
}

func (node *Node) DisplayPendingLines(ts uint64) {
loop:
	for {
		select {
		case line := <-node.pendingLines:
			PrintConsole(line)
		default:
			break loop
		}
	}
}

func (node *Node) DisplayPendingLogEntries(ts uint64) {
	isExiting := node.S.ctx.Err() != nil
loop:
	for {
		select {
		case e := <-node.logEntries:
			line := getTimestampedLogMessage(ts, e.Msg)
			_ = node.writeToLogFile(line)

			s := node.S
			if isExiting && e.Level <= WatchCritLevel {
				e.Level = WatchInfoLevel // avoid error display while exiting - it's normal to lose connectivity to nodes.
			}
			if e.IsWatch {
				// watch messages may get increased level/visibility
				if e.Level <= s.Dispatcher().GetWatchLevel(node.Id) { // IF it must be shown
					if s.logLevel < e.Level && s.logLevel >= WatchInfoLevel { // HOW it can be shown
						e.Level = s.logLevel
					}
					PrintLog(e.Level, node.String()+line)
				}
			} else if e.Level <= s.logLevel {
				PrintLog(e.Level, node.String()+line)
			}
		default:
			break loop
		}
	}
}

func (node *Node) writeToLogFile(line string) error {
	if node.logFile == nil {
		return nil
	}
	_, err := node.logFile.WriteString(line + "\n")
	if err != nil {
		_ = node.logFile.Close()
		node.logFile = nil
		node.logError(fmt.Errorf("couldn't write to node log file, closing it (%s)", node.logFile.Name()))
	}
	return err
}
