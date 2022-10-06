package radiomodel

/* bla
import (
	. "github.com/openthread/ot-ns/types"
	"github.com/simonlingoogle/go-simplelogger"
)

// RadioModelMutualInterference is a radio model where a transmission may interfere with another transmission
// ongoing on the same channel, depending on the relative level (Rx energy in dBm) of signals. Also CCA and
// energy scanning are supported.
type RadioModelMutualInterference struct {
	// Configured minimum Signal-to-Interference (SIR) ratio in dB that is required to receive a signal
	// in presence of at least one interfering, other signal.
	MinSirDb DbmValue

	nodes                 map[NodeId]*RadioNode
	activeTransmitters    map[ChannelId]map[NodeId]*RadioNode
	activeChannelSamplers map[ChannelId]map[NodeId]*RadioNode
}

func (rm *RadioModelMutualInterference) AddNode(nodeid NodeId, radioNode *RadioNode) {
	rm.nodes[nodeid] = radioNode
}

func (rm *RadioModelMutualInterference) DeleteNode(nodeid NodeId) {
	delete(rm.nodes, nodeid)
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
		rm.updateChannelSamplingNodes(node, evt) // all channel-sampling nodes detect the new Tx
	case EventTypeRadioTxDone:
		rm.txStop(node, evt)
	case EventTypeRadioChannelSample:
		// take 1st channel sample
		node.rssiSampleMax = rm.getRssiOnChannel(node, evt.ChanData.Channel)
	case EventTypeChannelSampleDone:
		// take final channel sample
		r := rm.getRssiOnChannel(node, evt.ChanData.Channel)
		if r > node.rssiSampleMax {
			node.rssiSampleMax = r
		}
	default:
		break // Unknown events not handled.
	}
}

func (rm *RadioModelMutualInterference) OnEventDispatch(src *RadioNode, dst *RadioNode, evt *Event) bool {
	switch evt.Type {
	case EventTypeRadioCommRx:
		// check if the destination node was listening since the start. If not, don't dispatch.
		if dst.receivingFrom != src.Id {
			return false
		}
		// compute the RSSI and store in the event
		evt.RxData = RxEventData{
			Channel: src.RadioChannel,
			Error:   OT_ERROR_NONE,
			Rssi:    rm.GetTxRssi(src, dst),
		}
		rm.applyInterference(src, dst, evt)
	case EventTypeRadioTxDone:
		// mark transmission of the node as success
		evt.TxDoneData = TxDoneEventData{
			Channel: src.RadioChannel,
			Error:   OT_ERROR_NONE,
		}
	case EventTypeChannelSampleDone:
		evt.ChanDoneData = ChanSampleDoneEventData{
			Channel: src.RadioChannel,
			Rssi:    src.rssiSampleMax,
		}
	default:
		break
	}
	return true
}

func (rm *RadioModelMutualInterference) GetName() string {
	return "MutualInterference"
}

func (rm *RadioModelMutualInterference) init() {
	rm.nodes = map[NodeId]*RadioNode{}
	rm.activeTransmitters = map[ChannelId]map[NodeId]*RadioNode{}
	rm.activeChannelSamplers = map[ChannelId]map[NodeId]*RadioNode{}
	for c := MinChannelNumber; c <= MaxChannelNumber; c++ {
		rm.activeTransmitters[c] = map[NodeId]*RadioNode{}
		rm.activeChannelSamplers[c] = map[NodeId]*RadioNode{}
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
		node.interferedBy[id] = interferingTransmitter
		interferingTransmitter.interferedBy[evt.NodeId] = node
	}

	rm.activeTransmitters[evt.TxData.Channel][evt.NodeId] = node

	// mark other nodes in Rx state, in-range, active-Rx, and on same channel, and not already
	// receiving, as receivers that may get the frame in the end.
	for _, dstNode := range rm.nodes {
		if dstNode.RadioChannel == node.RadioChannel && rm.CheckRadioReachable(node, dstNode) &&
			dstNode.RadioState == RadioRx && dstNode.receivingFrom == 0 {
			dstNode.receivingFrom = node.Id
		}
	}

}

func (rm *RadioModelMutualInterference) txStop(node *RadioNode, evt *Event) {
	_, nodeTransmits := rm.activeTransmitters[evt.TxData.Channel][evt.NodeId]
	simplelogger.AssertTrue(nodeTransmits)
	delete(rm.activeTransmitters[evt.TxData.Channel], evt.NodeId)
}

func (rm *RadioModelMutualInterference) applyInterference(src *RadioNode, dst *RadioNode, evt *Event) {
	// Apply interference.
	for _, interferer := range src.interferedBy {
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

// update sample value for all channel-sampling nodes that may detect the new source src.
func (rm *RadioModelMutualInterference) updateChannelSamplingNodes(src *RadioNode, evt *Event) {
	simplelogger.AssertTrue(evt.Type == EventTypeRadioCommTx)
	for _, samplingNode := range rm.activeChannelSamplers[evt.TxData.Channel] {
		r := rm.GetTxRssi(src, samplingNode)
		if r > samplingNode.rssiSampleMax && r != RssiInvalid {
			samplingNode.rssiSampleMax = r // TODO accurate method of energy combining.
		}
	}
}
*/
