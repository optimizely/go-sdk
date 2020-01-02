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
	"github.com/optimizely/go-sdk/pkg/metrics"
	"net/http"
	"sync"
	"time"

	"github.com/optimizely/go-sdk/pkg/logging"
	"github.com/optimizely/go-sdk/pkg/utils"
)

const maxRetries = 3
const defaultQueueSize = 1000
const sleepTime = 1 * time.Second

var dispatcherLogger = logging.GetLogger("EventDispatcher")

// Dispatcher dispatches events
type Dispatcher interface {
	DispatchEvent(event LogEvent) (bool, error)
}

// HTTPEventDispatcher is the HTTP implementation of the Dispatcher interface
type HTTPEventDispatcher struct {
	requester *utils.HTTPRequester
}

// DispatchEvent dispatches event with callback
func (ed *HTTPEventDispatcher) DispatchEvent(event LogEvent) (bool, error) {

	_, _, code, err := ed.requester.Post(event.EndPoint, event.Event)

	// also check response codes
	// resp.StatusCode == 400 is an error
	var success bool
	if err != nil {
		dispatcherLogger.Error("http.Post failed:", err)
		success = false
	} else {
		if code == http.StatusNoContent {
			success = true
		} else {
			dispatcherLogger.Error(fmt.Sprintf("http.Post invalid response %d", code), nil)
			success = false
		}
	}
	return success, err
}

// QueueEventDispatcher is a queued version of the event Dispatcher that queues, returns success, and dispatches events in the background
type QueueEventDispatcher struct {
	eventQueue     Queue
	eventFlushLock sync.Mutex
	Dispatcher     Dispatcher

	stats metrics.GenericMetrics
}

// DispatchEvent queues event with callback and calls flush in a go routine.
func (ed *QueueEventDispatcher) DispatchEvent(event LogEvent) (bool, error) {
	ed.eventQueue.Add(event)
	go func() {
		ed.flushEvents()
	}()
	return true, nil
}

// flush the events
func (ed *QueueEventDispatcher) flushEvents() {

	ed.eventFlushLock.Lock()

	defer func() {
		ed.eventFlushLock.Unlock()
	}()

	retryCount := 0

	ed.stats.Set("queueSize", int64(ed.eventQueue.Size()))
	for ed.eventQueue.Size() > 0 {
		if retryCount > maxRetries {
			dispatcherLogger.Error(fmt.Sprintf("event failed to send %d times. It will retry on next event sent", maxRetries), nil)
			ed.stats.Inc("failFlush")
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
			dispatcherLogger.Error("invalid type passed to event Dispatcher", nil)
			ed.eventQueue.Remove(1)
			ed.stats.Inc("failFlush")
			continue
		}

		success, err := ed.Dispatcher.DispatchEvent(event)

		if err == nil {
			if success {
				dispatcherLogger.Debug(fmt.Sprintf("Dispatched log event %+v", event))
				ed.eventQueue.Remove(1)
				retryCount = 0
				ed.stats.Inc("successFlush")
			} else {
				dispatcherLogger.Warning("dispatch event failed")
				// we failed.  Sleep some seconds and try again.
				time.Sleep(sleepTime)
				// increase retryCount.  We exit if we have retried x times.
				// we will retry again next event that is added.
				retryCount++
				ed.stats.Inc("retryFlush")
			}
		} else {
			dispatcherLogger.Error("Error dispatching ", err)
			// we failed.  Sleep some seconds and try again.
			time.Sleep(sleepTime)
			// increase retryCount.  We exit if we have retried x times.
			// we will retry again next event that is added.
			retryCount++
			ed.stats.Inc("retryFlush")
		}
	}
	ed.stats.Set("queueSize", int64(ed.eventQueue.Size()))
}

// NewQueueEventDispatcher creates a Dispatcher that queues in memory and then sends via go routine.
func NewQueueEventDispatcher(stats metrics.GenericMetrics) *QueueEventDispatcher {

	var dispatcherStats metrics.GenericMetrics
	if stats != nil {
		dispatcherStats = stats
	} else {
		dispatcherStats = metrics.NewMetrics() // protective code to set
	}
	return &QueueEventDispatcher{eventQueue: NewInMemoryQueue(defaultQueueSize), Dispatcher: &HTTPEventDispatcher{requester: utils.NewHTTPRequester()}, stats: dispatcherStats}
}
