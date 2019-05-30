package event

// Processor processes events
type Processor interface {
	processImpressionEvent(event Impression)
}

// DefaultEventProcessor is used out of the box by the SDK
type DefaultEventProcessor struct {
}

// ProcessImpression processes the given impression event
func (*DefaultEventProcessor) ProcessImpression(event Impression) {
	// TODO(mng): do something with the impression event
}
