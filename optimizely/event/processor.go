package event

import (
	"bytes"
	"sync"
	"time"
	"net/http"
	"encoding/json"
	"fmt"
)

type EventDispatcher interface {
	DispatchEvent(event interface{}, callback func(success bool))
}

// Processor processes events
type EventProcessor interface {
	ProcessEvent(event interface {} )
}

// DefaultEventProcessor is used out of the box by the SDK
type HttpEventProcessor struct {
	maxQueueSize int // max size of the queue before flush
	flushInterval time.Duration // in milliseconds
	batchSize int
	queue [] interface{}
	mux sync.Mutex
	ticker *time.Ticker

}

func (HttpEventProcessor) New(queueSize int, flushInterval time.Duration ) *HttpEventProcessor {
	p := &HttpEventProcessor{maxQueueSize:queueSize, flushInterval:flushInterval, queue:make([] interface{}, queueSize)}
	p.startTicker()
	return p
}

// ProcessImpression processes the given impression event
func (p *HttpEventProcessor) ProcessImpression(event Impression) {
	p.mux.Lock()
	p.queue = append(p.queue, event)
	p.mux.Unlock()
}

func (p *HttpEventProcessor) eventsCount() int {
	p.mux.Lock()
	defer p.mux.Unlock()
	return cap(p.queue)
}

func (p *HttpEventProcessor) getEvents(count int) []interface{} {
	p.mux.Lock()
	defer p.mux.Unlock()
	return p.queue[:count]
}

func (p *HttpEventProcessor) remove(count int) []interface{} {
	p.mux.Lock()
	defer p.mux.Unlock()
	elem := p.queue[:count]
	p.queue = p.queue[count:]
	return elem
}

func (p *HttpEventProcessor) startTicker() {
	if p.ticker != nil {
		return
	}
	p.ticker = time.NewTicker(p.flushInterval * time.Millisecond)
	go func() {
		for _ = range p.ticker.C {
			p.flushEvents()
		}
	}()
}

// ProcessImpression processes the given impression event
func (p *HttpEventProcessor) flushEvents() {
	for cap(p.queue) > 0 {
		events := p.getEvents(1)
		if len(events) > 0 {
			go p.sendEvent(events[0])
		}
 	}

}

func (p *HttpEventProcessor) sendEvent(event interface{}) {
	impression, ok := event.(Impression)
	if ok {
		jsonValue, _ := json.Marshal(impression)
		resp, err := http.Post("https://logx.optimizely.com/v1/events", "application/json", bytes.NewBuffer(jsonValue))
		fmt.Println(resp)
		if err != nil {
			fmt.Println(err)
		} else {

		}
	}
}
