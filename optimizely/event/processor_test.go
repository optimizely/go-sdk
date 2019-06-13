package event

import "testing"


func TestDefaultEventProcessor_ProcessImpression(*testing.T) {
	processor := NewEventProcessor(100, 100)

	impression := Impression{}

	processor.ProcessImpression(impression)

}