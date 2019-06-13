package event

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

type Dispatcher interface {
	DispatchEvent(event interface{}, callback func(success bool))
}

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

type HttpEventDispatcher struct {
}

func (*HttpEventDispatcher) DispatchEvent(event interface{}, callback func(success bool)) {
	impression, ok := event.(Impression)
	if ok {
		jsonValue, _ := json.Marshal(impression)
		resp, err := http.Post("https://logx.optimizely.com/v1/events", "application/json", bytes.NewBuffer(jsonValue))
		fmt.Println(resp)
		if err != nil {
			fmt.Println(err)
			callback(false)
		} else {
			callback(true)
		}
	}
}

func NewEventProcessor(queueSize int, flushInterval time.Duration ) Processor {
	p := &DefaultEventProcessor{MaxQueueSize: queueSize, FlushInterval:flushInterval, Queue:make([] interface{}, queueSize), EventDispatcher:&HttpEventDispatcher{}}
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
	for len(p.Queue) > 0 {
		events := p.getEvents(1)
		if len(events) > 0 {
			go p.EventDispatcher.DispatchEvent(events[0], func(success bool) {
				fmt.Println(success)
			})
		}
 	}

}
