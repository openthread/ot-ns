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
	if src != dst && dst.RadioState == RadioRx {
		dist := src.GetDistanceTo(dst)
		if dist <= src.RadioRange {
			rssi := rm.GetTxRssi(src, dst)
			if rssi >= RssiMin && rssi <= RssiMax && rssi >= dst.RxSensitivity {
				return true
			}
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
	case EventTypeRadioRxDone:
		fallthrough
	case EventTypeRadioComm:
		// compute the RSSI and store it in the event
		evt.RadioCommData.PowerDbm = rm.GetTxRssi(src, dst)
	case EventTypeRadioChannelSample:
		evt.RadioCommData.PowerDbm = src.rssiSampleMax
	}
	return true
}

func (rm *RadioModelIdeal) HandleEvent(node *RadioNode, q EventQueue, evt *Event) {
	switch evt.Type {
	case EventTypeRadioComm:
		rm.txStart(node, q, evt)
	case EventTypeRadioTxDone:
		rm.txStop(node, q, evt)
	case EventTypeRadioChannelSample:
		rm.channelSample(node, q, evt)
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

	srcNode.TxPower = evt.RadioCommData.PowerDbm // get last node's properties from the OT node's event params.
	srcNode.SetChannel(evt.RadioCommData.Channel)

	// dispatch radio event RadioComm 'start of frame Rx' to listening nodes.
	evt2 := evt.Copy()
	evt2.Type = EventTypeRadioComm
	evt2.MustDispatch = true
	q.Add(&evt2)

	// schedule new internal event to call txStop()
	txDoneEvt := evt.Copy()
	txDoneEvt.Type = EventTypeRadioTxDone
	txDoneEvt.Timestamp += evt.RadioCommData.Duration
	q.Add(&txDoneEvt)
}

func (rm *RadioModelIdeal) txStop(node *RadioNode, q EventQueue, evt *Event) {
	simplelogger.AssertTrue(node.RadioState == RadioTx)

	// Dispatch TxDone event back to the source
	txDoneEvt := evt.Copy()
	txDoneEvt.Type = EventTypeRadioTxDone
	txDoneEvt.MustDispatch = true
	q.Add(&txDoneEvt)

	// Create RxDone event, to signal nearby nodes the frame Rx is done.
	rxDoneEvt := evt.Copy()
	rxDoneEvt.Type = EventTypeRadioRxDone
	rxDoneEvt.MustDispatch = true
	q.Add(&rxDoneEvt)
}

func (rm *RadioModelIdeal) channelSample(srcNode *RadioNode, q EventQueue, evt *Event) {
	simplelogger.AssertTrue(srcNode.RadioState == RadioRx)
	srcNode.rssiSampleMax = RssiMinusInfinity // Ideal model never has CCA failure.

	// dispatch event with result back to node, when channel sampling stops.
	sampleDoneEvt := evt.Copy()
	sampleDoneEvt.Type = EventTypeRadioChannelSample
	sampleDoneEvt.Timestamp += evt.RadioCommData.Duration
	sampleDoneEvt.MustDispatch = true
	q.Add(&sampleDoneEvt)
}
