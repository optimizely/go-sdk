/****************************************************************************
 * Copyright 2019, Optimizely, Inc. and contributors                        *
 *                                                                          *
 * Licensed under the Apache License, Version 2.0 (the "License");          *
 * you may not use this file except in compliance with the License.         *
 * You may obtain a copy of the License at                                  *
 *                                                                          *
 *    http://www.apache.org/licenses/LICENSE-2.0                            *
 *                                                                          *
 * Unless required by applicable law or agreed to in writing, software      *
 * distributed under the License is distributed on an "AS IS" BASIS,        *
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. *
 * See the License for the specific language governing permissions and      *
 * limitations under the License.                                           *
 ***************************************************************************/

// Package event //
package event

import (
	"errors"
	"fmt"
	"github.com/optimizely/go-sdk/pkg/logging"
	"github.com/optimizely/go-sdk/pkg/utils"
	"github.com/stretchr/testify/assert"
	"math"
	"testing"
	"time"
)

type CountingDispatcher struct {
	eventCount int
	visitorCount int
}

func (c *CountingDispatcher) DispatchEvent(event LogEvent) (bool, error) {
	c.eventCount++
	c.visitorCount += len(event.Event.Visitors)
	return true, nil
}


type MockDispatcher struct {
	ShouldFail bool
	Events     Queue
}

func (m *MockDispatcher) DispatchEvent(event LogEvent) (bool, error) {
	if m.ShouldFail {
		return false, errors.New("Failed to dispatch")
	}

	m.Events.Add(event)
	return true, nil
}

func NewMockDispatcher(queueSize int, shouldFail bool) *MockDispatcher {
	return &MockDispatcher{Events: NewInMemoryQueue(queueSize), ShouldFail: shouldFail}
}

func TestDefaultEventProcessor_ProcessImpression(t *testing.T) {
	exeCtx := utils.NewCancelableExecutionCtx()
	processor := NewBatchEventProcessor()
	processor.EventDispatcher = nil
	processor.Start(exeCtx)
	processor.Start(exeCtx)

	impression := BuildTestImpressionEvent()

	processor.ProcessEvent(impression)

	assert.Equal(t, 1, processor.eventsCount())

	exeCtx.TerminateAndWait()

	assert.NotNil(t, processor.Ticker)

	assert.Equal(t, 0, processor.eventsCount())
}

func TestCustomEventProcessor_Create(t *testing.T) {
	exeCtx := utils.NewCancelableExecutionCtx()
	processor := NewBatchEventProcessor(
		WithEventDispatcher(NewMockDispatcher(100, false)),
		WithQueueSize(10),
		WithFlushInterval(100))
	processor.Start(exeCtx)

	impression := BuildTestImpressionEvent()

	processor.ProcessEvent(impression)

	assert.Equal(t, 1, processor.eventsCount())

	exeCtx.TerminateAndWait()

	assert.NotNil(t, processor.Ticker)

	assert.Equal(t, 0, processor.eventsCount())
}

func TestDefaultEventProcessor_LogEventNotification(t *testing.T) {
	exeCtx := utils.NewCancelableExecutionCtx()
	processor := NewBatchEventProcessor(
		WithQueueSize(100),
		WithQueue(NewInMemoryQueue(100)),
		WithEventDispatcher(NewMockDispatcher(100, false)),
		WithSDKKey("fakeSDKKey"))

	var logEvent LogEvent

	id, _ := processor.OnEventDispatch(func(eventNotification LogEvent) {
		logEvent = eventNotification
	})
	processor.Start(exeCtx)

	impression := BuildTestImpressionEvent()
	conversion := BuildTestConversionEvent()

	processor.ProcessEvent(impression)
	processor.ProcessEvent(impression)
	processor.ProcessEvent(conversion)
	processor.ProcessEvent(conversion)

	assert.Equal(t, 4, processor.eventsCount())

	exeCtx.TerminateAndWait()

	assert.NotNil(t, logEvent)
	assert.Equal(t, 4, len(logEvent.Event.Visitors))

	err := processor.RemoveOnEventDispatch(id)

	assert.Nil(t, err)
}

func TestDefaultEventProcessor_DefaultConfig(t *testing.T) {
	exeCtx := utils.NewCancelableExecutionCtx()
	processor := NewBatchEventProcessor(
		WithEventDispatcher(NewMockDispatcher(100, false)),
		// here we are setting the timing interval so that we don't have to wait the default 30 seconds
		WithFlushInterval(500*time.Millisecond))
	processor.Start(exeCtx)

	impression := BuildTestImpressionEvent()
	conversion := BuildTestConversionEvent()

	processor.ProcessEvent(impression)
	processor.ProcessEvent(impression)
	processor.ProcessEvent(conversion)
	processor.ProcessEvent(conversion)

	assert.Equal(t, 4, processor.eventsCount())

	// sleep for 1 second here. to allow event processor to run.
	time.Sleep(1 * time.Second)

	assert.NotNil(t, processor.Ticker)

	assert.Equal(t, 0, processor.eventsCount())

	result, ok := (processor.EventDispatcher).(*MockDispatcher)

	if ok {
		assert.Equal(t, 1, result.Events.Size())
		evs := result.Events.Get(1)
		logEvent, _ := evs[0].(LogEvent)
		assert.Equal(t, 4, len(logEvent.Event.Visitors))
	}
}

func TestDefaultEventProcessor_ProcessBatch(t *testing.T) {
	exeCtx := utils.NewCancelableExecutionCtx()
	processor := NewBatchEventProcessor(
		WithFlushInterval(1*time.Second),
		WithQueueSize(100),
		WithQueue(NewInMemoryQueue(100)),
		WithEventDispatcher(NewMockDispatcher(100, false)))
	processor.Start(exeCtx)

	impression := BuildTestImpressionEvent()
	conversion := BuildTestConversionEvent()

	processor.ProcessEvent(impression)
	processor.ProcessEvent(impression)
	processor.ProcessEvent(conversion)
	processor.ProcessEvent(conversion)

	assert.Equal(t, 4, processor.eventsCount())

	time.Sleep(1500 * time.Millisecond)

	assert.NotNil(t, processor.Ticker)

	assert.Equal(t, 0, processor.eventsCount())

	result, ok := (processor.EventDispatcher).(*MockDispatcher)

	if ok {
		assert.Equal(t, 1, result.Events.Size())
		evs := result.Events.Get(1)
		logEvent, _ := evs[0].(LogEvent)
		assert.Equal(t, 4, len(logEvent.Event.Visitors))
	}
}

func TestDefaultEventProcessor_BatchSizeMet(t *testing.T) {
	exeCtx := utils.NewCancelableExecutionCtx()
	processor := NewBatchEventProcessor(
		WithBatchSize(2),
		WithFlushInterval(1000*time.Millisecond),
		WithQueue(NewInMemoryQueue(2)),
		WithEventDispatcher(NewMockDispatcher(100, false)))
	processor.Start(exeCtx)

	impression := BuildTestImpressionEvent()
	conversion := BuildTestConversionEvent()

	processor.ProcessEvent(impression)
	processor.ProcessEvent(impression)

	assert.Equal(t, 2, processor.eventsCount())

	time.Sleep(100 * time.Millisecond)

	result, ok := (processor.EventDispatcher).(*MockDispatcher)

	if ok {
		assert.Equal(t, 1, result.Events.Size())
		evs := result.Events.Get(1)
		logEvent, _ := evs[0].(LogEvent)
		assert.Equal(t, 2, len(logEvent.Event.Visitors))

	}

	processor.ProcessEvent(conversion)
	processor.ProcessEvent(conversion)

	assert.Equal(t, 2, processor.eventsCount())

	exeCtx.TerminateAndWait()

	assert.NotNil(t, processor.Ticker)

	assert.Equal(t, 0, processor.eventsCount())

	if ok {
		assert.Equal(t, 2, result.Events.Size())
	}
}

func TestDefaultEventProcessor_BatchSizeLessThanQSize(t *testing.T) {
	processor := NewBatchEventProcessor(
		WithQueueSize(2),
		WithFlushInterval(1000*time.Millisecond),
		WithQueue(NewInMemoryQueue(100)),
		WithEventDispatcher(NewMockDispatcher(100, false)))

	assert.Equal(t, DefaultBatchSize, processor.BatchSize)
	assert.Equal(t, defaultQueueSize, processor.MaxQueueSize)

}

func TestDefaultEventProcessor_QSizeExceeded(t *testing.T) {
	exeCtx := utils.NewCancelableExecutionCtx()
	processor := NewBatchEventProcessor(
		WithQueueSize(2),
		WithBatchSize(2),
		WithFlushInterval(1000*time.Millisecond),
		WithQueue(NewInMemoryQueue(2)),
		WithEventDispatcher(NewMockDispatcher(100, true)))
	processor.Start(exeCtx)

	impression := BuildTestImpressionEvent()

	processor.ProcessEvent(impression)
	processor.ProcessEvent(impression)

	assert.Equal(t, 2, processor.eventsCount())

	processor.ProcessEvent(impression)
	processor.ProcessEvent(impression)

	assert.Equal(t, 2, processor.eventsCount())

}

func TestDefaultEventProcessor_FailedDispatch(t *testing.T) {
	exeCtx := utils.NewCancelableExecutionCtx()
	processor := NewBatchEventProcessor(
		WithQueueSize(100),
		WithFlushInterval(100),
		WithQueue(NewInMemoryQueue(100)),
		WithEventDispatcher(&MockDispatcher{ShouldFail: true, Events: NewInMemoryQueue(100)}))
	processor.Start(exeCtx)

	impression := BuildTestImpressionEvent()
	conversion := BuildTestConversionEvent()

	processor.ProcessEvent(impression)
	processor.ProcessEvent(impression)
	processor.ProcessEvent(conversion)
	processor.ProcessEvent(conversion)

	assert.Equal(t, 4, processor.eventsCount())

	exeCtx.TerminateAndWait()

	assert.NotNil(t, processor.Ticker)

	assert.Equal(t, 4, processor.eventsCount())

	result, ok := (processor.EventDispatcher).(*MockDispatcher)

	if ok {
		assert.Equal(t, 0, result.Events.Size())
	}
}

func TestBatchEventProcessor_FlushesOnClose(t *testing.T) {
	exeCtx := utils.NewCancelableExecutionCtx()
	processor := NewBatchEventProcessor(
		WithQueueSize(100),
		WithQueue(NewInMemoryQueue(100)),
		WithEventDispatcher(NewMockDispatcher(100, false)))
	processor.Start(exeCtx)

	impression := BuildTestImpressionEvent()
	conversion := BuildTestConversionEvent()

	processor.ProcessEvent(impression)
	processor.ProcessEvent(impression)
	processor.ProcessEvent(conversion)
	processor.ProcessEvent(conversion)

	assert.Equal(t, 4, processor.eventsCount())

	// Triggers the flush in the processor
	exeCtx.TerminateAndWait()

	assert.Equal(t, 0, processor.eventsCount())
}

func TestDefaultEventProcessor_ProcessBatchRevisionMismatch(t *testing.T) {
	exeCtx := utils.NewCancelableExecutionCtx()
	processor := NewBatchEventProcessor(
		WithQueueSize(100),
		WithQueue(NewInMemoryQueue(100)),
		WithEventDispatcher(NewMockDispatcher(100, false)))

	processor.Start(exeCtx)

	impression := BuildTestImpressionEvent()
	conversion := BuildTestConversionEvent()

	processor.ProcessEvent(impression)
	impression.EventContext.Revision = "12112121"
	processor.ProcessEvent(impression)
	processor.ProcessEvent(conversion)
	processor.ProcessEvent(conversion)

	assert.Equal(t, 4, processor.eventsCount())

	exeCtx.TerminateAndWait()

	assert.NotNil(t, processor.Ticker)

	assert.Equal(t, 0, processor.eventsCount())

	result, ok := (processor.EventDispatcher).(*MockDispatcher)

	if ok {
		assert.Equal(t, 3, result.Events.Size())
		evs := result.Events.Get(3)
		logEvent, _ := evs[len(evs)-1].(LogEvent)
		assert.Equal(t, 2, len(logEvent.Event.Visitors))
	}
}

func TestDefaultEventProcessor_ProcessBatchProjectMismatch(t *testing.T) {
	exeCtx := utils.NewCancelableExecutionCtx()
	processor := NewBatchEventProcessor(
		WithQueueSize(100),
		WithQueue(NewInMemoryQueue(100)),
		WithEventDispatcher(NewMockDispatcher(100, false)))

	processor.Start(exeCtx)

	impression := BuildTestImpressionEvent()
	conversion := BuildTestConversionEvent()

	processor.ProcessEvent(impression)
	impression.EventContext.ProjectID = "121121211111"
	processor.ProcessEvent(impression)
	processor.ProcessEvent(conversion)
	processor.ProcessEvent(conversion)

	assert.Equal(t, 4, processor.eventsCount())

	exeCtx.TerminateAndWait()

	assert.NotNil(t, processor.Ticker)

	assert.Equal(t, 0, processor.eventsCount())

	result, ok := (processor.EventDispatcher).(*MockDispatcher)

	if ok {
		assert.Equal(t, 3, result.Events.Size())
		evs := result.Events.Get(3)
		logEvent, _ := evs[len(evs)-1].(LogEvent)
		assert.Equal(t, 2, len(logEvent.Event.Visitors))
	}
}

func TestChanQueueEventProcessor_ProcessImpression(t *testing.T) {
	exeCtx := utils.NewCancelableExecutionCtx()
	processor := NewBatchEventProcessor(
		WithQueueSize(100),
		WithQueue(NewInMemoryQueue(100)),
		WithEventDispatcher(&HTTPEventDispatcher{requester: utils.NewHTTPRequester()}))

	processor.Start(exeCtx)

	impression := BuildTestImpressionEvent()

	processor.ProcessEvent(impression)
	processor.ProcessEvent(impression)
	processor.ProcessEvent(impression)

	exeCtx.TerminateAndWait()
	assert.NotNil(t, processor.Ticker)

	assert.Equal(t, 0, processor.eventsCount())
}

func TestChanQueueEventProcessor_ProcessBatch(t *testing.T) {
	exeCtx := utils.NewCancelableExecutionCtx()
	processor := NewBatchEventProcessor(
		WithQueueSize(100),
		WithQueue(NewInMemoryQueue(100)),
		WithEventDispatcher(NewMockDispatcher(100, false)))
	processor.Start(exeCtx)

	impression := BuildTestImpressionEvent()
	conversion := BuildTestConversionEvent()

	processor.ProcessEvent(impression)
	processor.ProcessEvent(impression)
	processor.ProcessEvent(conversion)
	processor.ProcessEvent(conversion)

	exeCtx.TerminateAndWait()

	assert.NotNil(t, processor.Ticker)

	assert.Equal(t, 0, processor.eventsCount())

	result, ok := (processor.EventDispatcher).(*MockDispatcher)

	if ok {
		assert.Equal(t, 1, result.Events.Size())
		evs := result.Events.Get(1)
		logEvent, _ := evs[0].(LogEvent)
		assert.True(t, len(logEvent.Event.Visitors) >= 1)
	}
}

// The NoOpLogger is used during benchmarking so that results are printed nicely.
type NoOpLogger struct {
}

func (l *NoOpLogger) Log(level logging.LogLevel, message string, fields map[string]interface{}) {
}

func (l *NoOpLogger) SetLogLevel(level logging.LogLevel) {

}

/**
goos: darwin
goarch: amd64
pkg: github.com/optimizely/go-sdk/pkg/event
BenchmarkProcessor/InMemory/BatchSize-10/QueueSize-10-8         	 2531830	       456 ns/op
BenchmarkProcessor/InMemory/BatchSize-20/QueueSize-10-8         	 2966862	       398 ns/op
BenchmarkProcessor/InMemory/BatchSize-30/QueueSize-10-8         	 3224689	       372 ns/op
BenchmarkProcessor/InMemory/BatchSize-40/QueueSize-10-8         	 3283634	       384 ns/op
BenchmarkProcessor/InMemory/BatchSize-50/QueueSize-10-8         	 3368804	       352 ns/op
BenchmarkProcessor/InMemory/BatchSize-60/QueueSize-10-8         	 3468763	       336 ns/op
BenchmarkProcessor/InMemory/BatchSize-10/QueueSize-100-8        	 2581394	       464 ns/op
BenchmarkProcessor/InMemory/BatchSize-20/QueueSize-100-8        	 2911731	       408 ns/op
BenchmarkProcessor/InMemory/BatchSize-30/QueueSize-100-8        	 3224674	       375 ns/op
BenchmarkProcessor/InMemory/BatchSize-40/QueueSize-100-8        	 3262027	       366 ns/op
BenchmarkProcessor/InMemory/BatchSize-50/QueueSize-100-8        	 3094736	       354 ns/op
BenchmarkProcessor/InMemory/BatchSize-60/QueueSize-100-8        	 3523911	       338 ns/op
BenchmarkProcessor/InMemory/BatchSize-10/QueueSize-1000-8       	 2580465	       467 ns/op
BenchmarkProcessor/InMemory/BatchSize-20/QueueSize-1000-8       	 2940940	       415 ns/op
BenchmarkProcessor/InMemory/BatchSize-30/QueueSize-1000-8       	 3229284	       375 ns/op
BenchmarkProcessor/InMemory/BatchSize-40/QueueSize-1000-8       	 3280029	       367 ns/op
BenchmarkProcessor/InMemory/BatchSize-50/QueueSize-1000-8       	 3258297	       368 ns/op
BenchmarkProcessor/InMemory/BatchSize-60/QueueSize-1000-8       	 3484419	       336 ns/op
BenchmarkProcessor/InMemory/BatchSize-10/QueueSize-10000-8      	 2598885	       462 ns/op
BenchmarkProcessor/InMemory/BatchSize-20/QueueSize-10000-8      	 2907445	       414 ns/op
BenchmarkProcessor/InMemory/BatchSize-30/QueueSize-10000-8      	 3215616	       382 ns/op
BenchmarkProcessor/InMemory/BatchSize-40/QueueSize-10000-8      	 3243544	       367 ns/op
BenchmarkProcessor/InMemory/BatchSize-50/QueueSize-10000-8      	 3382228	       391 ns/op
BenchmarkProcessor/InMemory/BatchSize-60/QueueSize-10000-8      	 3503428	       354 ns/op
BenchmarkProcessor/InMemory/BatchSize-10/QueueSize-100000-8     	 2268799	       512 ns/op
BenchmarkProcessor/InMemory/BatchSize-20/QueueSize-100000-8     	 2788728	       429 ns/op
BenchmarkProcessor/InMemory/BatchSize-30/QueueSize-100000-8     	 2799598	       404 ns/op
BenchmarkProcessor/InMemory/BatchSize-40/QueueSize-100000-8     	 3010062	       368 ns/op
BenchmarkProcessor/InMemory/BatchSize-50/QueueSize-100000-8     	 3353461	       352 ns/op
BenchmarkProcessor/InMemory/BatchSize-60/QueueSize-100000-8     	 3429447	       342 ns/op
BenchmarkProcessor/Channel/BatchSize-20/QueueSize-10-8          	 1755082	       738 ns/op
BenchmarkProcessor/Channel/BatchSize-30/QueueSize-10-8          	 1770522	       687 ns/op
BenchmarkProcessor/Channel/BatchSize-40/QueueSize-10-8          	 1784857	       679 ns/op
BenchmarkProcessor/Channel/BatchSize-50/QueueSize-10-8          	 1805431	       663 ns/op
BenchmarkProcessor/Channel/BatchSize-60/QueueSize-10-8          	 1815903	       786 ns/op
BenchmarkProcessor/Channel/BatchSize-10/QueueSize-100-8         	 1640970	       726 ns/op
BenchmarkProcessor/Channel/BatchSize-20/QueueSize-100-8         	 1719675	       725 ns/op
BenchmarkProcessor/Channel/BatchSize-40/QueueSize-100-8         	 1761488	       687 ns/op
BenchmarkProcessor/Channel/BatchSize-50/QueueSize-100-8         	 1734388	       686 ns/op
BenchmarkProcessor/Channel/BatchSize-60/QueueSize-100-8         	 1772548	       736 ns/op
BenchmarkProcessor/Channel/BatchSize-10/QueueSize-1000-8        	 1476832	       744 ns/op
BenchmarkProcessor/Channel/BatchSize-20/QueueSize-1000-8        	 1444575	       725 ns/op
BenchmarkProcessor/Channel/BatchSize-30/QueueSize-1000-8        	 1707232	       757 ns/op
BenchmarkProcessor/Channel/BatchSize-50/QueueSize-1000-8        	 1712044	       700 ns/op
BenchmarkProcessor/Channel/BatchSize-10/QueueSize-10000-8       	 1415298	       836 ns/op
BenchmarkProcessor/Channel/BatchSize-20/QueueSize-10000-8       	 1607797	       758 ns/op
BenchmarkProcessor/Channel/BatchSize-30/QueueSize-10000-8       	 1723740	       679 ns/op
BenchmarkProcessor/Channel/BatchSize-50/QueueSize-10000-8       	 1750117	       741 ns/op
BenchmarkProcessor/Channel/BatchSize-60/QueueSize-10000-8       	 1784826	       712 ns/op
BenchmarkProcessor/Channel/BatchSize-10/QueueSize-100000-8      	 1448926	       849 ns/op
BenchmarkProcessor/Channel/BatchSize-20/QueueSize-100000-8      	 1639345	       783 ns/op
BenchmarkProcessor/Channel/BatchSize-30/QueueSize-100000-8      	 1643888	       720 ns/op
BenchmarkProcessor/Channel/BatchSize-40/QueueSize-100000-8      	 1771636	       827 ns/op
BenchmarkProcessor/Channel/BatchSize-50/QueueSize-100000-8      	 1607248	       790 ns/op
BenchmarkProcessor/Channel/BatchSize-60/QueueSize-100000-8      	 1714305	       749 ns/op
*/
func BenchmarkProcessor(b *testing.B) {
	// no op logger added to keep out extra discarded events
	logging.SetLogger(&NoOpLogger{})

	merges := []struct {
		name string
		fun  func(qSize int) Queue
	}{
		{"InMemory", NewInMemoryQueue},
		{"Channel", NewChanQueue},
	}

	for _, merge := range merges {
		for i := 1.; i <= 5; i++ {
			qs := int(math.Pow(10, i))
			for j := 1; j <= 6; j++ {
				bs := 10 * j
				b.Run(fmt.Sprintf("%s/BatchSize-%d/QueueSize-%d", merge.name, bs, qs), func(b *testing.B) {
					q := merge.fun(qs)
					benchmarkProcessor(q, bs, b)
				})
			}
		}
	}

}

func benchmarkProcessor(q Queue, bSize int, b *testing.B) {
	exeCtx := utils.NewCancelableExecutionCtx()
	dispatcher := &CountingDispatcher{}
	processor := NewBatchEventProcessor(
		WithQueue(q),
		WithBatchSize(bSize),
		WithEventDispatcher(dispatcher))
	processor.Start(exeCtx)

	conversion := BuildTestConversionEvent()

	for i := 0; i < b.N; i++ {
		var success = false
		for !success {
			success = processor.ProcessEvent(conversion)
		}
	}

	exeCtx.TerminateAndWait()

	if b.N != dispatcher.visitorCount {
		println("Total sent and run ", dispatcher.visitorCount, b.N)
		b.Fail()
	}
}
