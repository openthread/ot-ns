package radiomodel

import (
	"github.com/openthread/ot-ns/openthread"
	. "github.com/openthread/ot-ns/types"
	"github.com/simonlingoogle/go-simplelogger"
)

const frameTransmitTimeUs uint64 = 1 // the fixed frame transmit time.

// RadioModelIdeal is an ideal radio model with infinite capacity and always const transmit time.
type RadioModelIdeal struct {
	//
}

func (rm *RadioModelIdeal) GetTxRssi(evt *Event, srcNode *RadioNode, dstNode *RadioNode, distMeters float64) int8 {
	simplelogger.AssertTrue(evt.Type == EventTypeRadioReceived)
	rssi := ComputeIndoorRssi(distMeters, srcNode.TxPower, dstNode.RxSensitivity)
	return rssi
}

func (rm *RadioModelIdeal) TxStart(node *RadioNode, q EventQueue, evt *Event) {
	var nextEvt *Event
	simplelogger.AssertTrue(evt.Type == EventTypeRadioTx || evt.Type == EventTypeRadioTxAck)
	if evt.Version > 1 {
		node.TxPower = evt.Param1     // get the Tx power from the OT node's event param.
		node.CcaEdThresh = evt.Param2 // get CCA ED threshold also.
	}
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
	nextEvt.Type = EventTypeRadioReceived
	nextEvt.Timestamp += frameTransmitTimeUs
	nextEvt.Delay = frameTransmitTimeUs
	q.AddEvent(nextEvt)

}

func (rm *RadioModelIdeal) HandleEvent(node *RadioNode, q EventQueue, evt *Event) {
	switch evt.Type {
	case EventTypeRadioTxAck:
		fallthrough
	case EventTypeRadioTx:
		rm.TxStart(node, q, evt)
	default:
		simplelogger.Panicf("event type not implemented: %v", evt.Type)
	}
}
