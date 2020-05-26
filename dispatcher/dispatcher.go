// Copyright (c) 2020, The OTNS Authors.
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
	"math/rand"
	"strconv"
	"strings"
	"sync"

	"github.com/openthread/ot-ns/progctx"

	"github.com/openthread/ot-ns/dissectpkt"
	"github.com/openthread/ot-ns/dissectpkt/wpan"
	"github.com/openthread/ot-ns/pcap"
	"github.com/openthread/ot-ns/threadconst"
	"github.com/openthread/ot-ns/visualize"
	"github.com/simonlingoogle/go-simplelogger"

	"math"
	"net"
	"time"

	. "github.com/openthread/ot-ns/types"
)

const (
	Ever uint64 = math.MaxUint64 / 2
)

const (
	ProcessEventTimeErrorUs = 0
	MaxSimulateSpeed        = 1000000
)

type pcapFrameItem struct {
	Ustime uint64
	Data   []byte
}

type Config struct {
	Speed float64
}

func DefaultConfig() *Config {
	return &Config{
		Speed: 1,
	}
}

type CallbackHandler interface {
	OnNodeFail(nodeid NodeId)
	OnNodeRecover(nodeid NodeId)
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
	eventChan             chan *event
	waitGroup             sync.WaitGroup
	CurTime               uint64
	pauseTime             uint64
	alarmMgr              *alarmMgr
	sendQueue             *sendQueue
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

	Counters struct {
		// Event counters
		AlarmEvents      uint64
		RadioEvents      uint64
		StatusPushEvents uint64
		// Packet dispatching counters
		DispatchByExtAddrSucc   uint64
		DispatchByExtAddrFail   uint64
		DispatchByShortAddrSucc uint64
		DispatchByShortAddrFail uint64
		DispatchAllInRange      uint64
	}
	watchingNodes map[NodeId]struct{}
	stopped       bool
}

func NewDispatcher(ctx *progctx.ProgCtx, cfg *Config, cbHandler CallbackHandler) *Dispatcher {
	udpAddr, err := net.ResolveUDPAddr("udp4", "127.0.0.1:9000")
	simplelogger.FatalIfError(err, err)
	ln, err := net.ListenUDP("udp", udpAddr)
	simplelogger.FatalIfError(err, err)
	_ = ln.SetWriteBuffer(25 * 1024 * 1024)
	_ = ln.SetReadBuffer(25 * 1024 * 1024)
	simplelogger.Infof("listening on udp port 9000")

	simplelogger.AssertNil(err)

	vis := visualize.NewNopVisualizer()

	d := &Dispatcher{
		ctx:                ctx,
		cfg:                *cfg,
		cbHandler:          cbHandler,
		udpln:              ln,
		eventChan:          make(chan *event, 10000),
		alarmMgr:           newAlarmMgr(),
		sendQueue:          newSendQueue(),
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
	}
	d.speed = d.normalizeSpeed(d.speed)
	d.pcap, err = pcap.NewFile("current.pcap")
	simplelogger.PanicIfError(err)

	go d.eventsReader()
	go d.pcapFrameWriter()

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
	defer simplelogger.Infof("dispatcher exit.")
	d.ctx.Defer(func() {
		_ = d.udpln.Close()
	})

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
			_ = d.pcap.Sync()
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

		d.recvEvents()
		d.syncAliveNodes()

		// process the next event
		goon := d.processNextEvent()
		simplelogger.AssertTrue(d.CurTime <= d.pauseTime)

		if !goon && len(d.aliveNodes) == 0 {
			d.advanceTime(d.pauseTime)
		}
	}
}

func (d *Dispatcher) handleRecvEvent(evt *event) {
	// create new node if necessary
	nodeid := evt.NodeId
	if _, ok := d.nodes[nodeid]; !ok {
		if _, deleted := d.deletedNodes[nodeid]; !deleted {
			// TODO: node should push extaddr to dispatcher
			// can not find the node, and the node is not registered (created by OTNS)
			d.newNode(nodeid, 0, -1, -1, 10000, DefaultNodeMode())
		} else {
			// the node is already deleted, ignore this message
			return
		}
	}

	if d.isWatching(evt.NodeId) {
		simplelogger.Warnf("Node %d <<< %+v, cur time %d, node time %d, delay %d", evt.NodeId, *evt,
			d.CurTime, int64(d.nodes[nodeid].CurTime)-int64(d.CurTime), evt.Delay)
	}

	delay := evt.Delay
	var evtTime uint64
	if delay >= 2147483647 {
		evtTime = Ever
	} else {
		evtTime = d.CurTime + evt.Delay
	}

	switch evt.Type {
	case eventTypeAlarmFired:
		d.Counters.AlarmEvents += 1
		d.setSleeping(nodeid)
		d.alarmMgr.SetTimestamp(nodeid, evtTime)
	case eventTypeRadioReceived:
		simplelogger.AssertTrue(evt.Delay == 1)
		d.Counters.RadioEvents += 1
		d.sendQueue.Add(evtTime, nodeid, evt.Data)
	case eventTypeStatusPush:
		d.Counters.StatusPushEvents += 1
		d.handleStatusPush(evt.NodeId, string(evt.Data))
	default:
		simplelogger.Panicf("event type not implemented: %v", evt.Type)
	}
}

func (d *Dispatcher) recvEvents() int {
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

func (d *Dispatcher) processNextEvent() bool {
	simplelogger.AssertTrue(d.CurTime <= d.pauseTime)

	// we need to wait until all nodes are sleep
	nextAlarmTime := d.alarmMgr.NextTimestamp()
	nextSendtime := d.sendQueue.NextTimestamp()

	nextEventTime := nextAlarmTime
	if nextEventTime > nextSendtime {
		nextEventTime = nextSendtime
	}

	// nextEventTime <= d.pauseTime
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
			return true
		}
	}

	if nextEventTime > d.pauseTime {
		return false
	}

	simplelogger.AssertTrue(nextAlarmTime >= d.CurTime && nextSendtime >= d.CurTime)
	var procUntilTime uint64
	if nextAlarmTime <= nextSendtime {
		procUntilTime = nextAlarmTime + ProcessEventTimeErrorUs
	} else {
		procUntilTime = nextSendtime + ProcessEventTimeErrorUs
	}

	if procUntilTime > d.pauseTime {
		procUntilTime = d.pauseTime
	}

	for {
		if nextAlarmTime > procUntilTime && nextSendtime > procUntilTime {
			break
		}

		if nextAlarmTime <= nextSendtime {
			// process next alarm
			d.advanceTime(nextAlarmTime)
			nextAlarm := d.alarmMgr.NextAlarm()
			simplelogger.AssertNotNil(nextAlarm)
			d.advanceNodeTime(nextAlarm.NodeId, nextAlarm.Timestamp, false)
			// mark the node as alive in the alarm
		} else {
			// process the send event
			s := d.sendQueue.PopNext()
			simplelogger.AssertTrue(s.Timestamp == nextSendtime)
			d.advanceTime(nextSendtime)
			// construct the message
			d.pcapFrameChan <- pcapFrameItem{nextSendtime, s.Data[1:]}
			d.sendNodeMessage(s)
		}

		nextAlarmTime = d.alarmMgr.NextTimestamp()
		nextSendtime = d.sendQueue.NextTimestamp()
	}

	return len(d.nodes) > 0
}

func (d *Dispatcher) eventsReader() {
	udpln := d.udpln
	readbuf := make([]byte, 4096)

	for {
		// wait until all nodes are sleepd
		n, srcaddr, err := udpln.ReadFromUDP(readbuf)
		if err != nil {
			if nerr, ok := err.(net.Error); ok && nerr.Temporary() {
				time.Sleep(time.Millisecond * 100)
				continue
			}

			simplelogger.Infof("UDP events reader quit.")
			break
		}

		if n < 11 {
			simplelogger.Panicf("message length too sort: %d", n)
		}

		delay := binary.LittleEndian.Uint64(readbuf[:8])
		typ := readbuf[8]
		datalen := binary.LittleEndian.Uint16(readbuf[9:11])
		nodeid := srcaddr.Port - 9000

		data := make([]byte, n-11)
		copy(data, readbuf[11:n])
		evt := &event{
			NodeId:  nodeid,
			Delay:   delay,
			Type:    typ,
			DataLen: datalen,
			Data:    data,
		}

		d.eventChan <- evt
	}
}

func (d *Dispatcher) advanceNodeTime(id NodeId, timestamp uint64, force bool) {
	node := d.nodes[id]

	oldTime := node.CurTime
	elapsed := timestamp - oldTime
	if timestamp <= oldTime {
		// node time was already newer than the timestamp
		if !force {
			return
		} else {
			elapsed = 0
		}
	}

	msg := make([]byte, 11)
	binary.LittleEndian.PutUint64(msg[:8], elapsed)
	msg[8] = eventTypeAlarmFired
	binary.LittleEndian.PutUint16(msg[9:11], 0)
	node.SendMessage(msg)
	node.CurTime = timestamp
	if timestamp > oldTime {
		node.failureCtrl.OnTimeAdvanced(oldTime)
	}

	d.alarmMgr.SetNotified(id)
	d.setAlive(id)
	if d.isWatching(id) {
		simplelogger.Warnf("Node %d >>> advance time %v -> %v", id, oldTime, timestamp)
	}
}

func (d *Dispatcher) sendNodeMessage(sit *sendItem) {
	// send the message to all nodes
	srcnodeid := sit.NodeId
	srcnode := d.nodes[srcnodeid]
	if srcnode == nil {
		if _, ok := d.deletedNodes[srcnodeid]; !ok {
			simplelogger.Errorf("%s: node %d not found", d, srcnodeid)
		}
		return
	}

	// send to self as notify for tx done (should do even if the node is failed)
	d.sendOneMessage(sit, srcnode, srcnode)

	if srcnode.isFailed {
		return
	}

	pktinfo := dissectpkt.Dissect(sit.Data)
	pktframe := pktinfo.MacFrame

	// try to dispatch the message by extaddr directly
	dispatchedByDstAddr := false
	dstAddrMode := pktframe.FrameControl.DstAddrMode()

	if dstAddrMode == wpan.DstAddrModeExtended {
		// the message should only be dispatched to the target node with the extaddr
		dstnode := d.extaddrMap[pktframe.DstAddrExtended]
		if dstnode != srcnode && dstnode != nil {
			if d.checkRadioReachable(srcnode, dstnode) {
				d.sendOneMessage(sit, srcnode, dstnode)
				d.visSend(srcnodeid, dstnode.Id, pktframe)
			} else {
				d.visSend(srcnodeid, InvalidNodeId, pktframe)
			}

			d.Counters.DispatchByExtAddrSucc++
		} else {
			d.Counters.DispatchByExtAddrFail++
			d.visSend(srcnodeid, InvalidNodeId, pktframe)
		}

		dispatchedByDstAddr = true
	} else if dstAddrMode == wpan.DstAddrModeShort {
		if pktframe.DstAddrShort != threadconst.BroadcastRloc16 {
			// unicast message should only be dispatched to target node with the rloc16
			dstnodes := d.rloc16Map[pktframe.DstAddrShort]
			dispatchCnt := 0

			if len(dstnodes) > 0 {
				for _, dstnode := range dstnodes {
					if d.checkRadioReachable(srcnode, dstnode) {
						d.sendOneMessage(sit, srcnode, dstnode)
						d.visSend(srcnodeid, dstnode.Id, pktframe)
						dispatchCnt++
					}
				}
				d.Counters.DispatchByShortAddrSucc++
			} else {
				d.Counters.DispatchByShortAddrFail++
			}

			if dispatchCnt == 0 {
				d.visSend(srcnodeid, InvalidNodeId, pktframe)
			}

			dispatchedByDstAddr = true
		}
	}

	if !dispatchedByDstAddr {
		for _, dstnode := range d.nodes {
			if d.checkRadioReachable(srcnode, dstnode) {
				d.sendOneMessage(sit, srcnode, dstnode)
			}
		}

		d.visSend(srcnodeid, BroadcastNodeId, pktframe)
	}
}

func (d *Dispatcher) checkRadioReachable(src *Node, dst *Node) bool {
	return dst != src && src.GetDistanceTo(dst) <= src.radioRange
}

func (d *Dispatcher) sendOneMessage(sit *sendItem, srcnode *Node, dstnode *Node) {
	if srcnode != dstnode {
		// we should always send the message when srcnode == dstnode, because it is the TX done notify
		if dstnode.isFailed {
			return
		}

		if d.globalPacketLossRatio > 0 {
			datalen := len(sit.Data)
			succRate := math.Pow(1.0-d.globalPacketLossRatio, float64(datalen)/128.0)
			if rand.Float64() >= succRate {
				return
			}
		}
	}

	timestamp := sit.Timestamp
	var elapsed uint64

	oldTime := dstnode.CurTime
	if timestamp > oldTime {
		elapsed = timestamp - oldTime
	} else {
		elapsed = 0
	}

	dstnode.Send(elapsed, sit.Data)
	dstnode.CurTime = timestamp
	if timestamp > oldTime {
		dstnode.failureCtrl.OnTimeAdvanced(oldTime)
	}

	dstnodeid := dstnode.Id
	d.alarmMgr.SetNotified(dstnodeid)
	d.setAlive(dstnodeid)

	if d.isWatching(dstnodeid) {
		if dstnode == srcnode {
			simplelogger.Warnf("Node %d >>> TX DONE", dstnodeid)
		} else {
			simplelogger.Warnf("Node %d >>> received message from node %d", dstnodeid, srcnode.Id)
		}
	}
}

func (d *Dispatcher) newNode(nodeid NodeId, extaddr uint64, x, y int, radioRange int, mode NodeMode) (node *Node) {
	node = newNode(d, nodeid, extaddr, x, y, radioRange)
	d.nodes[nodeid] = node
	simplelogger.AssertNil(d.extaddrMap[extaddr])
	d.extaddrMap[extaddr] = node
	d.alarmMgr.AddNode(nodeid)
	d.setAlive(nodeid)

	d.vis.AddNode(nodeid, extaddr, x, y, radioRange, mode)
	return
}

func (d *Dispatcher) setAlive(nodeid NodeId) {
	d.aliveNodes[nodeid] = struct{}{}
}

func (d *Dispatcher) setSleeping(nodeid NodeId) {
	delete(d.aliveNodes, nodeid)
}

func (d *Dispatcher) syncAliveNodes() {
	if len(d.aliveNodes) == 0 {
		return
	}

	simplelogger.Warnf("syncing %d alive nodes: %v", len(d.aliveNodes), d.aliveNodes)
	for nodeid := range d.aliveNodes {
		d.advanceNodeTime(nodeid, d.CurTime, true)
	}
}

func (d *Dispatcher) syncAllNodes() {
	for nodeid := range d.nodes {
		d.advanceNodeTime(nodeid, d.CurTime, false)
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
}

func (d *Dispatcher) GetVisualizer() visualize.Visualizer {
	return d.vis
}

func (d *Dispatcher) handleStatusPush(srcid NodeId, data string) {
	simplelogger.Debugf("status push: %d: %#v", srcid, data)
	srcnode := d.nodes[srcid]
	if srcnode == nil {
		simplelogger.Warnf("node not found: %d", srcid)
		return
	}

	statuses := strings.Split(data, ";")
	for _, status := range statuses {
		sp := strings.Split(status, "=")
		if len(sp) != 2 {
			continue
		}
		if sp[0] == "role" {
			role, err := strconv.Atoi(sp[1])
			simplelogger.PanicIfError(err)
			d.vis.SetNodeRole(srcid, visualize.OtDeviceRole(role))
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
		} else if sp[0] == "parid" {
			// set partition id
			parid, err := strconv.ParseUint(sp[1], 16, 32)
			simplelogger.PanicIfError(err)
			srcnode.PartitionId = uint32(parid)
			d.vis.SetNodePartitionId(srcid, uint32(parid))
		} else if sp[0] == "router_added" {
			extaddr, err := strconv.ParseUint(sp[1], 16, 64)
			simplelogger.PanicIfError(err)
			d.vis.AddRouterTable(srcid, extaddr)
		} else if sp[0] == "router_removed" {
			extaddr, err := strconv.ParseUint(sp[1], 16, 64)
			simplelogger.PanicIfError(err)
			d.vis.RemoveRouterTable(srcid, extaddr)
		} else if sp[0] == "child_added" {
			extaddr, err := strconv.ParseUint(sp[1], 16, 64)
			simplelogger.PanicIfError(err)
			d.vis.AddChildTable(srcid, extaddr)
		} else if sp[0] == "child_removed" {
			extaddr, err := strconv.ParseUint(sp[1], 16, 64)
			simplelogger.PanicIfError(err)
			d.vis.RemoveChildTable(srcid, extaddr)
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
		} else {
			simplelogger.Warnf("unknown status push: %s=%s", sp[0], sp[1])
		}
	}
}

func (d *Dispatcher) AddNode(nodeid NodeId, extaddr uint64, x, y int, radioRange int, mode NodeMode) {
	simplelogger.AssertNil(d.nodes[nodeid])
	simplelogger.Infof("dispatcher add node %d", nodeid)
	d.newNode(nodeid, extaddr, x, y, radioRange, mode)
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

func (d *Dispatcher) visSend(srcid NodeId, dstid NodeId, pktframe *wpan.MacFrame) {
	d.vis.Send(srcid, dstid, &visualize.MsgVisualizeInfo{
		Channel:         pktframe.Channel,
		FrameControl:    pktframe.FrameControl,
		Seq:             pktframe.Seq,
		DstAddrShort:    pktframe.DstAddrShort,
		DstAddrExtended: pktframe.DstAddrExtended,
	})
}

func (d *Dispatcher) advanceTime(ts uint64) {
	simplelogger.AssertTrue(d.CurTime <= ts, "%v > %v", d.CurTime, ts)
	if d.CurTime < ts {
		oldTime := d.CurTime
		d.CurTime = ts
		elapsedTime := int64(d.CurTime - d.speedStartTime)
		elapsedRealTime := time.Since(d.speedStartRealTime) / time.Microsecond
		if elapsedRealTime > 0 && ts/1000000 != oldTime/1000000 {
			d.vis.AdvanceTime(ts, float64(elapsedTime)/float64(elapsedRealTime))
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
	simplelogger.AssertTrue(d.extaddrMap[node.ExtAddr] == node)
	delete(d.extaddrMap, node.ExtAddr)
	d.alarmMgr.DeleteNode(id)
	d.deletedNodes[id] = struct{}{}

	d.vis.DeleteNode(id)
}

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
	simplelogger.AssertTrue(d.extaddrMap[oldExtAddr] == node)
	simplelogger.AssertNil(d.extaddrMap[node.ExtAddr])

	delete(d.extaddrMap, oldExtAddr)
	d.extaddrMap[node.ExtAddr] = node
	d.vis.OnExtAddrChange(node.Id, node.ExtAddr)
}
