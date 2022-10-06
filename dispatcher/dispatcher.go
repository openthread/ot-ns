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
	"fmt"
	"math/rand"
	"math"
	"net"
	"time"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/openthread/ot-ns/progctx"
	"github.com/openthread/ot-ns/dissectpkt"
	"github.com/openthread/ot-ns/dissectpkt/wpan"
	"github.com/openthread/ot-ns/energy"
	"github.com/openthread/ot-ns/pcap"
	"github.com/openthread/ot-ns/radiomodel"
	"github.com/openthread/ot-ns/threadconst"
	"github.com/openthread/ot-ns/visualize"
	"github.com/simonlingoogle/go-simplelogger"
	. "github.com/openthread/ot-ns/types"
)

const (
	Ever uint64 = math.MaxUint64 / 2
)

const (
	MaxSimulateSpeed = 1000000
)

type pcapFrameItem struct {
	Ustime uint64
	Data   []byte
}

type Config struct {
	Speed       float64
	Real        bool
	Host        string
	Port        int
	DumpPackets bool
	NoPcap      bool
}

func DefaultConfig() *Config {
	return &Config{
		Speed:       1,
		Real:        false,
		Host:        "localhost",
		Port:        threadconst.InitialDispatcherPort,
		DumpPackets: false,
	}
}

type CallbackHandler interface {
	OnNodeFail(nodeid NodeId)
	OnNodeRecover(nodeid NodeId)

	// Notifies that the node's UART was written with data.
	OnUartWrite(nodeid NodeId, data []byte)
}

type goDuration struct {
	duration time.Duration
	done     chan struct{}
}

type Dispatcher struct {
	ctx                   *progctx.ProgCtx
	cfg                   Config
	cbHandler             CallbackHandler
	udpln                 *net.UDPConn
	eventChan             chan *Event
	waitGroup             sync.WaitGroup
	CurTime               uint64
	pauseTime             uint64
	alarmMgr              *alarmMgr
	eventQueue            *sendQueue
	nodes                 map[NodeId]*Node
	deletedNodes          map[NodeId]struct{}
	aliveNodes            map[NodeId]struct{}
	pcap                  *pcap.File
	pcapFrameChan         chan pcapFrameItem
	vis                   visualize.Visualizer
	taskChan              chan func()
	speed                 float64
	speedStartRealTime    time.Time
	speedStartTime        uint64
	extaddrMap            map[uint64]*Node
	rloc16Map             rloc16Map
	goDurationChan        chan goDuration
	globalPacketLossRatio float64
	visOptions            VisualizationOptions
	coaps                 *coapsHandler

	Counters struct {
		// Event counters
		AlarmEvents      uint64
		RadioEvents      uint64
		StatusPushEvents uint64
		UartWriteEvents  uint64
		CollisionEvents  uint64
		// Packet dispatching counters
		DispatchByExtAddrSucc   uint64
		DispatchByExtAddrFail   uint64
		DispatchByShortAddrSucc uint64
		DispatchByShortAddrFail uint64
		DispatchAllInRange      uint64
	}
	watchingNodes  map[NodeId]struct{}
	energyAnalyser *energy.EnergyAnalyser
	stopped        bool
	radioModel     radiomodel.RadioModel
}

func NewDispatcher(ctx *progctx.ProgCtx, cfg *Config, cbHandler CallbackHandler) *Dispatcher {
	simplelogger.AssertTrue(!cfg.Real || cfg.Speed == 1)

	udpAddr, err := net.ResolveUDPAddr("udp4", fmt.Sprintf("%s:%d", cfg.Host, cfg.Port))
	simplelogger.FatalIfError(err, err)
	ln, err := net.ListenUDP("udp", udpAddr)
	simplelogger.FatalIfError(err, err)
	_ = ln.SetWriteBuffer(25 * 1024 * 1024)
	_ = ln.SetReadBuffer(25 * 1024 * 1024)
	simplelogger.Infof("dispatcher listening on %s ...", udpAddr)

	simplelogger.AssertNil(err)

	vis := visualize.NewNopVisualizer()

	d := &Dispatcher{
		ctx:                ctx,
		cfg:                *cfg,
		cbHandler:          cbHandler,
		udpln:              ln,
		eventChan:          make(chan *Event, 10000),
		eventQueue:         newSendQueue(),
		alarmMgr:           newAlarmMgr(),
		nodes:              make(map[NodeId]*Node),
		deletedNodes:       map[NodeId]struct{}{},
		aliveNodes:         make(map[NodeId]struct{}),
		extaddrMap:         map[uint64]*Node{},
		rloc16Map:          rloc16Map{},
		pcapFrameChan:      make(chan pcapFrameItem, 100000),
		speed:              cfg.Speed,
		speedStartRealTime: time.Now(),
		vis:                vis,
		taskChan:           make(chan func(), 100),
		watchingNodes:      map[NodeId]struct{}{},
		goDurationChan:     make(chan goDuration, 10),
		visOptions:         defaultVisualizationOptions(),
	}
	d.speed = d.normalizeSpeed(d.speed)
	if !d.cfg.NoPcap {
		d.pcap, err = pcap.NewFile("current.pcap")
		simplelogger.PanicIfError(err)
		go d.pcapFrameWriter()
	}

	go d.eventsReader()

	d.vis.SetSpeed(d.speed)
	simplelogger.Infof("dispatcher started: cfg=%+v", *cfg)

	return d
}

func (d *Dispatcher) Stop() {
	if d.stopped {
		return
	}
	d.stopped = true
	close(d.pcapFrameChan)
	d.vis.Stop()
	d.waitGroup.Wait()
}

func (d *Dispatcher) Nodes() map[NodeId]*Node {
	return d.nodes
}

func (d *Dispatcher) Go(duration time.Duration) <-chan struct{} {
	done := make(chan struct{})
	d.goDurationChan <- goDuration{
		duration: duration,
		done:     done,
	}
	return done
}

func (d *Dispatcher) Run() {
	d.ctx.WaitAdd("dispatcher", 1)
	defer d.ctx.WaitDone("dispatcher")
	defer simplelogger.Debugf("dispatcher exit.")

	defer d.Stop()

	done := d.ctx.Done()
loop:
	for {
		select {
		case f := <-d.taskChan:
			f()
			break
		case duration := <-d.goDurationChan:
			// sync the speed start time with the current time
			if len(d.nodes) == 0 {
				// no nodes, sleep for a small duration to avoid high cpu
				d.RecvEvents()
				time.Sleep(time.Millisecond * 10)
				close(duration.done)
				break
			}

			d.speedStartRealTime = time.Now()
			d.speedStartTime = d.CurTime

			simplelogger.AssertTrue(d.CurTime == d.pauseTime)
			oldPauseTime := d.pauseTime
			d.pauseTime += uint64(duration.duration / time.Microsecond)
			if d.pauseTime > Ever || d.pauseTime < oldPauseTime {
				d.pauseTime = Ever
			}

			simplelogger.AssertTrue(d.CurTime <= d.pauseTime)
			d.goUntilPauseTime()

			if d.ctx.Err() != nil {
				close(duration.done)
				break loop
			}

			simplelogger.AssertTrue(d.CurTime == d.pauseTime)
			d.syncAllNodes()
			if d.pcap != nil {
				_ = d.pcap.Sync()
			}
			close(duration.done)
			break
		case <-done:
			break loop
		}
	}
}

func (d *Dispatcher) goUntilPauseTime() {
	for d.CurTime < d.pauseTime {
		d.handleTasks()

		if d.ctx.Err() != nil {
			break
		}

		// keep receiving events from OT nodes until all are asleep i.e. will not produce more events.
		d.RecvEvents()
		if !d.stopped {
			d.syncAliveNodes() // normally there should not be any alive nodes anymore.
		}
		if len(d.aliveNodes) == 0 {
			// all are asleep now - process the next Events in queue, either alarm or other type, for a single time.
			goon := d.processNextEvent()
			simplelogger.AssertTrue(d.CurTime <= d.pauseTime)

			if !goon && len(d.aliveNodes) == 0 {
				d.advanceTime(d.pauseTime) // if nothing more to do before d.pauseTime.
			}
		}
	}
}

// handleRecvEvent is the central handler for all events externally received from OpenThread nodes.
// It may only process events immediately that are to be executed at time d.CurTime. Future events
// will need to be queued (scheduled).
func (d *Dispatcher) handleRecvEvent(evt *Event) {
	nodeid := evt.NodeId
	node := d.nodes[nodeid]
	if node == nil {
		if !d.isDeleted(nodeid) {
			// can not find the node, and the node is not registered (created by OTNS)
			simplelogger.Warnf("Event (type %v) received from unknown Node %v, discarding.", evt.Type, evt.NodeId)
		}
		return // node was deleted already: just silently ignore event.
	}
	d.setAlive(nodeid)          // node stays alive until Alarm event is received.
	evt.Timestamp = d.CurTime   // timestamp incoming event
	node.peerAddr = evt.SrcAddr // assign source address from event to node

	// TODO document this use (for alarm messages)
	delay := evt.Delay
	if delay >= 2147483647 {
		delay = Ever
	}

	// should not receive alarm event and radio event in real mode
	if d.cfg.Real && (evt.Type == EventTypeAlarmFired || evt.Type == EventTypeRadioReceived ||
		evt.Type == EventTypeRadioComm || evt.Type == EventTypeRadioChannelSample ||
		evt.Type == EventTypeRadioState) {
		simplelogger.Warnf("unexpected event in real mode: %v", evt.Type)
		return
	}

	switch evt.Type {
	case EventTypeAlarmFired:
		d.Counters.AlarmEvents += 1
		d.setSleeping(node.Id)
		d.alarmMgr.SetTimestamp(nodeid, d.CurTime+delay) // schedule future wake-up of node
	case EventTypeRadioReceived:
		simplelogger.Panicf("legacy EventTypeRadioReceived received")
	case EventTypeRadioComm:
		fallthrough
	case EventTypeRadioChannelSample:
		d.Counters.RadioEvents += 1
		d.radioModel.HandleEvent(node.radioNode, d.eventQueue, evt)
	case EventTypeRadioState:
		d.Counters.RadioEvents += 1
		d.handleRadioState(node, evt)
		d.radioModel.HandleEvent(node.radioNode, d.eventQueue, evt)
	case EventTypeStatusPush:
		d.Counters.StatusPushEvents += 1
		d.handleStatusPush(node, string(evt.Data))
	case EventTypeUartWrite:
		d.Counters.UartWriteEvents += 1
		d.cbHandler.OnUartWrite(node.Id, evt.Data)
	default:
		simplelogger.Panicf("received event type not implemented: %v", evt.Type)
	}
}

// RecvEvents receives events from nodes, and handles these, until there is no more alive node.
func (d *Dispatcher) RecvEvents() int {
	blockTimeout := time.After(time.Second * 5)
	count := 0

loop:
	for {
		shouldBlock := len(d.aliveNodes) > 0

		if shouldBlock {
			select {
			case evt := <-d.eventChan:
				count += 1
				d.handleRecvEvent(evt)
			case <-blockTimeout:
				// timeout
				break loop
			}
		} else {
			select {
			case evt := <-d.eventChan:
				count += 1
				d.handleRecvEvent(evt)
			default:
				break loop
			}
		}
	}
	return count
}

// processNextEvent processes all next events from the eventQueue for the current time instant.
// Returns true if the simulation needs to continue, or false if not (e.g. it's time to pause).
func (d *Dispatcher) processNextEvent() bool {
	simplelogger.AssertTrue(d.CurTime <= d.pauseTime)

	// fetch time of next event
	nextAlarmTime := d.alarmMgr.NextTimestamp()
	nextSendTime := d.eventQueue.NextTimestamp()
	simplelogger.AssertTrue(nextSendTime >= d.CurTime && nextAlarmTime >= d.CurTime)

	nextEventTime := min(nextAlarmTime, nextSendTime)

	// convert nextEventTime to real time
	if d.speed < MaxSimulateSpeed {
		var sleepUntilTime = nextEventTime
		if sleepUntilTime > d.pauseTime {
			sleepUntilTime = d.pauseTime
		}

		var needSleepDuration time.Duration
		if d.speed <= 0 {
			needSleepDuration = time.Hour
		} else {
			needSleepDuration = time.Duration(float64(sleepUntilTime-d.speedStartTime)/d.speed) * time.Microsecond
		}
		sleepUntilRealTime := d.speedStartRealTime.Add(needSleepDuration)
		now := time.Now()
		sleepTime := sleepUntilRealTime.Sub(now)

		if sleepTime > 0 {
			if sleepTime > time.Millisecond*10 {
				sleepTime = time.Millisecond * 10
			}
			time.Sleep(sleepTime)

			if d.cfg.Real {
				curTime := d.speedStartTime + uint64(float64(time.Since(d.speedStartRealTime)/time.Microsecond)*d.speed)
				if curTime > d.pauseTime {
					curTime = d.pauseTime
				}
				d.advanceTime(curTime)
			}
			return true
		}
	}

	if nextEventTime > d.pauseTime {
		return false
	}

	// process (if any) all queued events, that happen at exactly procUntilTime
	procUntilTime := nextEventTime
	for nextEventTime <= procUntilTime {

		d.advanceTime(nextEventTime)

		if nextAlarmTime <= nextSendTime {
			// process next alarm
			nextAlarm := d.alarmMgr.NextAlarm()
			node := d.nodes[nextAlarm.NodeId]
			simplelogger.AssertNotNil(nextAlarm)
			if node != nil {
				d.advanceNodeTime(node, nextAlarm.Timestamp, false)
			}

		} else {
			// process next event from the queue
			evt := d.eventQueue.PopNext()
			simplelogger.AssertTrue(evt.Timestamp == nextEventTime)
			simplelogger.AssertTrue(nextAlarmTime == d.CurTime || nextSendTime == d.CurTime)
			node := d.nodes[evt.NodeId]
			if node != nil {
				// execute event - either a msg sent out, or handled by RadioModel.
				if !evt.MustDispatch {
					d.radioModel.HandleEvent(node.radioNode, d.eventQueue, evt)
				} else {
					switch evt.Type {
					case EventTypeRadioRxDone:
						fallthrough
					case EventTypeRadioComm:
						d.sendRadioCommRxEvents(node, evt)
					default:
						if d.radioModel.OnEventDispatch(node.radioNode, node.radioNode, evt) {
							node.sendEvent(evt)
						}
					}
				}

			} else {
				// node may have been deleted in the meantime.
				simplelogger.Debugf("processNextEvent() skipping event for deleted/unknown node %v: %v", evt.NodeId, evt)
			}

		}
		nextAlarmTime = d.alarmMgr.NextTimestamp()
		nextSendTime = d.eventQueue.NextTimestamp()
		nextEventTime = min(nextAlarmTime, nextSendTime)
	}

	return len(d.nodes) > 0
}

func (d *Dispatcher) eventsReader() {
	udpln := d.udpln
	readbuf := make([]byte, 4096)

	for {
		// loop and read events from UDP socket until all nodes are asleep
		n, srcaddr, err := udpln.ReadFromUDP(readbuf)
		if err != nil {
			if nerr, ok := err.(net.Error); ok && nerr.Timeout() {
				time.Sleep(time.Millisecond * 100)
				continue
			}

			simplelogger.Fatalf("UDP events reader quit.")
			break
		}

		evt := &Event{}
		evt.Deserialize(readbuf[0:n])
		evt.NodeId = srcaddr.Port - d.cfg.Port
		evt.SrcAddr = srcaddr

		d.eventChan <- evt
	}
}

func (d *Dispatcher) advanceNodeTime(node *Node, timestamp uint64, force bool) {
	simplelogger.AssertNotNil(node)
	if d.cfg.Real {
		node.CurTime = timestamp
		return
	}

	oldTime := node.CurTime
	if timestamp <= oldTime && !force {
		// node time was already newer than the requested timestamp
		return
	}
	msg := &Event{
		Type:      EventTypeAlarmFired,
		Timestamp: timestamp,
	}
	node.sendEvent(msg) // actively move the node's virtual-time to new time using an alarm-event msg.
}

// SendToUART sends data to virtual time UART of the target node.
func (d *Dispatcher) SendToUART(id NodeId, data []byte) {
	evt := &Event{
		Timestamp: InvalidTimestamp,
		Delay:     0,
		Type:      EventTypeUartWrite,
		Data:      data,
		NodeId:    id,
	}
	dstnode := d.nodes[id]
	if dstnode != nil {
		dstnode.sendEvent(evt)
	} else {
		simplelogger.Debugf("SendToUART() failed to send to deleted/unknown node: %v", id)
	}
}

// sendRadioCommRxEvents dispatches an event where >=1 nodes may receive a frame being transmitted,
// determining who receives it, and also doing the frame logging/pcap and visualization events.
func (d *Dispatcher) sendRadioCommRxEvents(srcNode *Node, evt *Event) {
	simplelogger.AssertTrue(evt.Type == EventTypeRadioComm || evt.Type == EventTypeRadioRxDone)
	if srcNode.isFailed {
		return // source node can't send - don't send, and don't log in pcap.
	}

	// record the sent frame in Pcap/Dump logs - once, at time of Tx start.
	isStartOfFrame := evt.Type == EventTypeRadioComm
	if isStartOfFrame {
		if !d.cfg.NoPcap {
			d.pcapFrameChan <- pcapFrameItem{evt.Timestamp, evt.Data[RadioMessagePsduOffset:]}
		}
		if d.cfg.DumpPackets {
			d.dumpPacket(evt)
		}
	}

	// try to dispatch the message by address directly to the right node
	srcNodeId := srcNode.Id
	pktinfo := dissectpkt.Dissect(evt.Data)
	pktFrame := pktinfo.MacFrame
	dispatchedByDstAddr := false
	dstAddrMode := pktFrame.FrameControl.DstAddrMode()

	if dstAddrMode == wpan.DstAddrModeExtended {
		// the message should only be dispatched to the target node with the extaddr
		dstnode := d.extaddrMap[pktFrame.DstAddrExtended]
		if dstnode != srcNode && dstnode != nil {
			if d.checkRadioReachable(srcNode, dstnode) {
				d.sendOneRadioFrame(evt, srcNode, dstnode)
				if isStartOfFrame {
					d.visSendFrame(srcNodeId, dstnode.Id, pktFrame)
				}
			} else if isStartOfFrame {
				d.visSendFrame(srcNodeId, InvalidNodeId, pktFrame)
			}
			d.Counters.DispatchByExtAddrSucc++
		} else {
			d.Counters.DispatchByExtAddrFail++
			if isStartOfFrame {
				d.visSendFrame(srcNodeId, InvalidNodeId, pktFrame)
			}
		}

		dispatchedByDstAddr = true
	} else if dstAddrMode == wpan.DstAddrModeShort {
		if pktFrame.DstAddrShort != threadconst.BroadcastRloc16 {
			// unicast message should only be dispatched to target node with the rloc16
			dstnodes := d.rloc16Map[pktFrame.DstAddrShort]
			dispatchCnt := 0

			if len(dstnodes) > 0 {
				for _, dstnode := range dstnodes {
					if d.checkRadioReachable(srcNode, dstnode) {
						d.sendOneRadioFrame(evt, srcNode, dstnode)
						if isStartOfFrame {
							d.visSendFrame(srcNodeId, dstnode.Id, pktFrame)
						}
						dispatchCnt++
					}
				}
				d.Counters.DispatchByShortAddrSucc++
			} else {
				d.Counters.DispatchByShortAddrFail++
			}

			if dispatchCnt == 0 && isStartOfFrame {
				d.visSendFrame(srcNodeId, InvalidNodeId, pktFrame)
			}

			dispatchedByDstAddr = true
		}
	}

	// if not dispatched yet, dispatch to all nodes able to receive. Works e.g. for Acks that don't have
	// a destination address.
	if !dispatchedByDstAddr {
		for _, dstNode := range d.nodes {
			if d.checkRadioReachable(srcNode, dstNode) {
				d.sendOneRadioFrame(evt, srcNode, dstNode)
			}
		}
		d.Counters.DispatchAllInRange++
		if isStartOfFrame {
			d.visSendFrame(srcNodeId, BroadcastNodeId, pktFrame)
		}
	}
}

func (d *Dispatcher) checkRadioReachable(src *Node, dst *Node) bool {
	return src != dst && src != nil && dst != nil &&
		d.radioModel.CheckRadioReachable(src.radioNode, dst.radioNode)
}

func (d *Dispatcher) sendOneRadioFrame(evt *Event, srcnode *Node, dstnode *Node) {
	simplelogger.AssertFalse(d.cfg.Real)
	simplelogger.AssertTrue(EventTypeRadioComm == evt.Type || EventTypeRadioRxDone == evt.Type)
	simplelogger.AssertTrue(srcnode != dstnode)

	// Tx failure cases below:
	//   1) 'failed' state of the dest node
	if dstnode.isFailed {
		return
	}

	//   2) dispatcher's random packet loss Event (separate from radio model)
	if d.globalPacketLossRatio > 0 {
		datalen := len(evt.Data)
		succRate := math.Pow(1.0-d.globalPacketLossRatio, float64(datalen)/128.0)
		if rand.Float64() >= succRate {
			return
		}
	}

	// create new Event copy for individual dispatch to dstNode.
	evt2 := evt.Copy()

	// Tx failure cases below:
	//   3) radio model indicates failure on this specific link (e.g. interference) now.
	// Below lets the radio model process every individual dispatch, to set RSSI, error, etc.
	if d.radioModel.OnEventDispatch(srcnode.radioNode, dstnode.radioNode, &evt2) {
		// send the event plus time keeping - moves dstnode's time to the current send-event's time.
		dstnode.sendEvent(&evt2)
	}
}

func (d *Dispatcher) setAlive(nodeid NodeId) {
	if d.cfg.Real {
		// real devices are always considered sleeping
		return
	}
	d.aliveNodes[nodeid] = struct{}{}
}

func (d *Dispatcher) isAlive(nodeid NodeId) bool {
	if _, ok := d.aliveNodes[nodeid]; ok {
		return true
	}
	return false
}

func (d *Dispatcher) isDeleted(nodeid NodeId) bool {
	if _, ok := d.deletedNodes[nodeid]; ok {
		return true
	}
	return false
}

func (d *Dispatcher) setSleeping(nodeid NodeId) {
	simplelogger.AssertFalse(d.cfg.Real)
	delete(d.aliveNodes, nodeid)
}

// syncAliveNodes advances the node's time of alive nodes only to current dispatcher time.
func (d *Dispatcher) syncAliveNodes() {
	if len(d.aliveNodes) == 0 {
		return
	}

	simplelogger.Warnf("syncing %d alive nodes: %v", len(d.aliveNodes), d.aliveNodes)
	for nodeid := range d.aliveNodes {
		d.advanceNodeTime(d.nodes[nodeid], d.CurTime, true)
	}
}

// syncAllNodes advances all of the node's time to current dispatcher time.
func (d *Dispatcher) syncAllNodes() {
	for nodeid := range d.nodes {
		d.advanceNodeTime(d.nodes[nodeid], d.CurTime, false)
	}
}

func (d *Dispatcher) pcapFrameWriter() {
	d.waitGroup.Add(1)
	defer d.waitGroup.Done()

	defer func() {
		err := d.pcap.Close()
		if err != nil {
			simplelogger.Errorf("failed to close pcap: %v", err)
		}
	}()
	for item := range d.pcapFrameChan {
		err := d.pcap.AppendFrame(item.Ustime, item.Data)
		if err != nil {
			simplelogger.Errorf("write pcap failed:%+v", err)
		}
	}
}

func (d *Dispatcher) SetVisualizer(vis visualize.Visualizer) {
	simplelogger.AssertNotNil(vis)
	d.vis = vis
	d.vis.SetSpeed(d.speed)
	d.vis.SetEnergyAnalyser(d.energyAnalyser)
}

func (d *Dispatcher) GetVisualizer() visualize.Visualizer {
	return d.vis
}

func (d *Dispatcher) handleStatusPush(srcnode *Node, data string) {
	simplelogger.Debugf("status push: %d: %#v ts=%v", srcnode.Id, data, d.CurTime)
	statuses := strings.Split(data, ";")
	srcid := srcnode.Id
	for _, status := range statuses {
		sp := strings.Split(status, "=")
		if len(sp) != 2 {
			continue
		}
		if sp[0] == "transmit" {
			d.visStatusPushTransmit(srcnode, sp[1])
		} else if sp[0] == "role" {
			role, err := strconv.Atoi(sp[1])
			simplelogger.PanicIfError(err)
			d.setNodeRole(srcnode, OtDeviceRole(role))
		} else if sp[0] == "rloc16" {
			rloc16, err := strconv.Atoi(sp[1])
			simplelogger.PanicIfError(err)
			d.setNodeRloc16(srcid, uint16(rloc16))
		} else if sp[0] == "ping_request" {
			// e.x. ping_request=fdde:ad00:beef:0:556:90c8:ffaf:b7a3$0$4026600960
			args := strings.Split(sp[1], ",")
			dstaddr := args[0]
			datasize, err := strconv.Atoi(args[1])
			simplelogger.PanicIfError(err)
			timestamp, err := strconv.ParseUint(args[2], 10, 64)
			simplelogger.PanicIfError(err)
			srcnode.onPingRequest(d.convertNodeMilliTime(srcnode, uint32(timestamp)), dstaddr, datasize)
		} else if sp[0] == "ping_reply" {
			//e.x.ping_reply=fdde:ad00:beef:0:556:90c8:ffaf:b7a3$0$0$64
			args := strings.Split(sp[1], ",")
			dstaddr := args[0]
			datasize, err := strconv.Atoi(args[1])
			simplelogger.PanicIfError(err)
			timestamp, err := strconv.ParseUint(args[2], 10, 64)
			simplelogger.PanicIfError(err)
			hoplimit, err := strconv.Atoi(args[3])
			simplelogger.PanicIfError(err)
			srcnode.onPingReply(d.convertNodeMilliTime(srcnode, uint32(timestamp)), dstaddr, datasize, hoplimit)
		} else if sp[0] == "coap" {
			d.handleCoapEvent(srcnode, sp[1])
		} else if sp[0] == "parid" {
			// set partition id
			parid, err := strconv.ParseUint(sp[1], 16, 32)
			simplelogger.PanicIfError(err)
			srcnode.PartitionId = uint32(parid)
			d.vis.SetNodePartitionId(srcid, uint32(parid))
		} else if sp[0] == "router_added" {
			extaddr, err := strconv.ParseUint(sp[1], 16, 64)
			simplelogger.PanicIfError(err)
			if d.visOptions.RouterTable {
				d.vis.AddRouterTable(srcid, extaddr)
			}
		} else if sp[0] == "router_removed" {
			extaddr, err := strconv.ParseUint(sp[1], 16, 64)
			simplelogger.PanicIfError(err)
			if d.visOptions.RouterTable {
				d.vis.RemoveRouterTable(srcid, extaddr)
			}
		} else if sp[0] == "child_added" {
			extaddr, err := strconv.ParseUint(sp[1], 16, 64)
			simplelogger.PanicIfError(err)
			if d.visOptions.ChildTable {
				d.vis.AddChildTable(srcid, extaddr)
			}
		} else if sp[0] == "child_removed" {
			extaddr, err := strconv.ParseUint(sp[1], 16, 64)
			simplelogger.PanicIfError(err)
			if d.visOptions.ChildTable {
				d.vis.RemoveChildTable(srcid, extaddr)
			}
		} else if sp[0] == "parent" {
			extaddr, err := strconv.ParseUint(sp[1], 16, 64)
			simplelogger.PanicIfError(err)
			d.vis.SetParent(srcid, extaddr)
		} else if sp[0] == "joiner_state" {
			joinerState, err := strconv.Atoi(sp[1])
			simplelogger.PanicIfError(err)
			srcnode.onJoinerState(OtJoinerState(joinerState))
		} else if sp[0] == "extaddr" {
			extaddr, err := strconv.ParseUint(sp[1], 16, 64)
			simplelogger.PanicIfError(err)
			srcnode.onStatusPushExtAddr(extaddr)
		} else if sp[0] == "mode" {
			mode := ParseNodeMode(sp[1])
			d.vis.SetNodeMode(srcid, mode)
		} else {
			simplelogger.Warnf("unknown status push: %s=%s", sp[0], sp[1])
		}
	}
}

func (d *Dispatcher) AddNode(nodeid NodeId, cfg *NodeConfig) {
	simplelogger.AssertNil(d.nodes[nodeid])
	simplelogger.Infof("dispatcher add node %d", nodeid)
	node := newNode(d, nodeid, cfg)
	d.nodes[nodeid] = node
	d.alarmMgr.AddNode(nodeid)
	d.energyAnalyser.AddNode(nodeid, d.CurTime)
	d.vis.AddNode(nodeid, cfg.X, cfg.Y, cfg.RadioRange)
	d.radioModel.AddNode(nodeid, node.radioNode)
	d.setAlive(nodeid)

	if !d.cfg.Real {
		d.detectNodeExtAddr(node)
	}
}

// detectNodeExtAddr waits until node's extended address is emitted. Not to be used for real devices.
// This helps OTNS to make sure that the child process is ready to receive UDP events.
func (d *Dispatcher) detectNodeExtAddr(node *Node) {
	t0 := time.Now()
	deadline := t0.Add(time.Second * 10)
	for node.ExtAddr == InvalidExtAddr && time.Now().Before(deadline) {
		d.RecvEvents()
	}

	if node.ExtAddr == InvalidExtAddr {
		simplelogger.Panicf("expect node %d's extaddr to be valid, but failed", node.Id)
	} else {
		takeTime := time.Since(t0)
		simplelogger.Debugf("node %d's extaddr becomes valid in %v", node.Id, takeTime)
	}
}

func (d *Dispatcher) setNodeRloc16(srcid NodeId, rloc16 uint16) {
	node := d.nodes[srcid]
	simplelogger.AssertNotNil(node)

	simplelogger.Debugf("set node rloc: %x -> %x", node.Rloc16, rloc16)
	oldRloc16 := node.Rloc16
	if oldRloc16 != threadconst.InvalidRloc16 {
		// remove node from old rloc map
		d.rloc16Map.Remove(oldRloc16, node)
	}

	node.Rloc16 = rloc16
	if rloc16 != threadconst.InvalidRloc16 {
		// add node to the new rloc map
		d.rloc16Map.Add(rloc16, node)
	}

	d.vis.SetNodeRloc16(srcid, rloc16)
}

func (d *Dispatcher) visStatusPushTransmit(srcnode *Node, s string) {
	var fcf wpan.FrameControl

	// only visualize `transmit` status emitting in real mode because simulation nodes already have radio events visualized
	if !d.cfg.Real {
		return
	}

	parts := strings.Split(s, ",")

	if len(parts) < 3 {
		simplelogger.Panicf("invalid status push: transmit=%s", s)
	}

	channel, err := strconv.Atoi(parts[0])
	simplelogger.PanicIfError(err)
	fcfval, err := strconv.ParseUint(parts[1], 16, 16)
	simplelogger.PanicIfError(err)
	fcf = wpan.FrameControl(fcfval)

	seq, err := strconv.Atoi(parts[2])
	simplelogger.PanicIfError(err)

	dstAddrMode := fcf.DstAddrMode()

	visInfo := &visualize.MsgVisualizeInfo{
		Channel:      uint8(channel),
		FrameControl: fcf,
		Seq:          uint8(seq),
	}

	if dstAddrMode == wpan.DstAddrModeExtended {
		dstExtend, err := strconv.ParseUint(parts[3], 16, 64)
		simplelogger.PanicIfError(err)

		visInfo.DstAddrExtended = dstExtend

		dstnode := d.extaddrMap[dstExtend]
		if dstnode != srcnode && dstnode != nil {
			d.visSend(srcnode.Id, dstnode.Id, visInfo)
		} else {
			d.visSend(srcnode.Id, InvalidNodeId, visInfo)
		}
	} else if dstAddrMode == wpan.DstAddrModeShort {
		dstShortVal, err := strconv.ParseUint(parts[3], 16, 16)
		simplelogger.PanicIfError(err)

		dstShort := uint16(dstShortVal)
		visInfo.DstAddrShort = dstShort

		if dstShort != threadconst.BroadcastRloc16 {
			// unicast message should only be dispatched to target node with the rloc16
			dstnodes := d.rloc16Map[dstShort]

			if len(dstnodes) > 0 {
				for _, dstnode := range dstnodes {
					d.visSend(srcnode.Id, dstnode.Id, visInfo)
				}
			} else {
				d.visSend(srcnode.Id, InvalidNodeId, visInfo)
			}
		} else {
			d.vis.Send(srcnode.Id, BroadcastNodeId, visInfo)
		}
	} else {
		d.vis.Send(srcnode.Id, BroadcastNodeId, visInfo)
	}
}

func (d *Dispatcher) visSendFrame(srcid NodeId, dstid NodeId, pktframe *wpan.MacFrame) {
	d.visSend(srcid, dstid, &visualize.MsgVisualizeInfo{
		Channel:         pktframe.Channel,
		FrameControl:    pktframe.FrameControl,
		Seq:             pktframe.Seq,
		DstAddrShort:    pktframe.DstAddrShort,
		DstAddrExtended: pktframe.DstAddrExtended,
	})
}

func (d *Dispatcher) visSend(srcid NodeId, dstid NodeId, visInfo *visualize.MsgVisualizeInfo) {
	if dstid == BroadcastNodeId {
		if visInfo.FrameControl.FrameType() == wpan.FrameTypeAck {
			if !d.visOptions.AckMessage {
				return
			}
		} else {
			if !d.visOptions.BroadcastMessage {
				return
			}
		}
	} else {
		if !d.visOptions.UnicastMessage {
			return
		}
	}

	d.vis.Send(srcid, dstid, visInfo)
}

func (d *Dispatcher) advanceTime(ts uint64) {
	simplelogger.AssertTrue(d.CurTime <= ts, "%v > %v", d.CurTime, ts)
	simplelogger.AssertTrue(d.CurTime <= d.eventQueue.NextTimestamp())
	if d.CurTime < ts {
		oldTime := d.CurTime
		d.CurTime = ts
		simplelogger.AssertTrue(d.CurTime <= d.eventQueue.NextTimestamp())
		elapsedTime := int64(d.CurTime - d.speedStartTime)
		elapsedRealTime := time.Since(d.speedStartRealTime) / time.Microsecond
		if elapsedRealTime > 0 && ts/1000000 != oldTime/1000000 {
			d.vis.AdvanceTime(ts, float64(elapsedTime)/float64(elapsedRealTime))

			if d.energyAnalyser != nil && ts%energy.ComputePeriod == 0 {
				d.energyAnalyser.StoreNetworkEnergy(ts)
				d.vis.UpdateNodesEnergy(d.energyAnalyser.GetLatestEnergyOfNodes(), ts, true)
			}
		}

		if d.cfg.Real {
			d.syncAllNodes()
		}
	}
}

func (d *Dispatcher) PostAsync(trivial bool, task func()) {
	if trivial {
		select {
		case d.taskChan <- task:
			break
		default:
			break
		}
	} else {
		d.taskChan <- task
	}
}

func (d *Dispatcher) handleTasks() {
	defer func() {
		err := recover()
		if err != nil {
			simplelogger.Errorf("dispatcher handle task failed: %+v", err)
		}
	}()

loop:
	for {
		select {
		case t := <-d.taskChan:
			t()
			// continue
		default:
			break loop
		}
	}
}

func (d *Dispatcher) WatchNode(nodeid NodeId) {
	d.watchingNodes[nodeid] = struct{}{}
}

func (d *Dispatcher) UnwatchNode(nodeid NodeId) {
	delete(d.watchingNodes, nodeid)
}

func (d *Dispatcher) isWatching(nodeid NodeId) bool {
	_, ok := d.watchingNodes[nodeid]
	return ok
}

func (d *Dispatcher) GetAliveCount() int {
	return len(d.aliveNodes)
}

func (d *Dispatcher) GetNode(id NodeId) *Node {
	return d.nodes[id]
}

func (d *Dispatcher) GetFailedCount() int {
	failCount := 0
	for _, dn := range d.nodes {
		if dn.IsFailed() {
			failCount += 1
		}
	}
	return failCount
}

func (d *Dispatcher) SetNodePos(id NodeId, x, y int) {
	node := d.nodes[id]
	simplelogger.AssertNotNil(node)

	node.X, node.Y = x, y
	node.radioNode.SetNodePos(x, y)
	d.vis.SetNodePos(id, x, y)
}

func (d *Dispatcher) DeleteNode(id NodeId) {
	node := d.nodes[id]
	simplelogger.AssertNotNil(node)

	delete(d.nodes, id)
	delete(d.aliveNodes, id)
	delete(d.watchingNodes, id)
	if node.Rloc16 != threadconst.InvalidRloc16 {
		d.rloc16Map.Remove(node.Rloc16, node)
	}
	if node.ExtAddr != InvalidExtAddr {
		simplelogger.AssertTrue(d.extaddrMap[node.ExtAddr] == node)
		delete(d.extaddrMap, node.ExtAddr)
	}
	d.alarmMgr.DeleteNode(id)
	d.deletedNodes[id] = struct{}{}
	d.energyAnalyser.DeleteNode(id)
	d.vis.DeleteNode(id)
	d.radioModel.DeleteNode(id)
}

// SetNodeFailed sets the radio of the node to failed (true) or operational (false) state.
// Setting this will disable the automatic failure control (FailureCtrl).
func (d *Dispatcher) SetNodeFailed(id NodeId, fail bool) {
	node := d.nodes[id]
	simplelogger.AssertNotNil(node)

	// if radio is set to on/off explicitly, failureCtrl should not be used anymore
	node.SetFailTime(NonFailTime)

	if fail {
		node.Fail()
	} else {
		node.Recover()
	}
}

func (d *Dispatcher) SetSpeed(f float64) {
	ns := d.normalizeSpeed(f)
	if ns == d.speed {
		return
	}

	// sync the speed start time with the current time
	d.speedStartRealTime = time.Now()
	d.speedStartTime = d.CurTime
	d.speed = ns
	d.vis.SetSpeed(ns)
}

func (d *Dispatcher) normalizeSpeed(f float64) float64 {
	if f <= 0 {
		f = 0
	} else if f >= MaxSimulateSpeed {
		f = MaxSimulateSpeed
	}
	return f
}

func (d *Dispatcher) GetSpeed() float64 {
	return d.speed
}

func (d *Dispatcher) GetGlobalMessageDropRatio() float64 {
	return d.globalPacketLossRatio
}

func (d *Dispatcher) SetGlobalPacketLossRatio(plr float64) {
	if plr > 1 {
		plr = 1
	} else if plr < 0 {
		plr = 0
	}
	d.globalPacketLossRatio = plr
}

func (d *Dispatcher) convertNodeMilliTime(node *Node, milliTime uint32) uint64 {
	ts := node.CreateTime + uint64(milliTime)*1000 // convert to us

	// because timestamp on node is uint32_t, so it can not exceed 1293 hours, after that the timestamp rewinds from zero
	// so we should calculate the real timestamp.
	// This assumes that the node is not far behind in time
	for ts+(0xffffffff*1000) < d.CurTime {
		ts += 0xffffffff * 1000
	}

	return ts
}

func (d *Dispatcher) onStatusPushExtAddr(node *Node, oldExtAddr uint64) {
	if oldExtAddr == InvalidExtAddr {
		simplelogger.AssertTrue(d.extaddrMap[oldExtAddr] == nil)
	} else {
		simplelogger.AssertTrue(d.extaddrMap[oldExtAddr] == node)
		delete(d.extaddrMap, oldExtAddr)
	}
	simplelogger.AssertNil(d.extaddrMap[node.ExtAddr])

	d.extaddrMap[node.ExtAddr] = node
	d.vis.OnExtAddrChange(node.Id, node.ExtAddr)
}

func (d *Dispatcher) GetVisualizationOptions() VisualizationOptions {
	return d.visOptions
}

func (d *Dispatcher) SetVisualizationOptions(opts VisualizationOptions) {
	simplelogger.Debugf("dispatcher set visualization options: %+v", opts)
	d.visOptions = opts
}

// NotifyExit notifies the dispatcher that the node process has exited.
func (d *Dispatcher) NotifyExit(nodeid NodeId) {
	if !d.cfg.Real {
		d.setSleeping(nodeid)
	}
}

func (d *Dispatcher) NotifyCommand(nodeid NodeId) {
	d.setAlive(nodeid)
}

func (d *Dispatcher) dumpPacket(item *Event) {
	sb := strings.Builder{}
	_, _ = fmt.Fprintf(&sb, "DUMP:PACKET:%d:%d:", item.Timestamp, item.NodeId)
	for _, b := range item.Data {
		_, _ = fmt.Fprintf(&sb, "%02X", b)
	}

	_, _ = fmt.Fprintf(os.Stdout, "%s\n", sb.String())
}

func (d *Dispatcher) setNodeRole(node *Node, role OtDeviceRole) {
	node.Role = role
	d.vis.SetNodeRole(node.Id, role)
}

func (d *Dispatcher) handleCoapEvent(node *Node, argsStr string) {
	var err error

	if d.coaps == nil {
		// Coaps not enabled
		return
	}

	args := strings.Split(argsStr, ",")
	simplelogger.AssertTrue(len(args) > 0)
	action := args[0]

	if action == "send" || action == "recv" || action == "send_error" {
		var messageId, coapType, coapCode, port int

		simplelogger.AssertTrue(len(args) >= 7)

		messageId, err = strconv.Atoi(args[1])
		simplelogger.PanicIfError(err)

		coapType, err = strconv.Atoi(args[2])
		simplelogger.PanicIfError(err)

		coapCode, err = strconv.Atoi(args[3])
		simplelogger.PanicIfError(err)

		uri := args[4]

		ip := args[5]

		port, err = strconv.Atoi(args[6])
		simplelogger.PanicIfError(err)

		if action == "send" {
			d.coaps.OnSend(d.CurTime, node.Id, messageId, CoapType(coapType), CoapCode(coapCode), uri, ip, port)
		} else if action == "recv" {
			d.coaps.OnRecv(d.CurTime, node.Id, messageId, CoapType(coapType), CoapCode(coapCode), uri, ip, port)
		} else {
			simplelogger.AssertTrue(len(args) >= 7)
			threadError := args[6]

			d.coaps.OnSendError(node.Id, messageId, CoapType(coapType), CoapCode(coapCode), uri, ip, port, threadError)
		}
	} else {
		simplelogger.Warnf("unknown coap event: %+v", args)
	}
}

func (d *Dispatcher) EnableCoaps() {
	if d.coaps == nil {
		d.coaps = newCoapsHandler()
	}
}

func (d *Dispatcher) CollectCoapMessages() []*CoapMessage {
	if d.coaps != nil {
		return d.coaps.DumpMessages()
	} else {
		return nil
	}
}

func (d *Dispatcher) SetEnergyAnalyser(e *energy.EnergyAnalyser) {
	d.energyAnalyser = e
}

func (d *Dispatcher) GetRadioModel() radiomodel.RadioModel {
	return d.radioModel
}

func (d *Dispatcher) SetRadioModel(model radiomodel.RadioModel) {
	d.radioModel = model
}

func (d *Dispatcher) handleRadioState(node *Node, evt *Event) {
	simplelogger.AssertNotNil(node)
	if node.radioNode.RadioState != evt.RadioStateData.State {
		simplelogger.Debugf("%v hrs RadioState=%+v ts=%v", node.Id, evt.RadioStateData.State, evt.Timestamp)
	}
	node.radioNode.SetRadioState(evt.RadioStateData.State)
	node.radioNode.SetChannel(evt.RadioStateData.Channel)

	if d.energyAnalyser != nil {
		radioEnergy := d.energyAnalyser.GetNode(node.Id)
		simplelogger.AssertNotNil(radioEnergy)
		radioEnergy.SetRadioState(node.radioNode.RadioState, d.CurTime)
	}
}

func min(t1 uint64, t2 uint64) uint64 {
	if t1 <= t2 {
		return t1
	}
	return t2
}
