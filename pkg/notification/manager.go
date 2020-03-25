/****************************************************************************
 * Copyright 2019-2020, Optimizely, Inc. and contributors                   *
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
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/optimizely/go-sdk/pkg/logging"
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
	lock     sync.RWMutex
	logger   logging.OptimizelyLogProducer
}

// NewAtomicManager creates a new instance of the atomic manager
func NewAtomicManager(logger logging.OptimizelyLogProducer) *AtomicManager {
	return &AtomicManager{
		logger: logger,
		handlers: make(map[uint32]func(interface{})),
	}
}

// Add adds the given handler
func (am *AtomicManager) Add(newHandler func(interface{})) (int, error) {
	am.lock.Lock()
	defer am.lock.Unlock()

	atomic.AddUint32(&am.counter, 1)
	am.handlers[am.counter] = newHandler
	return int(am.counter), nil
}

// Remove removes handler with the given id
func (am *AtomicManager) Remove(id int) {
	am.lock.Lock()
	defer am.lock.Unlock()

	handlerID := uint32(id)
	if _, ok := am.handlers[handlerID]; ok {
		delete(am.handlers, handlerID)
		return
	}
	am.logger.Debug(fmt.Sprintf("Handler for id:%d not found", id))

}

// Send sends the notification to the registered handlers
func (am *AtomicManager) Send(notification interface{}) {
	// copying handler to avoid race condition
	handlers := am.copyHandlers()
	for _, handler := range handlers {
		handler(notification)
	}
}

// Return a copy of the given handlers
func (am *AtomicManager) copyHandlers() (handlers []func(interface{})) {
	am.lock.RLock()
	defer am.lock.RUnlock()
	for _, v := range am.handlers {
		handlers = append(handlers, v)
	}
	return handlers
}
