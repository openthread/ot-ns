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

package radiomodel

import (
	. "github.com/openthread/ot-ns/types"
)

// RadioModelIdeal is an ideal radio model with infinite parallel transmission capacity per
// channel. RSSI at the receiver can be set to an ideal constant RSSI value, or to a value
// based on an average RF propagation model. There is a hard stop of reception beyond the
// radioRange of the node i.e. ideal disc model.
type RadioModelIdeal struct {
	Name            string
	UseVariableRssi bool // when true uses distance-dependent RSSI model, else FixedRssi.
	FixedRssi       DbmValue
	IndoorParams    *IndoorModelParams

	nodes map[NodeId]*RadioNode
}

func (rm *RadioModelIdeal) AddNode(nodeid NodeId, radioNode *RadioNode) {
	rm.nodes[nodeid] = radioNode
}

func (rm *RadioModelIdeal) DeleteNode(nodeid NodeId) {
	delete(rm.nodes, nodeid)
}

func (rm *RadioModelIdeal) CheckRadioReachable(src *RadioNode, dst *RadioNode) bool {
	if src != dst && dst.RadioState == RadioRx {
		dist := src.GetDistanceTo(dst)
		if dist <= src.RadioRange { // simple disc radio model
			return true
		}
	}
	return false
}

func (rm *RadioModelIdeal) GetTxRssi(srcNode *RadioNode, dstNode *RadioNode) DbmValue {
	rssi := rm.FixedRssi // in the most ideal case, always assume a good RSSI up until the max range.
	if rm.UseVariableRssi {
		rssi = computeIndoorRssi(srcNode.RadioRange, srcNode.GetDistanceTo(dstNode), srcNode.TxPower, rm.IndoorParams)
	}
	return rssi
}

func (rm *RadioModelIdeal) OnEventDispatch(src *RadioNode, dst *RadioNode, evt *Event) bool {
	switch evt.Type {
	case EventTypeRadioCommStart:
		fallthrough
	case EventTypeRadioRxDone:
		// compute the RSSI and store it in the event
		evt.RadioCommData.PowerDbm = rm.GetTxRssi(src, dst)
	case EventTypeRadioChannelSample:
		// store the final sampled RSSI in the event
		evt.RadioCommData.PowerDbm = src.rssiSampleMax
	}
	return true
}

func (rm *RadioModelIdeal) HandleEvent(node *RadioNode, q EventQueue, evt *Event) {
	switch evt.Type {
	case EventTypeRadioCommStart:
		rm.txStart(node, q, evt)
	case EventTypeRadioTxDone:
		rm.txStop(node, q, evt)
	case EventTypeRadioChannelSample:
		rm.channelSampleStart(node, q, evt)
	}
}

func (rm *RadioModelIdeal) GetName() string {
	return rm.Name
}

func (rm *RadioModelIdeal) init() {
	rm.nodes = map[NodeId]*RadioNode{}
}

func (rm *RadioModelIdeal) txStart(srcNode *RadioNode, q EventQueue, evt *Event) {
	srcNode.TxPower = evt.RadioCommData.PowerDbm // get last node's properties from the OT node's event params.
	srcNode.SetChannel(int(evt.RadioCommData.Channel))

	// dispatch radio event RadioComm 'start of frame Rx' to listening nodes.
	rxStartEvt := evt.Copy()
	rxStartEvt.Type = EventTypeRadioCommStart
	rxStartEvt.RadioCommData.Error = OT_ERROR_NONE
	rxStartEvt.MustDispatch = true
	q.Add(&rxStartEvt)

	// schedule new internal event to call txStop() at end of duration.
	txDoneEvt := evt.Copy()
	txDoneEvt.Type = EventTypeRadioTxDone
	txDoneEvt.RadioCommData.Error = OT_ERROR_NONE
	txDoneEvt.MustDispatch = false
	txDoneEvt.Timestamp += evt.RadioCommData.Duration
	q.Add(&txDoneEvt)
}

func (rm *RadioModelIdeal) txStop(node *RadioNode, q EventQueue, evt *Event) {
	// Dispatch TxDone event back to the source
	txDoneEvt := evt.Copy()
	txDoneEvt.Type = EventTypeRadioTxDone
	txDoneEvt.RadioCommData.Error = OT_ERROR_NONE
	txDoneEvt.MustDispatch = true
	q.Add(&txDoneEvt)

	// Create RxDone event, to signal nearby node(s) the frame Rx is done.
	rxDoneEvt := evt.Copy()
	rxDoneEvt.Type = EventTypeRadioRxDone
	rxDoneEvt.MustDispatch = true
	q.Add(&rxDoneEvt)
}

func (rm *RadioModelIdeal) channelSampleStart(node *RadioNode, q EventQueue, evt *Event) {
	node.rssiSampleMax = RssiMinusInfinity // Ideal model never has CCA failure.
	node.SetChannel(int(evt.RadioCommData.Channel))

	// dispatch event with result back to node, when channel sampling stops.
	sampleDoneEvt := evt.Copy()
	sampleDoneEvt.Type = EventTypeRadioChannelSample
	sampleDoneEvt.Timestamp += evt.RadioCommData.Duration
	sampleDoneEvt.MustDispatch = true
	q.Add(&sampleDoneEvt)
}
