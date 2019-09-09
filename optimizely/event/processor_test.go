package event

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDefaultEventProcessor_ProcessImpression(t *testing.T) {
	ctx := context.Background()

	processor := NewEventProcessor(ctx, 10, 100, 100)

	impression := BuildTestImpressionEvent()

	processor.ProcessEvent(impression)

	assert.Equal(t, 1, processor.EventsCount())

	time.Sleep(2000 * time.Millisecond)

	assert.NotNil(t, processor.Ticker)

	assert.Equal(t, 0, processor.EventsCount())
}

type MockDispatcher struct {
	Events []LogEvent
}

func (f *MockDispatcher) DispatchEvent(event LogEvent) (bool, error) {
	f.Events = append(f.Events, event)
	return true, nil
}

func TestDefaultEventProcessor_ProcessBatch(t *testing.T) {
	processor := &QueueingEventProcessor{
		MaxQueueSize:    100,
		FlushInterval:   100,
		Q:               NewInMemoryQueue(100),
		EventDispatcher: &MockDispatcher{},
	}
	processor.BatchSize = 10
	processor.StartTicker(context.TODO())

	impression := BuildTestImpressionEvent()
	conversion := BuildTestConversionEvent()

	processor.ProcessEvent(impression)
	processor.ProcessEvent(impression)
	processor.ProcessEvent(conversion)
	processor.ProcessEvent(conversion)

	assert.Equal(t, 4, processor.EventsCount())

	time.Sleep(200 * time.Millisecond)

	assert.NotNil(t, processor.Ticker)

	assert.Equal(t, 0, processor.EventsCount())

	result, ok := (processor.EventDispatcher).(*MockDispatcher)

	if ok {
		assert.Equal(t, 1, len(result.Events))
		evs := result.Events[0]
		assert.Equal(t, 4, len(evs.Event.Visitors))
	}
}

func TestBatchEventProcessor_FlushesOnClose(t *testing.T) {
	ctx, cancelFn := context.WithCancel(context.Background())
	processor := &QueueingEventProcessor{
		MaxQueueSize:    100,
		FlushInterval:   30 * time.Second,
		Q:               NewInMemoryQueue(100),
		EventDispatcher: &MockDispatcher{},
	}
	processor.BatchSize = 10
	processor.StartTicker(ctx)

	impression := BuildTestImpressionEvent()
	conversion := BuildTestConversionEvent()

	processor.ProcessEvent(impression)
	processor.ProcessEvent(impression)
	processor.ProcessEvent(conversion)
	processor.ProcessEvent(conversion)

	assert.Equal(t, 4, processor.EventsCount())

	time.Sleep(500 * time.Millisecond)

	// Triggers the flush in the processor
	cancelFn()

	time.Sleep(500 * time.Millisecond)

	assert.Equal(t, 0, processor.EventsCount())
}

func TestDefaultEventProcessor_ProcessBatchRevisionMismatch(t *testing.T) {
	processor := &QueueingEventProcessor{
		MaxQueueSize:    100,
		FlushInterval:   100,
		Q:               NewInMemoryQueue(100),
		EventDispatcher: &MockDispatcher{},
	}
	processor.BatchSize = 10
	processor.StartTicker(context.TODO())

	impression := BuildTestImpressionEvent()
	conversion := BuildTestConversionEvent()

	processor.ProcessEvent(impression)
	impression.EventContext.Revision = "12112121"
	processor.ProcessEvent(impression)
	processor.ProcessEvent(conversion)
	processor.ProcessEvent(conversion)

	assert.Equal(t, 4, processor.EventsCount())

	time.Sleep(200 * time.Millisecond)

	assert.NotNil(t, processor.Ticker)

	assert.Equal(t, 0, processor.EventsCount())

	result, ok := (processor.EventDispatcher).(*MockDispatcher)

	if ok {
		assert.Equal(t, 3, len(result.Events))
		evs := result.Events[len(result.Events)-1]
		assert.Equal(t, 2, len(evs.Event.Visitors))
	}
}

func TestDefaultEventProcessor_ProcessBatchProjectMismatch(t *testing.T) {
	processor := &QueueingEventProcessor{
		MaxQueueSize:    100,
		FlushInterval:   100,
		Q:               NewInMemoryQueue(100),
		EventDispatcher: &MockDispatcher{},
	}
	processor.BatchSize = 10
	processor.StartTicker(context.TODO())

	impression := BuildTestImpressionEvent()
	conversion := BuildTestConversionEvent()

	processor.ProcessEvent(impression)
	impression.EventContext.ProjectID = "121121211111"
	processor.ProcessEvent(impression)
	processor.ProcessEvent(conversion)
	processor.ProcessEvent(conversion)

	assert.Equal(t, 4, processor.EventsCount())

	time.Sleep(200 * time.Millisecond)

	assert.NotNil(t, processor.Ticker)

	assert.Equal(t, 0, processor.EventsCount())

	result, ok := (processor.EventDispatcher).(*MockDispatcher)

	if ok {
		assert.Equal(t, 3, len(result.Events))
		evs := result.Events[len(result.Events)-1]
		assert.Equal(t, 2, len(evs.Event.Visitors))
	}
}

func TestChanQueueEventProcessor_ProcessImpression(t *testing.T) {
	processor:= &QueueingEventProcessor{
		MaxQueueSize:    100,
		FlushInterval:   100,
		Q:               NewChanQueue(100),
		EventDispatcher: &HTTPEventDispatcher{},
	}
	processor.BatchSize = 10
	processor.StartTicker(context.TODO())

	impression := BuildTestImpressionEvent()

	processor.ProcessEvent(impression)
	processor.ProcessEvent(impression)
	processor.ProcessEvent(impression)

	time.Sleep(3000 * time.Millisecond)

	assert.NotNil(t, processor.Ticker)

	assert.Equal(t, 0, processor.EventsCount())
}

func TestChanQueueEventProcessor_ProcessBatch(t *testing.T) {
	processor := &QueueingEventProcessor{MaxQueueSize: 100, FlushInterval: 100, Q: NewChanQueue(100), EventDispatcher: &MockDispatcher{}}
	processor.BatchSize = 10
	processor.StartTicker(context.TODO())

	impression := BuildTestImpressionEvent()
	conversion := BuildTestConversionEvent()

	processor.ProcessEvent(impression)
	processor.ProcessEvent(impression)
	processor.ProcessEvent(conversion)
	processor.ProcessEvent(conversion)

	time.Sleep(3000 * time.Millisecond)

	assert.NotNil(t, processor.Ticker)

	assert.Equal(t, 0, processor.EventsCount())

	time.Sleep(3000 * time.Millisecond)

	result, ok := (processor.EventDispatcher).(*MockDispatcher)

	if ok {
		assert.Equal(t, 1, len(result.Events))
		evs := result.Events[0]
		assert.True(t, len(evs.Event.Visitors) >= 1)
	}
}
