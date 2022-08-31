package radiomodel

import (
	. "github.com/openthread/ot-ns/types"
	"github.com/simonlingoogle/go-simplelogger"
)

// RadioModelIdeal is an ideal radio model with infinite parallel transmission capacity. Frame
// transmit time can be set constant, or to realistic 802.15.4 frame time. RSSI at the receiver
// can be set to an ideal constant RSSI value, or variable based on an average RF propagation model.
type RadioModelIdeal struct {
	Name string
	// UseVariableRssi when true uses distance-dependent RSSI model, else fixed RSSI.
	UseVariableRssi      bool
	FixedRssi            int8
	UseRealFrameDuration bool
	FixedFrameDuration   uint64 // only used if UseRealFramDuration == false
}

func (rm *RadioModelIdeal) GetTxRssi(evt *Event, srcNode *RadioNode, dstNode *RadioNode, distMeters float64) int8 {
	simplelogger.AssertTrue(evt.Type == EventTypeRadioReceived)
	rssi := rm.FixedRssi // in the most ideal case, always assume a good RSSI up until the max range.
	if rm.UseVariableRssi {
		rssi = ComputeIndoorRssi(distMeters, srcNode.TxPower, dstNode.RxSensitivity)
	}
	return rssi
}

func (rm *RadioModelIdeal) TxStart(node *RadioNode, q EventQueue, evt *Event) {
	var nextEvt *Event
	simplelogger.AssertTrue(evt.Type == EventTypeRadioTx)
	node.TxPower = evt.Param1     // get the Tx power from the OT node's event param.
	node.CcaEdThresh = evt.Param2 // get CCA ED threshold also.

	frameDuration := rm.FixedFrameDuration
	if rm.UseRealFrameDuration {
		frameDuration = getFrameDurationUs(evt)
	}

	// signal Tx Done event to sender.
	nextEvt = &Event{
		Type:      EventTypeRadioTxDone,
		Timestamp: evt.Timestamp + frameDuration,
		Param1:    OT_ERROR_NONE,
		NodeId:    evt.NodeId,
	}
	q.AddEvent(nextEvt)

	// let other radios of reachable Nodes receive the data (after N us propagation delay)
	nextEvt = &Event{
		Type:      EventTypeRadioReceived,
		Timestamp: evt.Timestamp + frameDuration,
		Data:      evt.Data,
		NodeId:    evt.NodeId,
	}
	q.AddEvent(nextEvt)
}

func (rm *RadioModelIdeal) HandleEvent(node *RadioNode, q EventQueue, evt *Event) {
	switch evt.Type {
	case EventTypeRadioTx:
		rm.TxStart(node, q, evt)
	default:
		simplelogger.Panicf("event type not implemented: %v", evt.Type)
	}
}

func (rm *RadioModelIdeal) GetName() string {
	return rm.Name
}
