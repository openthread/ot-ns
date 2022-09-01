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
	activeTransmitters map[NodeId]*RadioNode
}

func (rm *RadioModelMutualInterference) GetTxRssi(evt *Event, srcNode *RadioNode, dstNode *RadioNode) int8 {
	simplelogger.AssertTrue(srcNode != dstNode)
	distMeters := srcNode.GetDistanceTo(dstNode) * 0.1 // FIXME
	rssi := ComputeIndoorRssi(distMeters, srcNode.TxPower, dstNode.RxSensitivity)
	return rssi
}

func (rm *RadioModelMutualInterference) TxStart(node *RadioNode, q EventQueue, evt *Event) {
	simplelogger.AssertTrue(evt.Type == EventTypeRadioTx)
	node.TxPower = evt.Param1     // get the Tx power from the OT node's event param.
	node.CcaEdThresh = evt.Param2 // get CCA ED threshold also.
	isAck := dissectpkt.IsAckFrame(evt.Data)

	// check if a transmission is already ongoing by itself? If so signal OT_ERROR_ABORT back
	// to the OT stack, which will retry later without marking it as 'CCA failure'.
	if node.TxPhase > 0 {
		nextEvt := &Event{
			Type:      EventTypeRadioTxDone,
			Timestamp: evt.Timestamp + 1,
			Param1:    OT_ERROR_ABORT,
			NodeId:    evt.NodeId,
			Data:      evt.Data,
		}
		q.AddEvent(nextEvt)
		return
	}

	node.IsCcaFailed = false
	node.IsTxFailed = false
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
	nextEvt.Timestamp += delay
	nextEvt.Type = EventTypeRadioTxOngoing
	q.AddEvent(nextEvt)
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
		// re-use the current event as the next event for CCA period end, updating timing only.
		nextEvt := evt.Copy()
		nextEvt.Timestamp += ccaTimeUs
		q.AddEvent(nextEvt)

	case 3: // CCA second sample point and decision
		if !isAck && rm.ccaDetectsBusy(node, evt) {
			node.IsCcaFailed = true
		}
		if node.IsCcaFailed {
			// if CCA failed, then respond Tx Done event with error code ...
			nextEvt := &Event{
				Type:      EventTypeRadioTxDone,
				Timestamp: evt.Timestamp + 1,
				Param1:    OT_ERROR_CHANNEL_ACCESS_FAILURE,
				NodeId:    evt.NodeId,
				Data:      evt.Data,
			}
			q.AddEvent(nextEvt)

			// ... and reset back to start state.
			node.TxPhase = 0
			node.IsTxFailed = false
			node.IsCcaFailed = false
		} else {
			// Start frame transmission at time=now.
			rm.startTransmission(node, evt)

			// schedule the end-of-frame-transmission event, d us later.
			nextEvt := evt.Copy()
			d := getFrameDurationUs(evt)
			nextEvt.Timestamp += d
			q.AddEvent(nextEvt)
		}

	case 4: // End of frame transmit event
		rm.endTransmission(node, evt)

		// signal Tx Done event to sender.
		nextEvt := &Event{
			Type:      EventTypeRadioTxDone,
			Timestamp: evt.Timestamp + 1,
			Param1:    OT_ERROR_NONE,
			NodeId:    evt.NodeId,
			Data:      evt.Data,
		}
		q.AddEvent(nextEvt)

		// let other radios of Nodes receive the data (after 1 us propagation & processing delay)
		nextEvt2 := evt.Copy()
		nextEvt2.Type = EventTypeRadioReceived
		nextEvt2.Timestamp += 1
		nextEvt2.Param1 = OT_ERROR_NONE

		// Interference errors are later on applied by the radiomodel, during 1:1 event delivery.
		q.AddEvent(nextEvt2)
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

func (rm *RadioModelMutualInterference) AllowUnicastDispatch() bool {
	return false
}

func (rm *RadioModelMutualInterference) ccaDetectsBusy(node *RadioNode, evt *Event) bool {
	// loop all active transmitters, see if any one transmits above my CCA ED Threshold.
	for _, v := range rm.activeTransmitters {
		rssi := rm.GetTxRssi(evt, v, node)
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
	_, nodeTransmits := rm.activeTransmitters[evt.NodeId]
	simplelogger.AssertFalse(nodeTransmits)

	// mark what this new transmission will interfere with.
	for id, interferingTransmitter := range rm.activeTransmitters {
		node.InterferedBy[id] = interferingTransmitter
	}

	rm.activeTransmitters[evt.NodeId] = node

}

func (rm *RadioModelMutualInterference) endTransmission(node *RadioNode, evt *Event) {
	_, nodeTransmits := rm.activeTransmitters[evt.NodeId]
	simplelogger.AssertTrue(nodeTransmits)
	delete(rm.activeTransmitters, evt.NodeId)

	// reset values back for future transmission
	node.TimeLastTxEnded = evt.Timestamp     // for data frames and ACKs
	node.IsLastTxLong = IsLongDataframe(evt) // for data frames and ACKs
	node.TxPhase = 0                         // reset back
	node.IsTxFailed = false
	node.IsCcaFailed = false
	node.InterferedBy = make(map[NodeId]*RadioNode) // clear map
}

func (rm *RadioModelMutualInterference) ApplyInterference(evt *Event, src *RadioNode, dst *RadioNode) {
	// No interference modeled.
}
