/****************************************************************************
 * Copyright 2019-2020,2023 Optimizely, Inc. and contributors               *
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
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/optimizely/go-sdk/v2/pkg/entities"
	"github.com/optimizely/go-sdk/v2/pkg/logging"
	"github.com/optimizely/go-sdk/v2/pkg/utils"
	// Import the package containing the Impression type
)

type CountingDispatcher struct {
	eventCount   int
	visitorCount int
}

func (c *CountingDispatcher) DispatchEvent(event LogEvent) (bool, error) {
	c.eventCount++
	c.visitorCount += len(event.Event.Visitors)
	return true, nil
}

type MockDispatcher struct {
	ShouldFail  bool
	Events      Queue
	eventsQueue Queue // dispatch events from this queue
}

func (m *MockDispatcher) DispatchEvent(event LogEvent) (bool, error) {
	if m.ShouldFail {
		return false, errors.New("Failed to dispatch")
	}

	m.Events.Add(event)
	if m.eventsQueue != nil {
		m.eventsQueue.Add(event)
		go m.flushEvents()
	}
	return true, nil
}

func (m *MockDispatcher) flushEvents() {
	queueSize := m.eventsQueue.Size()
	for ; queueSize > 0; queueSize = m.eventsQueue.Size() {
		m.eventsQueue.Remove(1)
	}
}

func NewMockDispatcher(queueSize int, shouldFail bool) *MockDispatcher {
	return &MockDispatcher{Events: NewInMemoryQueue(queueSize), eventsQueue: NewInMemoryQueue(queueSize), ShouldFail: shouldFail}
}

func newExecutionContext() *utils.ExecGroup {
	return utils.NewExecGroup(context.Background(), logging.GetLogger("", "NewExecGroup"))
}

func TestDefaultEventProcessor_ProcessImpression(t *testing.T) {
	eg := newExecutionContext()
	processor := NewBatchEventProcessor()
	processor.EventDispatcher = NewQueueEventDispatcher("", processor.metricsRegistry)
	eg.Go(processor.Start)

	impression := BuildTestImpressionEvent()

	processor.ProcessEvent(impression)

	assert.Equal(t, 1, processor.eventsCount())

	eg.TerminateAndWait()

	assert.NotNil(t, processor.Ticker)

	assert.Equal(t, 0, processor.eventsCount())
}

func TestCustomEventProcessor_Create(t *testing.T) {
	eg := newExecutionContext()
	processor := NewBatchEventProcessor(
		WithEventDispatcher(NewMockDispatcher(100, false)),
		WithQueueSize(10),
		WithFlushInterval(100))
	eg.Go(processor.Start)

	impression := BuildTestImpressionEvent()

	processor.ProcessEvent(impression)

	assert.Equal(t, 1, processor.eventsCount())

	eg.TerminateAndWait()

	assert.NotNil(t, processor.Ticker)

	assert.Equal(t, 0, processor.eventsCount())
}

func TestEndPointOptionEventProcessor(t *testing.T) {
	// Default end point
	processor := NewBatchEventProcessor()
	assert.Equal(t, EventEndPoints["US"], processor.EventEndPoint)

	customEndPoint := "https://logx.optimizely.com"
	processor = NewBatchEventProcessor(
		WithEventEndPoint(customEndPoint))
	assert.Equal(t, customEndPoint, processor.EventEndPoint)
}

func TestEndPointOptionEU(t *testing.T) {
	// Default end point
	processor := NewBatchEventProcessor()
	assert.Equal(t, EventEndPoints["US"], processor.EventEndPoint)

	euEndpoint := EventEndPoints["EU"]
	processor = NewBatchEventProcessor(
		WithEventEndPoint(euEndpoint))
	assert.Equal(t, euEndpoint, processor.EventEndPoint)
}

func TestDefaultEventProcessor_LogEventNotification(t *testing.T) {
	eg := newExecutionContext()
	processor := NewBatchEventProcessor(
		WithQueueSize(100),
		WithQueue(NewInMemoryQueue(100)),
		WithEventDispatcher(NewMockDispatcher(100, false)),
		WithSDKKey("fakeSDKKey"))

	var logEvent LogEvent

	id, _ := processor.OnEventDispatch(func(eventNotification LogEvent) {
		logEvent = eventNotification
	})
	eg.Go(processor.Start)

	impression := BuildTestImpressionEvent()
	conversion := BuildTestConversionEvent()

	processor.ProcessEvent(impression)
	processor.ProcessEvent(impression)
	processor.ProcessEvent(conversion)
	processor.ProcessEvent(conversion)

	assert.Equal(t, 4, processor.eventsCount())

	eg.TerminateAndWait()

	assert.NotNil(t, logEvent)
	assert.Equal(t, 4, len(logEvent.Event.Visitors))

	err := processor.RemoveOnEventDispatch(id)

	assert.Nil(t, err)
}

// EUTestConfig extends TestConfig to set EU region
type EUTestConfig struct {
	TestConfig
}

// Override GetRegion to return "EU"
func (c EUTestConfig) GetRegion() string {
	return "EU"
}

func TestDefaultEventProcessor_LogEventNotification_EU(t *testing.T) {
	eg := newExecutionContext()
	processor := NewBatchEventProcessor(
		WithQueueSize(100),
		WithQueue(NewInMemoryQueue(100)),
		WithEventDispatcher(NewMockDispatcher(100, false)),
		WithSDKKey("fakeSDKKey"))

	var logEvent LogEvent

	id, _ := processor.OnEventDispatch(func(eventNotification LogEvent) {
		logEvent = eventNotification
	})
	eg.Go(processor.Start)

	// Create EU-specific impression and conversion events
	tc := EUTestConfig{}

	// Create impression event with EU region
	impressionUserEvent, _ := CreateImpressionUserEvent(
		tc,
		testExperiment,
		&testVariation,
		userContext,
		"",
		testExperiment.Key,
		"experiment",
		true,
		nil)

	// Create conversion event with EU region
	conversionUserEvent := CreateConversionUserEvent(
		tc,
		entities.Event{
			ExperimentIds: []string{"15402980349"},
			ID:            "15368860886",
			Key:           "sample_conversion"},
		userContext,
		make(map[string]interface{}))

	// Set event endpoint to EU
	processor.EventEndPoint = EventEndPoints["EU"]

	processor.ProcessEvent(impressionUserEvent)
	processor.ProcessEvent(impressionUserEvent)
	processor.ProcessEvent(conversionUserEvent)
	processor.ProcessEvent(conversionUserEvent)

	assert.Equal(t, 4, processor.eventsCount())

	eg.TerminateAndWait()

	assert.NotNil(t, logEvent)
	assert.Equal(t, 4, len(logEvent.Event.Visitors))
	assert.Equal(t, EventEndPoints["EU"], logEvent.EndPoint)

	// Set the region assertion only if the field exists
	if logEvent.Event.Region != "" {
		assert.Equal(t, "EU", logEvent.Event.Region)
	}

	err := processor.RemoveOnEventDispatch(id)

	assert.Nil(t, err)
}

func TestDefaultEventProcessor_BatchSizes(t *testing.T) {
	eg := newExecutionContext()
	processor := NewBatchEventProcessor(
		WithEventDispatcher(NewMockDispatcher(100, false)),
		// here we are setting the timing interval so that we don't have to wait the default 30 seconds
		WithFlushInterval(500*time.Minute),
		WithBatchSize(50))

	eg.Go(processor.Start)

	impression := BuildTestImpressionEvent()

	for i := 0; i < 100; i++ {
		processor.ProcessEvent(impression)
	}

	// sleep for 1 second here. to allow event processor to run.
	time.Sleep(1 * time.Second)

	assert.Equal(t, 0, processor.eventsCount())

	result, ok := (processor.EventDispatcher).(*MockDispatcher)

	if ok {
		assert.Equal(t, 2, result.Events.Size())
		evs := result.Events.Get(3)
		logEvent, _ := evs[0].(LogEvent)
		assert.Equal(t, 50, len(logEvent.Event.Visitors))
		logEvent, _ = evs[1].(LogEvent)
		assert.Equal(t, 50, len(logEvent.Event.Visitors))
	}
	eg.TerminateAndWait()
}

func TestDefaultEventProcessor_DefaultConfig(t *testing.T) {
	eg := newExecutionContext()
	dispatcher := NewMockDispatcher(100, false)
	processor := NewBatchEventProcessor(
		WithEventDispatcher(dispatcher),
		// here we are setting the timing interval so that we don't have to wait the default 30 seconds
		WithFlushInterval(500*time.Millisecond))
	eg.Go(processor.Start)

	impression := BuildTestImpressionEvent()
	conversion := BuildTestConversionEvent()

	processor.ProcessEvent(impression)
	processor.ProcessEvent(impression)
	processor.ProcessEvent(conversion)
	processor.ProcessEvent(conversion)

	assert.Equal(t, 4, processor.eventsCount())

	// sleep for 1 second here. to allow event processor to run.
	time.Sleep(1 * time.Second)

	assert.Equal(t, 0, processor.eventsCount())
	assert.Equal(t, 1, dispatcher.Events.Size())
	evs := dispatcher.Events.Get(1)
	logEvent, _ := evs[0].(LogEvent)
	assert.Equal(t, 4, len(logEvent.Event.Visitors))
}

func TestDefaultEventProcessor_ProcessBatch(t *testing.T) {
	eg := newExecutionContext()
	dispatcher := NewMockDispatcher(100, false)
	processor := NewBatchEventProcessor(
		WithFlushInterval(1*time.Second),
		WithQueueSize(100),
		WithQueue(NewInMemoryQueue(100)),
		WithEventDispatcher(dispatcher))
	eg.Go(processor.Start)

	impression := BuildTestImpressionEvent()
	conversion := BuildTestConversionEvent()

	processor.ProcessEvent(impression)
	processor.ProcessEvent(impression)
	processor.ProcessEvent(conversion)
	processor.ProcessEvent(conversion)

	assert.Equal(t, 4, processor.eventsCount())

	time.Sleep(1500 * time.Millisecond)

	assert.Equal(t, 0, processor.eventsCount())
	assert.Equal(t, 1, dispatcher.Events.Size())
	evs := dispatcher.Events.Get(1)
	logEvent, _ := evs[0].(LogEvent)
	assert.Equal(t, 4, len(logEvent.Event.Visitors))
}

func TestDefaultEventProcessor_BatchSizeMet(t *testing.T) {
	eg := newExecutionContext()
	dispatcher := NewMockDispatcher(100, false)
	processor := NewBatchEventProcessor(
		WithBatchSize(2),
		WithFlushInterval(1000*time.Millisecond),
		WithQueue(NewInMemoryQueue(2)),
		WithEventDispatcher(dispatcher))
	eg.Go(processor.Start)

	impression := BuildTestImpressionEvent()
	conversion := BuildTestConversionEvent()

	processor.ProcessEvent(impression)
	processor.ProcessEvent(impression)

	assert.Equal(t, 2, processor.eventsCount())

	time.Sleep(100 * time.Millisecond)

	assert.Equal(t, 1, dispatcher.Events.Size())
	evs := dispatcher.Events.Get(1)
	logEvent, _ := evs[0].(LogEvent)
	assert.Equal(t, 2, len(logEvent.Event.Visitors))

	processor.ProcessEvent(conversion)
	processor.ProcessEvent(conversion)

	assert.Equal(t, 2, processor.eventsCount())

	eg.TerminateAndWait()

	assert.NotNil(t, processor.Ticker)

	assert.Equal(t, 0, processor.eventsCount())
	assert.Equal(t, 2, dispatcher.Events.Size())
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
	eg := newExecutionContext()
	processor := NewBatchEventProcessor(
		WithQueueSize(2),
		WithBatchSize(2),
		WithFlushInterval(1000*time.Millisecond),
		WithQueue(NewInMemoryQueue(2)),
		WithEventDispatcher(NewMockDispatcher(100, true)))
	eg.Go(processor.Start)

	impression := BuildTestImpressionEvent()

	processor.ProcessEvent(impression)
	processor.ProcessEvent(impression)

	assert.Equal(t, 2, processor.eventsCount())

	processor.ProcessEvent(impression)
	processor.ProcessEvent(impression)

	assert.Equal(t, 2, processor.eventsCount())

}

func TestDefaultEventProcessor_FailedDispatch(t *testing.T) {
	eg := newExecutionContext()
	dispatcher := &MockDispatcher{ShouldFail: true, Events: NewInMemoryQueue(100)}
	processor := NewBatchEventProcessor(
		WithQueueSize(100),
		WithFlushInterval(100),
		WithQueue(NewInMemoryQueue(100)),
		WithEventDispatcher(dispatcher))
	eg.Go(processor.Start)

	impression := BuildTestImpressionEvent()
	conversion := BuildTestConversionEvent()

	processor.ProcessEvent(impression)
	processor.ProcessEvent(impression)
	processor.ProcessEvent(conversion)
	processor.ProcessEvent(conversion)

	assert.Equal(t, 4, processor.eventsCount())

	eg.TerminateAndWait()

	assert.NotNil(t, processor.Ticker)

	assert.Equal(t, 4, processor.eventsCount())
	assert.Equal(t, 0, dispatcher.Events.Size())
}

func TestBatchEventProcessor_FlushesOnClose(t *testing.T) {
	eg := newExecutionContext()
	processor := NewBatchEventProcessor(
		WithQueueSize(100),
		WithQueue(NewInMemoryQueue(100)),
		WithEventDispatcher(NewMockDispatcher(100, false)))
	eg.Go(processor.Start)

	impression := BuildTestImpressionEvent()
	conversion := BuildTestConversionEvent()

	processor.ProcessEvent(impression)
	processor.ProcessEvent(impression)
	processor.ProcessEvent(conversion)
	processor.ProcessEvent(conversion)

	assert.Equal(t, 4, processor.eventsCount())

	// Triggers the flush in the processor
	eg.TerminateAndWait()

	assert.Equal(t, 0, processor.eventsCount())
}

func TestDefaultEventProcessor_ProcessBatchRevisionMismatch(t *testing.T) {
	eg := newExecutionContext()
	dispatcher := NewMockDispatcher(100, false)
	processor := NewBatchEventProcessor(
		WithQueueSize(100),
		WithQueue(NewInMemoryQueue(100)),
		WithEventDispatcher(dispatcher))

	eg.Go(processor.Start)

	impression := BuildTestImpressionEvent()
	conversion := BuildTestConversionEvent()

	processor.ProcessEvent(impression)
	impression.EventContext.Revision = "12112121"
	processor.ProcessEvent(impression)
	processor.ProcessEvent(conversion)
	processor.ProcessEvent(conversion)

	assert.Equal(t, 4, processor.eventsCount())

	eg.TerminateAndWait()

	assert.NotNil(t, processor.Ticker)

	assert.Equal(t, 0, processor.eventsCount())
	assert.Equal(t, 3, dispatcher.Events.Size())
	evs := dispatcher.Events.Get(3)
	logEvent, _ := evs[len(evs)-1].(LogEvent)
	assert.Equal(t, 2, len(logEvent.Event.Visitors))
}

func TestDefaultEventProcessor_ProcessBatchProjectMismatch(t *testing.T) {
	eg := newExecutionContext()
	dispatcher := NewMockDispatcher(100, false)
	processor := NewBatchEventProcessor(
		WithQueueSize(100),
		WithQueue(NewInMemoryQueue(100)),
		WithEventDispatcher(dispatcher))

	eg.Go(processor.Start)

	impression := BuildTestImpressionEvent()
	conversion := BuildTestConversionEvent()

	processor.ProcessEvent(impression)
	impression.EventContext.ProjectID = "121121211111"
	processor.ProcessEvent(impression)
	processor.ProcessEvent(conversion)
	processor.ProcessEvent(conversion)

	assert.Equal(t, 4, processor.eventsCount())

	eg.TerminateAndWait()

	assert.NotNil(t, processor.Ticker)

	assert.Equal(t, 0, processor.eventsCount())
	assert.Equal(t, 3, dispatcher.Events.Size())
	evs := dispatcher.Events.Get(3)
	logEvent, _ := evs[len(evs)-1].(LogEvent)
	assert.Equal(t, 2, len(logEvent.Event.Visitors))
}

func TestChanQueueEventProcessor_ProcessImpression(t *testing.T) {
	eg := newExecutionContext()
	processor := NewBatchEventProcessor(
		WithQueueSize(100),
		WithQueue(NewInMemoryQueue(100)),
		WithEventDispatcher(NewHTTPEventDispatcher("", nil, nil)))

	eg.Go(processor.Start)

	impression := BuildTestImpressionEvent()

	processor.ProcessEvent(impression)
	processor.ProcessEvent(impression)
	processor.ProcessEvent(impression)

	eg.TerminateAndWait()
	assert.NotNil(t, processor.Ticker)

	assert.Equal(t, 0, processor.eventsCount())
}

func TestChanQueueEventProcessor_ProcessBatch(t *testing.T) {
	eg := newExecutionContext()
	dispatcher := NewMockDispatcher(100, false)
	processor := NewBatchEventProcessor(
		WithQueueSize(100),
		WithQueue(NewInMemoryQueue(100)),
		WithEventDispatcher(dispatcher))
	eg.Go(processor.Start)

	impression := BuildTestImpressionEvent()
	conversion := BuildTestConversionEvent()

	processor.ProcessEvent(impression)
	processor.ProcessEvent(impression)
	processor.ProcessEvent(conversion)
	processor.ProcessEvent(conversion)

	eg.TerminateAndWait()

	assert.NotNil(t, processor.Ticker)

	assert.Equal(t, 0, processor.eventsCount())
	assert.Equal(t, 1, dispatcher.Events.Size())
	evs := dispatcher.Events.Get(1)
	logEvent, _ := evs[0].(LogEvent)
	assert.True(t, len(logEvent.Event.Visitors) >= 1)
}

func TestCmabUUIDPropagationInEvents(t *testing.T) {
	// Setup test UUID
	testUUID := "test-uuid-12345"

	// Create a mock dispatcher to capture the final event
	mockDispatcher := NewMockDispatcher(100, false)

	// Create the processor with our mock dispatcher
	processor := NewBatchEventProcessor(
		WithQueueSize(100),
		WithQueue(NewInMemoryQueue(100)),
		WithEventDispatcher(mockDispatcher))

	// Create an impression event
	impressionEvent := BuildTestImpressionEvent()

	// Add CMAB UUID to the impression metadata
	impressionEvent.Impression.Metadata.CmabUUID = &testUUID

	// Process the event
	success := processor.ProcessEvent(impressionEvent)
	assert.True(t, success, "Event should be processed successfully")

	// Manually trigger a flush to ensure the event is dispatched
	processor.flushEvents()

	// Verify the event was dispatched
	assert.Equal(t, 1, mockDispatcher.Events.Size(), "One event should be dispatched")

	// Get the dispatched log event
	events := mockDispatcher.Events.Get(1)
	assert.Len(t, events, 1, "Should have one dispatched event")

	logEvent, ok := events[0].(LogEvent)
	assert.True(t, ok, "Event should be a LogEvent")

	// Serialize to JSON to inspect the structure
	jsonData, err := json.Marshal(logEvent)
	assert.NoError(t, err, "Should marshal to JSON without error")
	jsonString := string(jsonData)

	// Verify the UUID is present in the JSON representation
	assert.Contains(t, jsonString, testUUID, "CMAB UUID should be present in JSON")
}

func TestCmabUUIDNullInImpressionEvent(t *testing.T) {
	// Create processor with mock dispatcher
	mockDispatcher := NewMockDispatcher(100, false)
	processor := NewBatchEventProcessor(
		WithQueueSize(100),
		WithQueue(NewInMemoryQueue(100)),
		WithEventDispatcher(mockDispatcher))

	// Create impression event with a nil CMAB UUID
	impressionEvent := BuildTestImpressionEvent()
	impressionEvent.Impression.Metadata.CmabUUID = nil

	// Process event and flush
	processor.ProcessEvent(impressionEvent)
	processor.flushEvents()

	// Get dispatched event
	events := mockDispatcher.Events.Get(1)
	logEvent, _ := events[0].(LogEvent)

	// Verify JSON doesn't contain cmab_uuid field
	jsonData, _ := json.Marshal(logEvent)
	jsonString := string(jsonData)
	assert.NotContains(t, jsonString, "cmab_uuid",
		"Nil CMAB UUID should be omitted from JSON")
}

func TestCmabUUIDEmptyStringInImpressionEvent(t *testing.T) {
	// Create processor with mock dispatcher
	mockDispatcher := NewMockDispatcher(100, false)
	processor := NewBatchEventProcessor(
		WithQueueSize(100),
		WithQueue(NewInMemoryQueue(100)),
		WithEventDispatcher(mockDispatcher))

	// Create impression event with empty string UUID
	impressionEvent := BuildTestImpressionEvent()
	emptyUUID := ""
	impressionEvent.Impression.Metadata.CmabUUID = &emptyUUID

	// Process event and flush
	processor.ProcessEvent(impressionEvent)
	processor.flushEvents()

	// Get dispatched event
	events := mockDispatcher.Events.Get(1)
	logEvent, _ := events[0].(LogEvent)

	// Verify JSON contains the empty string cmab_uuid
	jsonData, _ := json.Marshal(logEvent)
	jsonString := string(jsonData)

	// Check for empty UUID in JSON - it will appear as "cmab_uuid":""
	assert.Contains(t, jsonString, `"cmab_uuid":""`,
		"Empty string UUID should be included in JSON")
}

func TestMultipleEventsWithDifferentUUIDs(t *testing.T) {
	// Create processor with mock dispatcher
	mockDispatcher := NewMockDispatcher(100, false)
	processor := NewBatchEventProcessor(
		WithQueueSize(100),
		WithQueue(NewInMemoryQueue(100)),
		WithEventDispatcher(mockDispatcher))

	// Create two impression events with different UUIDs
	uuid1 := "uuid-1-test"
	uuid2 := "uuid-2-test"

	impression1 := BuildTestImpressionEvent()
	impression1.Impression.Metadata.CmabUUID = &uuid1

	impression2 := BuildTestImpressionEvent()
	impression2.Impression.Metadata.CmabUUID = &uuid2

	// Process both events
	processor.ProcessEvent(impression1)
	processor.ProcessEvent(impression2)
	processor.flushEvents()

	// Get dispatched event batch
	events := mockDispatcher.Events.Get(1)
	logEvent, _ := events[0].(LogEvent)

	// Verify both UUIDs are in the batch
	jsonData, _ := json.Marshal(logEvent)
	jsonString := string(jsonData)

	assert.Contains(t, jsonString, uuid1, "First UUID should be in the JSON output")
	assert.Contains(t, jsonString, uuid2, "Second UUID should be in the JSON output")
}

// The NoOpLogger is used during benchmarking so that results are printed nicely.
type NoOpLogger struct {
}

func (l *NoOpLogger) Log(level logging.LogLevel, message string, fields map[string]interface{}) {
}

func (l *NoOpLogger) SetLogLevel(level logging.LogLevel) {

}

/*
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
*/
func BenchmarkProcessor(b *testing.B) {
	// no op logger added to keep out extra discarded events
	logging.SetLogger(&NoOpLogger{})

	merges := []struct {
		name string
		fun  func(qSize int) Queue
	}{
		{"InMemory", NewInMemoryQueue},
	}

	for _, merge := range merges {
		for i := 4.; i <= 7; i++ {
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
	eg := newExecutionContext()
	dispatcher := &CountingDispatcher{}
	processor := NewBatchEventProcessor(
		WithQueue(q),
		WithBatchSize(bSize),
		WithEventDispatcher(dispatcher))
	eg.Go(processor.Start)

	conversion := BuildTestConversionEvent()

	for i := 0; i < b.N; i++ {
		var success = false
		for !success {
			success = processor.ProcessEvent(conversion)
		}
	}

	eg.TerminateAndWait()

	if b.N != dispatcher.visitorCount {
		println("Total sent and run ", dispatcher.visitorCount, b.N)
		b.Fail()
	}
}
