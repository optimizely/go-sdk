package event

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)


func TestDefaultEventProcessor_ProcessImpression(t *testing.T) {
	processor := NewEventProcessor(100, 100)

	impression := Impression{}

	processor.ProcessImpression(impression)

	result, ok := processor.(*DefaultEventProcessor)

	if ok {
		assert.Equal(t, 1, result.eventsCount())

		time.Sleep(1000 * time.Millisecond)

		assert.NotNil(t, result.Ticker)

		assert.Equal(t, 0, result.eventsCount())
	} else {
		assert.Equal(t, true, false)
	}

}

func TestDefaultEventProcessor_ProcessImpressions(t *testing.T) {
	processor := NewEventProcessor(100, 100)

	impression := Impression{}

	processor.ProcessImpression(impression)
	processor.ProcessImpression(impression)

	result, ok := processor.(*DefaultEventProcessor)

	if ok {
		assert.Equal(t, 2, result.eventsCount())

		time.Sleep(1000 * time.Millisecond)

		assert.NotNil(t, result.Ticker)

		assert.Equal(t, 1, result.eventsCount())

		time.Sleep(1000 * time.Millisecond)

		assert.Equal(t, 0, result.eventsCount())

	} else {
		assert.Equal(t, true, false)
	}

}