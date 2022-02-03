package radiomodel

import (
	"github.com/openthread/ot-ns/openthread"
	. "github.com/openthread/ot-ns/types"
	"github.com/simonlingoogle/go-simplelogger"
)

type RadioModelInterfereAll struct {
	isRfBusy bool
}

func (rm *RadioModelInterfereAll) GetTxRssi(evt *Event, srcNode *RadioNode, dstNode *RadioNode, distMeters float64) int8 {
	simplelogger.AssertTrue(evt.Type == EventTypeRadioFrameToNode)
	simplelogger.AssertTrue(srcNode != dstNode)
	rssi := ComputeIndoorRssi(distMeters, srcNode.TxPower, dstNode.RxSensitivity)
	return rssi
}

func (rm *RadioModelInterfereAll) TxStart(node *RadioNode, q EventQueue, evt *Event) {
	var nextEvt *Event
	simplelogger.AssertTrue(evt.Type == EventTypeRadioFrameToSim || evt.Type == EventTypeRadioFrameAckToSim)
	node.TxPower = evt.Param1     // get the Tx power from the OT node's event param.
	node.CcaEdThresh = evt.Param2 // get CCA ED threshold also.
	isAck := evt.Type == EventTypeRadioFrameAckToSim

	// check if a transmission is already ongoing? If so return OT_ERROR_ABORT.
	if node.TxPhase > 0 {
		// FIXME: submac layer of OpenThread shouldn't start Tx when it's still ongoing.
		nextEvt = &Event{
			Type:      EventTypeRadioTxDone,
			Timestamp: evt.Timestamp + 1,
			Delay:     1,
			Data:      []byte{openthread.OT_ERROR_ABORT},
			NodeId:    evt.NodeId,
		}
		q.AddEvent(nextEvt)
		return
	}

	node.IsCcaFailed = false
	node.IsTxFailed = false
	node.TxPhase++

	// node starts Tx - first phase is to wait any mandatory 802.15.4 silence time (LIFS/SIFS)
	// before CCA can commence.
	var delay uint64 = 0
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
	// re-use the current event as the next event, updating timing and marking it internally-sourced.
	nextEvt = evt
	nextEvt.Timestamp += delay
	nextEvt.Delay = delay
	nextEvt.IsInternal = true
	q.AddEvent(nextEvt)
}

func (rm *RadioModelInterfereAll) TxOngoing(node *RadioNode, q EventQueue, evt *Event) {

	simplelogger.AssertTrue(evt.Type == EventTypeRadioFrameToSim || evt.Type == EventTypeRadioFrameAckToSim)
	isAck := evt.Type == EventTypeRadioFrameAckToSim

	node.TxPhase++
	if node.TxPhase == 2 && isAck {
		node.TxPhase++ // for ACKs, CCA not applied but fixed delay in phase 1.
	}

	switch node.TxPhase {
	case 2: // CCA first sample point
		//perform CCA check at current time point 1.
		if rm.isRfBusy {
			node.IsCcaFailed = true // TODO use the node.CcaEdThresh to determine this.
		}
		// re-use the current event as the next event for CCA period end, updating timing only.
		nextEvt := evt
		nextEvt.Timestamp += ccaTimeUs
		nextEvt.Delay = ccaTimeUs
		q.AddEvent(nextEvt)

	case 3: // CCA second sample point and decision
		if rm.isRfBusy && !isAck {
			node.IsCcaFailed = true // TODO use the node.CcaEdThresh to determine this.
		}
		if node.IsCcaFailed && !isAck {
			// if CCA fails, then respond Tx Done event with error code.
			nextEvt := &Event{
				Type:      EventTypeRadioTxDone,
				Timestamp: evt.Timestamp + 1,
				Delay:     1,
				Data:      []byte{openthread.OT_ERROR_CHANNEL_ACCESS_FAILURE},
				NodeId:    evt.NodeId,
			}
			q.AddEvent(nextEvt)
			node.TxPhase = 0 // reset back

		} else {
			if rm.isRfBusy && isAck {
				// if ACK collides with existing transmission, mark ACK as failed.
				node.IsTxFailed = true
			} else {
				// CCA was successful, or it's an ACK, so start frame transmission now.
				rm.isRfBusy = true
			}
			// schedule the end-of-frame-transmission event.
			nextEvt := evt
			d := rm.getFrameDurationUs(evt)
			nextEvt.Timestamp += d
			nextEvt.Delay = d
			q.AddEvent(nextEvt)
		}

	case 4: // End of frame transmit event

		// signal Tx Done event to sender.
		nextEvt := &Event{
			Type:      EventTypeRadioTxDone,
			Timestamp: evt.Timestamp + 1,
			Delay:     1,
			Data:      []byte{openthread.OT_ERROR_NONE},
			NodeId:    evt.NodeId,
		}
		q.AddEvent(nextEvt)

		// bookkeeping
		node.TimeLastTxEnded = evt.Timestamp     // for data frames and ACKs
		node.IsLastTxLong = IsLongDataframe(evt) // for data frames and ACKs
		node.TxPhase = 0                         // reset back
		rm.isRfBusy = false

		// let other radios of Nodes receive the data (after 1 us propagation delay)
		nextEvt = evt
		nextEvt.Type = EventTypeRadioFrameToNode
		nextEvt.IsInternal = false
		nextEvt.Timestamp += 1
		nextEvt.Delay = 1
		if node.IsTxFailed { // mark as interfered packet in case of failure
			nextEvt.Type = EventTypeRadioFrameToNodeInterfered
		}
		q.AddEvent(nextEvt)
	}
}

func (rm *RadioModelInterfereAll) HandleEvent(node *RadioNode, q EventQueue, evt *Event) {
	switch evt.Type {
	case EventTypeRadioFrameAckToSim:
		fallthrough
	case EventTypeRadioFrameToSim:
		if !evt.IsInternal {
			rm.TxStart(node, q, evt)
		} else {
			rm.TxOngoing(node, q, evt)
		}
	default:
		simplelogger.Panicf("event type not implemented: %v", evt.Type)
	}
}

// getFrameDurationUs gets the duration of the PHY frame in us indicated by evt of type eventTypeRadioFrame*
func (rm *RadioModelInterfereAll) getFrameDurationUs(evt *Event) uint64 {
	var n uint64
	simplelogger.AssertTrue(len(evt.Data) >= RadioMessagePsduOffset)
	n = (uint64)(len(evt.Data) - RadioMessagePsduOffset) // PSDU size 5..127
	n += phyHeaderSize                                   // add PHY preamble, sfd, PHR bytes
	return n * symbolTimeUs * symbolsPerOctet
}
