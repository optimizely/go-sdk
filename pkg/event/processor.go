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
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"golang.org/x/sync/semaphore"

	"github.com/optimizely/go-sdk/pkg/logging"
	"github.com/optimizely/go-sdk/pkg/metrics"
	"github.com/optimizely/go-sdk/pkg/notification"
	"github.com/optimizely/go-sdk/pkg/registry"
)

// Processor processes events
type Processor interface {
	ProcessEvent(event UserEvent) bool
}

// BatchEventProcessor is used out of the box by the SDK to queue up and batch events to be sent to the Optimizely
// log endpoint for results processing.
type BatchEventProcessor struct {
	sdkKey          string
	MaxQueueSize    int           // max size of the queue before flush
	FlushInterval   time.Duration // in milliseconds
	BatchSize       int
	Q               Queue
	flushLock       sync.Mutex
	Ticker          *time.Ticker
	EventDispatcher Dispatcher
	processing      *semaphore.Weighted

	metricsRegistry metrics.Registry
}

// DefaultBatchSize holds the default value for the batch size
const DefaultBatchSize = 10

// DefaultEventQueueSize holds the default value for the event queue size
const DefaultEventQueueSize = 2000

// DefaultEventFlushInterval holds the default value for the event flush interval
const DefaultEventFlushInterval = 30 * time.Second

const maxFlushWorkers = 1

var pLogger = logging.GetLogger("EventProcessor")

// BPOptionConfig is the BatchProcessor options that give you the ability to add one more more options before the processor is initialized.
type BPOptionConfig func(qp *BatchEventProcessor)

// WithBatchSize sets the batch size as a config option to be passed into the NewProcessor method
func WithBatchSize(bsize int) BPOptionConfig {
	return func(qp *BatchEventProcessor) {
		qp.BatchSize = bsize
	}
}

// WithQueueSize sets the queue size as a config option to be passed into the NewProcessor method
func WithQueueSize(qsize int) BPOptionConfig {
	return func(qp *BatchEventProcessor) {
		qp.MaxQueueSize = qsize
	}
}

// WithFlushInterval sets the flush interval as a config option to be passed into the NewProcessor method
func WithFlushInterval(flushInterval time.Duration) BPOptionConfig {
	return func(qp *BatchEventProcessor) {
		qp.FlushInterval = flushInterval
	}
}

// WithQueue sets the Processor Queue as a config option to be passed into the NewProcessor method
func WithQueue(q Queue) BPOptionConfig {
	return func(qp *BatchEventProcessor) {
		qp.Q = q
	}
}

// WithEventDispatcher sets the Processor Dispatcher as a config option to be passed into the NewProcessor method
func WithEventDispatcher(d Dispatcher) BPOptionConfig {
	return func(qp *BatchEventProcessor) {
		qp.EventDispatcher = d
	}
}

// WithSDKKey sets the SDKKey used to register for notifications.  This should be removed when the project
// config supports sdk key.
func WithSDKKey(sdkKey string) BPOptionConfig {
	return func(qp *BatchEventProcessor) {
		qp.sdkKey = sdkKey
	}
}

// WithEventDispatcherMetrics sets metrics into the NewProcessor method
func WithEventDispatcherMetrics(metricsRegistry metrics.Registry) BPOptionConfig {
	return func(qp *BatchEventProcessor) {
		qp.metricsRegistry = metricsRegistry
	}
}

// NewBatchEventProcessor returns a new instance of BatchEventProcessor with queueSize and flushInterval
func NewBatchEventProcessor(options ...BPOptionConfig) *BatchEventProcessor {
	p := &BatchEventProcessor{processing: semaphore.NewWeighted(int64(maxFlushWorkers))}

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

	if p.BatchSize > p.MaxQueueSize {
		pLogger.Warning(
			fmt.Sprintf("Batch size %d is larger than queue size %d.  Setting to defaults",
				p.BatchSize, p.MaxQueueSize))

		p.BatchSize = DefaultBatchSize
		p.MaxQueueSize = defaultQueueSize
	}

	if p.Q == nil {
		p.Q = NewInMemoryQueue(p.MaxQueueSize)
	}

	return p
}

// Start initializes the event processor
func (p *BatchEventProcessor) Start(ctx context.Context) {
	if p.EventDispatcher == nil {
		dispatcher := NewQueueEventDispatcher(p.metricsRegistry)
		defer dispatcher.flushEvents()
		p.EventDispatcher = dispatcher
	}

	pLogger.Debug("Batch event processor started")
	p.startTicker(ctx)
}

// ProcessEvent takes the given user event (can be an impression or conversion event) and queues it up to be dispatched
// to the Optimizely log endpoint. A dispatch happens when we flush the events, which can happen on a set interval or
// when the specified batch size (defaulted to 10) is reached.
func (p *BatchEventProcessor) ProcessEvent(event UserEvent) bool {

	if p.Q.Size() >= p.MaxQueueSize {
		pLogger.Warning("MaxQueueSize has been met. Discarding event")
		return false
	}

	p.Q.Add(event)

	if p.Q.Size() < p.BatchSize {
		return true
	}

	if p.processing.TryAcquire(1) {
		// it doesn't matter if the timer has kicked in here.
		// we just want to start one go routine when the batch size is met.
		pLogger.Debug("batch size reached.  Flushing routine being called")
		go func() {
			p.flushEvents()
			p.processing.Release(1)
		}()
	}

	return true
}

// eventsCount returns size of an event queue
func (p *BatchEventProcessor) eventsCount() int {
	return p.Q.Size()
}

// getEvents returns events from event queue for count
func (p *BatchEventProcessor) getEvents(count int) []interface{} {
	return p.Q.Get(count)
}

// remove removes events from queue for count
func (p *BatchEventProcessor) remove(count int) []interface{} {
	return p.Q.Remove(count)
}

// StartTicker starts new ticker for flushing events
func (p *BatchEventProcessor) startTicker(ctx context.Context) {
	if p.Ticker != nil {
		return
	}
	p.Ticker = time.NewTicker(p.FlushInterval)

	for {
		select {
		case <-p.Ticker.C:
			p.flushEvents()
		case <-ctx.Done():
			pLogger.Debug("Event processor stopped, flushing events.")
			p.flushEvents()
			d, ok := p.EventDispatcher.(*QueueEventDispatcher)
			if ok {
				d.flushEvents()
			}
			return
		}
	}
}

// check if user event can be batched in the current batch
func (p *BatchEventProcessor) canBatch(current *Batch, user UserEvent) bool {
	if current.ProjectID == user.EventContext.ProjectID &&
		current.Revision == user.EventContext.Revision {
		return true
	}

	return false
}

// add the visitor to the current batch
func (p *BatchEventProcessor) addToBatch(current *Batch, visitor Visitor) {
	visitors := append(current.Visitors, visitor)
	current.Visitors = visitors
}

// flushEvents flushes events in queue
func (p *BatchEventProcessor) flushEvents() {
	// we flush when queue size is reached.
	// however, if there is a ticker cycle already processing, we should wait
	p.flushLock.Lock()
	defer p.flushLock.Unlock()

	var batchEvent Batch
	var batchEventCount = 0
	var failedToSend = false

	for p.eventsCount() > 0 {
		if failedToSend {
			pLogger.Error("last Event Batch failed to send; retry on next flush", errors.New("dispatcher failed"))
			break
		}
		events := p.getEvents(p.BatchSize)

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
				p.remove(batchEventCount)
				batchEventCount = 0
				batchEvent = Batch{}
			} else {
				pLogger.Warning("Failed to dispatch event successfully")
				failedToSend = true
			}
		}
	}
}

// OnEventDispatch registers a handler for LogEvent notifications
func (p *BatchEventProcessor) OnEventDispatch(callback func(logEvent LogEvent)) (int, error) {
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
func (p *BatchEventProcessor) RemoveOnEventDispatch(id int) error {
	notificationCenter := registry.GetNotificationCenter(p.sdkKey)

	if err := notificationCenter.RemoveHandler(id, notification.LogEvent); err != nil {
		pLogger.Warning("Problem with removing notification handler.")
		return err
	}
	return nil
}
