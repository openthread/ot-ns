package dispatcher

type (
	RadioModel interface {
		IsTxSuccess(node *Node, evt *event) bool

		// HandleEvent handles all internal radio-model events.
		HandleEvent(node *Node, evt *event)
	}
)
