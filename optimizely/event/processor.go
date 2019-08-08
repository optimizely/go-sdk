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
	p.BatchSize = 20
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

func (p *QueueingEventProcessor) Append(current *Batch, new Batch) bool {
	if current.ProjectID == new.ProjectID &&
		current.Revision == new.Revision {
		visitors := append(current.Visitors, new.Visitors[0])
		current.Visitors = visitors
		return true
	}

	return false
}

// ProcessEvent processes the given impression event
func (p *QueueingEventProcessor) FlushEvents() {
	// we flush when queue size is reached.
	// however, if there is a ticker cycle already processing, we should wait
	p.Mux.Lock()
	var batchEvent Batch
	var batchEventCount = 0
	var failedToSend = false

	for p.EventsCount() > 0 {
		if failedToSend {
			break
		}
		events := p.GetEvents(p.BatchSize)

		if len(events) > 0 {
			for i := 0; i < len(events); i++ {
				userEvent, ok := events[i].(UserEvent)
				if ok {
					var eventToAdd Batch
					if userEvent.Conversion != nil {
						eventToAdd = createConversionBatchEvent(userEvent)
					} else if userEvent.Impression != nil {
							eventToAdd = createImpressionBatchEvent(userEvent)
					}
					if batchEventCount == 0 {
						batchEvent = eventToAdd
						batchEventCount = 1
					} else if !p.Append(&batchEvent, eventToAdd) {
						break
					} else {
						batchEventCount++
					}

					if batchEventCount >= p.BatchSize {
						break
					}
				}
			}
		}
		if batchEventCount > 0 {
			p.EventDispatcher.DispatchEvent(createLogEvent(batchEvent), func(success bool) {
				fmt.Println(success)
				if success {
					p.Remove(batchEventCount)
					batchEventCount = 0
					batchEvent = Batch{}
				} else {
					failedToSend = true
				}
			})
		}
	}
	p.Mux.Unlock()
}
