/****************************************************************************
 * Copyright 2019, Optimizely, Inc. and contributors                        *
 *                                                                          *
 * Licensed under the Apache License, Version 2.0 (the "License");          *
 * you may not use this file except in compliance with the License.         *
 * You may obtain a copy of the License at                                  *
 *                                                                          *
 *    http://www.apache.org/licenses/LICENSE-2.0                            *
 *                                                                          *
 * Unless required by applicable law or agreed to in writing, software      *
 * distributed under the License is distributed on an "AS IS" BASIS,        *
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. *
 * See the License for the specific language governing permissions and      *
 * limitations under the License.                                           *
 ***************************************************************************/

// Package event //
package event

import (
	"errors"
	"fmt"
	"github.com/optimizely/go-sdk/pkg/logging"
	"github.com/optimizely/go-sdk/pkg/notification"
	"github.com/optimizely/go-sdk/pkg/registry"
	"github.com/optimizely/go-sdk/pkg/utils"
	"sync"
	"time"
)

// Processor processes events
type Processor interface {
	ProcessEvent(event UserEvent)
}

// QueueingEventProcessor is used out of the box by the SDK
type QueueingEventProcessor struct {
	sdkKey 			string
	MaxQueueSize    int           // max size of the queue before flush
	FlushInterval   time.Duration // in milliseconds
	BatchSize       int
	Q               Queue
	Mux             sync.Mutex
	Ticker          *time.Ticker
	EventDispatcher Dispatcher
}

// DefaultBatchSize holds the default value for the batch size
const DefaultBatchSize = 10

// DefaultEventQueueSize holds the default value for the event queue size
const DefaultEventQueueSize = 100

// DefaultEventFlushInterval holds the default value for the event flush interval
const DefaultEventFlushInterval = 30 * time.Second

var pLogger = logging.GetLogger("EventProcessor")

// QPConfigOption is the QueuingProcessor options that give you the ability to add one more more options before the processor is initialized.
type QPConfigOption func(qp *QueueingEventProcessor)

// BatchSize sets the batch size as a config option to be passed into the NewProcessor method
func BatchSize(bsize int) QPConfigOption {
	return func(qp *QueueingEventProcessor) {
		qp.BatchSize = bsize
	}
}

// QueueSize sets the queue size as a config option to be passed into the NewProcessor method
func QueueSize(qsize int) QPConfigOption {
	return func(qp *QueueingEventProcessor) {
		qp.MaxQueueSize = qsize
	}
}

// FlushInterval sets the flush interval as a config option to be passed into the NewProcessor method
func FlushInterval(flushInterval time.Duration) QPConfigOption {
	return func(qp *QueueingEventProcessor) {
		qp.FlushInterval = flushInterval
	}
}

// PQ sets the Processor Queue as a config option to be passed into the NewProcessor method
func PQ(q Queue) QPConfigOption {
	return func(qp *QueueingEventProcessor) {
		qp.Q = q
	}
}

// PDispatcher sets the Processor Dispatcher as a config option to be passed into the NewProcessor method
func PDispatcher(d Dispatcher) QPConfigOption {
	return func(qp *QueueingEventProcessor) {
		qp.EventDispatcher = d
	}
}

// SDKKey sets the SDKKey used to register for notifications.  This should be removed when the project
// config supports sdk key.
func SDKKey(sdkKey string) QPConfigOption {
	return func(qp *QueueingEventProcessor) {
		qp.sdkKey = sdkKey
	}
}

// NewEventProcessor returns a new instance of QueueingEventProcessor with queueSize and flushInterval
func NewEventProcessor(options ...QPConfigOption) *QueueingEventProcessor {
	p := &QueueingEventProcessor{}

	for _, opt := range options {
		opt(p)
	}

	if p.MaxQueueSize == 0 {
		p.MaxQueueSize = defaultQueueSize
	}

	if p.FlushInterval == 0 {
		p.FlushInterval = DefaultEventFlushInterval
	}

	if p.BatchSize == 0 {
		p.BatchSize = DefaultBatchSize
	}

	if p.Q == nil {
		p.Q = NewInMemoryQueue(p.MaxQueueSize)
	}

	return p
}

// Start initializes the event processor
func (p *QueueingEventProcessor) Start(exeCtx utils.ExecutionCtx) {
	if p.EventDispatcher == nil {
		p.EventDispatcher = NewQueueEventDispatcher(exeCtx.GetContext())
	}

	p.startTicker(exeCtx)
	pLogger.Debug("Batch event processor started")
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
func (p *QueueingEventProcessor) startTicker(exeCtx utils.ExecutionCtx) {
	if p.Ticker != nil {
		return
	}
	p.Ticker = time.NewTicker(p.FlushInterval * time.Millisecond)
	wg := exeCtx.GetWaitSync()
	wg.Add(1)
	go func() {

		defer wg.Done()
		for {
			select {
			case <-p.Ticker.C:
				p.FlushEvents()
			case <-exeCtx.GetContext().Done():
				pLogger.Debug("Event processor stopped, flushing events.")
				p.FlushEvents()
				d, ok := p.EventDispatcher.(*QueueEventDispatcher)
				if ok {
					d.flushEvents()
				}
				return
			}
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
			pLogger.Error("last Event Batch failed to send; retry on next flush", errors.New("dispatcher failed"))
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
					} else {
						if !p.canBatch(&batchEvent, userEvent) {
							// this could happen if the project config was updated for instance.
							pLogger.Info("Can't batch last event. Sending current batch.")
							break
						} else {
							p.addToBatch(&batchEvent, createVisitorFromUserEvent(userEvent))
							batchEventCount++
						}
					}

					if batchEventCount >= p.BatchSize {
						// the batch size is reached so take the current batchEvent and send it.
						break
					}
				}
			}
		}
		if batchEventCount > 0 {
			// TODO: figure out what to do with the error
			logEvent := createLogEvent(batchEvent)
			notificationCenter := registry.GetNotificationCenter(p.sdkKey)

			err := notificationCenter.Send(notification.LogEvent, logEvent)

			if err != nil {
				pLogger.Error("Send Log Event notification failed.", err)
			}
			if success, _ := p.EventDispatcher.DispatchEvent(logEvent); success {
				pLogger.Debug("Dispatched event successfully")
				p.Remove(batchEventCount)
				batchEventCount = 0
				batchEvent = Batch{}
			} else {
				pLogger.Warning("Failed to dispatch event successfully")
				failedToSend = true
			}
		}
	}
	p.Mux.Unlock()
}

// OnEventDispatch registers a handler for LogEvent notifications
func (p *QueueingEventProcessor) OnEventDispatch(callback func(logEvent LogEvent)) (int, error) {
	notificationCenter := registry.GetNotificationCenter(p.sdkKey)

	handler := func(payload interface{}) {
		if ev, ok := payload.(LogEvent); ok {
			callback(ev)
		} else {
			pLogger.Warning(fmt.Sprintf("Unable to convert notification payload %v into LogEventNotification", payload))
		}
	}
	id, err := notificationCenter.AddHandler(notification.LogEvent, handler)
	if err != nil {
		pLogger.Error("Problem with adding notification handler.", err)
		return 0, err
	}
	return id, nil
}

// RemoveOnEventDispatch removes handler for LogEvent notification with given id
func (p *QueueingEventProcessor) RemoveOnEventDispatch(id int) error {
	notificationCenter := registry.GetNotificationCenter(p.sdkKey)

	if err := notificationCenter.RemoveHandler(id, notification.LogEvent); err != nil {
		pLogger.Warning("Problem with removing notification handler.")
		return err
	}
	return nil
}
