package radiomodel

import (
	"github.com/openthread/ot-ns/dissectpkt"
	. "github.com/openthread/ot-ns/types"
	"github.com/simonlingoogle/go-simplelogger"
)

// RadioModelMutualInterference is a radio model where a transmission may interfere with another transmission
// ongoing, depending on the relative level (Rx energy in dBm) of signals. Also CCA is implemented in nodes
// with a CCA ED threshold as set by the node individually (or a default in case the node doesn't communicate
// the CCA ED threshold to the simulator.)
type RadioModelMutualInterference struct {
	activeTransmitters map[ChannelId]map[NodeId]*RadioNode

	// Configured minimum Signal-to-Interference (SIR) ratio in dB that is required to receive a signal
	// in presence of at least one interfering, other signal.
	MinSirDb int
}

func (rm *RadioModelMutualInterference) CheckRadioReachable(evt *Event, src *RadioNode, dst *RadioNode) bool {
	rssi := rm.GetTxRssi(evt, src, dst)
	return rssi >= RssiMin && rssi <= RssiMax && rssi >= dst.RxSensitivity
}

func (rm *RadioModelMutualInterference) GetTxRssi(evt *Event, srcNode *RadioNode, dstNode *RadioNode) int8 {
	simplelogger.AssertTrue(srcNode != dstNode)
	rssi := computeIndoorRssi(srcNode.RadioRange, srcNode.GetDistanceTo(dstNode), srcNode.TxPower, dstNode.RxSensitivity)
	return rssi
}

func (rm *RadioModelMutualInterference) TxStart(node *RadioNode, q EventQueue, evt *Event) {
	simplelogger.AssertTrue(evt.Type == EventTypeRadioTx)

	// check if a transmission is already ongoing by itself? If so signal OT_ERROR_ABORT back
	// to the OT stack, which will retry later without marking it as 'CCA failure'.
	if node.TxPhase > 0 {
		nextEvt := evt.Copy()
		nextEvt.Type = EventTypeRadioTxDone
		nextEvt.Timestamp += 1
		nextEvt.TxDoneData = TxDoneEventData{
			Channel: evt.TxData.Channel,
			Error:   OT_ERROR_ABORT,
		}
		q.AddEvent(&nextEvt)
		return
	}

	// init before new transmission
	node.TxPower = evt.TxData.TxPower        // get the Tx power from the OT node's event param.
	node.CcaEdThresh = evt.TxData.CcaEdTresh // get CCA ED threshold also.
	node.IsCcaFailed = false
	node.IsTxFailed = false
	node.IsLastTxLong = isLongDataframe(evt)
	node.FrameTxInfo = dissectpkt.Dissect(evt.Data)
	node.InterferedBy = make(map[NodeId]*RadioNode) // clear map
	node.TxPhase++

	// node starts Tx - first phase is to wait any mandatory 802.15.4 silence time (LIFS/SIFS)
	// before Tx can commence.
	var delay uint64
	if dissectpkt.IsAckFrame(node.FrameTxInfo) {
		var timeStartCca uint64 = evt.Timestamp
		if node.TimeNextTx > ccaTimeUs+turnaroundTimeUs { // check to avoid negative uint64
			timeStartCca = node.TimeNextTx - ccaTimeUs - turnaroundTimeUs
		}
		if evt.Timestamp < timeStartCca {
			// must delay additionally until allowed to send again
			delay = timeStartCca - evt.Timestamp
		} else {
			delay = 0 // No delay needed, proceed straight to CCA start
		}
	} else {
		delay = aifsTimeUs // ack is sent after fixed delay and no CCA.
	}
	// create an internal event to continue the transmission procedure after the delay.
	nextEvt := evt.Copy()
	nextEvt.Type = EventTypeRadioTxOngoing
	nextEvt.Timestamp += delay
	q.AddEvent(&nextEvt)
}

func (rm *RadioModelMutualInterference) txOngoing(node *RadioNode, q EventQueue, evt *Event) {
	simplelogger.AssertTrue(evt.Type == EventTypeRadioTxOngoing)
	isAck := dissectpkt.IsAckFrame(node.FrameTxInfo)

	node.TxPhase++
	if node.TxPhase == 2 && isAck {
		node.TxPhase = 4 // for ACKs, CCA not applied
	}

	switch node.TxPhase {
	case 2: // CCA first sample point
		//perform CCA check at current time point 1.
		if rm.ccaDetectsBusy(node, evt) {
			node.IsCcaFailed = true
		}
		// next event is for CCA period end
		nextEvt := evt.Copy()
		nextEvt.Timestamp += ccaTimeUs
		q.AddEvent(&nextEvt)

	case 3: // CCA second sample point and decision
		if rm.ccaDetectsBusy(node, evt) {
			node.IsCcaFailed = true
		}
		if node.IsCcaFailed {
			// if CCA failed, then respond Tx Done event with error code ...
			node.IsTxFailed = true
			nextEvt := evt.Copy()
			nextEvt.Type = EventTypeRadioTxDone
			nextEvt.Timestamp += 1
			nextEvt.TxDoneData = TxDoneEventData{
				Channel: evt.TxData.Channel,
				Error:   OT_ERROR_CHANNEL_ACCESS_FAILURE,
			}
			q.AddEvent(&nextEvt)

			// ... and reset back to start state.
			node.TxPhase = 0
		} else {
			// move to next state after turnAroundTime
			nextEvt := evt.Copy()
			nextEvt.Timestamp += turnaroundTimeUs
			q.AddEvent(&nextEvt)
		}

	case 4:
		// Start frame transmission
		rm.startTransmission(node, evt)

		// schedule the end-of-frame-transmission event, d us later.
		nextEvt := evt.Copy()
		nextEvt.Timestamp += getFrameDurationUs(evt)
		q.AddEvent(&nextEvt)

	case 5: // End of frame transmit event
		// signal Tx Done event to sender.
		nextEvt := evt.Copy()
		nextEvt.Type = EventTypeRadioTxDone
		nextEvt.Timestamp += 1
		nextEvt.TxDoneData = TxDoneEventData{
			Channel: evt.TxData.Channel,
			Error:   OT_ERROR_NONE,
		}
		q.AddEvent(&nextEvt)

		// let other radios of Nodes receive the data
		nextEvt2 := evt.Copy()
		nextEvt2.Type = EventTypeRadioRx
		nextEvt2.Timestamp += 1
		nextEvt2.RxData = RxEventData{
			Channel: evt.TxData.Channel,
			Error:   OT_ERROR_NONE,
			Rssi:    RssiInvalid, // Rssi will be computed upon individual event delivery to node.
		}
		q.AddEvent(&nextEvt2)

		rm.endTransmission(node, evt)
	}
}

func (rm *RadioModelMutualInterference) HandleEvent(node *RadioNode, q EventQueue, evt *Event) {
	switch evt.Type {
	case EventTypeRadioTx:
		rm.TxStart(node, q, evt)
	case EventTypeRadioTxOngoing:
		rm.txOngoing(node, q, evt)
	default:
		simplelogger.Errorf("Radiomodel event type not implemented: %v", evt.Type)
	}
}

func (rm *RadioModelMutualInterference) GetName() string {
	return "MutualInterference"
}

func (rm *RadioModelMutualInterference) init() {
	rm.activeTransmitters = map[ChannelId]map[NodeId]*RadioNode{}
	for c := minChannelNumber; c <= maxChannelNumber; c++ {
		rm.activeTransmitters[c] = map[NodeId]*RadioNode{}
	}
}

func (rm *RadioModelMutualInterference) ccaDetectsBusy(node *RadioNode, evt *Event) bool {
	// loop all active transmitters, see if any one transmits above my CCA ED Threshold.
	// This is only CCA Mode 1 (ED) per 802.15.4-2015.
	for _, v := range rm.activeTransmitters[evt.TxData.Channel] {
		rssi := rm.GetTxRssi(nil, v, node)
		if rssi == RssiInvalid {
			continue
		}
		if rssi >= node.CcaEdThresh {
			return true
		}
	}
	return false
}

func (rm *RadioModelMutualInterference) startTransmission(node *RadioNode, evt *Event) {
	_, nodeTransmits := rm.activeTransmitters[evt.TxData.Channel][evt.NodeId]
	simplelogger.AssertFalse(nodeTransmits)

	// mark what this new transmission will interfere with.
	for id, interferingTransmitter := range rm.activeTransmitters[evt.TxData.Channel] {
		node.InterferedBy[id] = interferingTransmitter
		interferingTransmitter.InterferedBy[evt.NodeId] = node
	}

	rm.activeTransmitters[evt.TxData.Channel][evt.NodeId] = node
}

func (rm *RadioModelMutualInterference) endTransmission(node *RadioNode, evt *Event) {
	_, nodeTransmits := rm.activeTransmitters[evt.TxData.Channel][evt.NodeId]
	simplelogger.AssertTrue(nodeTransmits)
	delete(rm.activeTransmitters[evt.TxData.Channel], evt.NodeId)

	// set values for future transmission
	node.TimeLastTxEnded = evt.Timestamp
	if isLongDataframe(evt) {
		node.TimeNextTx = evt.Timestamp + lifsTimeUs
	} else {
		node.TimeNextTx = evt.Timestamp + sifsTimeUs
	}
	node.TxPhase = 0 // reset the phase back.
}

func (rm *RadioModelMutualInterference) OnRxEventDispatch(evt *Event, src *RadioNode, dst *RadioNode) {
	simplelogger.AssertTrue(evt.Type == EventTypeRadioRx)

	// Apply interference.
	for _, interferer := range src.InterferedBy {
		if interferer == dst { // if dst node was at some point transmitting itself, fail the Rx
			evt.RxData.Error = OT_ERROR_ABORT
			return
		}
		rssiInterferer := int(rm.GetTxRssi(nil, interferer, dst))
		rssi := int(evt.RxData.Rssi)
		sirDb := rssi - rssiInterferer
		if sirDb < rm.MinSirDb {
			// interfering signal gets too close to the transmission rssi, impacts the signal.
			evt.Data = interferePsduData(evt.Data, float64(sirDb))
			evt.RxData.Error = OT_ERROR_FCS
		}
	}

	// In case of Ack being delivered to the node that did an Ack-request with same seq nr:
	// adjust IFS info of the dst node. Per 802.15.4-2015, IFS is applied after Ack Rx.
	isAck := dissectpkt.IsAckFrame(src.FrameTxInfo)
	if isAck && dst.FrameTxInfo != nil && dst.FrameTxInfo.MacFrame.Seq == src.FrameTxInfo.MacFrame.Seq &&
		dst.FrameTxInfo.MacFrame.FrameControl.AckRequest() {
		if dst.IsLastTxLong {
			if dst.TimeNextTx < evt.Timestamp+lifsTimeUs {
				dst.TimeNextTx = evt.Timestamp + lifsTimeUs
			}
		} else {
			if dst.TimeNextTx < evt.Timestamp+sifsTimeUs {
				dst.TimeNextTx = evt.Timestamp + sifsTimeUs
			}
		}
	}

}
