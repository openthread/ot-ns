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

package dispatcher

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
	"net"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/openthread/ot-ns/dissectpkt"
	"github.com/openthread/ot-ns/dissectpkt/wpan"
	"github.com/openthread/ot-ns/energy"
	. "github.com/openthread/ot-ns/event"
	"github.com/openthread/ot-ns/logger"
	"github.com/openthread/ot-ns/pcap"
	"github.com/openthread/ot-ns/prng"
	"github.com/openthread/ot-ns/progctx"
	"github.com/openthread/ot-ns/radiomodel"
	. "github.com/openthread/ot-ns/types"
	"github.com/openthread/ot-ns/visualize"
)

type CallbackHandler interface {
	// OnUartWrite Notifies that the node's UART was written with data.
	OnUartWrite(nodeid NodeId, data []byte)

	// OnLogWrite Notifies that a log item wsa written to the node's log.
	OnLogWrite(nodeid NodeId, data []byte)

	// OnNextEventTime Notifies that the Dispatcher simulation will move shortly to the next event time.
	OnNextEventTime(nextTimeUs uint64)

	// OnRfSimEvent Notifies that Dispatcher received an OT-RFSIM platform event that it didn't handle itself.
	OnRfSimEvent(nodeid NodeId, evt *Event)
}

// goDuration represents a particular duration of the simulation at a given speed.
type goDuration struct {
	duration time.Duration
	done     chan error // signals end of the simulation duration and success (nil) or some error.
	speed    float64    // a speedup value, or DefaultDispatcherSpeed.
	cancel   bool       // if set to true, the simulation duration will be cancelled early.
}

type Dispatcher struct {
	CurTime  uint64
	Counters struct {
		// Received event counters
		AlarmEvents      uint64
		RadioEvents      uint64
		StatusPushEvents uint64
		UartWriteEvents  uint64
		LogWriteEvents   uint64
		OtherEvents      uint64
		// Packet-related event dispatching counters
		DispatchByExtAddrSucc   uint64
		DispatchByExtAddrFail   uint64
		DispatchByShortAddrSucc uint64
		DispatchByShortAddrFail uint64
		DispatchAllInRange      uint64
		// Other counters
		TopologyChanges uint64
	}

	ctx                   *progctx.ProgCtx
	cfg                   Config
	cbHandler             CallbackHandler
	udpln                 net.Listener
	socketName            string
	eventChan             chan *Event
	waitGroup             sync.WaitGroup
	waitGroupNodes        sync.WaitGroup
	currentGoDuration     goDuration
	pauseTime             uint64
	alarmMgr              *alarmMgr
	eventQueue            *sendQueue
	nodes                 map[NodeId]*Node
	nodesArray            []*Node
	deletedNodes          map[NodeId]struct{}
	aliveNodes            map[NodeId]struct{}
	pcap                  pcap.File
	pcapFrameChan         chan pcap.Frame
	vis                   visualize.Visualizer
	taskChan              chan func()
	speed                 float64
	speedStartRealTime    time.Time
	lastEnergyVizTime     uint64
	speedStartTime        uint64
	extaddrMap            map[uint64]*Node
	rloc16Map             rloc16Map
	goDurationChan        chan goDuration
	globalPacketLossRatio float64
	visOptions            VisualizationOptions
	coaps                 *coapsHandler
	watchingNodes         map[NodeId]struct{}
	energyAnalyser        *energy.EnergyAnalyser
	stopped               bool
	radioModel            radiomodel.RadioModel
	oldStats              NodeStats
	timeWinStats          TimeWindowStats
}

func NewDispatcher(ctx *progctx.ProgCtx, cfg *Config, cbHandler CallbackHandler) *Dispatcher {
	logger.AssertTrue(!cfg.Realtime || cfg.Speed == 1)
	var err error
	ln, unixSocketFile := newUnixSocket(cfg.SimulationId)
	vis := visualize.NewNopVisualizer()

	d := &Dispatcher{
		ctx:                ctx,
		cfg:                *cfg,
		cbHandler:          cbHandler,
		udpln:              ln,
		socketName:         unixSocketFile,
		eventChan:          make(chan *Event, 10000),
		eventQueue:         newSendQueue(),
		alarmMgr:           newAlarmMgr(),
		nodes:              make(map[NodeId]*Node),
		nodesArray:         make([]*Node, 0),
		deletedNodes:       map[NodeId]struct{}{},
		aliveNodes:         make(map[NodeId]struct{}),
		extaddrMap:         map[uint64]*Node{},
		rloc16Map:          rloc16Map{},
		pcapFrameChan:      make(chan pcap.Frame, 100000),
		speed:              cfg.Speed,
		speedStartRealTime: time.Now(),
		vis:                vis,
		taskChan:           make(chan func(), 10000),
		watchingNodes:      map[NodeId]struct{}{},
		goDurationChan:     make(chan goDuration, 1),
		visOptions:         defaultVisualizationOptions(),
		stopped:            false,
		oldStats:           NodeStats{},
		timeWinStats:       defaultTimeWindowStats(),
	}
	d.speed = d.normalizeSpeed(d.speed)
	if d.cfg.PcapEnabled {
		d.pcap, err = pcap.NewFile("current.pcap", cfg.PcapFrameType)
		logger.PanicIfError(err)
		d.waitGroup.Add(1)
		go d.pcapFrameWriter()
	}

	d.waitGroup.Add(1)
	go d.eventsReader()

	d.vis.SetSpeed(d.speed)
	logger.Infof("dispatcher started: cfg=%+v", *cfg)

	return d
}

func newUnixSocket(socketId int) (net.Listener, string) {
	err := os.MkdirAll("/tmp/otns", 0777)
	logger.FatalIfError(err, err)
	unixSocketFile := fmt.Sprintf("/tmp/otns/socket_dispatcher_%d", socketId) // remove old one
	err = os.RemoveAll(unixSocketFile)
	logger.FatalIfError(err, err)
	ln, err := net.Listen("unix", unixSocketFile)
	logger.FatalIfError(err, err)
	return ln, unixSocketFile
}

func (d *Dispatcher) Stop() {
	if d.stopped {
		return
	}

	d.stopped = true
	defer logger.Debugf("dispatcher exit.")
	logger.Debugf("stopping dispatcher ...")
	d.ctx.Cancel("dispatcher-stop")
	d.GoCancel()        // cancel current simulation period
	_ = d.udpln.Close() // close socket to stop d.eventsReader accepting new clients.
	if d.cfg.PhyTxStats {
		d.finalizeTimeWindowStats()
	}

	d.vis.Stop()
	close(d.pcapFrameChan)
	logger.Tracef("waiting for dispatcher threads to stop ...")
	d.waitGroup.Wait()
}

func (d *Dispatcher) isStopping() bool {
	select {
	case <-d.ctx.Done():
		return true
	default:
		return false
	}
}

func (d *Dispatcher) GetConfig() *Config {
	return &d.cfg
}

func (d *Dispatcher) GetUnixSocketName() string {
	return d.socketName
}

// Nodes returns an array of dispatcher Node pointers, sorted on NodeId.
func (d *Dispatcher) Nodes() []*Node {
	return d.nodesArray
}

func (d *Dispatcher) reconstructNodesArray() {
	d.nodesArray = make([]*Node, 0, len(d.nodes))
	nodeIds := make([]NodeId, 0, len(d.nodes))
	for nodeId := range d.nodes {
		nodeIds = append(nodeIds, nodeId)
	}
	sort.Ints(nodeIds)
	for _, nodeId := range nodeIds {
		d.nodesArray = append(d.nodesArray, d.nodes[nodeId])
	}
}

func (d *Dispatcher) Go(duration time.Duration) <-chan error {
	logger.AssertTrue(duration >= 0)
	done := make(chan error, 1)
	d.goDurationChan <- goDuration{
		duration: duration,
		done:     done,
		speed:    DefaultDispatcherSpeed,
	}
	return done
}

func (d *Dispatcher) GoAtSpeed(duration time.Duration, speed float64) <-chan error {
	logger.AssertTrue(speed >= 0.0 && duration >= 0)
	done := make(chan error, 1)
	d.goDurationChan <- goDuration{
		duration: duration,
		done:     done,
		speed:    speed,
	}
	return done
}

// GoCancel cancels the current Go....() operation and pauses the simulation at d.CurTime.
func (d *Dispatcher) GoCancel() <-chan error {
	d.currentGoDuration.cancel = true
	return d.currentGoDuration.done
}

func (d *Dispatcher) Run() {
	defer d.ctx.WaitDone("dispatcher")

	done := d.ctx.Done()
loop:
	for {
		select {
		case t := <-d.taskChan:
			t()
		case duration := <-d.goDurationChan:
			d.currentGoDuration = duration
			if len(d.nodes) == 0 || duration.duration < 0 {
				// no nodes or no sim progress, sleep for a small duration to avoid high cpu
				d.RecvEvents()
				time.Sleep(time.Millisecond * 10)
			} else {
				logger.AssertTrue(d.CurTime == d.pauseTime)
				d.goSimulateForDuration(duration)
				logger.AssertTrue(d.CurTime == d.pauseTime)

				d.syncAllNodes()
				if d.pcap != nil {
					_ = d.pcap.Sync()
				}
			}
			close(duration.done)
		case <-done:
			break loop
		}
	}

	// handle all remaining tasks - other goroutines may be blocking on task completion.
	// handle all remaining go-duration requests.
loop2:
	for {
		select {
		case t := <-d.taskChan:
			t()
		case duration := <-d.goDurationChan:
			close(duration.done)
		default:
			break loop2
		}
	}
}

func (d *Dispatcher) goSimulateForDuration(duration goDuration) {
	var postSpeed float64

	// sync the speed start time with the current time
	d.speedStartRealTime = time.Now()
	d.speedStartTime = d.CurTime

	if duration.speed != DefaultDispatcherSpeed {
		postSpeed = d.speed
		d.SetSpeed(duration.speed) // adapt speed for particular period.
	}

	// determine pauseTime (after duration)
	d.pauseTime = d.CurTime + uint64(duration.duration/time.Microsecond)
	if d.pauseTime > Ever {
		d.pauseTime = Ever
	}

	logger.AssertTrue(d.CurTime <= d.pauseTime)

	for d.CurTime <= d.pauseTime {
		d.handleTasks()

		if d.currentGoDuration.cancel || d.isStopping() {
			break
		}

		// keep receiving events from OT nodes until all are asleep i.e. will not produce more events.
		d.RecvEvents()
		d.syncAliveNodes() // normally there should not be any alive nodes anymore.

		if len(d.aliveNodes) == 0 {
			// all are asleep now - process the next Events in queue, either alarm or other type, for a single time.
			goon := d.processNextEvent(d.speed)
			logger.AssertTrue(d.CurTime <= d.pauseTime)

			if !goon && len(d.aliveNodes) == 0 {
				d.cbHandler.OnNextEventTime(d.pauseTime)
				d.radioModel.OnNextEventTime(d.pauseTime)
				d.advanceTime(d.pauseTime) // if nothing more to do before d.pauseTime.
				break
			}
		}
	}

	if duration.speed != DefaultDispatcherSpeed { // restore original speed after period with custom speed set.
		d.SetSpeed(postSpeed)
	}
	if d.pauseTime > d.CurTime { // if we e.g. cancelled period simulation early, and pauseTime not reached.
		d.pauseTime = d.CurTime
	}
}

// handleRecvEvent is the central handler for all events externally received from OpenThread nodes.
// It may only process events immediately that are to be executed at time d.CurTime. Future events
// will need to be queued (scheduled).
func (d *Dispatcher) handleRecvEvent(evt *Event) {
	nodeid := evt.NodeId
	node := d.nodes[nodeid]
	if node == nil {
		logger.Warnf("Event (type %v) received from unknown Node %v, discarding.", evt.Type, evt.NodeId)
		return
	}

	node.conn = evt.Conn      // store socket connection for this node.
	evt.Timestamp = d.CurTime // timestamp the incoming event

	// TODO document this use (for alarm messages)
	delay := evt.Delay
	if delay >= 2147483647 {
		delay = Ever
	}

	switch evt.Type {
	case EventTypeAlarmFired:
		d.Counters.AlarmEvents += 1
		if evt.MsgId == node.msgId { // if OT-node has seen my last sent event (so is done processing)
			d.setSleeping(node.Id)
		}
		d.alarmMgr.SetTimestamp(nodeid, d.CurTime+delay) // schedule future wake-up of node
	case EventTypeRadioCommStart,
		EventTypeRadioState,
		EventTypeRadioChannelSample:
		d.Counters.RadioEvents += 1
		d.eventQueue.Add(evt)
	case EventTypeStatusPush:
		d.Counters.StatusPushEvents += 1
		d.handleStatusPush(node, string(evt.Data))
	case EventTypeUartWrite:
		d.Counters.UartWriteEvents += 1
		d.cbHandler.OnUartWrite(node.Id, evt.Data)
	case EventTypeLogWrite:
		d.Counters.LogWriteEvents += 1
		d.cbHandler.OnLogWrite(node.Id, evt.Data)
	case EventTypeExtAddr:
		d.Counters.OtherEvents += 1
		var extaddr = binary.BigEndian.Uint64(evt.Data[0:8])
		node.onStatusPushExtAddr(extaddr)
	case EventTypeNodeInfo:
		d.Counters.OtherEvents += 1
	case EventTypeNodeDisconnected:
		d.Counters.OtherEvents += 1
		logger.Debugf("%s socket disconnected.", node)
		d.setSleeping(node.Id)
		d.alarmMgr.SetTimestamp(node.Id, Ever)
	default:
		d.Counters.OtherEvents += 1
		d.cbHandler.OnRfSimEvent(node.Id, evt)
	}
}

// RecvEvents receives events from nodes, and handles these, until there is no more alive node.
func (d *Dispatcher) RecvEvents() int {
	done := d.ctx.Done()
	count := 0
	isExiting := false
	blockTimeout := time.After(DefaultReadTimeout)

loop:
	for {
		shouldBlock := len(d.aliveNodes) > 0
		if shouldBlock {
			select {
			case evt := <-d.eventChan: // get new event
				count += 1
				d.handleRecvEvent(evt)
			case <-blockTimeout: // timeout
				break loop
			case <-done:
				if !isExiting {
					blockTimeout = time.After(time.Millisecond * 250)
					isExiting = true
				}
				time.Sleep(time.Millisecond * 10)
				break
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

// processNextEvent processes all next events from the eventQueue for the next time instant.
// Returns true if the simulation needs to continue, or false if not (e.g. it's time to pause).
func (d *Dispatcher) processNextEvent(simSpeed float64) bool {
	logger.AssertTrue(d.CurTime <= d.pauseTime)
	logger.AssertTrue(simSpeed >= 0)

	// fetch time of next event
	nextAlarmTime := d.alarmMgr.NextTimestamp()
	nextSendTime := d.eventQueue.NextTimestamp()
	logger.AssertTrue(nextSendTime >= d.CurTime && nextAlarmTime >= d.CurTime)

	nextEventTime := min(nextAlarmTime, nextSendTime)

	// convert nextEventTime to real time
	if simSpeed < MaxSimulateSpeed {
		var sleepUntilTime = nextEventTime
		if sleepUntilTime > d.pauseTime {
			sleepUntilTime = d.pauseTime
		}

		var needSleepDuration time.Duration
		if simSpeed <= 0 {
			needSleepDuration = time.Hour
		} else {
			needSleepDuration = time.Duration(float64(sleepUntilTime-d.speedStartTime)/simSpeed) * time.Microsecond
		}
		sleepUntilRealTime := d.speedStartRealTime.Add(needSleepDuration)
		now := time.Now()
		sleepTime := sleepUntilRealTime.Sub(now)

		if sleepTime > 0 {
			if sleepTime > time.Millisecond*10 {
				sleepTime = time.Millisecond * 10 // max cap to keep program responsive
			}
			time.Sleep(sleepTime)

			// move simulation time ahead at speed, even during periods without sim events.
			curTime := d.speedStartTime + uint64(float64(time.Since(d.speedStartRealTime)/time.Microsecond)*simSpeed)
			if curTime > d.pauseTime {
				curTime = d.pauseTime
			}
			if curTime < nextEventTime {
				d.advanceTime(curTime)
			}
			return true
		}
	}

	if nextEventTime > d.pauseTime {
		return false
	}
	d.cbHandler.OnNextEventTime(nextEventTime)
	d.radioModel.OnNextEventTime(nextEventTime)
	d.advanceTime(nextEventTime)

	// process (if any) all queued events, that happen at exactly procUntilTime
	procUntilTime := nextEventTime
	for nextEventTime <= procUntilTime {
		if nextAlarmTime <= nextSendTime {
			// process next alarm
			nextAlarm := d.alarmMgr.NextAlarm()
			logger.AssertNotNil(nextAlarm)
			node := d.nodes[nextAlarm.NodeId]
			if node != nil {
				d.advanceNodeTime(node, nextAlarm.Timestamp, false)
			}
		} else {
			// process next event from the queue
			evt := d.eventQueue.PopNext()
			logger.AssertTrue(evt.Timestamp == nextEventTime)
			logger.AssertTrue(nextAlarmTime == d.CurTime || nextSendTime == d.CurTime)
			node := d.nodes[evt.NodeId]
			if node != nil {
				// execute event - either a msg to be dispatched, or handled internally.
				if !evt.MustDispatch {
					switch evt.Type {
					case EventTypeAlarmFired:
						d.advanceNodeTime(node, evt.Timestamp, false)
					case EventTypeRadioLog:
						node.logger.Tracef("%s", string(evt.Data))
					case EventTypeRadioCommStart:
						if evt.RadioCommData.Error == OT_TX_TYPE_INTF {
							// for interference transmissions, visualized here.
							d.visSendInterference(evt.NodeId, BroadcastNodeId, evt.RadioCommData)
						}
						d.radioModel.HandleEvent(node.RadioNode, d.eventQueue, evt)
					case EventTypeRadioState:
						d.handleRadioState(node, evt)
						d.radioModel.HandleEvent(node.RadioNode, d.eventQueue, evt)
					default:
						d.radioModel.HandleEvent(node.RadioNode, d.eventQueue, evt)
					}
				} else {
					switch evt.Type {
					case EventTypeRadioCommStart:
						d.sendRadioCommRxStartEvents(node, evt)
					case EventTypeRadioRxDone:
						d.sendRadioCommRxDoneEvents(node, evt)
					default:
						if d.radioModel.OnEventDispatch(node.RadioNode, node.RadioNode, evt) {
							node.sendEvent(evt)
						}
					}
				}
			} else if evt.NodeId > 0 {
				logger.Warnf("processNextEvent() with deleted/unknown node %v: %v", evt.NodeId, evt)
			}
		}
		nextAlarmTime = d.alarmMgr.NextTimestamp()
		nextSendTime = d.eventQueue.NextTimestamp()
		nextEventTime = min(nextAlarmTime, nextSendTime)
	}

	return len(d.nodes) > 0
}

func (d *Dispatcher) eventsReader() {
	defer d.waitGroup.Done()
	defer logger.Tracef("dispatcher node socket threads stopped.")
	defer os.RemoveAll(d.socketName) // delete Unix socket file when done.
	defer d.udpln.Close()

	logger.Debugf("dispatcher listening on socket %s ...", d.socketName)
	for {
		// Wait for OT nodes to connect.
		conn, err := d.udpln.Accept()
		if err != nil || d.isStopping() {
			if conn != nil {
				_ = conn.Close()
			}
			if !d.isStopping() {
				logger.Panicf("connection Accept() failed: %v", err)
			}
			break
		}

		// Handle the new connection in a separate goroutine.
		d.waitGroupNodes.Add(1)
		go func(myConn net.Conn) {
			defer d.waitGroupNodes.Done()
			defer myConn.Close()

			buf := make([]byte, 65536)
			myNodeId := 0

			for {
				n, err := myConn.Read(buf)

				if errors.Is(err, io.EOF) {
					break
				} else if err != nil {
					logger.NodeLogf(myNodeId, logger.ErrorLevel, "closing socket after read error: %+v", err)
					break
				}

				bufIdx := 0
				for bufIdx < n {
					evt := &Event{}
					nextEventOffset := evt.Deserialize(buf[bufIdx:n])
					if nextEventOffset == 0 { // a complete event wasn't found.
						logger.Panicf("Node %d - Too many events, or incorrect event data - may increase 'buf' size in Dispatcher.eventsReader()", myNodeId)
					}
					bufIdx += nextEventOffset
					// First event received should be NodeInfo type. From this, we learn nodeId.
					if myNodeId == 0 && evt.Type == EventTypeNodeInfo {
						myNodeId = evt.NodeInfoData.NodeId
						logger.AssertTrue(myNodeId > 0)
						logger.Debugf("Init event received from new Node %d", myNodeId)
					}
					evt.NodeId = myNodeId
					evt.Conn = myConn
					d.eventChan <- evt
				}

				if n > len(buf)/2 { // increase buf size when needed
					buf = make([]byte, len(buf)*2)
					logger.NodeLogf(myNodeId, logger.WarnLevel, "increasing eventsReader() buf size to: %d KB", len(buf)/1024)
				}
			}

			// Once the socket is disconnected, signal one last event.
			d.eventChan <- &Event{
				Delay:  0,
				Type:   EventTypeNodeDisconnected,
				NodeId: myNodeId,
				Conn:   nil,
			}
		}(conn)
	}

	logger.Tracef("waiting for dispatcher node socket threads to stop ...")
	d.waitGroupNodes.Wait() // wait for all node goroutines to stop before closing eventsReader.
}

func (d *Dispatcher) advanceNodeTime(node *Node, timestamp uint64, force bool) {
	logger.AssertNotNil(node)

	oldTime := node.CurTime
	if timestamp <= oldTime && !force {
		// node time was already equal to or newer than the requested timestamp
		return
	}

	msg := &Event{
		Type:      EventTypeAlarmFired,
		Timestamp: timestamp,
	}
	node.sendEvent(msg) // move the OT-node's virtual-time to new time using an alarm msg.
}

// sendRadioCommRxStartEvents dispatches an event to nearby nodes eligible to receiving the frame.
// It also logs the frame (pcap and/or dump) and visualizes the sending.
func (d *Dispatcher) sendRadioCommRxStartEvents(srcNode *Node, evt *Event) {
	logger.AssertTrue(evt.Type == EventTypeRadioCommStart)
	if srcNode.isFailed {
		return // failed source node can't send - don't send
	}

	// record the to-be-received frame in Pcap file
	if d.cfg.PcapEnabled {
		d.pcapFrameChan <- pcap.Frame{
			Timestamp: evt.Timestamp,
			Data:      evt.Data[RadioMessagePsduOffset:],
			Channel:   evt.RadioCommData.Channel,
			Rssi:      float32(evt.RadioCommData.PowerDbm), // uses Tx power as virtual-sniffer's RSSI.
		}
	}

	// record the sent frame in Dump logging - once, at time of Tx start.
	if d.cfg.DumpPackets {
		d.dumpPacket(evt)
	}

	// dispatch the message to all in range that are receiving.
	neighborNodes := map[NodeId]*Node{}
	for _, dstNode := range d.nodesArray {
		if d.checkRadioReachable(srcNode, dstNode) {
			d.sendOneRadioFrame(evt, srcNode, dstNode)
			neighborNodes[dstNode.Id] = dstNode
		}
	}
	d.Counters.DispatchAllInRange++

	// visualize the transmission and (intended) reception of the frame, based on addressing.
	pktinfo := dissectpkt.Dissect(evt.Data)
	pktFrame := pktinfo.MacFrame
	dstAddrMode := pktFrame.FrameControl.DestAddrMode()

	if dstAddrMode == wpan.AddrModeExtended {
		// unicast ExtAddr frame
		dstNode := d.extaddrMap[pktFrame.DstAddrExtended]
		if dstNode != nil && neighborNodes[dstNode.Id] != nil {
			d.visSendFrame(srcNode.Id, dstNode.Id, pktFrame, evt.RadioCommData)
		} else {
			// extAddr didn't exist or was out of range
			d.visSendFrame(srcNode.Id, InvalidNodeId, pktFrame, evt.RadioCommData)
		}
	} else if dstAddrMode == wpan.AddrModeShort && pktFrame.DstAddrShort != BroadcastRloc16 {
		// unicast short addr frame. May go to multiple if multiple nodes use same short addr.
		dstNodes := d.rloc16Map[pktFrame.DstAddrShort]

		if len(dstNodes) > 0 {
			for _, dstNode := range dstNodes {
				if neighborNodes[dstNode.Id] != nil {
					d.visSendFrame(srcNode.Id, dstNode.Id, pktFrame, evt.RadioCommData)
				}
			}
		} else {
			d.visSendFrame(srcNode.Id, InvalidNodeId, pktFrame, evt.RadioCommData)
		}
	} else {
		// broadcast frame
		d.visSendFrame(srcNode.Id, BroadcastNodeId, pktFrame, evt.RadioCommData)
	}
}

// sendRadioCommRxDoneEvents dispatches an event where >=1 nodes may receive a frame that is done
// being transmitted, and determines who receives it.
func (d *Dispatcher) sendRadioCommRxDoneEvents(srcNode *Node, evt *Event) {
	logger.AssertTrue(evt.Type == EventTypeRadioRxDone)

	if srcNode.isFailed {
		return // source node can't send - don't send, and don't log in pcap.
	}

	// try to dispatch the message by address directly to the right node
	pktinfo := dissectpkt.Dissect(evt.Data)
	pktFrame := pktinfo.MacFrame
	dispatchedByDstAddr := false
	dstAddrMode := pktFrame.FrameControl.DestAddrMode()

	if dstAddrMode == wpan.AddrModeExtended {
		// the message should only be dispatched to the target node with the extaddr
		dstnode := d.extaddrMap[pktFrame.DstAddrExtended]
		if dstnode != srcNode && dstnode != nil {
			if d.checkRadioReachable(srcNode, dstnode) {
				d.sendOneRadioFrame(evt, srcNode, dstnode)
			}
			d.Counters.DispatchByExtAddrSucc++
		} else {
			d.Counters.DispatchByExtAddrFail++
		}
		dispatchedByDstAddr = true
	} else if dstAddrMode == wpan.AddrModeShort &&
		pktFrame.DstAddrShort != BroadcastRloc16 {
		// unicast message should only be dispatched to target node(s) with the rloc16
		dstNodes := d.rloc16Map[pktFrame.DstAddrShort]
		dispatchCnt := 0

		if len(dstNodes) > 0 {
			for _, dstNode := range dstNodes {
				if d.checkRadioReachable(srcNode, dstNode) {
					d.sendOneRadioFrame(evt, srcNode, dstNode)
					dispatchCnt++
				}
			}
			d.Counters.DispatchByShortAddrSucc++
		} else {
			d.Counters.DispatchByShortAddrFail++
		}
		dispatchedByDstAddr = true
	}

	// if not dispatched yet, dispatch to all nodes able to receive. Works e.g. for Acks that don't have
	// a destination address.
	if !dispatchedByDstAddr {
		for _, dstNode := range d.nodesArray {
			if d.checkRadioReachable(srcNode, dstNode) {
				d.sendOneRadioFrame(evt, srcNode, dstNode)
			}
		}
		d.Counters.DispatchAllInRange++
	}
}

func (d *Dispatcher) checkRadioReachable(src *Node, dst *Node) bool {
	// the RadioModel will check distance and radio-state of receivers.
	return src != dst && src != nil && dst != nil &&
		d.radioModel.CheckRadioReachable(src.RadioNode, dst.RadioNode)
}

func (d *Dispatcher) sendOneRadioFrame(evt *Event, srcnode *Node, dstnode *Node) {
	logger.AssertTrue(EventTypeRadioCommStart == evt.Type || EventTypeRadioRxDone == evt.Type)
	logger.AssertTrue(srcnode != dstnode)

	// Tx failure cases below:
	//   1) 'failed' state of the dest node
	if dstnode.isFailed {
		return
	}

	//   2) dispatcher's random packet loss Event (applied separate from radio model)
	if d.globalPacketLossRatio > 0 {
		datalenMac := len(evt.Data) - 1
		succRate := math.Pow(1.0-d.globalPacketLossRatio, float64(datalenMac)/MacFrameLenBytes)
		if prng.NewUnitRandom() >= succRate {
			return
		}
	}

	// create new Event copy for individual dispatch to dstNode.
	evt2 := evt.Copy()
	evt2.NodeId = dstnode.Id

	// Tx failure cases below:
	//   3) radio model indicates failure on this specific link (e.g. interference) now.
	// Below lets the radio model process every individual dispatch, to set RSSI, error, etc.
	if d.radioModel.OnEventDispatch(srcnode.RadioNode, dstnode.RadioNode, &evt2) {
		// send the event plus time keeping - moves dstnode's time to the current send-event's time.
		dstnode.sendEvent(&evt2)
	}
}

func (d *Dispatcher) setAlive(nodeid NodeId) {
	logger.AssertFalse(d.isDeleted(nodeid))
	d.aliveNodes[nodeid] = struct{}{}
}

func (d *Dispatcher) IsAlive(nodeid NodeId) bool {
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
	logger.AssertFalse(d.isDeleted(nodeid))
	delete(d.aliveNodes, nodeid)
}

// syncAliveNodes advances the node's time of alive nodes only to current dispatcher time.
func (d *Dispatcher) syncAliveNodes() {
	if len(d.aliveNodes) == 0 || d.isStopping() {
		return
	}

	logger.Warnf("syncing %d alive nodes: %v", len(d.aliveNodes), d.aliveNodes)
	for nodeid := range d.aliveNodes {
		d.advanceNodeTime(d.nodes[nodeid], d.CurTime, true)
	}
}

// syncAllNodes advances all the node's time to current dispatcher time.
func (d *Dispatcher) syncAllNodes() {
	for _, node := range d.nodesArray {
		d.advanceNodeTime(node, d.CurTime, false)
	}
	d.RecvEvents() // blocks until all nodes asleep again.
}

func (d *Dispatcher) pcapFrameWriter() {
	defer d.waitGroup.Done()

	defer func() {
		err := d.pcap.Close()
		if err != nil {
			logger.Errorf("failed to close pcap: %v", err)
		}
	}()

	for item := range d.pcapFrameChan {
		err := d.pcap.AppendFrame(item)
		if err != nil {
			logger.Errorf("write pcap frame failed: %+v", err)
		}
	}
}

func (d *Dispatcher) SetVisualizer(vis visualize.Visualizer) {
	logger.AssertNotNil(vis)
	d.vis = vis
	d.vis.SetSpeed(d.speed)
	d.vis.SetEnergyAnalyser(d.energyAnalyser)
}

func (d *Dispatcher) GetVisualizer() visualize.Visualizer {
	return d.vis
}

func (d *Dispatcher) handleStatusPush(node *Node, data string) {
	node.logger.Tracef("status push: %#v", data)
	statuses := strings.Split(data, ";")
	srcid := node.Id
	oldTopologyCounter := d.Counters.TopologyChanges

	for _, status := range statuses {
		sp := strings.Split(status, "=")
		if len(sp) != 2 {
			continue
		}
		if sp[0] == "transmit" {
			// 'transmit' status is currently not visualized: This is already done by OTNS based on
			// radio frames transmitted.
		} else if sp[0] == "role" {
			role, err := strconv.Atoi(sp[1])
			logger.PanicIfError(err)
			d.setNodeRole(node, OtDeviceRole(role))
			d.Counters.TopologyChanges++
		} else if sp[0] == "rloc16" {
			rloc16, err := strconv.Atoi(sp[1])
			logger.PanicIfError(err)
			d.setNodeRloc16(srcid, uint16(rloc16))
		} else if sp[0] == "ping_request" {
			// e.x. ping_request=fdde:ad00:beef:0:556:90c8:ffaf:b7a3$0$4026600960
			args := strings.Split(sp[1], ",")
			dstaddr := args[0]
			datasize, err := strconv.Atoi(args[1])
			logger.PanicIfError(err)
			timestamp, err := strconv.ParseUint(args[2], 10, 64)
			logger.PanicIfError(err)
			node.onPingRequest(d.convertNodeMilliTime(node, uint32(timestamp)), dstaddr, datasize)
		} else if sp[0] == "ping_reply" {
			//e.x.ping_reply=fdde:ad00:beef:0:556:90c8:ffaf:b7a3$0$0$64
			args := strings.Split(sp[1], ",")
			dstaddr := args[0]
			datasize, err := strconv.Atoi(args[1])
			logger.PanicIfError(err)
			timestamp, err := strconv.ParseUint(args[2], 10, 64)
			logger.PanicIfError(err)
			hoplimit, err := strconv.Atoi(args[3])
			logger.PanicIfError(err)
			node.onPingReply(d.convertNodeMilliTime(node, uint32(timestamp)), dstaddr, datasize, hoplimit)
		} else if sp[0] == "coap" {
			d.handleCoapEvent(node, sp[1])
		} else if sp[0] == "parid" {
			// set partition id
			parid, err := strconv.ParseUint(sp[1], 16, 32)
			logger.PanicIfError(err)
			node.PartitionId = uint32(parid)
			d.vis.SetNodePartitionId(srcid, uint32(parid))
			d.Counters.TopologyChanges++
		} else if sp[0] == "router_added" {
			extaddr, err := strconv.ParseUint(sp[1], 16, 64)
			logger.PanicIfError(err)
			if d.visOptions.RouterTable {
				d.vis.AddRouterTable(srcid, extaddr)
			}
			d.Counters.TopologyChanges++
		} else if sp[0] == "router_removed" {
			extaddr, err := strconv.ParseUint(sp[1], 16, 64)
			logger.PanicIfError(err)
			if d.visOptions.RouterTable {
				d.vis.RemoveRouterTable(srcid, extaddr)
			}
			d.Counters.TopologyChanges++
		} else if sp[0] == "child_added" {
			extaddr, err := strconv.ParseUint(sp[1], 16, 64)
			logger.PanicIfError(err)
			if d.visOptions.ChildTable {
				d.vis.AddChildTable(srcid, extaddr)
			}
			d.Counters.TopologyChanges++
		} else if sp[0] == "child_removed" {
			extaddr, err := strconv.ParseUint(sp[1], 16, 64)
			logger.PanicIfError(err)
			if d.visOptions.ChildTable {
				d.vis.RemoveChildTable(srcid, extaddr)
			}
			d.Counters.TopologyChanges++
		} else if sp[0] == "parent" {
			extaddr, err := strconv.ParseUint(sp[1], 16, 64)
			logger.PanicIfError(err)
			d.vis.SetParent(srcid, extaddr)
			d.Counters.TopologyChanges++
		} else if sp[0] == "joiner_state" {
			joinerState, err := strconv.Atoi(sp[1])
			logger.PanicIfError(err)
			node.onJoinerState(OtJoinerState(joinerState))
		} else if sp[0] == "extaddr" {
			extaddr, err := strconv.ParseUint(sp[1], 16, 64)
			logger.PanicIfError(err)
			node.onStatusPushExtAddr(extaddr)
		} else if sp[0] == "mode" {
			mode := ParseNodeMode(sp[1])
			node.Mode = mode
			d.vis.SetNodeMode(srcid, mode)
			d.Counters.TopologyChanges++
		} else {
			logger.Errorf("received unknown status push: %s=%s", sp[0], sp[1])
		}
	}

	// check if a re-evaluation of node topology statistics is needed.
	if d.Counters.TopologyChanges > oldTopologyCounter {
		d.updateNodeStats()
	}
}

func (d *Dispatcher) AddNode(nodeid NodeId, cfg *NodeConfig) *Node {
	logger.AssertNil(d.nodes[nodeid])
	logger.Debugf("dispatcher AddNode id=%d", nodeid)
	delete(d.deletedNodes, nodeid)

	node := newNode(d, nodeid, cfg)
	d.nodes[nodeid] = node
	d.reconstructNodesArray()
	d.Counters.TopologyChanges++
	d.alarmMgr.AddNode(nodeid)
	d.energyAnalyser.AddNode(nodeid, d.CurTime)
	d.vis.AddNode(nodeid, cfg)
	d.radioModel.AddNode(node.RadioNode)
	d.setAlive(nodeid)
	d.updateNodeStats()

	if d.cfg.DefaultWatchOn {
		lev, err := logger.ParseLevelString(d.cfg.DefaultWatchLevel)
		if err == nil {
			d.WatchNode(nodeid, lev)
		} else {
			logger.Error(err)
		}
	}
	return node
}

func (d *Dispatcher) setNodeRloc16(srcid NodeId, rloc16 uint16) {
	node := d.nodes[srcid]
	logger.AssertNotNil(node)

	node.logger.Debugf("set node RLOC16: %x -> %x", node.Rloc16, rloc16)
	oldRloc16 := node.Rloc16
	if oldRloc16 != InvalidRloc16 {
		// remove node from old rloc map
		d.rloc16Map.Remove(oldRloc16, node)
	}

	node.Rloc16 = rloc16
	if rloc16 != InvalidRloc16 {
		// add node to the new rloc map
		d.rloc16Map.Add(rloc16, node)
	}

	d.vis.SetNodeRloc16(srcid, rloc16)
}

func (d *Dispatcher) visSendFrame(srcid NodeId, dstid NodeId, pktframe *wpan.MacFrame, commData RadioCommEventData) {
	d.visSend(srcid, dstid, &visualize.MsgVisualizeInfo{
		Channel:         pktframe.Channel,
		FrameControl:    pktframe.FrameControl,
		Seq:             pktframe.Seq,
		DstAddrShort:    pktframe.DstAddrShort,
		DstAddrExtended: pktframe.DstAddrExtended,
		SendDurationUs:  uint32(commData.Duration),
		PowerDbm:        commData.PowerDbm,
		FrameSizeBytes:  pktframe.LengthBytes + pktframe.PhyHdrLength, // Note: count total frame len (PHY+MAC)
	})
}

func (d *Dispatcher) visSendInterference(srcid NodeId, dstid NodeId, commData RadioCommEventData) {
	d.visSend(srcid, dstid, &visualize.MsgVisualizeInfo{
		Channel:         commData.Channel,
		FrameControl:    0x04,
		Seq:             0,
		DstAddrShort:    0,
		DstAddrExtended: 0,
		SendDurationUs:  uint32(commData.Duration),
		PowerDbm:        commData.PowerDbm,
		FrameSizeBytes:  0,
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
	logger.AssertTrue(d.CurTime <= ts, "%v > %v", d.CurTime, ts)
	d.CurTime = ts

	elapsedTime := int64(d.CurTime - d.speedStartTime)
	elapsedRealTime := time.Since(d.speedStartRealTime) / time.Microsecond
	if elapsedRealTime > 0 {
		d.vis.AdvanceTime(ts, float64(elapsedTime)/float64(elapsedRealTime))
	} else {
		d.vis.AdvanceTime(ts, MaxSimulateSpeed)
	}

	if d.energyAnalyser != nil && (ts >= d.lastEnergyVizTime+energy.ComputePeriod || (d.lastEnergyVizTime == 0 && ts > 0)) {
		d.energyAnalyser.StoreNetworkEnergy(ts)
		d.vis.UpdateNodesEnergy(d.energyAnalyser.GetLatestEnergyOfNodes(), ts, true)
		d.lastEnergyVizTime = ts
	}

	if d.cfg.PhyTxStats {
		d.updateTimeWindowStats()
	}
}

func (d *Dispatcher) PostAsync(task func()) {
	d.taskChan <- task
}

func (d *Dispatcher) handleTasks() {
	defer func() {
		err := recover()
		if err != nil {
			logger.TraceError("dispatcher handle task failed: %+v", err)
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

func (d *Dispatcher) WatchNode(nodeid NodeId, watchLevel logger.Level) {
	d.watchingNodes[nodeid] = struct{}{}
	node := d.nodes[nodeid]
	if node != nil {
		node.logger.SetDisplayLevel(watchLevel)
	}
}

func (d *Dispatcher) UnwatchNode(nodeid NodeId) {
	node := d.nodes[nodeid]
	if node != nil {
		node.logger.SetDisplayLevel(logger.ErrorLevel)
	}
	delete(d.watchingNodes, nodeid)
}

func (d *Dispatcher) GetWatchingNodes() []NodeId {
	watchingNodeIds := make([]NodeId, len(d.watchingNodes))
	j := 0
	for k := range d.watchingNodes {
		watchingNodeIds[j] = k
		j++
	}
	sort.Ints(watchingNodeIds)
	return watchingNodeIds
}

func (d *Dispatcher) GetNode(id NodeId) *Node {
	return d.nodes[id]
}

func (d *Dispatcher) GetFailedCount() int {
	failCount := 0
	for _, dn := range d.nodesArray {
		if dn.IsFailed() {
			failCount += 1
		}
	}
	return failCount
}

func (d *Dispatcher) SetNodePos(id NodeId, x, y, z int) {
	node := d.nodes[id]
	logger.AssertNotNil(node)

	node.X, node.Y, node.Z = x, y, z
	node.RadioNode.SetNodePos(x, y, z)
	d.vis.SetNodePos(id, x, y, z)
}

func (d *Dispatcher) DeleteNode(id NodeId) {
	node := d.nodes[id]
	logger.AssertNotNil(node)

	delete(d.nodes, id)
	d.reconstructNodesArray()
	d.Counters.TopologyChanges++
	delete(d.aliveNodes, id)
	delete(d.watchingNodes, id)
	if node.Rloc16 != InvalidRloc16 {
		d.rloc16Map.Remove(node.Rloc16, node)
	}
	if node.ExtAddr != InvalidExtAddr {
		logger.AssertTrue(d.extaddrMap[node.ExtAddr] == node)
		delete(d.extaddrMap, node.ExtAddr)
	}
	d.alarmMgr.DeleteNode(id)
	d.deletedNodes[id] = struct{}{}
	d.energyAnalyser.DeleteNode(id)
	d.vis.DeleteNode(id)
	d.radioModel.DeleteNode(id)
	d.eventQueue.DisableEventsForNode(id)
	d.updateNodeStats()
}

// SetNodeFailed sets the radio of the node to failed (true) or operational (false) state.
// Setting this will disable the automatic failure control (FailureCtrl).
func (d *Dispatcher) SetNodeFailed(id NodeId, fail bool) {
	node := d.nodes[id]
	logger.AssertNotNil(node)

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
		logger.AssertTrue(d.extaddrMap[oldExtAddr] == nil)
	} else {
		logger.AssertTrue(d.extaddrMap[oldExtAddr] == node)
		delete(d.extaddrMap, oldExtAddr)
	}
	logger.AssertNil(d.extaddrMap[node.ExtAddr])

	d.extaddrMap[node.ExtAddr] = node
	d.vis.OnExtAddrChange(node.Id, node.ExtAddr)
}

func (d *Dispatcher) GetVisualizationOptions() VisualizationOptions {
	return d.visOptions
}

func (d *Dispatcher) SetVisualizationOptions(opts VisualizationOptions) {
	logger.Debugf("dispatcher set visualization options: %+v", opts)
	d.visOptions = opts
}

func (d *Dispatcher) NotifyCommand(nodeid NodeId) {
	d.setAlive(nodeid)
}

// NotifyNodeFailure is called by other goroutines to notify the dispatcher that a node process has
// failed. From failed nodes, we don't expect further messages and they can't be alive.
func (d *Dispatcher) NotifyNodeProcessFailure(nodeid NodeId) {
	d.eventChan <- &Event{
		Delay:  0,
		Type:   EventTypeNodeDisconnected,
		NodeId: nodeid,
		Conn:   nil,
	}
}

func (d *Dispatcher) dumpPacket(item *Event) {
	sb := strings.Builder{}
	_, _ = fmt.Fprintf(&sb, "DUMP:PACKET:%d:%d:", item.Timestamp, item.NodeId)
	for _, b := range item.Data {
		_, _ = fmt.Fprintf(&sb, "%02X", b)
	}

	logger.Println(sb.String())
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
	logger.AssertTrue(len(args) > 0)
	action := args[0]

	if action == "send" || action == "recv" || action == "send_error" {
		var messageId, coapType, coapCode, port int

		logger.AssertTrue(len(args) >= 7)

		messageId, err = strconv.Atoi(args[1])
		logger.PanicIfError(err)

		coapType, err = strconv.Atoi(args[2])
		logger.PanicIfError(err)

		coapCode, err = strconv.Atoi(args[3])
		logger.PanicIfError(err)

		uri := args[4]

		ip := args[5]

		port, err = strconv.Atoi(args[6])
		logger.PanicIfError(err)

		if action == "send" {
			d.coaps.OnSend(d.CurTime, node.Id, messageId, CoapType(coapType), CoapCode(coapCode), uri, ip, port)
		} else if action == "recv" {
			d.coaps.OnRecv(d.CurTime, node.Id, messageId, CoapType(coapType), CoapCode(coapCode), uri, ip, port)
		} else {
			logger.AssertTrue(len(args) >= 7)
			threadError := args[6]

			d.coaps.OnSendError(node.Id, messageId, CoapType(coapType), CoapCode(coapCode), uri, ip, port, threadError)
		}
	} else {
		logger.Warnf("unknown coap event: %+v", args)
	}
}

// EnableCoaps enables CoAP message tracking (if already enabled, it does nothing)
func (d *Dispatcher) EnableCoaps() {
	if d.coaps == nil {
		d.coaps = newCoapsHandler()
	}
}

func (d *Dispatcher) CollectCoapMessages(clearCollectedMessages bool) []*CoapMessage {
	if d.coaps != nil {
		return d.coaps.DumpMessages(clearCollectedMessages)
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
	if d.radioModel != model && d.radioModel != nil {
		// when setting a new model, transfer all nodes into it.
		for _, node := range d.nodesArray {
			d.radioModel.DeleteNode(node.Id)
			model.AddNode(node.RadioNode)
		}
	}
	d.radioModel = model
}

func (d *Dispatcher) handleRadioState(node *Node, evt *Event) {
	logger.AssertNotNil(node)
	subState := evt.RadioStateData.SubState
	state := evt.RadioStateData.State
	energyState := evt.RadioStateData.EnergyState

	const hdr = "(OTNS)       [T] RadioState----:"
	node.logger.Tracef("%s EnergyState=%s SubState=%s RadioState=%s RadioTime=%d NextStTime=+%d",
		hdr, energyState, subState, state, evt.RadioStateData.RadioTime, evt.Delay)

	if d.energyAnalyser != nil {
		radioEnergy := d.energyAnalyser.GetNode(node.Id)
		logger.AssertNotNil(radioEnergy)
		radioEnergy.SetRadioState(energyState, d.CurTime)
	}

	// if a next radio-state transition time is indicated, make sure to schedule node wake-up for that time.
	// This is independent from any alarm-time set by the node which is the OT's stack next-operation time.
	if evt.Delay > 0 {
		d.eventQueue.Add(&Event{
			Type:      EventTypeAlarmFired,
			NodeId:    node.Id,
			Timestamp: d.CurTime + evt.Delay,
		})
	}
}
