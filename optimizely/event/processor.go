package event

import (
	"errors"
	"github.com/optimizely/go-sdk/optimizely/logging"
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

var pLogger = logging.GetLogger("EventProcessor")

// NewEventProcessor returns a new instance of QueueingEventProcessor with queueSize and flushInterval
func NewEventProcessor(queueSize int, flushInterval time.Duration ) Processor {
	p := &QueueingEventProcessor{MaxQueueSize: queueSize, FlushInterval:flushInterval, Q:NewInMemoryQueue(queueSize), EventDispatcher:&HTTPEventDispatcher{}}
	p.BatchSize = 10
	p.StartTicker()
	return p
}

// NewEventProcessor returns a new instance of QueueingEventProcessor with queueSize and flushInterval
func NewEventProcessorNSQ(queueSize int, flushInterval time.Duration ) Processor {
	p := &QueueingEventProcessor{MaxQueueSize: queueSize, FlushInterval:flushInterval, Q:NewNSQueue(queueSize), EventDispatcher:&HTTPEventDispatcher{}}
	p.BatchSize = 10
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

// EventsCount returns size of an event queue
func (p *QueueingEventProcessor) EventsCount() int {
	return p.Q.Size()
}

// GetEvents returns events from event queue for count
func (p *QueueingEventProcessor) GetEvents(count int) []interface{} {
	return p.Q.Get(count)
}

// Remove removes events from queue for count
func (p *QueueingEventProcessor) Remove(count int) []interface{} {
	return p.Q.Remove(count)
}

// StartTicker starts new ticker for flushing events
func (p *QueueingEventProcessor) StartTicker() {
	if p.Ticker != nil {
		return
	}
	p.Ticker = time.NewTicker(p.FlushInterval * time.Millisecond)
	go func() {
		for range p.Ticker.C {
			p.FlushEvents()
		}
	}()
}

// check if user event can be batched in the current batch
func (p *QueueingEventProcessor) canBatch(current *Batch, user UserEvent) bool {
	if current.ProjectID == user.EventContext.ProjectID &&
		current.Revision == user.EventContext.Revision {
		return true
	}

	return false
}

// add the visitor to the current batch
func (p *QueueingEventProcessor) addToBatch(current *Batch, visitor Visitor) {
	visitors := append(current.Visitors, visitor)
	current.Visitors = visitors
}

// FlushEvents flushes events in queue
func (p *QueueingEventProcessor) FlushEvents() {
	// we flush when queue size is reached.
	// however, if there is a ticker cycle already processing, we should wait
	p.Mux.Lock()
	var batchEvent Batch
	var batchEventCount = 0
	var failedToSend = false

	for p.EventsCount() > 0 {
		if failedToSend {
			pLogger.Error("Last Event Batch failed to send. Retry on next Flush", errors.New("Dispatcher failed"))
			break
		}
		events := p.GetEvents(p.BatchSize)

		if len(events) > 0 {
			for i := 0; i < len(events); i++ {
				userEvent, ok := events[i].(UserEvent)
				if ok {
					if batchEventCount == 0 {
						batchEvent = createBatchEvent(userEvent, createVisitorFromUserEvent(userEvent))
						batchEventCount = 1
					} else if !p.canBatch(&batchEvent, userEvent) {
						// this could happen if the project config was updated for instance.
						pLogger.Info("Can't batch last event. Sending current batch.")
						break
					} else {
						p.addToBatch(&batchEvent, createVisitorFromUserEvent(userEvent))
						batchEventCount++
					}

					if batchEventCount >= p.BatchSize {
						// the batch size is reached so take the current batchEvent and send it.
						break
					}
				}
			}
		}
		if batchEventCount > 0 {
			p.EventDispatcher.DispatchEvent(createLogEvent(batchEvent), func(success bool) {
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
