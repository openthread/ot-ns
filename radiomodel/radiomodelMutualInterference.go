package radiomodel

import (
	"github.com/openthread/ot-ns/dissectpkt"
	. "github.com/openthread/ot-ns/types"
	"github.com/simonlingoogle/go-simplelogger"
)

// RadioModelMutualInterference is a somewhat pessimistic radio model where one transmitter will always interfere
// with all other transmitters in the simulation, regardless of distance. This means no 2 or more nodes
// can transmit at the same time. It's useful to evaluate capacity-limited situations.
type RadioModelMutualInterference struct {
	ActiveTransmitters map[NodeId]*RadioNode
	MinSirDb           int
}

func (rm *RadioModelMutualInterference) CheckRadioReachable(evt *Event, src *RadioNode, dst *RadioNode) bool {
	rssi := rm.GetTxRssi(evt, src, dst)
	return rssi >= RssiMin && rssi <= RssiMax && rssi >= dst.RxSensitivity
}

func (rm *RadioModelMutualInterference) GetTxRssi(evt *Event, srcNode *RadioNode, dstNode *RadioNode) int8 {
	simplelogger.AssertTrue(srcNode != dstNode)
	rssi := ComputeIndoorRssi(srcNode.RadioRange, srcNode.GetDistanceTo(dstNode), srcNode.TxPower, dstNode.RxSensitivity)
	return rssi
}

func (rm *RadioModelMutualInterference) TxStart(node *RadioNode, q EventQueue, evt *Event) {
	simplelogger.AssertTrue(evt.Type == EventTypeRadioTx)
	if node.TxPower != evt.TxData.TxPower {
		node.TxPower = evt.TxData.TxPower // get the Tx power from the OT node's event param.
	}
	node.CcaEdThresh = evt.TxData.CcaEdTresh // get CCA ED threshold also.
	isAck := dissectpkt.IsAckFrame(evt.Data)

	// check if a transmission is already ongoing by itself? If so signal OT_ERROR_ABORT back
	// to the OT stack, which will retry later without marking it as 'CCA failure'.
	if node.TxPhase > 0 {
		nextEvt := evt.Copy()
		nextEvt.Type = EventTypeRadioTxDone
		nextEvt.Timestamp += 1
		nextEvt.TxDoneData.Error = OT_ERROR_ABORT
		q.AddEvent(&nextEvt)
		return
	}

	node.IsCcaFailed = false
	node.IsTxFailed = false
	node.InterferedBy = make(map[NodeId]*RadioNode) // clear map
	node.TxPhase++

	// node starts Tx - first phase is to wait any mandatory 802.15.4 silence time (LIFS/SIFS)
	// before CCA can commence.
	var delay uint64
	if !isAck {
		var ifs uint64 = sifsTimeUs
		if node.IsLastTxLong {
			ifs = lifsTimeUs
		}
		if ifs > ccaTimeUs {
			ifs -= ccaTimeUs // CCA time may be part of the IFS. TODO check vs 15.4 standards
		}
		if evt.Timestamp >= ifs && node.TimeLastTxEnded > (evt.Timestamp-ifs) {
			// must delay additionally until allowed to send again
			delay = ifs - (evt.Timestamp - node.TimeLastTxEnded)
		} else {
			delay = 0 // No delay needed, proceed straight to CCA start
		}
	} else {
		delay = aifsTimeUs // ack is sent after fixed delay and no CCA.
	}
	// create an internal event to continue the transmission procedure later on
	nextEvt := evt.Copy()
	nextEvt.Type = EventTypeRadioTxOngoing
	nextEvt.Timestamp += delay
	q.AddEvent(&nextEvt)
}

func (rm *RadioModelMutualInterference) TxOngoing(node *RadioNode, q EventQueue, evt *Event) {
	simplelogger.AssertTrue(evt.Type == EventTypeRadioTxOngoing)
	isAck := dissectpkt.IsAckFrame(evt.Data)

	node.TxPhase++
	if node.TxPhase == 2 && isAck {
		node.TxPhase++ // for ACKs, CCA not applied but fixed delay in phase 1.
	}

	switch node.TxPhase {
	case 2: // CCA first sample point
		//perform CCA check at current time point 1.
		if !isAck && rm.ccaDetectsBusy(node, evt) {
			node.IsCcaFailed = true
		}
		// next event is for CCA period end
		nextEvt := evt.Copy()
		nextEvt.Timestamp += ccaTimeUs
		q.AddEvent(&nextEvt)

	case 3: // CCA second sample point and decision
		if !isAck && rm.ccaDetectsBusy(node, evt) {
			node.IsCcaFailed = true
		}
		if node.IsCcaFailed {
			// if CCA failed, then respond Tx Done event with error code ...
			node.IsTxFailed = true
			nextEvt := evt.Copy()
			nextEvt.Type = EventTypeRadioTxDone
			nextEvt.Timestamp += 1
			nextEvt.TxDoneData.Error = OT_ERROR_CHANNEL_ACCESS_FAILURE
			q.AddEvent(&nextEvt)

			// ... and reset back to start state.
			node.TxPhase = 0
		} else {
			// Start frame transmission at time=now.
			rm.startTransmission(node, evt)

			// schedule the end-of-frame-transmission event, d us later.
			nextEvt := evt.Copy()
			nextEvt.Timestamp += getFrameDurationUs(evt)
			q.AddEvent(&nextEvt)
		}

	case 4: // End of frame transmit event
		// signal Tx Done event to sender.
		nextEvt := evt.Copy()
		nextEvt.Type = EventTypeRadioTxDone
		nextEvt.Timestamp += 1
		nextEvt.TxDoneData.Error = OT_ERROR_NONE
		q.AddEvent(&nextEvt)

		// let other radios of Nodes receive the data
		nextEvt2 := evt.Copy()
		nextEvt2.Type = EventTypeRadioRx
		nextEvt2.Timestamp += 1
		nextEvt2.TxDoneData.Error = OT_ERROR_NONE
		q.AddEvent(&nextEvt2)

		rm.endTransmission(node, evt)
	}
}

func (rm *RadioModelMutualInterference) HandleEvent(node *RadioNode, q EventQueue, evt *Event) {
	switch evt.Type {
	case EventTypeRadioTx:
		rm.TxStart(node, q, evt)
	case EventTypeRadioTxOngoing:
		rm.TxOngoing(node, q, evt)
	default:
		simplelogger.Panicf("event type not implemented: %v", evt.Type)
	}
}

func (rm *RadioModelMutualInterference) GetName() string {
	return "MutualInterference"
}

func (rm *RadioModelMutualInterference) ccaDetectsBusy(node *RadioNode, evt *Event) bool {
	// loop all active transmitters, see if any one transmits above my CCA ED Threshold.
	for _, v := range rm.ActiveTransmitters {
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
	_, nodeTransmits := rm.ActiveTransmitters[evt.NodeId]
	simplelogger.AssertFalse(nodeTransmits)

	// mark what this new transmission will interfere with.
	for id, interferingTransmitter := range rm.ActiveTransmitters {
		node.InterferedBy[id] = interferingTransmitter
		interferingTransmitter.InterferedBy[evt.NodeId] = node
	}

	rm.ActiveTransmitters[evt.NodeId] = node
}

func (rm *RadioModelMutualInterference) endTransmission(node *RadioNode, evt *Event) {
	_, nodeTransmits := rm.ActiveTransmitters[evt.NodeId]
	simplelogger.AssertTrue(nodeTransmits)
	delete(rm.ActiveTransmitters, evt.NodeId)

	// set values for future transmission
	node.TimeLastTxEnded = evt.Timestamp     // for data frames and ACKs
	node.IsLastTxLong = IsLongDataframe(evt) // for data frames and ACKs
	node.TxPhase = 0                         // reset the phase back.
}

func (rm *RadioModelMutualInterference) ApplyInterference(evt *Event, src *RadioNode, dst *RadioNode) {
	simplelogger.AssertTrue(evt.Type == EventTypeRadioRx)
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
			evt.Data = InterferePsduData(evt.Data, float64(sirDb))
			evt.RxData.Error = OT_ERROR_FCS
		}
	}
}
