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
	"sync"

	"github.com/optimizely/go-sdk/pkg/logging"
)

// Queue represents a queue
type Queue interface {
	Add(item interface{}) // TODO Should return a bool
	Remove(count int) []interface{}
	Get(count int) []interface{}
	Size() int
}

// InMemoryQueue represents a in-memory queue
type InMemoryQueue struct {
	logger  logging.OptimizelyLogProducer
	MaxSize int
	Queue   []interface{}
	Mux     sync.Mutex
}

// Get returns queue for given count size
func (q *InMemoryQueue) Get(count int) []interface{} {
	q.Mux.Lock()
	defer q.Mux.Unlock()

	count = q.getSafeCount(count)
	return q.Queue[:count]
}

// Add appends item to queue
func (q *InMemoryQueue) Add(item interface{}) {
	q.Mux.Lock()
	defer q.Mux.Unlock()

	if len(q.Queue) >= q.MaxSize {
		q.logger.Warning("MaxQueueSize has been met. Discarding event")
		return
	}

	q.Queue = append(q.Queue, item)
}

// Remove removes item from queue and returns elements slice
func (q *InMemoryQueue) Remove(count int) []interface{} {
	q.Mux.Lock()
	defer q.Mux.Unlock()

	count = q.getSafeCount(count)
	elem := q.Queue[:count]
	q.Queue = q.Queue[count:]
	return elem
}

func (q *InMemoryQueue) getSafeCount(count int) int {
	if size := len(q.Queue); size < count {
		return size
	}

	return count
}

// Size returns size of queue
func (q *InMemoryQueue) Size() int {
	q.Mux.Lock()
	defer q.Mux.Unlock()
	return len(q.Queue)
}

// NewInMemoryQueue returns new InMemoryQueue with given queueSize
func NewInMemoryQueue(queueSize int) Queue {
	logger := logging.GetLogger("", "InMemoryQueue")
	return NewInMemoryQueueWithLogger(queueSize, logger)
}

// NewInMemoryQueueWithLogger returns new InMemoryQueue with given queueSize and logger
func NewInMemoryQueueWithLogger(queueSize int, logger logging.OptimizelyLogProducer) Queue {
	return &InMemoryQueue{
		logger:  logger,
		MaxSize: queueSize,
		Queue:   make([]interface{}, 0, queueSize),
	}
}
