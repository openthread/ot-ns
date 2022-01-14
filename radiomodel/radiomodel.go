package radiomodel

import . "github.com/openthread/ot-ns/types"

// EventQueue is the abstraction of the queue where the radio model sends its outgoing (new) events to.
type EventQueue interface {
	AddEvent(*Event)
}

// RadioModel provides access to any type of radio model.
type RadioModel interface {
	// IsTxSuccess checks whether the radio frame Tx indicated by evt is successful or not, according to the radio model
	IsTxSuccess(node *RadioNode, evt *Event) bool

	// HandleEvent handles all radio-model events coming out of the simulator event queue.
	// node must be the RadioNode object equivalent to the evt.NodeId node. Newly generated events may go back into
	// the EventQueue q.
	HandleEvent(node *RadioNode, q EventQueue, evt *Event)
}
