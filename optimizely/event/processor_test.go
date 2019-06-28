package event

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)


func TestDefaultEventProcessor_ProcessImpression(t *testing.T) {
	processor := NewEventProcessor(100, 100)

	impression := LogEvent{}

	processor.ProcessImpression(impression)

	result, ok := processor.(*DefaultEventProcessor)

	if ok {
		assert.Equal(t, 1, result.EventsCount())

		time.Sleep(2000 * time.Millisecond)

		assert.NotNil(t, result.Ticker)

		assert.Equal(t, 0, result.EventsCount())
	} else {
		assert.Equal(t, true, false)
	}

}

func TestDefaultEventProcessor_ProcessImpressions(t *testing.T) {
	processor := NewEventProcessor(100, 100)

	impression := LogEvent{}

	processor.ProcessImpression(impression)
	processor.ProcessImpression(impression)

	result, ok := processor.(*DefaultEventProcessor)

	if ok {
		assert.Equal(t, 2, result.EventsCount())

		time.Sleep(2000 * time.Millisecond)

		assert.NotNil(t, result.Ticker)

		assert.Equal(t, 0, result.EventsCount())

	} else {
		assert.Equal(t, true, false)
	}

}