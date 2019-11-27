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
	"github.com/optimizely/go-sdk/pkg/logging"
	"github.com/optimizely/go-sdk/pkg/utils"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

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
BenchmarkWithQueueSize/QueueSize100-8         	 2000000	       774 ns/op
BenchmarkWithQueueSize/QueueSize500-8         	 2000000	       716 ns/op
BenchmarkWithQueueSize/QueueSize1000-8        	 2000000	       690 ns/op
BenchmarkWithQueueSize/QueueSize2000-8        	 2000000	       732 ns/op
BenchmarkWithQueueSize/QueueSize3000-8        	 2000000	       735 ns/op
BenchmarkWithQueueSize/QueueSize4000-8        	 2000000	       740 ns/op

BenchmarkWithBatchSize/BatchSize10-8          	 2000000	       665 ns/op
BenchmarkWithBatchSize/BatchSize20-8          	 2000000	       597 ns/op
BenchmarkWithBatchSize/BatchSize30-8          	 3000000	       556 ns/op
BenchmarkWithBatchSize/BatchSize40-8          	 3000000	       566 ns/op
BenchmarkWithBatchSize/BatchSize50-8          	 3000000	       546 ns/op
BenchmarkWithBatchSize/BatchSize60-8          	 3000000	       504 ns/op

BenchmarkWithQueue/InMemoryQueue-8            	 2000000	       674 ns/op
BenchmarkWithQueue/ChannelQueue-8             	 2000000	       937 ns/op

*/
func BenchmarkWithQueueSize(b *testing.B) {
	// no op logger added to keep out extra discarded events
	logging.SetLogger(&NoOpLogger{})

	merges := []struct {
		name  string
		qSize int
	}{
		{"QueueSize100", 100},
		{"QueueSize500", 500},
		{"QueueSize1000", 1000},
		{"QueueSize2000", 2000},
		{"QueueSize3000", 3000},
		{"QueueSize4000", 4000},
	}

	for _, merge := range merges {
		var totalSent = 0
		var numberRun = 0
		b.Run(merge.name, func(b *testing.B) {
			if numberRun == 0 {
				numberRun = b.N
			}
			totalSent += benchmarkProcessorWithQueueSize(merge.qSize, b)
		})
		if totalSent < numberRun {
			println("Total sent and run ", totalSent, numberRun)
			b.Fail()
		}
	}
}

func BenchmarkWithBatchSize(b *testing.B) {
	logging.SetLogger(&NoOpLogger{})

	merges := []struct {
		name      string
		batchSize int
	}{
		{"BatchSize10", 10},
		{"BatchSize20", 20},
		{"BatchSize30", 30},
		{"BatchSize40", 40},
		{"BatchSize50", 50},
		{"BatchSize60", 60},
	}

	for _, merge := range merges {
		b.Run(merge.name, func(b *testing.B) {
			benchmarkProcessorWithBatchSize(merge.batchSize, b)
		})
	}

}

func BenchmarkWithQueue(b *testing.B) {
	logging.SetLogger(&NoOpLogger{})

	b.Run("InMemoryQueue", func(b *testing.B) {
		benchmarkProcessorWithQueue(NewInMemoryQueue(defaultQueueSize), b)
	})

	b.Run("ChannelQueue", func(b *testing.B) {
		benchmarkProcessorWithQueue(NewChanQueue(defaultQueueSize), b)
	})

}

func benchmarkProcessorWithQueueSize(qSize int, b *testing.B) int {
	exeCtx := utils.NewCancelableExecutionCtx()
	dispatcher := NewMockDispatcher(100, false)
	processor := NewBatchEventProcessor(
		WithQueueSize(qSize),
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

	return dispatcher.Events.Size()
}

func benchmarkProcessorWithQueue(q Queue, b *testing.B) int {
	exeCtx := utils.NewCancelableExecutionCtx()
	dispatcher := NewMockDispatcher(100, false)
	processor := NewBatchEventProcessor(
		WithQueue(q),
		WithEventDispatcher(dispatcher))
	processor.Start(exeCtx)

	conversion := BuildTestConversionEvent()

	for i := 0; i < b.N; i++ {
		var success = false
		for !success {
			success = processor.ProcessEvent(conversion)
		}
		//time.Sleep(benchmarkSleep)
	}

	exeCtx.TerminateAndWait()

	return dispatcher.Events.Size()
}

func benchmarkProcessorWithBatchSize(bs int, b *testing.B) int {
	exeCtx := utils.NewCancelableExecutionCtx()
	dispatcher := NewMockDispatcher(100, false)
	processor := NewBatchEventProcessor(
		WithBatchSize(bs),
		WithEventDispatcher(dispatcher))
	processor.Start(exeCtx)

	conversion := BuildTestConversionEvent()

	for i := 0; i < b.N; i++ {
		var success = false
		for !success {
			success = processor.ProcessEvent(conversion)
		}
		//time.Sleep(benchmarkSleep)
	}

	exeCtx.TerminateAndWait()

	return dispatcher.Events.Size()
}
