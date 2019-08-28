package event

type ChanQueue struct {
	ch chan interface{}
	messages Queue
	isValid func(value interface{}) bool
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

// NewNSQueue returns new NSQ based queue with given queueSize
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

