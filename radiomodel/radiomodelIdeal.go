package radiomodel

import (
	"github.com/openthread/ot-ns/openthread"
	. "github.com/openthread/ot-ns/types"
	"github.com/simonlingoogle/go-simplelogger"
)

const frameTransmitTimeUs uint64 = 1 // the fixed frame transmit time.

// RadioModelIdeal is an ideal radio model with infinite capacity and always fixed N us transmit time.
type RadioModelIdeal struct {
	//
}

func (rm *RadioModelIdeal) IsTxSuccess(evt *Event, srcNode *RadioNode, dstNode *RadioNode, distMeters float64) int8 {
	simplelogger.AssertTrue(evt.Type == EventTypeRadioFrameToNode)
	rssi := ComputeFsplRssi(distMeters, evt.Param)
	return rssi
}

func (rm *RadioModelIdeal) TxStart(node *RadioNode, q EventQueue, evt *Event) {
	var nextEvt *Event
	simplelogger.AssertTrue(evt.Type == EventTypeRadioFrameToSim || evt.Type == EventTypeRadioFrameAckToSim)
	node.TxPower = evt.Param // get the Tx power from the OT node's event param.

	// signal Tx Done event to sender.
	nextEvt = &Event{
		Type:      EventTypeRadioTxDone,
		Timestamp: evt.Timestamp + frameTransmitTimeUs,
		Delay:     frameTransmitTimeUs,
		Data:      []byte{openthread.OT_ERROR_NONE},
		NodeId:    evt.NodeId,
	}
	q.AddEvent(nextEvt)

	// let other radios of reachable Nodes receive the data (after N us propagation delay)
	nextEvt = evt
	nextEvt.Type = EventTypeRadioFrameToNode
	nextEvt.Timestamp += frameTransmitTimeUs
	nextEvt.Delay = frameTransmitTimeUs
	q.AddEvent(nextEvt)

}

func (rm *RadioModelIdeal) HandleEvent(node *RadioNode, q EventQueue, evt *Event) {
	switch evt.Type {
	case EventTypeRadioFrameAckToSim:
		fallthrough
	case EventTypeRadioFrameToSim:
		rm.TxStart(node, q, evt)
	default:
		simplelogger.Panicf("event type not implemented: %v", evt.Type)
	}
}
