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
	Name            string
	UseVariableRssi bool // when true uses distance-dependent RSSI model, else FixedRssi.
	FixedRssi       DbmValue

	nodes map[NodeId]*RadioNode
}

func (rm *RadioModelIdeal) AddNode(nodeid NodeId, radioNode *RadioNode) {
	rm.nodes[nodeid] = radioNode
}

func (rm *RadioModelIdeal) DeleteNode(nodeid NodeId) {
	delete(rm.nodes, nodeid)
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

func (rm *RadioModelIdeal) OnEventDispatch(src *RadioNode, dst *RadioNode, evt *Event) bool {
	switch evt.Type {
	case EventTypeRadioCommRx:
		// compute the RSSI and store in the event
		evt.RxData = RxEventData{
			Channel: src.RadioChannel,
			Error:   OT_ERROR_NONE,
			Rssi:    rm.GetTxRssi(src, dst),
		}
	case EventTypeRadioTxDone:
		// mark transmission of the node as success in the event.
		simplelogger.AssertTrue(src.RadioState == RadioTx)
		evt.TxDoneData = TxDoneEventData{
			Channel: src.RadioChannel,
			Error:   OT_ERROR_NONE,
		}
	case EventTypeChannelSampleDone:
		// Ideal model always detects 'no channel activity'.
		evt.ChanDoneData = ChanDoneEventData{
			Channel: evt.ChanData.Channel,
			Rssi:    RssiMinusInfinity,
		}
	default:
		break
	}
	return true
}

func (rm *RadioModelIdeal) HandleEvent(node *RadioNode, q EventQueue, evt *Event) {
	switch evt.Type {
	case EventTypeRadioCommTx:
		rm.txStart(node, q, evt)
	case EventTypeRadioTxDone:
		rm.txStop(node, q, evt)
	case EventTypeChannelSample:
		rm.channelSample(node, q, evt)
		node.rssiSampleMax = RssiMinusInfinity
		break
	case EventTypeChannelSampleDone:
		break
	default:
		break
	}
}

func (rm *RadioModelIdeal) GetName() string {
	return rm.Name
}

func (rm *RadioModelIdeal) init() {
	rm.nodes = map[NodeId]*RadioNode{}
}

func (rm *RadioModelIdeal) txStart(srcNode *RadioNode, q EventQueue, evt *Event) {
	simplelogger.AssertTrue(srcNode.RadioState == RadioTx)

	srcNode.TxPower = evt.TxData.TxPower // get last node's properties from the OT node's event params.
	srcNode.SetChannel(evt.TxData.Channel)
	srcNode.SetRadioState(RadioTx)

	// schedule new event for when tx is done, after evt.Delay us.
	txDoneEvt := evt.Copy()
	txDoneEvt.Type = EventTypeRadioTxDone
	txDoneEvt.Timestamp += evt.Delay
	q.Add(&txDoneEvt)
}

func (rm *RadioModelIdeal) txStop(node *RadioNode, q EventQueue, evt *Event) {
	simplelogger.AssertTrue(node.RadioState == RadioTx)

	// Create rx event, to let nearby nodes receive the frame.
	evt2 := evt.Copy()
	evt2.Type = EventTypeRadioCommRx
	q.Add(&evt2)
}

func (rm *RadioModelIdeal) channelSample(srcNode *RadioNode, q EventQueue, evt *Event) {
	srcNode.rssiSampleMax = RssiMinusInfinity

	// schedule event when channel sampling stops.
	evt2 := evt.Copy()
	evt2.Type = EventTypeChannelSampleDone
	evt2.Timestamp += evt.Delay
	q.Add(&evt2)
}
