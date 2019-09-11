package event

import (
	"context"
	"github.com/optimizely/go-sdk/optimizely/utils"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func close(wg *sync.WaitGroup, cancelFn context.CancelFunc) {
	cancelFn()
	wg.Wait()
}
func TestDefaultEventProcessor_ProcessImpression(t *testing.T) {
	exeCtx := utils.NewCancelableExecutionCtxExecutionCtx()
	processor := NewEventProcessor(exeCtx, 10, 100, 100)

	impression := BuildTestImpressionEvent()

	processor.ProcessEvent(impression)

	assert.Equal(t, 1, processor.EventsCount())

	exeCtx.TerminateAndWait()

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
	exeCtx := utils.NewCancelableExecutionCtxExecutionCtx()
	processor := &QueueingEventProcessor{
		MaxQueueSize:    100,
		FlushInterval:   100,
		Q:               NewInMemoryQueue(100),
		EventDispatcher: &MockDispatcher{},
		wg:              exeCtx.GetWaitSync(),
	}
	processor.BatchSize = 10
	processor.StartTicker(exeCtx.GetContext())

	impression := BuildTestImpressionEvent()
	conversion := BuildTestConversionEvent()

	processor.ProcessEvent(impression)
	processor.ProcessEvent(impression)
	processor.ProcessEvent(conversion)
	processor.ProcessEvent(conversion)

	assert.Equal(t, 4, processor.EventsCount())

	exeCtx.TerminateAndWait()

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
	exeCtx := utils.NewCancelableExecutionCtxExecutionCtx()
	processor := &QueueingEventProcessor{
		MaxQueueSize:    100,
		FlushInterval:   30 * time.Second,
		Q:               NewInMemoryQueue(100),
		EventDispatcher: &MockDispatcher{},
		wg:              exeCtx.GetWaitSync(),
	}
	processor.BatchSize = 10
	processor.StartTicker(exeCtx.GetContext())

	impression := BuildTestImpressionEvent()
	conversion := BuildTestConversionEvent()

	processor.ProcessEvent(impression)
	processor.ProcessEvent(impression)
	processor.ProcessEvent(conversion)
	processor.ProcessEvent(conversion)

	assert.Equal(t, 4, processor.EventsCount())

	// Triggers the flush in the processor
	exeCtx.TerminateAndWait()

	assert.Equal(t, 0, processor.EventsCount())
}

func TestDefaultEventProcessor_ProcessBatchRevisionMismatch(t *testing.T) {
	exeCtx := utils.NewCancelableExecutionCtxExecutionCtx()
	processor := &QueueingEventProcessor{
		MaxQueueSize:    100,
		FlushInterval:   100,
		Q:               NewInMemoryQueue(100),
		EventDispatcher: &MockDispatcher{},
		wg:              exeCtx.GetWaitSync(),
	}
	processor.BatchSize = 10
	processor.StartTicker(exeCtx.GetContext())

	impression := BuildTestImpressionEvent()
	conversion := BuildTestConversionEvent()

	processor.ProcessEvent(impression)
	impression.EventContext.Revision = "12112121"
	processor.ProcessEvent(impression)
	processor.ProcessEvent(conversion)
	processor.ProcessEvent(conversion)

	assert.Equal(t, 4, processor.EventsCount())

	exeCtx.TerminateAndWait()

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
	exeCtx := utils.NewCancelableExecutionCtxExecutionCtx()
	processor := &QueueingEventProcessor{
		MaxQueueSize:    100,
		FlushInterval:   100,
		Q:               NewInMemoryQueue(100),
		EventDispatcher: &MockDispatcher{},
		wg:              exeCtx.GetWaitSync(),
	}
	processor.BatchSize = 10
	processor.StartTicker(exeCtx.GetContext())

	impression := BuildTestImpressionEvent()
	conversion := BuildTestConversionEvent()

	processor.ProcessEvent(impression)
	impression.EventContext.ProjectID = "121121211111"
	processor.ProcessEvent(impression)
	processor.ProcessEvent(conversion)
	processor.ProcessEvent(conversion)

	assert.Equal(t, 4, processor.EventsCount())

	exeCtx.TerminateAndWait()

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
	exeCtx := utils.NewCancelableExecutionCtxExecutionCtx()
	processor := &QueueingEventProcessor{
		MaxQueueSize:    100,
		FlushInterval:   100,
		Q:               NewChanQueue(100),
		EventDispatcher: &HTTPEventDispatcher{},
		wg:              exeCtx.GetWaitSync(),
	}
	processor.BatchSize = 10
	processor.StartTicker(exeCtx.GetContext())

	impression := BuildTestImpressionEvent()

	processor.ProcessEvent(impression)
	processor.ProcessEvent(impression)
	processor.ProcessEvent(impression)

	exeCtx.TerminateAndWait()
	assert.NotNil(t, processor.Ticker)

	assert.Equal(t, 0, processor.EventsCount())
}

func TestChanQueueEventProcessor_ProcessBatch(t *testing.T) {
	exeCtx := utils.NewCancelableExecutionCtxExecutionCtx()
	processor := &QueueingEventProcessor{MaxQueueSize: 100, FlushInterval: 100, Q: NewChanQueue(100), EventDispatcher: &MockDispatcher{}, wg: exeCtx.GetWaitSync()}
	processor.BatchSize = 10
	processor.StartTicker(exeCtx.GetContext())

	impression := BuildTestImpressionEvent()
	conversion := BuildTestConversionEvent()

	processor.ProcessEvent(impression)
	processor.ProcessEvent(impression)
	processor.ProcessEvent(conversion)
	processor.ProcessEvent(conversion)

	exeCtx.TerminateAndWait()

	assert.NotNil(t, processor.Ticker)

	assert.Equal(t, 0, processor.EventsCount())

	result, ok := (processor.EventDispatcher).(*MockDispatcher)

	if ok {
		assert.Equal(t, 1, len(result.Events))
		evs := result.Events[0]
		assert.True(t, len(evs.Event.Visitors) >= 1)
	}
}
