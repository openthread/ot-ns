package radiomodel

import (
	"github.com/openthread/ot-ns/openthread"
	. "github.com/openthread/ot-ns/types"
	"github.com/simonlingoogle/go-simplelogger"
)

// RadioModelIdeal is an ideal radio model with infinite capacity and always 1 us transmit time.
type RadioModelIdeal struct {
}

func (rm *RadioModelIdeal) IsTxSuccess(node *RadioNode, evt *Event) bool {
	return true
}

func (rm *RadioModelIdeal) TxStart(node *RadioNode, q EventQueue, evt *Event) {
	var nextEvt *Event
	simplelogger.AssertTrue(evt.Type == EventTypeRadioFrameToSim || evt.Type == EventTypeRadioFrameAckToSim)

	// signal Tx Done event to sender.
	nextEvt = &Event{
		Type:      EventTypeRadioTxDone,
		Timestamp: evt.Timestamp + 1,
		Delay:     1,
		Data:      []byte{openthread.OT_ERROR_NONE},
		NodeId:    evt.NodeId,
	}
	q.AddEvent(nextEvt)

	// let other radios of Nodes receive the data (after 1 us propagation delay)
	nextEvt = evt
	nextEvt.Type = EventTypeRadioFrameToNode
	nextEvt.Timestamp += 1
	nextEvt.Delay = 1
	q.AddEvent(nextEvt)

}

func (rm *RadioModelIdeal) HandleEvent(node *RadioNode, q EventQueue, evt *Event) {
	switch evt.Type {
	case EventTypeRadioFrameAckToSim:
	case EventTypeRadioFrameToSim:
		rm.TxStart(node, q, evt)
	default:
		simplelogger.Panicf("event type not implemented: %v", evt.Type)
	}
}
