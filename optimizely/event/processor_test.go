package event

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)


func TestDefaultEventProcessor_ProcessImpression(t *testing.T) {
	processor := NewEventProcessor(100, 100)

	impression := BuildTestImpressionEvent()

	processor.ProcessEvent(impression)

	result, ok := processor.(*QueueingEventProcessor)

	if ok {
		assert.Equal(t, 1, result.EventsCount())

		time.Sleep(2000 * time.Millisecond)

		assert.NotNil(t, result.Ticker)

		assert.Equal(t, 0, result.EventsCount())
	} else {
		assert.Equal(t, true, false)
	}

}

func TestNSQEventProcessor_ProcessImpression(t *testing.T) {
	processor := NewEventProcessorNSQ(100, 100)

	impression := BuildTestImpressionEvent()

	processor.ProcessEvent(impression)
	processor.ProcessEvent(impression)
	processor.ProcessEvent(impression)

	result, ok := processor.(*QueueingEventProcessor)

	if ok {
		time.Sleep(3000 * time.Millisecond)

		assert.NotNil(t, result.Ticker)

		assert.Equal(t, 0, result.EventsCount())
	} else {
		assert.Equal(t, true, false)
	}

}

type MockDispatcher struct {
	Events []LogEvent
}

func (f *MockDispatcher)DispatchEvent(event LogEvent, callback func(success bool)) {
	f.Events = append(f.Events, event)
	callback(true)
}

func TestNSQEventProcessor_ProcessBatch(t *testing.T) {
	processor := &QueueingEventProcessor{MaxQueueSize: 100, FlushInterval: 100, Q: NewNSQueue(100), EventDispatcher: &MockDispatcher{}}
	processor.BatchSize = 10
	processor.StartTicker()

	impression := BuildTestImpressionEvent()
	conversion := BuildTestConversionEvent()

	processor.ProcessEvent(impression)
	processor.ProcessEvent(impression)
	processor.ProcessEvent(conversion)
	processor.ProcessEvent(conversion)

	time.Sleep(3000 * time.Millisecond)

	assert.NotNil(t, processor.Ticker)

	assert.Equal(t, 0, processor.EventsCount())

	result, ok := (processor.EventDispatcher).(*MockDispatcher)

	if ok {
		assert.Equal(t, 1, len(result.Events))
		evs := result.Events[0]
		assert.Equal(t, 4, len(evs.event.Visitors))
	}
}

func TestDefaultEventProcessor_ProcessBatch(t *testing.T) {
	processor := &QueueingEventProcessor{MaxQueueSize: 100, FlushInterval: 100, Q: NewInMemoryQueue(100), EventDispatcher: &MockDispatcher{}}
	processor.BatchSize = 10
	processor.StartTicker()

	impression := BuildTestImpressionEvent()
	conversion := BuildTestConversionEvent()

	processor.ProcessEvent(impression)
	processor.ProcessEvent(impression)
	processor.ProcessEvent(conversion)
	processor.ProcessEvent(conversion)

	assert.Equal(t, 4, processor.EventsCount())

	time.Sleep(2000 * time.Millisecond)

	assert.NotNil(t, processor.Ticker)

	assert.Equal(t, 0, processor.EventsCount())

	result, ok := (processor.EventDispatcher).(*MockDispatcher)

	if ok {
		assert.Equal(t, 1, len(result.Events))
		evs := result.Events[0]
		assert.Equal(t, 4, len(evs.event.Visitors))
	}
}

func TestDefaultEventProcessor_ProcessBatchRevisionMismatch(t *testing.T) {
	processor := &QueueingEventProcessor{MaxQueueSize: 100, FlushInterval: 100, Q: NewInMemoryQueue(100), EventDispatcher: &MockDispatcher{}}
	processor.BatchSize = 10
	processor.StartTicker()

	impression := BuildTestImpressionEvent()
	conversion := BuildTestConversionEvent()

	processor.ProcessEvent(impression)
	impression.EventContext.Revision = "12112121"
	processor.ProcessEvent(impression)
	processor.ProcessEvent(conversion)
	processor.ProcessEvent(conversion)

	assert.Equal(t, 4, processor.EventsCount())

	time.Sleep(2000 * time.Millisecond)

	assert.NotNil(t, processor.Ticker)

	assert.Equal(t, 0, processor.EventsCount())

	result, ok := (processor.EventDispatcher).(*MockDispatcher)

	if ok {
		assert.Equal(t, 3, len(result.Events))
		evs := result.Events[len(result.Events)-1]
		assert.Equal(t, 2, len(evs.event.Visitors))
	}
}

func TestDefaultEventProcessor_ProcessBatchProjectMismatch(t *testing.T) {
	processor := &QueueingEventProcessor{MaxQueueSize: 100, FlushInterval: 100, Q: NewInMemoryQueue(100), EventDispatcher: &MockDispatcher{}}
	processor.BatchSize = 10
	processor.StartTicker()

	impression := BuildTestImpressionEvent()
	conversion := BuildTestConversionEvent()

	processor.ProcessEvent(impression)
	impression.EventContext.ProjectID = "121121211111"
	processor.ProcessEvent(impression)
	processor.ProcessEvent(conversion)
	processor.ProcessEvent(conversion)

	assert.Equal(t, 4, processor.EventsCount())

	time.Sleep(2000 * time.Millisecond)

	assert.NotNil(t, processor.Ticker)

	assert.Equal(t, 0, processor.EventsCount())

	result, ok := (processor.EventDispatcher).(*MockDispatcher)

	if ok {
		assert.Equal(t, 3, len(result.Events))
		evs := result.Events[len(result.Events)-1]
		assert.Equal(t, 2, len(evs.event.Visitors))
	}
}