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

func (rm *RadioModelIdeal) CheckRadioReachable(evt *Event, src *RadioNode, dst *RadioNode) bool {
	simplelogger.AssertTrue(src != dst)
	dist := src.GetDistanceTo(dst)
	if dist <= src.RadioRange {
		rssi := rm.GetTxRssi(evt, src, dst)
		if rssi >= RssiMin && rssi <= RssiMax && rssi >= dst.RxSensitivity {
			return true
		}
	}
	return false
}

func (rm *RadioModelIdeal) GetTxRssi(evt *Event, srcNode *RadioNode, dstNode *RadioNode) int8 {
	simplelogger.AssertTrue(evt.Type == EventTypeRadioRx)
	rssi := rm.FixedRssi // in the most ideal case, always assume a good RSSI up until the max range.
	if rm.UseVariableRssi {
		rssi = computeIndoorRssi(srcNode.RadioRange, srcNode.GetDistanceTo(dstNode), srcNode.TxPower, dstNode.RxSensitivity)
	}
	return rssi
}

func (rm *RadioModelIdeal) TxStart(node *RadioNode, q EventQueue, evt *Event) {
	simplelogger.AssertTrue(evt.Type == EventTypeRadioTx)
	node.TxPower = evt.TxData.TxPower // get last node's properties from the OT node's event params.
	node.CcaEdThresh = evt.TxData.CcaEdTresh

	frameDuration := rm.FixedFrameDuration
	if rm.UseRealFrameDuration {
		frameDuration = getFrameDurationUs(evt)
	}

	// signal Tx Done event to sender.
	nextEvt := evt.Copy()
	nextEvt.Type = EventTypeRadioTxDone
	nextEvt.Timestamp += frameDuration
	nextEvt.TxDoneData = TxDoneEventData{
		Channel: evt.TxData.Channel,
		Error:   OT_ERROR_NONE,
	}
	q.AddEvent(&nextEvt)
	node.TimeLastTxEnded = evt.Timestamp

	// let other radios of reachable Nodes receive the data (after N us propagation delay)
	nextEvt2 := evt.Copy()
	nextEvt2.Type = EventTypeRadioRx
	nextEvt2.Timestamp += frameDuration
	nextEvt2.RxData = RxEventData{
		Channel: evt.TxData.Channel,
		Error:   OT_ERROR_NONE,
		Rssi:    RssiInvalid,
	}
	q.AddEvent(&nextEvt2)
}

func (rm *RadioModelIdeal) ApplyInterference(evt *Event, src *RadioNode, dst *RadioNode) {
	// No interference modeled.
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

func (rm *RadioModelIdeal) init() {
	// Nothing to init.
}
