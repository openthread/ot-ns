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
	isRfBusy bool
}

func (rm *RadioModelMutualInterference) GetTxRssi(evt *Event, srcNode *RadioNode, dstNode *RadioNode, distMeters float64) int8 {
	simplelogger.AssertTrue(evt.Type == EventTypeRadioReceived)
	simplelogger.AssertTrue(srcNode != dstNode)
	rssi := ComputeIndoorRssi(distMeters, srcNode.TxPower, dstNode.RxSensitivity)
	return rssi
}

func (rm *RadioModelMutualInterference) TxStart(node *RadioNode, q EventQueue, evt *Event) {
	simplelogger.AssertTrue(evt.Type == EventTypeRadioTx)
	node.TxPower = evt.Param1     // get the Tx power from the OT node's event param.
	node.CcaEdThresh = evt.Param2 // get CCA ED threshold also.
	isAck := dissectpkt.IsAckFrame(evt.Data)

	// check if a transmission is already ongoing? If so signal OT_ERROR_ABORT back
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
	// create an internal event to continue the transmission later on
	nextEvt := evt.Copy()
	nextEvt.Timestamp += delay
	nextEvt.IsInternal = true
	q.AddEvent(nextEvt)
}

func (rm *RadioModelMutualInterference) TxOngoing(node *RadioNode, q EventQueue, evt *Event) {
	simplelogger.AssertTrue(evt.Type == EventTypeRadioTx)
	isAck := dissectpkt.IsAckFrame(evt.Data)

	node.TxPhase++
	if node.TxPhase == 2 && isAck {
		node.TxPhase++ // for ACKs, CCA not applied but fixed delay in phase 1.
	}

	switch node.TxPhase {
	case 2: // CCA first sample point
		//perform CCA check at current time point 1.
		if rm.isRfBusy && !isAck {
			node.IsCcaFailed = true // TODO use the node.CcaEdThresh to determine this.
		}
		// re-use the current event as the next event for CCA period end, updating timing only.
		nextEvt := evt.Copy()
		nextEvt.Timestamp += ccaTimeUs
		q.AddEvent(nextEvt)

	case 3: // CCA second sample point and decision
		if rm.isRfBusy && !isAck {
			node.IsCcaFailed = true // TODO use the node.CcaEdThresh to determine this.
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
			if rm.isRfBusy {
				// if collides with existing ongoing transmission, mark failed.
				// FIXME mark the other ongoing transmission as failed, too.
				node.IsTxFailed = true
			}

			// Start frame transmission at time=now.
			rm.isRfBusy = true

			// schedule the end-of-frame-transmission event, d us later.
			nextEvt := evt.Copy()
			d := getFrameDurationUs(evt)
			nextEvt.Timestamp += d
			q.AddEvent(nextEvt)
		}

	case 4: // End of frame transmit event

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

		if node.IsTxFailed { // mark as interfered packet in case of failure
			nextEvt2.Param1 = OT_ERROR_FCS
		}
		q.AddEvent(nextEvt2)

		// final state done, bookkeeping for node and globally
		node.TimeLastTxEnded = evt.Timestamp     // for data frames and ACKs
		node.IsLastTxLong = IsLongDataframe(evt) // for data frames and ACKs
		node.TxPhase = 0                         // reset back
		node.IsTxFailed = false
		node.IsCcaFailed = false
		rm.isRfBusy = false
	}
}

func (rm *RadioModelMutualInterference) HandleEvent(node *RadioNode, q EventQueue, evt *Event) {
	switch evt.Type {
	case EventTypeRadioTx:
		if !evt.IsInternal {
			rm.TxStart(node, q, evt)
		} else {
			rm.TxOngoing(node, q, evt)
		}
	default:
		simplelogger.Panicf("event type not implemented: %v", evt.Type)
	}
}

func (rm *RadioModelMutualInterference) GetName() string {
	return "MutualInterference"
}
