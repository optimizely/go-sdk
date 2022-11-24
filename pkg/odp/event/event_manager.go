/****************************************************************************
 * Copyright 2022, Optimizely, Inc. and contributors                        *
 *                                                                          *
 * Licensed under the Apache License, Version 2.0 (the "License");          *
 * you may not use this file except in compliance with the License.         *
 * You may obtain a copy of the License at                                  *
 *                                                                          *
 *    https://www.apache.org/licenses/LICENSE-2.0                           *
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

	guuid "github.com/google/uuid"
	"github.com/optimizely/go-sdk/pkg/event"
	"github.com/optimizely/go-sdk/pkg/logging"
	"github.com/optimizely/go-sdk/pkg/odp/config"
	"github.com/optimizely/go-sdk/pkg/odp/utils"
	"golang.org/x/sync/semaphore"
)

// EMOptionFunc are the EventManager options that give you the ability to add one more more options before the event manager is initialized.
type EMOptionFunc func(em *BatchEventManager)

const maxFlushWorkers = 1
const maxRetries = 3

// Manager represents the event manager.
type Manager interface {
	Start(ctx context.Context)
	IdentifyUser(userID string)
	ProcessEvent(odpEvent Event) bool
	FlushEvents()
}

// BatchEventManager represents default implementation of BatchEventManager
type BatchEventManager struct {
	sdkKey        string
	maxQueueSize  int           // max size of the queue before flush
	flushInterval time.Duration // in milliseconds
	batchSize     int
	eventQueue    event.Queue
	flushLock     sync.Mutex
	ticker        *time.Ticker
	apiManager    APIManager
	OdpConfig     config.Config
	processing    *semaphore.Weighted
	logger        logging.OptimizelyLogProducer
}

// WithBatchSize sets the batch size as a config option to be passed into the NewBatchEventManager method
func WithBatchSize(bsize int) EMOptionFunc {
	return func(bm *BatchEventManager) {
		bm.batchSize = bsize
	}
}

// WithQueueSize sets the queue size as a config option to be passed into the NewBatchEventManager method
func WithQueueSize(qsize int) EMOptionFunc {
	return func(bm *BatchEventManager) {
		bm.maxQueueSize = qsize
	}
}

// WithFlushInterval sets the flush interval as a config option to be passed into the NewBatchEventManager method
func WithFlushInterval(flushInterval time.Duration) EMOptionFunc {
	return func(bm *BatchEventManager) {
		bm.flushInterval = flushInterval
	}
}

// WithQueue sets the Queue as a config option to be passed into the NewBatchEventManager method
func WithQueue(q event.Queue) EMOptionFunc {
	return func(bm *BatchEventManager) {
		bm.eventQueue = q
	}
}

// WithSDKKey sets the SDKKey used for logging
func WithSDKKey(sdkKey string) EMOptionFunc {
	return func(bm *BatchEventManager) {
		bm.sdkKey = sdkKey
	}
}

// WithAPIManager sets apiManager as a config option to be passed into the NewBatchEventManager method
func WithAPIManager(apiManager APIManager) EMOptionFunc {
	return func(bm *BatchEventManager) {
		bm.apiManager = apiManager
	}
}

// WithOdpConfig sets odpConfig option to be passed into the NewBatchEventManager method
func WithOdpConfig(odpConfig config.Config) EMOptionFunc {
	return func(bm *BatchEventManager) {
		bm.OdpConfig = odpConfig
	}
}

// NewBatchEventManager returns a new instance of BatchEventManager with options
func NewBatchEventManager(options ...EMOptionFunc) *BatchEventManager {
	bm := &BatchEventManager{processing: semaphore.NewWeighted(int64(maxFlushWorkers))}

	for _, opt := range options {
		opt(bm)
	}

	bm.logger = logging.GetLogger(bm.sdkKey, "BatchEventManager")

	if bm.maxQueueSize == 0 {
		bm.maxQueueSize = utils.DefaultEventQueueSize
	}

	if bm.flushInterval == 0 {
		bm.flushInterval = utils.DefaultEventFlushInterval
	}

	if bm.batchSize == 0 {
		bm.batchSize = utils.DefaultBatchSize
	}

	if bm.batchSize > bm.maxQueueSize {
		bm.logger.Warning(
			fmt.Sprintf("Batch size %d is larger than queue size %d.  Setting to defaults",
				bm.batchSize, bm.maxQueueSize))

		bm.batchSize = utils.DefaultBatchSize
		bm.maxQueueSize = utils.DefaultEventQueueSize
	}

	if bm.eventQueue == nil {
		bm.eventQueue = event.NewInMemoryQueueWithLogger(bm.maxQueueSize, bm.logger)
	}

	if bm.apiManager == nil {
		bm.apiManager = NewEventAPIManager(bm.sdkKey, nil)
	}

	return bm
}

// Start does not do any initialization, just starts the ticker
func (bm *BatchEventManager) Start(ctx context.Context) {
	if !bm.IsOdpServiceIntegrated() {
		return
	}
	bm.startTicker(ctx)
}

// IdentifyUser associates a full-stack userid with an established VUID
func (bm *BatchEventManager) IdentifyUser(userID string) {
	if !bm.IsOdpServiceIntegrated() {
		bm.logger.Debug(utils.IdentityOdpNotIntegrated)
	}
	identifiers := map[string]string{utils.OdpFSUserIDKey: userID}
	odpEvent := Event{
		Type:        utils.OdpEventType,
		Action:      utils.OdpActionIdentified,
		Identifiers: identifiers,
	}
	bm.ProcessEvent(odpEvent)
}

// ProcessEvent takes the given odp event and queues it up to be dispatched.
// A dispatch happens when we flush the events, which can happen on a set interval or
// when the specified batch size is reached.
func (bm *BatchEventManager) ProcessEvent(odpEvent Event) bool {
	if !bm.IsOdpServiceIntegrated() {
		bm.logger.Debug(utils.OdpNotIntegrated)
		return false
	}

	if !utils.IsValidOdpData(odpEvent.Data) {
		bm.logger.Error(utils.OdpInvalidData, errors.New("invalid event data"))
		return false
	}

	if bm.eventQueue.Size() >= bm.maxQueueSize {
		bm.logger.Warning("maxQueueSize has been met. Discarding event")
		return false
	}

	bm.addCommonData(&odpEvent)
	bm.eventQueue.Add(odpEvent)

	if bm.eventQueue.Size() < bm.batchSize {
		return true
	}

	if bm.processing.TryAcquire(1) {
		// it doesn't matter if the timer has kicked in here.
		// we just want to start one go routine when the batch size is met.
		bm.logger.Debug("batch size reached.  Flushing routine being called")
		go func() {
			bm.FlushEvents()
			bm.processing.Release(1)
		}()
	}

	return true
}

// StartTicker starts new ticker for flushing events
func (bm *BatchEventManager) startTicker(ctx context.Context) {
	// Make sure multiple go-routines dont reinitialize ticker
	bm.flushLock.Lock()
	if bm.ticker != nil {
		bm.flushLock.Unlock()
		return
	}

	bm.logger.Info("Batch event manager started")
	bm.ticker = time.NewTicker(bm.flushInterval)
	bm.flushLock.Unlock()

	for {
		select {
		case <-bm.ticker.C:
			bm.FlushEvents()
		case <-ctx.Done():
			bm.logger.Debug("BatchEventManager stopped, flushing events.")
			bm.FlushEvents()
			return
		}
	}
}

// FlushEvents flushes events in queue
func (bm *BatchEventManager) FlushEvents() {
	// we flush when queue size is reached.
	// however, if there is a ticker cycle already processing, we should wait
	bm.flushLock.Lock()
	defer bm.flushLock.Unlock()

	var batchEvent []Event
	var batchEventCount = 0
	var failedToSend = false

	for bm.eventQueue.Size() > 0 {
		if failedToSend {
			bm.logger.Error("last Event Batch failed to send; retry on next flush", errors.New("dispatcher failed"))
			break
		}

		events := bm.eventQueue.Get(bm.batchSize)
		if len(events) > 0 {
			for _, event := range events {
				odpEvent, ok := event.(Event)
				if ok {
					if batchEventCount == 0 {
						batchEvent = []Event{odpEvent}
						batchEventCount = 1
					} else {
						batchEvent = append(batchEvent, odpEvent)
						batchEventCount++
					}

					if batchEventCount >= bm.batchSize {
						// the batch size is reached so take the current batchEvent and send it.
						break
					}
				}
			}
		}

		// Only send event if batch is available
		if batchEventCount > 0 {
			retryCount := 0
			// Retry till maxRetries reached
			for retryCount < maxRetries {
				failedToSend = true
				shouldRetry, err := bm.apiManager.SendOdpEvents(bm.OdpConfig, batchEvent)
				// Remove events from queue if dispatch failed and retrying is not suggested
				if !shouldRetry {
					bm.eventQueue.Remove(batchEventCount)
					batchEventCount = 0
					batchEvent = []Event{}
					if err == nil {
						bm.logger.Debug("Dispatched odp event successfully")
						failedToSend = false
					} else {
						bm.logger.Warning(err.Error())
					}
					break
				}
				retryCount++
			}
		}
	}
}

// IsOdpServiceIntegrated returns true if odp service is integrated
func (bm *BatchEventManager) IsOdpServiceIntegrated() bool {
	if bm.OdpConfig == nil || !bm.OdpConfig.IsOdpServiceIntegrated() {
		// ensure empty queue
		bm.eventQueue.Remove(bm.eventQueue.Size())
		return false
	}

	return true
}

func (bm *BatchEventManager) addCommonData(odpEvent *Event) {
	commonData := map[string]interface{}{
		"idempotence_id":      guuid.New().String(),
		"data_source_type":    "sdk",
		"data_source":         event.ClientName,
		"data_source_version": event.Version,
	}
	// Override common data with user provided data
	for k, v := range odpEvent.Data {
		commonData[k] = v
	}
	odpEvent.Data = commonData
}
