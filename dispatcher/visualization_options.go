package dispatcher

type VisualizationOptions struct {
	BroadcastMessage bool
	UnicastMessage   bool
	AckMessage       bool
	RouterTable      bool
	ChildTable       bool
}

func defaultVisualizationOptions() VisualizationOptions {
	return VisualizationOptions{
		BroadcastMessage: true,
		UnicastMessage:   true,
		AckMessage:       true,
		RouterTable:      true,
		ChildTable:       true,
	}
}
