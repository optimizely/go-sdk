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
	"fmt"
	"net/http"
	"time"

	"golang.org/x/sync/semaphore"

	"github.com/optimizely/go-sdk/pkg/logging"
	"github.com/optimizely/go-sdk/pkg/metrics"
	"github.com/optimizely/go-sdk/pkg/utils"
)

const maxWorkers = int64(1)
const maxRetries = 3
const defaultQueueSize = 1000
const sleepTime = 1 * time.Second

// Dispatcher dispatches events
type Dispatcher interface {
	DispatchEvent(event LogEvent) (bool, error)
}

// httpEventDispatcher is the HTTP implementation of the Dispatcher interface
type httpEventDispatcher struct {
	requester *utils.HTTPRequester
	logger    logging.OptimizelyLogProducer
}

// DispatchEvent dispatches event with callback
func (ed *httpEventDispatcher) DispatchEvent(event LogEvent) (bool, error) {

	_, _, code, err := ed.requester.Post(event.EndPoint, event.Event)

	// also check response codes
	// resp.StatusCode == 400 is an error
	var success bool
	if err != nil {
		ed.logger.Error("http.Post failed:", err)
		success = false
	} else {
		if code == http.StatusNoContent {
			success = true
		} else {
			ed.logger.Error(fmt.Sprintf("http.Post invalid response %d", code), nil)
			success = false
		}
	}
	return success, err
}

// NewHTTPEventDispatcher creates a full http dispatcher. The requester and logger parameters can be nil.
func NewHTTPEventDispatcher(sdkKey string, requester *utils.HTTPRequester, logger logging.OptimizelyLogProducer) Dispatcher {
	if requester == nil {
		requester = utils.NewHTTPRequester(logging.GetLogger(sdkKey, "HTTPRequester"))
	}
	if logger == nil {
		logger = logging.GetLogger(sdkKey, "httpEventDispatcher")
	}

	return &httpEventDispatcher{requester: requester, logger: logger}
}

// QueueEventDispatcher is a queued version of the event Dispatcher that queues, returns success, and dispatches events in the background
type QueueEventDispatcher struct {
	eventQueue Queue
	processing *semaphore.Weighted
	Dispatcher Dispatcher
	logger     logging.OptimizelyLogProducer

	// metrics
	queueSizeGauge     metrics.Gauge
	sucessFlushCounter metrics.Counter
	failFlushCounter   metrics.Counter
	retryFlushCounter  metrics.Counter
}

// DispatchEvent queues event with callback and calls flush in a go routine.
func (ed *QueueEventDispatcher) DispatchEvent(event LogEvent) (bool, error) {
	ed.eventQueue.Add(event)
	go ed.flushEvents()
	return true, nil
}

// flush the events
func (ed *QueueEventDispatcher) flushEvents() {

	// Limit flushing to a single worker
	if !ed.processing.TryAcquire(1) {
		return
	}
	defer ed.processing.Release(1)

	retryCount := 0
	queueSize := ed.eventQueue.Size()
	for ; queueSize > 0; queueSize = ed.eventQueue.Size() {
		ed.queueSizeGauge.Set(float64(queueSize))
		if retryCount > maxRetries {
			ed.logger.Error(fmt.Sprintf("event failed to send %d times. It will retry on next event sent", maxRetries), nil)
			ed.failFlushCounter.Add(1)
			break
		}

		items := ed.eventQueue.Get(1)
		if len(items) == 0 {
			// something happened.  Just continue and you should expect size to be zero.
			continue
		}
		event, ok := items[0].(LogEvent)
		if !ok {
			// remove it
			ed.logger.Error("invalid type passed to event Dispatcher", nil)
			ed.eventQueue.Remove(1)
			ed.failFlushCounter.Add(1)
			continue
		}

		success, err := ed.Dispatcher.DispatchEvent(event)

		if err == nil {
			if success {
				ed.logger.Debug("dispatch log event succeeded")
				ed.eventQueue.Remove(1)
				retryCount = 0
				ed.sucessFlushCounter.Add(1)
			} else {
				ed.logger.Warning("dispatch event failed")
				// we failed.  Sleep some seconds and try again.
				// increase retryCount.  We exit if we have retried x times.
				// we will retry again next event that is added.
				retryCount++
				ed.retryFlushCounter.Add(1)
				time.Sleep(sleepTime)
			}
		} else {
			ed.logger.Error("Error dispatching ", err)
			// we failed.  Sleep some seconds and try again.
			time.Sleep(sleepTime)
			// increase retryCount.  We exit if we have retried x times.
			// we will retry again next event that is added.
			retryCount++
			ed.retryFlushCounter.Add(1)
		}
	}
	ed.queueSizeGauge.Set(float64(queueSize))
}

// NewQueueEventDispatcher creates a Dispatcher that queues in memory and then sends via go routine.
func NewQueueEventDispatcher(sdkKey string, metricsRegistry metrics.Registry) *QueueEventDispatcher {

	var dispatcherMetricsRegistry metrics.Registry
	if metricsRegistry != nil {
		dispatcherMetricsRegistry = metricsRegistry
	} else {
		dispatcherMetricsRegistry = metrics.NewNoopRegistry() // protective code to set
	}

	logger := logging.GetLogger(sdkKey, "QueueEventDispatcher")
	return &QueueEventDispatcher{
		eventQueue:         NewInMemoryQueueWithLogger(defaultQueueSize, logger),
		Dispatcher:         NewHTTPEventDispatcher(sdkKey, nil, nil),
		queueSizeGauge:     dispatcherMetricsRegistry.GetGauge(metrics.DispatcherQueueSize),
		retryFlushCounter:  dispatcherMetricsRegistry.GetCounter(metrics.DispatcherRetryFlush),
		failFlushCounter:   dispatcherMetricsRegistry.GetCounter(metrics.DispatcherFailedFlush),
		sucessFlushCounter: dispatcherMetricsRegistry.GetCounter(metrics.DispatcherSuccessFlush),
		logger:             logger,
		processing:         semaphore.NewWeighted(maxWorkers),
	}
}
