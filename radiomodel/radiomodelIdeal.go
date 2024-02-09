// Copyright (c) 2022-2024, The OTNS Authors.
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

package radiomodel

import (
	. "github.com/openthread/ot-ns/event"
	"github.com/openthread/ot-ns/logger"
	. "github.com/openthread/ot-ns/types"
)

// RadioModelIdeal is an ideal radio model with infinite parallel transmission capacity per
// channel. RSSI at the receiver can be set to an ideal constant RSSI value, or to a value
// based on an average RF propagation model. There is a hard stop of reception beyond the
// radioRange of the node i.e. ideal disc model.
type RadioModelIdeal struct {
	name         string
	params       *RadioModelParams
	nodes        map[NodeId]*RadioNode
	eventQ       EventQueue
	channelStats map[ChannelId]*ChannelStats
}

func (rm *RadioModelIdeal) AddNode(radioNode *RadioNode) {
	rm.nodes[radioNode.Id] = radioNode
}

func (rm *RadioModelIdeal) DeleteNode(nodeid NodeId) {
	delete(rm.nodes, nodeid)
}

func (rm *RadioModelIdeal) CheckRadioReachable(src *RadioNode, dst *RadioNode) bool {
	if src != dst && dst.RadioState == RadioRx && src.RadioChannel == dst.RadioChannel {
		dist := src.GetDistanceTo(dst)
		if dist <= src.RadioRange { // simple disc radio model
			return true
		}
	}
	return false
}

func (rm *RadioModelIdeal) GetTxRssi(srcNode *RadioNode, dstNode *RadioNode) DbValue {
	var rssi DbValue
	if rm.params.RssiMinDbm < rm.params.RssiMaxDbm {
		rssi = computeIndoorRssiItu(srcNode.GetDistanceTo(dstNode), srcNode.TxPower, rm.params)
		if rssi < rm.params.RssiMinDbm {
			rssi = rm.params.RssiMinDbm
		} else if rssi > rm.params.RssiMaxDbm {
			rssi = rm.params.RssiMaxDbm
		}
	} else {
		rssi = rm.params.RssiMaxDbm
	}
	return rssi
}

func (rm *RadioModelIdeal) OnEventDispatch(src *RadioNode, dst *RadioNode, evt *Event) bool {
	switch evt.Type {
	case EventTypeRadioCommStart:
		fallthrough
	case EventTypeRadioRxDone:
		// compute the RSSI and store it in the event
		evt.RadioCommData.PowerDbm = clipRssi(rm.GetTxRssi(src, dst))
	case EventTypeRadioChannelSample:
		// store the final sampled RSSI in the event
		evt.RadioCommData.PowerDbm = clipRssi(src.rssiSampleMax)
	}
	return true
}

func (rm *RadioModelIdeal) OnNextEventTime(ts uint64) {
	//
}

func (rm *RadioModelIdeal) OnParametersModified() {
	//
}

func (rm *RadioModelIdeal) HandleEvent(node *RadioNode, q EventQueue, evt *Event) {
	rm.eventQ = q

	switch evt.Type {
	case EventTypeRadioCommStart:
		rm.txStart(node, evt)
		rm.statsTxStart(node, evt)
	case EventTypeRadioTxDone:
		rm.txStop(node, evt)
		rm.statsTxStop(node, evt)
	case EventTypeRadioChannelSample:
		rm.channelSampleStart(node, evt)
	case EventTypeRadioState:
		node.SetRadioState(evt.RadioStateData.EnergyState, evt.RadioStateData.SubState)
		node.SetChannel(evt.RadioStateData.Channel)
		node.SetRxSensitivity(DbValue(evt.RadioStateData.RxSensDbm))
	default:
		break
	}
}

func (rm *RadioModelIdeal) GetName() string {
	return rm.name
}

func (rm *RadioModelIdeal) GetParameters() *RadioModelParams {
	return rm.params
}

func (rm *RadioModelIdeal) GetChannelStats(channel ChannelId, curTimeUs uint64) *ChannelStats {
	if chanStats, ok := rm.channelStats[channel]; ok {
		// check if an operation is ongoing - if so, include portion of to-be-added Tx duration in stats.
		if chanStats.numTransmitters > 0 && curTimeUs > chanStats.txStartTime {
			chanStats.TxTimeUs += curTimeUs - chanStats.txStartTime
			chanStats.txStartTime = curTimeUs
		}
		return chanStats
	}
	return nil // no channel stats for this channel
}

func (rm *RadioModelIdeal) ResetChannelStats(channel ChannelId) {
	delete(rm.channelStats, channel)
}

func (rm *RadioModelIdeal) init() {
	rm.nodes = map[NodeId]*RadioNode{}
	rm.channelStats = make(map[ChannelId]*ChannelStats)
}

func (rm *RadioModelIdeal) txStart(srcNode *RadioNode, evt *Event) {
	srcNode.TxPower = DbValue(evt.RadioCommData.PowerDbm) // get last node's properties from the OT node's event params.
	srcNode.SetChannel(evt.RadioCommData.Channel)

	// dispatch radio event RadioComm 'start of frame Rx' to listening nodes.
	rxStartEvt := evt.Copy()
	rxStartEvt.Type = EventTypeRadioCommStart
	rxStartEvt.RadioCommData.Error = OT_ERROR_NONE
	rxStartEvt.MustDispatch = true
	rm.eventQ.Add(&rxStartEvt)

	// schedule new internal event to call txStop() at end of duration.
	txDoneEvt := evt.Copy()
	txDoneEvt.Type = EventTypeRadioTxDone
	txDoneEvt.RadioCommData.Error = OT_ERROR_NONE
	txDoneEvt.MustDispatch = false
	txDoneEvt.Timestamp += evt.RadioCommData.Duration
	rm.eventQ.Add(&txDoneEvt)
}

func (rm *RadioModelIdeal) txStop(node *RadioNode, evt *Event) {
	// Dispatch TxDone event back to the source
	txDoneEvt := evt.Copy()
	txDoneEvt.Type = EventTypeRadioTxDone
	txDoneEvt.RadioCommData.Error = OT_ERROR_NONE
	txDoneEvt.MustDispatch = true
	rm.eventQ.Add(&txDoneEvt)

	// Create RxDone event, to signal nearby node(s) the frame Rx is done.
	rxDoneEvt := evt.Copy()
	rxDoneEvt.Type = EventTypeRadioRxDone
	rxDoneEvt.MustDispatch = true
	rm.eventQ.Add(&rxDoneEvt)
}

func (rm *RadioModelIdeal) statsTxStart(node *RadioNode, evt *Event) {
	ch := evt.RadioCommData.Channel
	chStats, ok := rm.channelStats[ch]
	if !ok {
		chStats = &ChannelStats{
			Channel:         ch,
			TxTimeUs:        0,
			NumFrames:       0,
			numTransmitters: 0,
		}
		rm.channelStats[ch] = chStats
	}
	if chStats.numTransmitters == 0 {
		chStats.txStartTime = evt.Timestamp
	}
	chStats.numTransmitters++
	chStats.NumFrames++
}

func (rm *RadioModelIdeal) statsTxStop(node *RadioNode, evt *Event) {
	ch := evt.RadioCommData.Channel
	chStats, ok := rm.channelStats[ch]
	logger.AssertTrue(ok)
	logger.AssertTrue(chStats.numTransmitters > 0)

	chStats.numTransmitters--
	if chStats.numTransmitters == 0 {
		txDur := evt.Timestamp - chStats.txStartTime
		chStats.TxTimeUs += txDur
	}
}

func (rm *RadioModelIdeal) channelSampleStart(node *RadioNode, evt *Event) {
	node.rssiSampleMax = RssiMinusInfinity // Ideal model never has CCA failure.
	node.SetChannel(evt.RadioCommData.Channel)

	// dispatch event with result back to node, when channel sampling stops.
	sampleDoneEvt := evt.Copy()
	sampleDoneEvt.Type = EventTypeRadioChannelSample
	sampleDoneEvt.Timestamp += evt.RadioCommData.Duration
	sampleDoneEvt.MustDispatch = true
	rm.eventQ.Add(&sampleDoneEvt)
}
