/****************************************************************************
 * Copyright 2019-2020,2022-2023 Optimizely, Inc. and contributors          *
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
	"sync"
	"testing"
	"time"

	"github.com/optimizely/go-sdk/v2/pkg/entities"
	"github.com/optimizely/go-sdk/v2/pkg/metrics"

	"github.com/stretchr/testify/assert"
)

type MetricsRegistry struct {
	metricsCounterVars map[string]*MetricsCounter
	metricsGaugeVars   map[string]*MetricsGauge

	gaugeLock   sync.Mutex
	counterLock sync.Mutex
}

func (m *MetricsRegistry) GetCounter(key string) metrics.Counter {
	m.counterLock.Lock()
	defer m.counterLock.Unlock()
	if counter, ok := m.metricsCounterVars[key]; ok {
		return counter
	}

	counter := &MetricsCounter{}
	m.metricsCounterVars[key] = counter
	return counter
}

func (m *MetricsRegistry) GetGauge(key string) metrics.Gauge {
	m.gaugeLock.Lock()
	defer m.gaugeLock.Unlock()
	if gauge, ok := m.metricsGaugeVars[key]; ok {
		return gauge
	}
	gauge := &MetricsGauge{}
	m.metricsGaugeVars[key] = gauge
	return gauge
}

type MetricsCounter struct {
	f    float64
	lock sync.Mutex
}

func (m *MetricsCounter) Add(value float64) {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.f += value
}

func (m *MetricsCounter) Get() float64 {
	m.lock.Lock()
	defer m.lock.Unlock()
	return m.f
}

type MetricsGauge struct {
	f    float64
	lock sync.Mutex
}

func (m *MetricsGauge) Set(value float64) {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.f = value
}

func (m *MetricsGauge) Get() float64 {
	m.lock.Lock()
	defer m.lock.Unlock()
	return m.f
}
func NewMetricsRegistry() *MetricsRegistry {

	return &MetricsRegistry{
		metricsCounterVars: map[string]*MetricsCounter{},
		metricsGaugeVars:   map[string]*MetricsGauge{},
	}
}

func TestQueueEventDispatcher_DispatchEvent(t *testing.T) {
	metricsRegistry := NewMetricsRegistry()

	q := NewQueueEventDispatcher("", metricsRegistry)

	assert.True(t, q.Dispatcher != nil)
	if d, ok := q.Dispatcher.(*httpEventDispatcher); ok {
		assert.True(t, d.requester != nil && d.logger != nil)
	} else {
		assert.True(t, false)
	}
	sender := &MockDispatcher{Events: NewInMemoryQueue(100)}
	q.Dispatcher = sender

	eventTags := map[string]interface{}{"revenue": 55.0, "value": 25.1}
	config := TestConfig{}

	conversionUserEvent := CreateConversionUserEvent(config, entities.Event{ExperimentIds: []string{"15402980349"}, ID: "15368860886", Key: "sample_conversion"}, userContext, eventTags)

	batch := createBatchEvent(conversionUserEvent, createVisitorFromUserEvent(conversionUserEvent))
	assert.Equal(t, conversionUserEvent.Timestamp, batch.Visitors[0].Snapshots[0].Events[0].Timestamp)

	logEvent := createLogEvent(batch, EventEndPoints["US"])

	success, _ := q.DispatchEvent(logEvent)

	assert.True(t, success)

	// its been queued
	assert.True(t, (q.eventQueue.Size() == 1 && sender.Events.Size() == 0) || (q.eventQueue.Size() == 0 && sender.Events.Size() == 1))

	// give the queue a chance to run
	time.Sleep(1 * time.Second)

	// check the queue
	assert.Equal(t, 0, q.eventQueue.Size())
	assert.Equal(t, 1, sender.Events.Size())

	assert.Equal(t, float64(0), metricsRegistry.GetGauge(metrics.DispatcherQueueSize).(*MetricsGauge).Get())
	assert.Equal(t, float64(1), metricsRegistry.GetCounter(metrics.DispatcherSuccessFlush).(*MetricsCounter).Get())
	assert.Equal(t, float64(0), metricsRegistry.GetCounter(metrics.DispatcherFailedFlush).(*MetricsCounter).Get())
	assert.Equal(t, float64(0), metricsRegistry.GetCounter(metrics.DispatcherRetryFlush).(*MetricsCounter).Get())

}

func TestQueueEventDispatcher_InvalidEvent(t *testing.T) {
	metricsRegistry := NewMetricsRegistry()
	q := NewQueueEventDispatcher("", metricsRegistry)
	q.Dispatcher = &MockDispatcher{Events: NewInMemoryQueue(100)}

	config := TestConfig{}
	q.eventQueue.Add(config)

	assert.Equal(t, 1, q.eventQueue.Size())

	// give the queue a chance to run
	q.flushEvents()

	// check the queue. bad event type should be removed.  but, not sent.
	assert.Equal(t, 0, q.eventQueue.Size())

	assert.Equal(t, float64(0), metricsRegistry.GetGauge(metrics.DispatcherQueueSize).(*MetricsGauge).Get())
	assert.Equal(t, float64(0), metricsRegistry.GetCounter(metrics.DispatcherSuccessFlush).(*MetricsCounter).Get())
	assert.Equal(t, float64(1), metricsRegistry.GetCounter(metrics.DispatcherFailedFlush).(*MetricsCounter).Get())
	assert.Equal(t, float64(0), metricsRegistry.GetCounter(metrics.DispatcherRetryFlush).(*MetricsCounter).Get())

}

func TestQueueEventDispatcher_FailDispath(t *testing.T) {
	metricsRegistry := NewMetricsRegistry()
	q := NewQueueEventDispatcher("", metricsRegistry)
	q.Dispatcher = &MockDispatcher{ShouldFail: true, Events: NewInMemoryQueue(100)}

	eventTags := map[string]interface{}{"revenue": 55.0, "value": 25.1}
	config := TestConfig{}

	conversionUserEvent := CreateConversionUserEvent(config, entities.Event{ExperimentIds: []string{"15402980349"}, ID: "15368860886", Key: "sample_conversion"}, userContext, eventTags)

	batch := createBatchEvent(conversionUserEvent, createVisitorFromUserEvent(conversionUserEvent))
	assert.Equal(t, conversionUserEvent.Timestamp, batch.Visitors[0].Snapshots[0].Events[0].Timestamp)

	logEvent := createLogEvent(batch, EventEndPoints["US"])

	q.DispatchEvent(logEvent)

	assert.Equal(t, 1, q.eventQueue.Size())

	// give the queue a chance to run the queue is drained asynchronously
	retryCount, _ := metricsRegistry.GetCounter(metrics.DispatcherRetryFlush).(*MetricsCounter)
	assert.Eventually(t, func() bool { return retryCount.Get() > 1 }, 5*time.Second, 1*time.Second)

	// check the queue. the event should still be in the queue
	assert.Equal(t, 1, q.eventQueue.Size())

	assert.Equal(t, float64(1), metricsRegistry.GetGauge(metrics.DispatcherQueueSize).(*MetricsGauge).Get())
	assert.Equal(t, float64(0), metricsRegistry.GetCounter(metrics.DispatcherSuccessFlush).(*MetricsCounter).Get())
}

func TestQueueEventDispatcher_WaitForDispatchingEventsOnClose(t *testing.T) {
	metricsRegistry := NewMetricsRegistry()

	q := NewQueueEventDispatcher("", metricsRegistry)

	assert.True(t, q.Dispatcher != nil)
	if d, ok := q.Dispatcher.(*httpEventDispatcher); ok {
		assert.True(t, d.requester != nil && d.logger != nil)
	} else {
		assert.True(t, false)
	}
	sender := &MockDispatcher{Events: NewInMemoryQueue(100), eventsQueue: NewInMemoryQueue(100)}
	q.Dispatcher = sender

	eventTags := map[string]interface{}{"revenue": 55.0, "value": 25.1}
	config := TestConfig{}

	for i := 0; i < 10; i++ {
		conversionUserEvent := CreateConversionUserEvent(config, entities.Event{ExperimentIds: []string{"15402980349"}, ID: "15368860886", Key: "sample_conversion"}, userContext, eventTags)

		batch := createBatchEvent(conversionUserEvent, createVisitorFromUserEvent(conversionUserEvent))
		assert.Equal(t, conversionUserEvent.Timestamp, batch.Visitors[0].Snapshots[0].Events[0].Timestamp)

		logEvent := createLogEvent(batch, EventEndPoints["US"])

		success, _ := q.DispatchEvent(logEvent)

		assert.True(t, success)
	}

	// wait for the events to be dispatched
	q.waitForDispatchingEventsOnClose(10 * time.Second)

	// check the queue
	assert.Equal(t, 0, q.eventQueue.Size())
}

func TestGetRetryInterval(t *testing.T) {
	tests := []struct {
		name       string
		retryCount int
		expected   time.Duration
	}{
		{"first retry", 0, 200 * time.Millisecond},
		{"second retry", 1, 400 * time.Millisecond},
		{"third retry", 2, 800 * time.Millisecond},
		{"fourth retry capped at max", 3, 1 * time.Second},
		{"fifth retry capped at max", 4, 1 * time.Second},
		{"high retry count capped at max", 10, 1 * time.Second},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getRetryInterval(tt.retryCount)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetRetryInterval_ExponentialGrowth(t *testing.T) {
	// Verify exponential growth pattern: each interval should be double the previous
	prev := getRetryInterval(0)
	for i := 1; i <= 2; i++ {
		curr := getRetryInterval(i)
		assert.Equal(t, prev*2, curr, "retry %d should be double retry %d", i, i-1)
		prev = curr
	}
}

func TestQueueEventDispatcher_ExponentialBackoff(t *testing.T) {
	metricsRegistry := NewMetricsRegistry()
	q := NewQueueEventDispatcher("", metricsRegistry)
	q.Dispatcher = &MockDispatcher{ShouldFail: true, Events: NewInMemoryQueue(100)}

	eventTags := map[string]interface{}{"revenue": 55.0, "value": 25.1}
	config := TestConfig{}

	conversionUserEvent := CreateConversionUserEvent(config, entities.Event{ExperimentIds: []string{"15402980349"}, ID: "15368860886", Key: "sample_conversion"}, userContext, eventTags)
	batch := createBatchEvent(conversionUserEvent, createVisitorFromUserEvent(conversionUserEvent))
	logEvent := createLogEvent(batch, EventEndPoints["US"])

	startTime := time.Now()
	_, _ = q.DispatchEvent(logEvent)

	// Wait for 3 retries to complete (200ms + 400ms = 600ms minimum for 2 delays between 3 attempts)
	retryCounter := metricsRegistry.GetCounter(metrics.DispatcherRetryFlush).(*MetricsCounter)
	assert.Eventually(t, func() bool { return retryCounter.Get() >= 3 }, 5*time.Second, 100*time.Millisecond)

	elapsed := time.Since(startTime)
	// With exponential backoff: 200ms + 400ms = 600ms minimum for delays between 3 attempts
	// Allow some tolerance for test execution overhead
	assert.True(t, elapsed >= 500*time.Millisecond, "Expected at least 500ms elapsed for exponential backoff, got %v", elapsed)
}
