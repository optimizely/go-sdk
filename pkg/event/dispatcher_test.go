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
	"testing"
	"time"

	"github.com/optimizely/go-sdk/pkg/entities"
	"github.com/optimizely/go-sdk/pkg/metrics"
	"github.com/stretchr/testify/assert"
)

func TestQueueEventDispatcher_DispatchEvent(t *testing.T) {
	stats := metrics.NewMetrics()

	q := NewQueueEventDispatcher(stats)
	q.Dispatcher = &MockDispatcher{Events: NewInMemoryQueue(100)}

	eventTags := map[string]interface{}{"revenue": 55.0, "value": 25.1}
	config := TestConfig{}

	conversionUserEvent := CreateConversionUserEvent(config, entities.Event{ExperimentIds: []string{"15402980349"}, ID: "15368860886", Key: "sample_conversion"}, userContext, eventTags)

	batch := createBatchEvent(conversionUserEvent, createVisitorFromUserEvent(conversionUserEvent))

	logEvent := createLogEvent(batch)

	success, _ := q.DispatchEvent(logEvent)

	assert.True(t, success)

	// its been queued
	assert.Equal(t, 1, q.eventQueue.Size())

	// give the queue a chance to run
	time.Sleep(1 * time.Second)

	// check the queue
	assert.Equal(t, 0, q.eventQueue.Size())

	assert.Equal(t, int64(1), stats.Get("successFlush"))
	assert.Equal(t, int64(0), stats.Get("retryFlush"))
	assert.Equal(t, int64(0), stats.Get("queueSize"))
	assert.Equal(t, int64(0), stats.Get("failFlush"))

}

func TestQueueEventDispatcher_InvalidEvent(t *testing.T) {
	stats := metrics.NewMetrics()
	q := NewQueueEventDispatcher(stats)
	q.Dispatcher = &MockDispatcher{Events: NewInMemoryQueue(100)}

	config := TestConfig{}
	q.eventQueue.Add(config)

	assert.Equal(t, 1, q.eventQueue.Size())

	// give the queue a chance to run
	q.flushEvents()

	// check the queue. bad event type should be removed.  but, not sent.
	assert.Equal(t, 0, q.eventQueue.Size())

	assert.Equal(t, int64(0), stats.Get("successFlush"))
	assert.Equal(t, int64(0), stats.Get("retryFlush"))
	assert.Equal(t, int64(0), stats.Get("queueSize"))
	assert.Equal(t, int64(1), stats.Get("failFlush"))

}

func TestQueueEventDispatcher_FailDispath(t *testing.T) {
	stats := metrics.NewMetrics()
	q := NewQueueEventDispatcher(stats)
	q.Dispatcher = &MockDispatcher{ShouldFail: true, Events: NewInMemoryQueue(100)}

	eventTags := map[string]interface{}{"revenue": 55.0, "value": 25.1}
	config := TestConfig{}

	conversionUserEvent := CreateConversionUserEvent(config, entities.Event{ExperimentIds: []string{"15402980349"}, ID: "15368860886", Key: "sample_conversion"}, userContext, eventTags)

	batch := createBatchEvent(conversionUserEvent, createVisitorFromUserEvent(conversionUserEvent))

	logEvent := createLogEvent(batch)

	q.DispatchEvent(logEvent)

	assert.Equal(t, 1, q.eventQueue.Size())

	// give the queue a chance to run
	q.flushEvents()
	time.Sleep(1 * time.Second)

	// check the queue. bad event type should be removed.  but, not sent.
	assert.Equal(t, 1, q.eventQueue.Size())

	assert.Equal(t, int64(0), stats.Get("successFlush"))
	assert.True(t, stats.Get("retryFlush") > 1)
	assert.Equal(t, int64(1), stats.Get("queueSize"))

	q.flushEvents()

	// check the queue. bad event type should be removed.  but, not sent.
	assert.Equal(t, 1, q.eventQueue.Size())
}
