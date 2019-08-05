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
