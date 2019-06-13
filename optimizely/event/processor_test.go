package event

import (
	"github.com/stretchr/testify/assert"
	"testing"
)


func TestDefaultEventProcessor_ProcessImpression(t *testing.T) {
	processor := NewEventProcessor(100, 100)

	impression := Impression{}

	processor.ProcessImpression(impression)

	result, ok := processor.(*DefaultEventProcessor)

	if ok {
		assert.Equal(t, result.eventsCount(),1)
	} else {
		assert.Equal(t, true, false)
	}

}