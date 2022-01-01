package dispatcher

import (
	"github.com/openthread/ot-ns/openthread"
	"github.com/simonlingoogle/go-simplelogger"
)

const CCA_TIME_US uint64 = 128

type RadioModelInterfereAll struct {
	isRfBusy bool
}

func (rm *RadioModelInterfereAll) IsTxSuccess(node *Node, evt *event) bool {
	return true
}

func (rm *RadioModelInterfereAll) TxStart(node *Node, evt *event) {
	var nextEvt *event

	// check if a transmission is already ongoing? If so return busy error.
	if node.txPhase > 0 {
		nextEvt := &event{
			Type:      eventTypeRadioTxDone,
			Timestamp: evt.Timestamp + 1,
			Data:      []byte{openthread.OT_ERROR_BUSY},
			NodeId:    node.Id,
		}
		node.D.evtQueue.AddEvent(nextEvt)
		return
	}

	node.isCcaFailed = false
	node.txPhase++

	// node starts Tx - perform CCA check at current time point 1.
	if rm.isRfBusy {
		node.isCcaFailed = true
	}

	// re-use the current event as the next event, updating type and timing.
	nextEvt = evt
	nextEvt.Type = eventTypeRadioFrameSimInternal
	nextEvt.Timestamp = node.D.CurTime + CCA_TIME_US
	node.D.evtQueue.AddEvent(nextEvt)
}

func (rm *RadioModelInterfereAll) TxOngoing(node *Node, evt *event) {

	node.txPhase++
	switch node.txPhase {
	case 2: // CCA second sample point and decision
		if rm.isRfBusy {
			node.isCcaFailed = true
		}
		if node.isCcaFailed {
			// if CCA fails, then respond Tx Done with error code.
			nextEvt := &event{
				Type:      eventTypeRadioTxDone,
				Timestamp: evt.Timestamp + 1, // TODO check if +0 could work here
				Data:      []byte{openthread.OT_ERROR_CHANNEL_ACCESS_FAILURE},
				NodeId:    node.Id,
			}
			node.D.evtQueue.AddEvent(nextEvt)
			node.txPhase = 0 // reset back
		} else {
			// CCA was successful, start frame transmission now.
			rm.isRfBusy = true
			// schedule the end-of-frame-transmission event.
			nextEvt := evt
			nextEvt.Type = eventTypeRadioFrameSimInternal
			nextEvt.Timestamp += rm.getFrameDurationUs(evt)
			node.D.evtQueue.AddEvent(nextEvt)
		}
		node.isCcaFailed = false
	case 3: // End of frame transmit event

		// signal Tx Done to sender.
		nextEvt := &event{
			Type:      eventTypeRadioTxDone,
			Timestamp: evt.Timestamp + 1, // TODO check if +0 could work here.
			Data:      []byte{openthread.OT_ERROR_NONE},
			NodeId:    node.Id,
		}
		node.D.evtQueue.AddEvent(nextEvt)

		// let other radios of Nodes receive the data.
		nextEvt = evt
		nextEvt.Type = eventTypeRadioFrameToNode
		nextEvt.Timestamp += 1 // TODO check if +0 could work here.
		node.D.evtQueue.AddEvent(nextEvt)

		node.txPhase = 0 // reset back
		node.isCcaFailed = false
		rm.isRfBusy = false
	}
}

func (rm *RadioModelInterfereAll) HandleEvent(node *Node, evt *event) {
	switch evt.Type {
	case eventTypeRadioFrameToSim:
		rm.TxStart(node, evt)
	case eventTypeRadioFrameSimInternal:
		rm.TxOngoing(node, evt)
	default:
		simplelogger.Panicf("event type not implemented: %v", evt.Type)
	}
}

func (rm *RadioModelInterfereAll) getFrameDurationUs(evt *event) uint64 {
	var n uint64
	n = (uint64)(len(evt.Data) - 1) // PSDU size 5..127
	n += 6                          // add PHY preamble, sfd, PHR bytes
	return n * 8 * 1000000 / 250000
}

/*
func temp {
	d.Counters.RadioEvents += 1
	// perform CCA sampling nr 1
	if d.radioModel.isChannelClear(node) {
		// set up event to sample again at end of CCA window
		evtCca := &event{
			Timestamp: d.CurTime + 128, // TODO move 128us to constants - CCA time
			Type:      eventTypeRadioTxDone,
		}
		d.sendQueue.AddEvent(evtCca)

	} else {
		evtCca := &event{
			Timestamp: d.CurTime + 128, // TODO move 128us to constants - CCA time
			Type:      eventTypeRadioTxDone,
			Data:      []byte{openthread.OT_ERROR_CHANNEL_ACCESS_FAILURE},
		}
		d.sendOneRadioEvent(evtCca, node, node)
	}
	// Tx frame from node triggers Rx-start in other nodes near instantly
	// FIXME TODO send to radio model.
	// here it triggers CCA period of ...
	// call radiomodel.
	// if clear CCA, then reserve the medium and put a 'tx done' event in my own queue (to put in tx-done
	// status here, and then send onward to the Tx-node.
	// also schedule rx-start event right after CCA. That goes to receiving nodes.
	// after packet duration, tx-done is sent and also radio-frame-event to rx nodes. And medium is cleared
	// again.
	d.sendQueue.Add(d.CurTime+1, nodeid, evt.Data)
	// Tx frame (and Rx-frame in other nodes) is done after some time
	//xyz
}

func temp2 {
case eventTypeRadioTxDone:
// 2nd sample point of cca.
if d.radioModel.isChannelClear(node) {
// yes, 2nd point and 1st point are clear - start tx of frame.
//FIXME get linked frame event and also start rx event to all nodes.
evtCca := &event{
Timestamp: d.CurTime + 128, // TODO move 128us to constants - CCA time
Type:      eventTypeRadioTxDone,
Data:      []byte{openthread.OT_ERROR_CHANNEL_ACCESS_FAILURE},
}
d.sendOneRadioFrameEvent(evtCca, node, node)
} else {
evtCca := &event{
Timestamp: d.CurTime + 128, // TODO move 128us to constants - CCA time
Type:      eventTypeRadioTxDone,
Data:      []byte{openthread.OT_ERROR_CHANNEL_ACCESS_FAILURE},
}
d.sendOneRadioFrameEvent(evtCca, node, node)
}

}
*/
