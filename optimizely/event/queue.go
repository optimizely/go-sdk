package event

import (
	"sync"
)

type Queue interface {
	Add(item interface{})
	Remove(count int) [] interface{}
	Get(count int) [] interface{}
	Size() int
}

type InMemoryQueue struct {
	Queue           [] interface{}
	Mux             sync.Mutex
}

func (i *InMemoryQueue) Get(count int) [] interface{} {
	if i.Size() < count {
		count = i.Size()
	}
	i.Mux.Lock()
	defer i.Mux.Unlock()
	return i.Queue[:count]
}

func (i *InMemoryQueue) Add(item interface{}) {
	i.Mux.Lock()
	i.Queue = append(i.Queue, item)
	i.Mux.Unlock()
}

func (i *InMemoryQueue) Remove(count int) [] interface{} {
	if i.Size() < count {
		count = i.Size()
	}
	i.Mux.Lock()
	defer i.Mux.Unlock()
	elem := i.Queue[:count]
	i.Queue = i.Queue[count:]
	return elem
}

func (i *InMemoryQueue) Size() int {
	i.Mux.Lock()
	defer i.Mux.Unlock()
	return len(i.Queue)
}


func NewInMemoryQueue(queueSize int) Queue {
	i := &InMemoryQueue{Queue:make([] interface{}, 0, queueSize)}
	return i
}
