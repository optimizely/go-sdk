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

// ChanQueue is a go channel based queue that takes things from the channel and puts them in a in memory queue
type ChanQueue struct {
	ch chan interface{}
	messages Queue
}

// Get returns queue for given count size
func (i *ChanQueue) Get(count int) []interface{} {
	return i.messages.Get(count)
}

// Add appends item to queue
func (i *ChanQueue) Add(item interface{}) {
	i.ch <- item
}

// Remove removes item from queue and returns elements slice
func (i *ChanQueue) Remove(count int) []interface{} {
	return i.messages.Remove(count)

}

// Size returns size of queue
func (i *ChanQueue) Size() int {
	return i.messages.Size()
}

// NewChanQueue returns new go channel based queue with given in memory queueSize
func NewChanQueue(queueSize int) Queue {

	ch := make(chan interface{})

	i := &ChanQueue{ch:ch, messages: NewInMemoryQueue(queueSize)}

	go func() {
		for item := range i.ch {
			i.messages.Add(item)
		}
	}()

	return i
}

