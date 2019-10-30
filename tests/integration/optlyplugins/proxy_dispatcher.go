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

package optlyplugins

import (
	"github.com/optimizely/go-sdk/pkg/event"
)

// EventReceiver returns dispatched events
type EventReceiver interface {
	GetEvents() []event.Batch
}

// ProxyEventDispatcher represents a valid HTTP implementation of the Dispatcher interface
type ProxyEventDispatcher struct {
	events []event.Batch
}

// DispatchEvent dispatches event with callback
func (d *ProxyEventDispatcher) DispatchEvent(event event.LogEvent) (bool, error) {
	d.events = append(d.events, event.Event)
	return true, nil
}

// GetEvents returns dispatched events
func (d *ProxyEventDispatcher) GetEvents() []event.Batch {
	if d.events == nil {
		d.events = []event.Batch{}
	}
	return d.events
}

// ClearEvents deletes dispatched events
func (d *ProxyEventDispatcher) ClearEvents() {
	d.events = nil
}
