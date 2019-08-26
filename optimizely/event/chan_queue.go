package event

var done = make(chan bool)

type ChanQueue struct {
	ch chan UserEvent
	messages Queue
}

// Get returns queue for given count size
func (i *ChanQueue) Get(count int) []interface{} {

	events := []interface{}{}

	messages := i.messages.Get(count)
	for _, message := range messages {
		events = append(events, message)
	}

	return events
}

func (i *ChanQueue) publish(userEvent UserEvent) {
	i.ch <- userEvent
}

// Add appends item to queue
func (i *ChanQueue) Add(item interface{}) {
	event, ok := item.(UserEvent)
	if !ok {
		// cannot add non-user events
		return
	}

	i.publish(event)
}

// Remove removes item from queue and returns elements slice
func (i *ChanQueue) Remove(count int) []interface{} {
	userEvents := make([]interface{},0, count)
	events := i.messages.Remove(count)
	for _,message := range events {
		userEvents = append(userEvents, message)
	}
	return userEvents
}

// Size returns size of queue
func (i *ChanQueue) Size() int {
	return i.messages.Size()
}

// NewNSQueue returns new NSQ based queue with given queueSize
func NewChanQueue(queueSize int) Queue {

	ch := make(chan UserEvent)

	i := &ChanQueue{ch:ch, messages: NewInMemoryQueue(queueSize)}

	go func() {
		for item := range i.ch {
			i.messages.Add(item)
		}
	}()

	return i
}

