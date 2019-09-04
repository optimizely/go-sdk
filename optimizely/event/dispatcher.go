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
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/optimizely/go-sdk/optimizely/logging"
)

const jsonContentType = "application/json"
var dispatcherLogger = logging.GetLogger("EventDispatcher")

// Dispatcher dispatches events
type Dispatcher interface {
	DispatchEvent(event LogEvent, callback func(success bool))
}

// HTTPEventDispatcher is the HTTP implementation of the Dispatcher interface
type HTTPEventDispatcher struct {
}

// DispatchEvent dispatches event with callback
func (*HTTPEventDispatcher) DispatchEvent(event LogEvent, callback func(bool)) {
	jsonValue, _ := json.Marshal(event.event)
	resp, err := http.Post(event.endPoint, jsonContentType, bytes.NewBuffer(jsonValue))
	// also check response codes
	// resp.StatusCode == 400 is an error
	var success bool
	if err != nil {
		dispatcherLogger.Error("http.Post failed:", err)
		success = false
	} else {
		if resp.StatusCode == 204 {
			success = true
		} else {
			dispatcherLogger.Error(fmt.Sprintf("http.Post invalid response %d", resp.StatusCode), nil)
			success = false
		}
	}
	callback(success)
}

type QueueEventDispatcher struct {
	eventQueue Queue
	mux        sync.Mutex
	dispatcher *HTTPEventDispatcher
}

func (ed *QueueEventDispatcher) DispatchEvent(event LogEvent, callback func(bool)) {

	ed.eventQueue.Add(event)
	callback(true)
	go func() {
		ed.flushEvents()
	}()
}

func (ed *QueueEventDispatcher) flushEvents() {

	ed.mux.Lock()

	defer func(){
		ed.mux.Unlock()
	}()

	for ed.eventQueue.Size() > 0 {
		items := ed.eventQueue.Get(1)
		if len(items) == 0 {
			time.Sleep(5 * time.Second)
			continue
		}
		event, ok := items[0].(LogEvent)
		if !ok {
			//remove it?
			dispatcherLogger.Error("invalid type passed to event dispatcher", nil)
			ed.eventQueue.Remove(1)
			continue
		}

		ed.dispatcher.DispatchEvent(event, func(success bool) {
			if success {
				ed.eventQueue.Remove(1)
			} else {
				// we failed.  Sleep 5 seconds and try again.  Or, should we exit and try the next time
				// a log event is added?
				time.Sleep(5 * time.Second)
			}
		})
	}
}

func NewQueueEventDispatcher(ctx context.Context) Dispatcher {
	dispatcher := &QueueEventDispatcher{eventQueue: NewInMemoryQueue(1000), dispatcher:&HTTPEventDispatcher{}}

	go func() {
		_ = <-ctx.Done()
		dispatcher.flushEvents()
		return
	}()

	return dispatcher
}
