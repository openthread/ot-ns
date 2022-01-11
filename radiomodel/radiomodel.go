package radiomodel

import "github.com/openthread/ot-ns/types"

type RadioModel interface {
	IsTxSuccess(node types.RadioNode, evt *types.Event) bool

	// HandleEvent handles all internal radio-model events.
	HandleEvent(node types.RadioNode, evt *types.Event)
}
