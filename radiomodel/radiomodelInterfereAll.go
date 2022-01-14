package radiomodel

import (
	"github.com/openthread/ot-ns/openthread"
	. "github.com/openthread/ot-ns/types"
	"github.com/simonlingoogle/go-simplelogger"
)

const CCA_TIME_US uint64 = 128
const SYMBOL_TIME_US uint64 = 16
const ACK_AIFS_TIME_US uint64 = SYMBOL_TIME_US * 12

type RadioModelInterfereAll struct {
	isRfBusy bool
}

func (rm *RadioModelInterfereAll) IsTxSuccess(node *RadioNode, evt *Event) bool {
	return true
}

func (rm *RadioModelInterfereAll) TxStart(node *RadioNode, q EventQueue, evt *Event) {
	var nextEvt *Event

	// check if a transmission is already ongoing? If so return OT_ERROR_ABORT.
	if node.TxPhase > 0 {
		// FIXME: submac layer of OpenThread shouldn't start Tx when it's still ongoing.
		nextEvt = &Event{
			Type:      EventTypeRadioTxDone,
			Timestamp: evt.Timestamp + 0,
			Delay:     0,
			Data:      []byte{openthread.OT_ERROR_ABORT},
			NodeId:    evt.NodeId,
		}
		q.AddEvent(nextEvt)
		return
	}

	node.IsCcaFailed = false
	node.TxPhase++

	// node starts Tx - perform CCA check at current time point 1.
	if rm.isRfBusy {
		node.IsCcaFailed = true
	}

	// re-use the current event as the next event, updating type and timing.
	nextEvt = evt
	nextEvt.Type = EventTypeRadioFrameSimInternal
	nextEvt.Timestamp += CCA_TIME_US
	nextEvt.Delay = CCA_TIME_US
	q.AddEvent(nextEvt)
}

func (rm *RadioModelInterfereAll) TxOngoing(node *RadioNode, q EventQueue, evt *Event) {

	node.TxPhase++
	switch node.TxPhase {
	case 2: // CCA second sample point and decision
		if rm.isRfBusy {
			node.IsCcaFailed = true
		}
		if node.IsCcaFailed {
			// if CCA fails, then respond Tx Done with error code.
			nextEvt := &Event{
				Type:      EventTypeRadioTxDone,
				Timestamp: evt.Timestamp,
				Delay:     0,
				Data:      []byte{openthread.OT_ERROR_CHANNEL_ACCESS_FAILURE},
				NodeId:    evt.NodeId,
			}
			q.AddEvent(nextEvt)
			node.TxPhase = 0 // reset back
		} else {
			// CCA was successful, start frame transmission now.
			rm.isRfBusy = true
			// schedule the end-of-frame-transmission event.
			nextEvt := evt
			nextEvt.Type = EventTypeRadioFrameSimInternal
			d := rm.getFrameDurationUs(evt)
			nextEvt.Timestamp += d
			nextEvt.Delay = d
			q.AddEvent(nextEvt)
		}
		node.IsCcaFailed = false
	case 3: // End of frame transmit event

		// signal Tx Done to sender.
		nextEvt := &Event{
			Type:      EventTypeRadioTxDone,
			Timestamp: evt.Timestamp,
			Delay:     0,
			Data:      []byte{openthread.OT_ERROR_NONE},
			NodeId:    evt.NodeId,
		}
		q.AddEvent(nextEvt)

		// let other radios of Nodes receive the data.
		nextEvt = evt
		nextEvt.Type = EventTypeRadioFrameToNode
		nextEvt.Timestamp += 0
		nextEvt.Delay = 0
		q.AddEvent(nextEvt)

		node.TxPhase = 0 // reset back
		node.IsCcaFailed = false
		rm.isRfBusy = false
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
	n = (uint64)(len(evt.Data) - 1) // PSDU size 5..127
	n += 6                          // add PHY preamble, sfd, PHR bytes
	return n * 8 * 1000000 / 250000
}
