package dispatcher

import (
	"github.com/openthread/ot-ns/openthread"
	"github.com/simonlingoogle/go-simplelogger"
)

const CCA_TIME_US uint64 = 128
const SYMBOL_TIME_US uint64 = 16
const ACK_AIFS_TIME_US uint64 = SYMBOL_TIME_US * 12

type RadioModelInterfereAll struct {
	isRfBusy bool
}

func (rm *RadioModelInterfereAll) IsTxSuccess(node *Node, evt *event) bool {
	return true
}

func (rm *RadioModelInterfereAll) TxStart(node *Node, evt *event) {
	var nextEvt *event

	// check if a transmission is already ongoing? If so return OT_ERROR_ABORT.
	if node.txPhase > 0 {
		// FIXME: submac layer of OpenThread shouldn't start Tx when it's still ongoing.
		nextEvt = &event{
			Type:      eventTypeRadioTxDone,
			Timestamp: evt.Timestamp + 0,
			Delay:     0,
			Data:      []byte{openthread.OT_ERROR_ABORT},
			NodeId:    node.Id,
		}
		node.D.evtQueue.AddEvent(nextEvt)
		return
	}

	node.isCcaFailed = false
	node.txPhase++

	// node starts Tx - perform CCA check at current time point 1.
	if rm.isRfBusy {
		node.isCcaFailed = true
	}

	// re-use the current event as the next event, updating type and timing.
	nextEvt = evt
	nextEvt.Type = eventTypeRadioFrameSimInternal
	nextEvt.Timestamp = node.D.CurTime + CCA_TIME_US
	nextEvt.Delay = CCA_TIME_US
	node.D.evtQueue.AddEvent(nextEvt)
}

func (rm *RadioModelInterfereAll) TxOngoing(node *Node, evt *event) {

	node.txPhase++
	switch node.txPhase {
	case 2: // CCA second sample point and decision
		if rm.isRfBusy {
			node.isCcaFailed = true
		}
		if node.isCcaFailed {
			// if CCA fails, then respond Tx Done with error code.
			nextEvt := &event{
				Type:      eventTypeRadioTxDone,
				Timestamp: evt.Timestamp,
				Delay:     0,
				Data:      []byte{openthread.OT_ERROR_CHANNEL_ACCESS_FAILURE},
				NodeId:    node.Id,
			}
			node.D.evtQueue.AddEvent(nextEvt)
			node.txPhase = 0 // reset back
		} else {
			// CCA was successful, start frame transmission now.
			rm.isRfBusy = true
			// schedule the end-of-frame-transmission event.
			nextEvt := evt
			nextEvt.Type = eventTypeRadioFrameSimInternal
			d := rm.getFrameDurationUs(evt)
			nextEvt.Timestamp += d
			nextEvt.Delay = d
			node.D.evtQueue.AddEvent(nextEvt)
		}
		node.isCcaFailed = false
	case 3: // End of frame transmit event

		// signal Tx Done to sender.
		nextEvt := &event{
			Type:      eventTypeRadioTxDone,
			Timestamp: evt.Timestamp,
			Delay:     0,
			Data:      []byte{openthread.OT_ERROR_NONE},
			NodeId:    node.Id,
		}
		node.D.evtQueue.AddEvent(nextEvt)

		// let other radios of Nodes receive the data.
		nextEvt = evt
		nextEvt.Type = eventTypeRadioFrameToNode
		//nextEvt.Timestamp += 0
		nextEvt.Delay = 0
		node.D.evtQueue.AddEvent(nextEvt)

		node.txPhase = 0 // reset back
		node.isCcaFailed = false
		rm.isRfBusy = false
	}
}

func (rm *RadioModelInterfereAll) HandleEvent(node *Node, evt *event) {
	switch evt.Type {
	case eventTypeRadioFrameToSim:
		rm.TxStart(node, evt)
	case eventTypeRadioFrameSimInternal:
		rm.TxOngoing(node, evt)
	default:
		simplelogger.Panicf("event type not implemented: %v", evt.Type)
	}
}

// getFrameDurationUs gets the duration of the frame indicated by evt of type eventTypeRadioFrame*
func (rm *RadioModelInterfereAll) getFrameDurationUs(evt *event) uint64 {
	var n uint64
	n = (uint64)(len(evt.Data) - 1) // PSDU size 5..127
	n += 6                          // add PHY preamble, sfd, PHR bytes
	return n * 8 * 1000000 / 250000
}
