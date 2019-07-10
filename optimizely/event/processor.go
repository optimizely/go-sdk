package event

import (
	"fmt"
	"sync"
	"time"
)

// Processor processes events
type Processor interface {
	ProcessEvent(event UserEvent)
}

// QueueingEventProcessor is used out of the box by the SDK
type QueueingEventProcessor struct {
	MaxQueueSize    int           // max size of the queue before flush
	FlushInterval   time.Duration // in milliseconds
	BatchSize       int
	Q               Queue
	Mux             sync.Mutex
	Ticker          *time.Ticker
	EventDispatcher Dispatcher
}

func NewEventProcessor(queueSize int, flushInterval time.Duration ) Processor {
	p := &QueueingEventProcessor{MaxQueueSize: queueSize, FlushInterval:flushInterval, Q:NewInMemoryQueue(queueSize), EventDispatcher:&HttpEventDispatcher{}}
	p.StartTicker()
	return p
}

// ProcessEvent processes the given impression event
func (p *QueueingEventProcessor) ProcessEvent(event UserEvent) {
	p.Q.Add(event)

	if p.Q.Size() >= p.MaxQueueSize {
		go func() {
			p.FlushEvents()
		}()
	}
}

func (p *QueueingEventProcessor) EventsCount() int {
	return p.Q.Size()
}

func (p *QueueingEventProcessor) GetEvents(count int) []interface{} {
	return p.Q.Get(count)
}

func (p *QueueingEventProcessor) Remove(count int) []interface{} {
	return p.Q.Remove(count)
}

func (p *QueueingEventProcessor) StartTicker() {
	if p.Ticker != nil {
		return
	}
	p.Ticker = time.NewTicker(p.FlushInterval * time.Millisecond)
	go func() {
		for _ = range p.Ticker.C {
			p.FlushEvents()
		}
	}()
}

// ProcessEvent processes the given impression event
func (p *QueueingEventProcessor) FlushEvents() {
	// we flush when queue size is reached.
	// however, if there is a ticker cycle already processing, we should wait
	p.Mux.Lock()
	for p.EventsCount() > 0 {
		events := p.GetEvents(1)
		if len(events) > 0 {
			userEvent, ok := events[0].(UserEvent)
			if ok {
				if userEvent.Conversion != nil {
					eventBatch := createConversionBatchEvent(userEvent)
					p.EventDispatcher.DispatchEvent(createLogEvent(eventBatch), func(success bool) {
						fmt.Println(success)
						if success {
							p.Remove(1)
						}
					})
				} else if userEvent.Impression != nil {
					eventBatch := createImpressionBatchEvent(userEvent)
					p.EventDispatcher.DispatchEvent(createLogEvent(eventBatch), func(success bool) {
						fmt.Println(success)
						if success {
							p.Remove(1)
						}
					})
				}
			}
		}
 	}
	p.Mux.Unlock()
}
