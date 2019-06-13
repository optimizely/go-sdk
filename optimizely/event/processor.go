package event

import (
	"fmt"
	"sync"
	"time"
)

// Processor processes events
type Processor interface {
	ProcessImpression(event Impression)
}

// DefaultEventProcessor is used out of the box by the SDK
type DefaultEventProcessor struct {
	MaxQueueSize    int           // max size of the queue before flush
	FlushInterval   time.Duration // in milliseconds
	BatchSize       int
	Queue           [] interface{}
	Mux             sync.Mutex
	Ticker          *time.Ticker
	EventDispatcher Dispatcher
}

func NewEventProcessor(queueSize int, flushInterval time.Duration ) Processor {
	p := &DefaultEventProcessor{MaxQueueSize: queueSize, FlushInterval:flushInterval, Queue:make([] interface{}, 0, queueSize), EventDispatcher:&HttpEventDispatcher{}}
	p.startTicker()
	return p
}

// ProcessImpression processes the given impression event
func (p *DefaultEventProcessor) ProcessImpression(event Impression) {
	p.Mux.Lock()
	p.Queue = append(p.Queue, event)
	p.Mux.Unlock()
}

func (p *DefaultEventProcessor) eventsCount() int {
	p.Mux.Lock()
	defer p.Mux.Unlock()
	return len(p.Queue)
}

func (p *DefaultEventProcessor) getEvents(count int) []interface{} {
	p.Mux.Lock()
	defer p.Mux.Unlock()
	return p.Queue[:count]
}

func (p *DefaultEventProcessor) remove(count int) []interface{} {
	p.Mux.Lock()
	defer p.Mux.Unlock()
	elem := p.Queue[:count]
	p.Queue = p.Queue[count:]
	return elem
}

func (p *DefaultEventProcessor) startTicker() {
	if p.Ticker != nil {
		return
	}
	p.Ticker = time.NewTicker(p.FlushInterval * time.Millisecond)
	go func() {
		for _ = range p.Ticker.C {
			p.flushEvents()
		}
	}()
}

// ProcessImpression processes the given impression event
func (p *DefaultEventProcessor) flushEvents() {
	for p.eventsCount() > 0 {
		events := p.getEvents(1)
		if len(events) > 0 {
			p.EventDispatcher.DispatchEvent(events[0], func(success bool) {
				fmt.Println(success)
				if success {
					p.remove(1)
				}
			})
		}
 	}

}
