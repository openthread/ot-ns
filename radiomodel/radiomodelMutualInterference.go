// Copyright (c) 2022-2023, The OTNS Authors.
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
	"math"

	. "github.com/openthread/ot-ns/event"
	"github.com/openthread/ot-ns/logger"
	. "github.com/openthread/ot-ns/types"
)

// RadioModelMutualInterference is a radio model where a transmission may interfere with another transmission
// ongoing on the same channel, depending on the relative level (Rx energy in dBm) of signals. Also, CCA and
// energy scanning are supported. There is no hard stop of reception beyond the radioRange of the node; although
// the radioRange of the node represents the distance at which a minimally workable Thread link can operate, there
// is also radio reception possible beyond the radioRange. Also, devices with better Rx sensitivity will receive
// radio frames at longer distances beyond the radioRange.
type RadioModelMutualInterference struct {
	name       string
	params     *RadioModelParams
	prevParams RadioModelParams
	fading     *fadingModel

	nodes                 map[NodeId]*RadioNode
	activeTransmitters    map[ChannelId]map[NodeId]*RadioNode
	activeChannelSamplers map[ChannelId]map[NodeId]*RadioNode
	interferedBy          map[NodeId]map[NodeId]*RadioNode
	eventQ                EventQueue
}

func (rm *RadioModelMutualInterference) AddNode(nodeid NodeId, radioNode *RadioNode) {
	rm.nodes[nodeid] = radioNode
	rm.interferedBy[nodeid] = map[NodeId]*RadioNode{}
}

func (rm *RadioModelMutualInterference) DeleteNode(nodeid NodeId) {
	delete(rm.nodes, nodeid)
	for c := MinChannelNumber; c <= MaxChannelNumber; c++ {
		delete(rm.activeTransmitters[c], nodeid)
		delete(rm.activeChannelSamplers[c], nodeid)
	}
	rm.interferedBy[nodeid] = map[NodeId]*RadioNode{} // clear map
}

func (rm *RadioModelMutualInterference) CheckRadioReachable(src *RadioNode, dst *RadioNode) bool {
	if src == dst || dst.RadioState != RadioRx || src.RadioChannel != dst.RadioChannel {
		return false
	}
	if rm.params.IsDiscLimit && src.GetDistanceTo(dst) > src.RadioRange {
		return false
	}
	rssi := rm.GetTxRssi(src, dst)
	floorDbm := math.Max(dst.RxSensitivity, rm.params.NoiseFloorDbm) + rm.params.SnrMinThresholdDb
	return rssi >= RssiMin && rssi <= RssiMax && rssi >= floorDbm
}

func (rm *RadioModelMutualInterference) GetTxRssi(src *RadioNode, dst *RadioNode) DbValue {
	var rssi DbValue

	dist := src.GetDistanceTo(dst)
	if rm.params.IsDiscLimit && dist > src.RadioRange {
		return RssiMinusInfinity
	}

	if rm.params.RssiMinDbm < rm.params.RssiMaxDbm {
		rssi = computeIndoorRssi3gpp(dist, src.TxPower, rm.params)
		if rm.params.ShadowFadingSigmaDb > 0 || rm.params.TimeFadingSigmaMaxDb > 0 {
			rssi -= rm.fading.computeFading(src, dst, rm.params)
		}
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

func (rm *RadioModelMutualInterference) OnEventDispatch(src *RadioNode, dst *RadioNode, evt *Event) bool {
	switch evt.Type {
	case EventTypeRadioCommStart:
		// compute the RSSI and store in the event.
		evt.RadioCommData.PowerDbm = clipRssi(rm.GetTxRssi(src, dst))

	case EventTypeRadioRxDone:
		// compute the RSSI and store in the event
		evt.RadioCommData.PowerDbm = clipRssi(rm.GetTxRssi(src, dst))

		// check for interference by other signals and apply to event.
		rm.applyInterference(src, dst, evt)

	case EventTypeRadioChannelSample:
		// take final channel sample
		if evt.RadioCommData.Error == OT_ERROR_NONE {
			r := rm.getRssiOnChannel(src, evt.RadioCommData.Channel)
			if r > src.rssiSampleMax {
				src.rssiSampleMax = r
			}
			// store the final sampled RSSI in the event
			evt.RadioCommData.PowerDbm = clipRssi(src.rssiSampleMax)
		} else {
			evt.RadioCommData.PowerDbm = int8(RssiInvalid)
		}

	default:
		break
	}
	return true
}

func (rm *RadioModelMutualInterference) OnNextEventTime(ts uint64) {
	rm.fading.onAdvanceTime(ts)
}

func (rm *RadioModelMutualInterference) OnParametersModified() {
	// for specific parameter changes, clear cache, so that values may be rebuilt in conformance with latest
	// global parameter settings.
	if rm.prevParams.TimeFadingSigmaMaxDb != rm.params.TimeFadingSigmaMaxDb ||
		rm.prevParams.ShadowFadingSigmaDb != rm.params.ShadowFadingSigmaDb ||
		rm.prevParams.MeterPerUnit != rm.params.MeterPerUnit ||
		rm.prevParams.MeanTimeFadingChange != rm.params.MeanTimeFadingChange {
		rm.fading.clearCaches()
	}
	rm.prevParams = *rm.params
}

func (rm *RadioModelMutualInterference) HandleEvent(node *RadioNode, q EventQueue, evt *Event) {
	rm.eventQ = q

	switch evt.Type {
	case EventTypeRadioCommStart:
		rm.txStart(node, evt)
		rm.updateChannelSamplingNodes(node, evt) // all channel-sampling nodes detect the new Tx
	case EventTypeRadioTxDone:
		rm.txStop(node, evt)
	case EventTypeRadioChannelSample:
		rm.channelSampleStart(node, evt)
	case EventTypeRadioState:
		node.SetRadioState(evt.RadioStateData.EnergyState, evt.RadioStateData.SubState)
		node.SetChannel(evt.RadioStateData.Channel)
		node.SetRxSensitivity(DbValue(evt.RadioStateData.RxSensDbm))
	default:
		break // Unknown events not handled.
	}
}

func (rm *RadioModelMutualInterference) GetName() string {
	return rm.name
}

func (rm *RadioModelMutualInterference) GetParameters() *RadioModelParams {
	return rm.params
}

func (rm *RadioModelMutualInterference) init() {
	rm.nodes = map[NodeId]*RadioNode{}
	rm.activeTransmitters = map[ChannelId]map[NodeId]*RadioNode{}
	rm.activeChannelSamplers = map[ChannelId]map[NodeId]*RadioNode{}
	for c := MinChannelNumber; c <= MaxChannelNumber; c++ {
		rm.activeTransmitters[c] = map[NodeId]*RadioNode{}
		rm.activeChannelSamplers[c] = map[NodeId]*RadioNode{}
	}
	rm.interferedBy = map[NodeId]map[NodeId]*RadioNode{}
}

func (rm *RadioModelMutualInterference) getRssiAmbientNoise() DbValue {
	return rm.params.NoiseFloorDbm
}

func (rm *RadioModelMutualInterference) getRssiOnChannel(node *RadioNode, channel ChannelId) DbValue {
	rssiMax := rm.getRssiAmbientNoise()
	// loop all active transmitters
	for _, v := range rm.activeTransmitters[channel] {
		rssi := rm.GetTxRssi(v, node)
		if rssi == RssiInvalid {
			continue
		}
		rssiMax = addSignalPowersDbm(rssi, rssiMax)
	}
	return rssiMax
}

func (rm *RadioModelMutualInterference) txStart(node *RadioNode, evt *Event) {
	// verify node doesn't already transmit or sample on this channel.
	ch := evt.RadioCommData.Channel // move to the (new) channel for this Tx
	_, nodeTransmits := rm.activeTransmitters[ch][node.Id]
	_, nodeSamples := rm.activeChannelSamplers[ch][node.Id]
	if nodeTransmits || nodeSamples {
		// schedule new event dispatch to indicate tx is done with error.
		txDoneEvt := evt.Copy()
		txDoneEvt.Type = EventTypeRadioTxDone
		txDoneEvt.RadioCommData.Error = OT_ERROR_ABORT
		txDoneEvt.MustDispatch = true
		txDoneEvt.Timestamp += 1
		rm.eventQ.Add(&txDoneEvt)
		return
	}

	node.TxPower = DbValue(evt.RadioCommData.PowerDbm)
	node.SetChannel(ch)

	// reset interferedBy bookkeeping, remove data from last time.
	rm.interferedBy[node.Id] = map[NodeId]*RadioNode{} // clear map

	// mark what this new transmission will interfere with and will be interfered by.
	for id, interferingTransmitter := range rm.activeTransmitters[ch] {
		logger.AssertTrue(id != node.Id) // sanity check
		rm.interferedBy[node.Id][id] = interferingTransmitter
		rm.interferedBy[id][node.Id] = node
	}

	rm.activeTransmitters[ch][node.Id] = node

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

func (rm *RadioModelMutualInterference) txStop(node *RadioNode, evt *Event) {
	ch := evt.RadioCommData.Channel
	// if channel changed during operation, we need to stop it also at the old channel.
	isChannelChangedDuringTx := ch != node.RadioChannel
	if isChannelChangedDuringTx {
		delete(rm.activeTransmitters[node.RadioChannel], node.Id)
	}

	// stop active transmission
	delete(rm.activeTransmitters[ch], node.Id)

	// Dispatch TxDone event back to the source, at time==now
	txDoneEvt := evt.Copy()
	txDoneEvt.Type = EventTypeRadioTxDone
	txDoneEvt.RadioCommData.Error = OT_ERROR_NONE
	txDoneEvt.MustDispatch = true
	rm.eventQ.Add(&txDoneEvt)

	// Create RxDone event, to signal nearby node(s) that the frame Rx is done, at time==now
	rxDoneEvt := evt.Copy()
	rxDoneEvt.Type = EventTypeRadioRxDone
	rxDoneEvt.RadioCommData.Error = OT_ERROR_NONE
	if isChannelChangedDuringTx {
		rxDoneEvt.RadioCommData.Error = OT_ERROR_FCS
	}
	rxDoneEvt.MustDispatch = true
	rm.eventQ.Add(&rxDoneEvt)
}

func (rm *RadioModelMutualInterference) applyInterference(src *RadioNode, dst *RadioNode, evt *Event) {
	// Apply interference. Loop all interferers that were active during Tx by 'src' and add their signal powers.
	powIntfMax := rm.getRssiAmbientNoise()
	for _, interferer := range rm.interferedBy[src.Id] {
		if interferer == dst { // if dst node was at some point transmitting itself, fail the Rx
			rm.log(evt.Timestamp, dst.Id, "Detected self-transmission of Node, set Rx OT_ERROR_ABORT")
			evt.RadioCommData.Error = OT_ERROR_ABORT
			return
		}
		// calculate how strong the interferer was, as seen by dst
		powIntf := rm.GetTxRssi(interferer, dst)
		powIntfMax = addSignalPowersDbm(powIntf, powIntfMax)
	}

	// probabilistic BER model
	rssi := rm.GetTxRssi(src, dst)
	sirDb := rssi - powIntfMax // the Signal-to-Interferer (SIR/SINR) ratio
	isLogMsg, logMsg := applyBerModel(sirDb, src.Id, evt)
	if isLogMsg {
		rm.log(evt.Timestamp, dst.Id, logMsg) // log it on dest node's log
	}
}

// update sample value for all channel-sampling nodes that may detect the new source src.
func (rm *RadioModelMutualInterference) updateChannelSamplingNodes(src *RadioNode, evt *Event) {
	logger.AssertTrue(evt.Type == EventTypeRadioCommStart)
	ch := evt.RadioCommData.Channel
	for _, samplingNode := range rm.activeChannelSamplers[ch] {
		r := rm.GetTxRssi(src, samplingNode)
		if r != RssiInvalid {
			samplingNode.rssiSampleMax = addSignalPowersDbm(r, samplingNode.rssiSampleMax)
		}
	}
}

func (rm *RadioModelMutualInterference) channelSampleStart(srcNode *RadioNode, evt *Event) {
	ch := evt.RadioCommData.Channel
	// verify node doesn't already transmit or sample on its channel.
	_, nodeTransmits := rm.activeTransmitters[ch][srcNode.Id]
	_, nodeSamples := rm.activeChannelSamplers[ch][srcNode.Id]
	if nodeTransmits || nodeSamples {
		evt.RadioCommData.Error = OT_ERROR_ABORT
	} else {
		// take 1st channel sample
		srcNode.SetChannel(ch)
		srcNode.rssiSampleMax = rm.getRssiOnChannel(srcNode, ch)
	}
	// dispatch event with result back to node, when channel sampling stops.
	sampleDoneEvt := evt.Copy()
	sampleDoneEvt.Type = EventTypeRadioChannelSample
	sampleDoneEvt.Timestamp += evt.RadioCommData.Duration
	sampleDoneEvt.MustDispatch = true
	rm.eventQ.Add(&sampleDoneEvt)
}

func (rm *RadioModelMutualInterference) log(ts uint64, id NodeId, msg string) {
	const hdr = "(OTNS)       [T] RadioModelMI--: "
	rm.eventQ.Add(&Event{
		Timestamp: ts,
		Type:      EventTypeRadioLog,
		NodeId:    id,
		Data:      []byte(hdr + msg),
	})
}
