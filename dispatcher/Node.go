// Copyright (c) 2022, The OTNS Authors.
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

package dispatcher

import (
	"encoding/binary"
	"fmt"
	"math"
	"net"

	"github.com/openthread/ot-ns/threadconst"
	. "github.com/openthread/ot-ns/types"
	"github.com/simonlingoogle/go-simplelogger"
)

const (
	maxPingResultCount = 1000
	maxJoinResultCount = 1000
	minChannel         = 11
	maxChannel         = 26
)

type pingRequest struct {
	Timestamp uint64
	Dst       string
	DataSize  int
}
type PingResult struct {
	Dst      string
	DataSize int
	Delay    uint64
}

type joinerSession struct {
	StartTime  uint64
	JoinedTime uint64
	StopTime   uint64
}

type JoinResult struct {
	JoinDuration    uint64
	SessionDuration uint64
}

type Node struct {
	D                  *Dispatcher
	Id                 NodeId
	X, Y               int
	PartitionId        uint32
	ExtAddr            uint64
	Rloc16             uint16
	CreateTime         uint64
	CurTime            uint64
	Role               OtDeviceRole
	RadioState         RadioStates
	RadioChannel       uint8
	radioLockState     bool
	rxBusyUntil        []uint64
	isInvalidReception []bool
	isWaitingAck       bool
	waitAckSN          uint8

	peerAddr      *net.UDPAddr
	failureCtrl   *FailureCtrl
	isFailed      bool
	radioRange    int
	pendingPings  []*pingRequest
	pingResults   []*PingResult
	joinerState   OtJoinerState
	joinerSession *joinerSession
	joinResults   []*JoinResult
}

func newNode(d *Dispatcher, nodeid NodeId, x, y int, radioRange int) *Node {
	simplelogger.AssertTrue(radioRange >= 0)

	nc := &Node{
		D:                  d,
		Id:                 nodeid,
		CurTime:            d.CurTime,
		CreateTime:         d.CurTime,
		X:                  x,
		Y:                  y,
		ExtAddr:            InvalidExtAddr,
		Rloc16:             threadconst.InvalidRloc16,
		Role:               OtDeviceRoleDisabled,
		peerAddr:           nil, // peer address will be set when the first event is received
		radioRange:         radioRange,
		joinerState:        OtJoinerStateIdle,
		RadioChannel:       minChannel,
		radioLockState:     false,
		isInvalidReception: make([]bool, maxChannel-minChannel+1), // one for each channel (11-26)
		rxBusyUntil:        make([]uint64, maxChannel-minChannel+1),
		isWaitingAck:       false,
	}

	nc.failureCtrl = newFailureCtrl(nc, NonFailTime)

	return nc
}

func (node *Node) String() string {
	return fmt.Sprintf("Node<%016x@%d,%d>", node.ExtAddr, node.X, node.Y)
}

func (node *Node) Send(elapsed uint64, data []byte) {
	msg := make([]byte, len(data)+11)
	binary.LittleEndian.PutUint64(msg[:8], elapsed)
	msg[8] = eventTypeRadioComm

	binary.LittleEndian.PutUint16(msg[9:11], uint16(len(data)))
	n := copy(msg[11:], data)
	simplelogger.AssertTrue(n == len(data))

	node.SendMessage(msg)
}

func (node *Node) SendTxDoneSignal(elapsed uint64, seq uint8) {
	msg := make([]byte, 12)
	binary.LittleEndian.PutUint64(msg[:8], elapsed)
	msg[8] = eventTypeRadioTxDone

	binary.LittleEndian.PutUint16(msg[9:11], uint16(1))
	msg[11] = seq

	node.SendMessage(msg)
}

func (node *Node) SendChannelActivity(channel uint8, value int8, elapsed uint64) {
	msg := make([]byte, 13)
	binary.LittleEndian.PutUint64(msg[:8], elapsed)
	msg[8] = eventTypeChannelActivity

	binary.LittleEndian.PutUint16(msg[9:11], uint16(9))
	msg[11] = channel
	msg[12] = uint8(value)
	node.SendMessage(msg)
}

func (node *Node) SendMessage(msg []byte) {
	if node.peerAddr != nil {
		_, _ = node.D.udpln.WriteToUDP(msg, node.peerAddr)
	} else {
		simplelogger.Errorf("%s does not have a peer address", node)
	}
}

func (node *Node) GetDistanceTo(other *Node) (dist int) {
	dx := other.X - node.X
	dy := other.Y - node.Y
	dist = int(math.Sqrt(float64(dx*dx + dy*dy)))
	return
}

func (node *Node) IsFailed() bool {
	return node.isFailed
}

func (node *Node) Fail() {
	if !node.isFailed {
		node.isFailed = true
		node.D.cbHandler.OnNodeFail(node.Id)
		node.D.vis.OnNodeFail(node.Id)
	}
}

func (node *Node) Recover() {
	if node.isFailed {
		node.isFailed = false
		node.D.cbHandler.OnNodeRecover(node.Id)
		node.D.vis.OnNodeRecover(node.Id)
	}
}

func (node *Node) DumpStat() string {
	d := node.D
	alarmTs := d.alarmMgr.GetTimestamp(node.Id)
	return fmt.Sprintf("CurTime=%v, AlarmTs=%v, Failed=%-5v, RecoverTS=%v", node.CurTime, alarmTs, node.isFailed, node.failureCtrl.recoverTs)
}

func (node *Node) SetFailTime(failTime FailTime) {
	node.failureCtrl.SetFailTime(failTime)
}

func (node *Node) onPingRequest(timestamp uint64, dstaddr string, datasize int) {
	if datasize < 4 {
		// if datasize < 4, timestamp is 0, these ping requests are ignored
		return
	}

	node.pendingPings = append(node.pendingPings, &pingRequest{
		Timestamp: timestamp,
		Dst:       dstaddr,
		DataSize:  datasize,
	})
}

func (node *Node) onPingReply(timestamp uint64, dstaddr string, datasize int, hoplimit int) {
	if datasize < 4 {
		// if datasize < 4, timestamp is 0, these ping replies are ignored
		return
	}
	const maxPingDelayUs uint64 = 10 * 1000000
	var leftPingRequests []*pingRequest
	for _, req := range node.pendingPings {
		if req.Timestamp == timestamp && req.Dst == dstaddr {
			// ping replied
			node.addPingResult(req.Dst, req.DataSize, node.D.CurTime-req.Timestamp)
		} else if req.Timestamp+maxPingDelayUs < node.D.CurTime {
			// ping timeout
			node.addPingResult(req.Dst, req.DataSize, maxPingDelayUs)
		} else {
			leftPingRequests = append(leftPingRequests, req)
		}
	}

	node.pendingPings = leftPingRequests
}

func (node *Node) addPingResult(dst string, datasize int, delay uint64) {
	node.pingResults = append(node.pingResults, &PingResult{
		Dst:      dst,
		DataSize: datasize,
		Delay:    delay,
	})

	if len(node.pingResults) > maxPingResultCount {
		node.pingResults = node.pingResults[1:]
	}
}

func (node *Node) CollectPings() []*PingResult {
	ret := node.pingResults
	node.pingResults = nil
	return ret
}

func (node *Node) CollectJoins() []*JoinResult {
	ret := node.joinResults
	node.joinResults = nil
	return ret
}

func (node *Node) onStatusPushExtAddr(extaddr uint64) {
	simplelogger.AssertTrue(extaddr != InvalidExtAddr)
	oldExtAddr := node.ExtAddr
	if oldExtAddr == extaddr {
		return
	}

	node.ExtAddr = extaddr
	node.D.onStatusPushExtAddr(node, oldExtAddr)
}

func (node *Node) onJoinerState(state OtJoinerState) {
	// A success join states: Idle -> Discover -> Connecting -> Connected -> Entrust -> Joined -> Idle
	// A failed join states: Idle -> Discover -> Connecting -> Idle
	if node.joinerState == state {
		return
	}

	node.joinerState = state
	if state == OtJoinerStateDiscover || state == OtJoinerStateConnect {
		// new joiner session started
		node.startNewJoinerSession()
	} else if state == OtJoinerStateJoined {
		if node.joinerSession != nil {
			node.joinerSession.JoinedTime = node.CurTime
		}
	} else if state == OtJoinerStateIdle {
		node.closeJoinerSession()
	}
}

func (node *Node) startNewJoinerSession() {
	if node.joinerSession != nil {
		return
	}

	node.joinerSession = &joinerSession{
		StartTime:  node.CurTime,
		JoinedTime: 0,
		StopTime:   0,
	}
}

func (node *Node) closeJoinerSession() {
	js := node.joinerSession
	if js == nil {
		return
	}

	js.StopTime = node.CurTime
	// collection join result
	node.addJoinResult(js)
	node.joinerSession = nil
}

func (node *Node) addJoinResult(js *joinerSession) {
	var joinDuration uint64
	if js.JoinedTime != 0 {
		joinDuration = js.JoinedTime - js.StartTime
	}

	sessionDuration := js.StopTime - js.StartTime

	node.joinResults = append(node.joinResults, &JoinResult{
		JoinDuration:    joinDuration,
		SessionDuration: sessionDuration,
	})

	if len(node.joinResults) > maxJoinResultCount {
		node.joinResults = node.joinResults[1:]
	}
}

func (node *Node) IsReceptionSuccess(channel uint8) bool {
	return !node.isFailed &&
		node.RadioState == RadioRx &&
		node.RadioChannel == channel &&
		node.D.CurTime >= node.rxBusyUntil[channel-minChannel] &&
		!node.isInvalidReception[channel-minChannel]
}

func (node *Node) IsCollisionEvent(src *Node) bool {
	return node.Id != src.Id && // Needed check
		node.RadioChannel == src.RadioChannel &&
		node.D.CurTime < node.rxBusyUntil[src.RadioChannel-minChannel] &&
		node.isInvalidReception[src.RadioChannel-minChannel]
}

func (node *Node) IsWaitingAck(seq uint8) bool {
	return node.isWaitingAck && node.waitAckSN == seq
}

func (node *Node) InformAckReceived(seq uint8) {
	if node.IsWaitingAck(seq) {
		node.isWaitingAck = false
		node.waitAckSN = 0
	}
}

func (node *Node) ReceivePacket(channel uint8, until uint64) {
	if !node.isInvalidReception[channel-minChannel] {
		if node.rxBusyUntil[channel-minChannel] > 0 || node.RadioChannel != channel || node.isFailed {
			node.isInvalidReception[channel-minChannel] = true
		}
	}

	if until > node.rxBusyUntil[channel-minChannel] {
		node.rxBusyUntil[channel-minChannel] = until
	}
}

func (node *Node) IsChannelBusy(channel uint8) bool {
	return !node.isFailed && node.rxBusyUntil[channel-minChannel] > 0
}

func (node *Node) SetInvalidReception() {
	node.isInvalidReception[node.RadioChannel-minChannel] = true
}

func (node *Node) UpdateCollisionCondition() {
	for i := 0; i < maxChannel-minChannel+1; i++ {
		if node.rxBusyUntil[i] <= node.D.CurTime {
			node.isInvalidReception[i] = false
			node.rxBusyUntil[i] = 0
		}
	}
}

func (node *Node) SetRadioStateFromString(s string, timestamp uint64) {
	radioEnergy := node.D.energyAnalyser.GetNode(node.Id)

	simplelogger.AssertNotNil(radioEnergy)
	simplelogger.AssertFalse(node.radioLockState, "radio state was locked")

	var state RadioStates
	switch s {
	case "off":
		state = RadioDisabled
	case "sleep":
		state = RadioSleep
	case "tx":
		state = RadioTx
	case "rx":
		state = RadioRx
	default:
		simplelogger.Panicf("unknown radio state: %s", s)
	}

	node.RadioState = state
	radioEnergy.SetRadioState(state, timestamp)
}

func (node *Node) LockRadioState(lock bool) {
	node.radioLockState = lock
}
