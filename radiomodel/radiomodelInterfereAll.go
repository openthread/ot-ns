package radiomodel

import (
	"github.com/openthread/ot-ns/openthread"
	. "github.com/openthread/ot-ns/types"
	"github.com/simonlingoogle/go-simplelogger"
)

type RadioModelInterfereAll struct {
	isRfBusy bool
}

func (rm *RadioModelInterfereAll) IsTxSuccess(node *RadioNode, evt *Event) bool {
	return true
}

func (rm *RadioModelInterfereAll) TxStart(node *RadioNode, q EventQueue, evt *Event) {
	var nextEvt *Event
	simplelogger.AssertTrue(evt.Type == EventTypeRadioFrameToSim)

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
	node.TxPhase++

	// node starts Tx - first phase is to wait any mandatory 802.15.4 silence time (LIFS/SIFS)
	// before CCA can commence.
	var delay uint64 = 0
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

	// re-use the current event as the next event, updating type and timing.
	nextEvt = evt
	nextEvt.Type = EventTypeRadioFrameSimInternal
	nextEvt.Timestamp += delay
	nextEvt.Delay = delay
	q.AddEvent(nextEvt)
}

func (rm *RadioModelInterfereAll) TxOngoing(node *RadioNode, q EventQueue, evt *Event) {

	simplelogger.AssertTrue(evt.Type == EventTypeRadioFrameSimInternal)

	node.TxPhase++
	switch node.TxPhase {
	case 2: // CCA first sample point
		//perform CCA check at current time point 1.
		if rm.isRfBusy {
			node.IsCcaFailed = true
		}
		// re-use the current event as the next event for CCA period end, updating timing only.
		nextEvt := evt
		nextEvt.Timestamp += ccaTimeUs
		nextEvt.Delay = ccaTimeUs
		q.AddEvent(nextEvt)

	case 3: // CCA second sample point and decision
		if rm.isRfBusy {
			node.IsCcaFailed = true
		}
		if node.IsCcaFailed {
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
			node.IsCcaFailed = false
		} else {
			// CCA was successful, start frame transmission now.
			rm.isRfBusy = true

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
		node.TimeLastTxEnded = evt.Timestamp
		node.IsLastTxLong = IsLongDataframe(evt)
		node.TxPhase = 0 // reset back
		node.IsCcaFailed = false
		rm.isRfBusy = false

		// let other radios of Nodes receive the data (after 1 us)
		nextEvt = evt
		nextEvt.Type = EventTypeRadioFrameToNode
		nextEvt.Timestamp += 1
		nextEvt.Delay = 1
		q.AddEvent(nextEvt)
	}
}

func (rm *RadioModelInterfereAll) HandleEvent(node *RadioNode, q EventQueue, evt *Event) {
	switch evt.Type {
	case EventTypeRadioFrameToSim:
		rm.TxStart(node, q, evt)
	case EventTypeRadioFrameSimInternal:
		rm.TxOngoing(node, q, evt)
	default:
		simplelogger.Panicf("event type not implemented: %v", evt.Type)
	}
}

// getFrameDurationUs gets the duration of the PHY frame indicated by evt of type eventTypeRadioFrame*
func (rm *RadioModelInterfereAll) getFrameDurationUs(evt *Event) uint64 {
	var n uint64
	simplelogger.AssertTrue(len(evt.Data) >= RadioMessagePsduOffset)
	n = (uint64)(len(evt.Data) - RadioMessagePsduOffset) // PSDU size 5..127
	n += phyHeaderSize                                   // add PHY preamble, sfd, PHR bytes
	return n * symbolTimeUs * symbolsPerOctet
}
