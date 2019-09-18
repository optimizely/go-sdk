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

// Package notification //
package notification

import (
	"sync/atomic"
)

// Manager is a generic interface for managing notifications of a particular type
type Manager interface {
	Add(func(interface{})) (int, error)
	Remove(id int)
	Send(message interface{})
}

// AtomicManager adds handlers atomically
type AtomicManager struct {
	handlers map[uint32]func(interface{})
	counter  uint32
}

// NewAtomicManager creates a new instance of the atomic manager
func NewAtomicManager() *AtomicManager {
	return &AtomicManager{
		handlers: make(map[uint32]func(interface{})),
	}
}

// Add adds the given handler
func (am *AtomicManager) Add(newHandler func(interface{})) (int, error) {
	atomic.AddUint32(&am.counter, 1)
	am.handlers[am.counter] = newHandler
	return int(am.counter), nil
}

// Remove removes handler with the given id
func (am *AtomicManager) Remove(id int) {
	handlerID := uint32(id)
	if _, ok := am.handlers[handlerID]; ok {
		delete(am.handlers, handlerID)
	}
}

// Send sends the notification to the registered handlers
func (am *AtomicManager) Send(notification interface{}) {
	for _, handler := range am.handlers {
		handler(notification)
	}
}
