package radiomodel

import (
	. "github.com/openthread/ot-ns/types"
	"github.com/simonlingoogle/go-simplelogger"
)

// RadioModelMutualInterference is a radio model where a transmission may interfere with another transmission
// ongoing on the same channel, depending on the relative level (Rx energy in dBm) of signals. Also CCA is
// supported for nodes using the CCA ED threshold as set by the node.
type RadioModelMutualInterference struct {
	// Configured minimum Signal-to-Interference (SIR) ratio in dB that is required to receive a signal
	// in presence of at least one interfering, other signal.
	MinSirDb DbmValue

	activeTransmitters map[ChannelId]map[NodeId]*RadioNode
}

func (rm *RadioModelMutualInterference) CheckRadioReachable(src *RadioNode, dst *RadioNode) bool {
	rssi := rm.GetTxRssi(src, dst)
	return rssi >= RssiMin && rssi <= RssiMax && rssi >= dst.RxSensitivity
}

func (rm *RadioModelMutualInterference) GetTxRssi(srcNode *RadioNode, dstNode *RadioNode) DbmValue {
	simplelogger.AssertTrue(srcNode != dstNode)
	rssi := computeIndoorRssi(srcNode.RadioRange, srcNode.GetDistanceTo(dstNode), srcNode.TxPower, dstNode.RxSensitivity)
	return rssi
}

func (rm *RadioModelMutualInterference) HandleEvent(node *RadioNode, q EventQueue, evt *Event) {
	switch evt.Type {
	case EventTypeRadioCommTx:
		rm.txStart(node, evt)
	case EventTypeRadioTxDone:
		rm.txStop(node, evt)
	case EventTypeChannelActivity:
		// take 1st channel sample
		node.rssiSampleMax = rm.getRssiOnChannel(node, evt.ChanData.Channel)
	case EventTypeChannelActivityDone:
		// take 2nd channel sample
		r := rm.getRssiOnChannel(node, evt.ChanData.Channel)
		if r > node.rssiSampleMax {
			node.rssiSampleMax = r
		}
	default:
		simplelogger.Errorf("Radiomodel event type not implemented: %v", evt.Type)
	}
}

func (rm *RadioModelMutualInterference) OnEventDispatch(evt *Event, src *RadioNode, dst *RadioNode) {
	switch evt.Type {
	case EventTypeRadioCommRx:
		// compute the RSSI and store in the event
		evt.RxData = RxEventData{
			Channel: src.RadioChannel,
			Error:   OT_ERROR_NONE,
			Rssi:    rm.GetTxRssi(src, dst),
		}
		rm.applyInterference(evt, src, dst)
	case EventTypeRadioTxDone:
		// mark transmission of the node as success
		simplelogger.AssertTrue(src.RadioState == RadioTx)
		evt.TxDoneData = TxDoneEventData{
			Channel: src.RadioChannel,
			Error:   OT_ERROR_NONE,
		}
	case EventTypeChannelActivityDone:
		evt.ChanDoneData = ChanDoneEventData{
			Channel: evt.ChanData.Channel,
			Rssi:    src.rssiSampleMax,
		}
	default:
		simplelogger.Panicf("Unexpected event type: %v", evt.Type)
	}
}

func (rm *RadioModelMutualInterference) GetName() string {
	return "MutualInterference"
}

func (rm *RadioModelMutualInterference) init() {
	rm.activeTransmitters = map[ChannelId]map[NodeId]*RadioNode{}
	for c := MinChannelNumber; c <= MaxChannelNumber; c++ {
		rm.activeTransmitters[c] = map[NodeId]*RadioNode{}
	}
}

func (rm *RadioModelMutualInterference) getRssiOnChannel(node *RadioNode, channel uint8) int8 {
	rssiMax := RssiMinusInfinity
	// loop all active transmitters
	for _, v := range rm.activeTransmitters[channel] {
		rssi := rm.GetTxRssi(v, node)
		if rssi == RssiInvalid {
			continue
		}
		if rssi > rssiMax {
			rssiMax = rssi // TODO combine signal energies in more realistic way.
		}
	}
	return rssiMax
}

func (rm *RadioModelMutualInterference) txStart(node *RadioNode, evt *Event) {
	_, nodeTransmits := rm.activeTransmitters[evt.TxData.Channel][evt.NodeId]
	simplelogger.AssertFalse(nodeTransmits)

	// mark what this new transmission will interfere with.
	for id, interferingTransmitter := range rm.activeTransmitters[evt.TxData.Channel] {
		node.InterferedBy[id] = interferingTransmitter
		interferingTransmitter.InterferedBy[evt.NodeId] = node
	}

	rm.activeTransmitters[evt.TxData.Channel][evt.NodeId] = node
}

func (rm *RadioModelMutualInterference) txStop(node *RadioNode, evt *Event) {
	_, nodeTransmits := rm.activeTransmitters[evt.TxData.Channel][evt.NodeId]
	simplelogger.AssertTrue(nodeTransmits)
	delete(rm.activeTransmitters[evt.TxData.Channel], evt.NodeId)

	node.TimeLastTxEnded = evt.Timestamp
}

func (rm *RadioModelMutualInterference) applyInterference(evt *Event, src *RadioNode, dst *RadioNode) {
	// Apply interference.
	for _, interferer := range src.InterferedBy {
		if interferer == dst { // if dst node was at some point transmitting itself, fail the Rx
			evt.RxData.Error = OT_ERROR_ABORT
			return
		}
		rssiInterferer := int(rm.GetTxRssi(interferer, dst))
		rssi := int(evt.RxData.Rssi)
		sirDb := rssi - rssiInterferer
		if sirDb < int(rm.MinSirDb) {
			// interfering signal gets too close to the transmission rssi, impacts the signal.
			evt.Data = interferePsduData(evt.Data, float64(sirDb))
			evt.RxData.Error = OT_ERROR_FCS
		}
	}
}
