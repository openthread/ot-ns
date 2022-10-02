package radiomodel

import (
	. "github.com/openthread/ot-ns/types"
	"github.com/simonlingoogle/go-simplelogger"
)

// RadioModelIdeal is an ideal radio model with infinite parallel transmission capacity per
// channel. RSSI at the receiver can be set to an ideal constant RSSI value, or to a value
// based on an average RF propagation model. There is a hard stop of reception beyond the
// radioRange of the node i.e. ideal disc model.
type RadioModelIdeal struct {
	Name string
	// UseVariableRssi when true uses distance-dependent RSSI model, else fixed RSSI.
	UseVariableRssi bool
	FixedRssi       int8
}

func (rm *RadioModelIdeal) CheckRadioReachable(src *RadioNode, dst *RadioNode) bool {
	simplelogger.AssertTrue(src != dst)
	dist := src.GetDistanceTo(dst)
	if dist <= src.RadioRange {
		rssi := rm.GetTxRssi(src, dst)
		if rssi >= RssiMin && rssi <= RssiMax && rssi >= dst.RxSensitivity {
			return true
		}
	}
	return false
}

func (rm *RadioModelIdeal) GetTxRssi(srcNode *RadioNode, dstNode *RadioNode) DbmValue {
	rssi := rm.FixedRssi // in the most ideal case, always assume a good RSSI up until the max range.
	if rm.UseVariableRssi {
		rssi = computeIndoorRssi(srcNode.RadioRange, srcNode.GetDistanceTo(dstNode), srcNode.TxPower, dstNode.RxSensitivity)
	}
	return rssi
}

func (rm *RadioModelIdeal) OnEventDispatch(evt *Event, src *RadioNode, dst *RadioNode) {
	switch evt.Type {
	case EventTypeRadioCommRx:
		// compute the RSSI and store in the event
		evt.RxData = RxEventData{
			Channel: src.RadioChannel,
			Error:   OT_ERROR_NONE,
			Rssi:    rm.GetTxRssi(src, dst),
		}
	case EventTypeRadioTxDone:
		// mark transmission of the node as success
		simplelogger.AssertTrue(src.RadioState == RadioTx)
		evt.TxDoneData = TxDoneEventData{
			Channel: src.RadioChannel,
			Error:   OT_ERROR_NONE,
		}
	case EventTypeChannelActivityDone:
		// Ideal model always detects 'no channel activity'.
		evt.ChanDoneData = ChanDoneEventData{
			Channel: evt.ChanData.Channel,
			Rssi:    RssiMinusInfinity,
		}
	default:
		simplelogger.Panicf("Unexpected event type: %v", evt.Type)
	}
}

func (rm *RadioModelIdeal) HandleEvent(node *RadioNode, q EventQueue, evt *Event) {
	switch evt.Type {
	case EventTypeRadioCommTx:
		rm.txStart(node, evt)
	case EventTypeRadioTxDone:
		rm.txStop(node, evt)
	case EventTypeChannelActivity:
		break
	case EventTypeChannelActivityDone:
		break
	default:
		simplelogger.Errorf("Radiomodel event type not implemented: %v", evt.Type)
	}
}

func (rm *RadioModelIdeal) GetName() string {
	return rm.Name
}

func (rm *RadioModelIdeal) init() {
	// Nothing to init.
}

func (rm *RadioModelIdeal) txStart(node *RadioNode, evt *Event) {
	simplelogger.AssertTrue(node.RadioState == RadioTx)
	simplelogger.AssertFalse(node.RadioLockState)
	node.LockRadioState(true)

	node.TxPower = evt.TxData.TxPower // get last node's properties from the OT node's event params.
	node.CcaEdThresh = evt.TxData.CcaEdTresh
}

func (rm *RadioModelIdeal) txStop(node *RadioNode, evt *Event) {
	simplelogger.AssertTrue(node.RadioState == RadioTx)
	simplelogger.AssertTrue(node.RadioLockState)
	node.LockRadioState(false)
}
