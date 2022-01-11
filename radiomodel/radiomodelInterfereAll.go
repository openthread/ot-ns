package radiomodel

import (
	"github.com/openthread/ot-ns/openthread"
	"github.com/openthread/ot-ns/types"
	"github.com/simonlingoogle/go-simplelogger"
)

const CCA_TIME_US uint64 = 128
const SYMBOL_TIME_US uint64 = 16
const ACK_AIFS_TIME_US uint64 = SYMBOL_TIME_US * 12

type RadioModelInterfereAll struct {
	isRfBusy bool
}

func (rm *RadioModelInterfereAll) IsTxSuccess(node types.RadioNode, evt *types.Event) bool {
	return true
}

func (rm *RadioModelInterfereAll) TxStart(node types.RadioNode, evt *types.Event) {
	var nextEvt *types.Event

	// check if a transmission is already ongoing? If so return OT_ERROR_ABORT.
	if node.TxPhase() > 0 {
		// FIXME: submac layer of OpenThread shouldn't start Tx when it's still ongoing.
		nextEvt = &types.Event{
			Type:      types.EventTypeRadioTxDone,
			Timestamp: evt.Timestamp + 0,
			Delay:     0,
			Data:      []byte{openthread.OT_ERROR_ABORT},
		}
		node.AddEvent(nextEvt)
		return
	}

	node.SetCcaFailed(false)
	node.AdvanceTxPhase()

	// node starts Tx - perform CCA check at current time point 1.
	if rm.isRfBusy {
		node.SetCcaFailed(true)
	}

	// re-use the current event as the next event, updating type and timing.
	nextEvt = evt
	nextEvt.Type = types.EventTypeRadioFrameSimInternal
	nextEvt.Timestamp += CCA_TIME_US
	nextEvt.Delay = CCA_TIME_US
	node.AddEvent(nextEvt)
}

func (rm *RadioModelInterfereAll) TxOngoing(node types.RadioNode, evt *types.Event) {

	node.AdvanceTxPhase()
	switch node.TxPhase() {
	case 2: // CCA second sample point and decision
		if rm.isRfBusy {
			node.SetCcaFailed(true)
		}
		if node.IsCcaFailed() {
			// if CCA fails, then respond Tx Done with error code.
			nextEvt := &types.Event{
				Type:      types.EventTypeRadioTxDone,
				Timestamp: evt.Timestamp,
				Delay:     0,
				Data:      []byte{openthread.OT_ERROR_CHANNEL_ACCESS_FAILURE},
			}
			node.AddEvent(nextEvt)
			node.SetTxPhase(0) // reset back
		} else {
			// CCA was successful, start frame transmission now.
			rm.isRfBusy = true
			// schedule the end-of-frame-transmission event.
			nextEvt := evt
			nextEvt.Type = types.EventTypeRadioFrameSimInternal
			d := rm.getFrameDurationUs(evt)
			nextEvt.Timestamp += d
			nextEvt.Delay = d
			node.AddEvent(nextEvt)
		}
		node.SetCcaFailed(false)
	case 3: // End of frame transmit event

		// signal Tx Done to sender.
		nextEvt := &types.Event{
			Type:      types.EventTypeRadioTxDone,
			Timestamp: evt.Timestamp,
			Delay:     0,
			Data:      []byte{openthread.OT_ERROR_NONE},
		}
		node.AddEvent(nextEvt)

		// let other radios of Nodes receive the data.
		nextEvt = evt
		nextEvt.Type = types.EventTypeRadioFrameToNode
		//nextEvt.Timestamp += 0
		nextEvt.Delay = 0
		node.AddEvent(nextEvt)

		node.SetTxPhase(0) // reset back
		node.SetCcaFailed(false)
		rm.isRfBusy = false
	}
}

func (rm *RadioModelInterfereAll) HandleEvent(node types.RadioNode, evt *types.Event) {
	switch evt.Type {
	case types.EventTypeRadioFrameToSim:
		rm.TxStart(node, evt)
	case types.EventTypeRadioFrameSimInternal:
		rm.TxOngoing(node, evt)
	default:
		simplelogger.Panicf("event type not implemented: %v", evt.Type)
	}
}

// getFrameDurationUs gets the duration of the frame indicated by evt of type eventTypeRadioFrame*
func (rm *RadioModelInterfereAll) getFrameDurationUs(evt *types.Event) uint64 {
	var n uint64
	n = (uint64)(len(evt.Data) - 1) // PSDU size 5..127
	n += 6                          // add PHY preamble, sfd, PHR bytes
	return n * 8 * 1000000 / 250000
}
