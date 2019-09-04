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
)

// Queue represents a queue
type Queue interface {
	Add(item interface{})
	Remove(count int) []interface{}
	Get(count int) []interface{}
	Size() int
}

// InMemoryQueue represents a in-memory queue
type InMemoryQueue struct {
	Queue []interface{}
	Mux   sync.Mutex
}

// Get returns queue for given count size
func (i *InMemoryQueue) Get(count int) []interface{} {
	if i.Size() < count {
		count = i.Size()
	}
	i.Mux.Lock()
	defer i.Mux.Unlock()
	return i.Queue[:count]
}

// Add appends item to queue
func (i *InMemoryQueue) Add(item interface{}) {
	i.Mux.Lock()
	i.Queue = append(i.Queue, item)
	i.Mux.Unlock()
}

// Remove removes item from queue and returns elements slice
func (i *InMemoryQueue) Remove(count int) []interface{} {
	if i.Size() < count {
		count = i.Size()
	}
	i.Mux.Lock()
	defer i.Mux.Unlock()
	elem := i.Queue[:count]
	i.Queue = i.Queue[count:]
	return elem
}

// Size returns size of queue
func (i *InMemoryQueue) Size() int {
	i.Mux.Lock()
	defer i.Mux.Unlock()
	return len(i.Queue)
}

// NewInMemoryQueue returns new InMemoryQueue with given queueSize
func NewInMemoryQueue(queueSize int) Queue {
	i := &InMemoryQueue{Queue: make([]interface{}, 0, queueSize)}
	return i
}
