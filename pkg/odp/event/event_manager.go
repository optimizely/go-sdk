/****************************************************************************
 * Copyright 2022-2023, Optimizely, Inc. and contributors                   *
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
	"strings"
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
	// odpConfig is required here since it can be updated anytime and ticker needs to be aware of latest changes
	Start(ctx context.Context, odpConfig config.Config)
	IdentifyUser(apiKey, apiHost, userID string)
	ProcessEvent(apiKey, apiHost string, odpEvent Event) error
	FlushEvents(apiKey, apiHost string)
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
	processing    *semaphore.Weighted
	logger        logging.OptimizelyLogProducer
}

// WithQueueSize sets the queue size as a config option to be passed into the NewBatchEventManager method
// default value is 10000
func WithQueueSize(qsize int) EMOptionFunc {
	return func(bm *BatchEventManager) {
		if qsize > 0 {
			bm.maxQueueSize = qsize
		}
	}
}

// WithFlushInterval sets the flush interval as a config option to be passed into the NewBatchEventManager method
// default value is 1 second
func WithFlushInterval(flushInterval time.Duration) EMOptionFunc {
	return func(bm *BatchEventManager) {
		if flushInterval >= 0 {
			// if flush interval is zero, send events immediately by setting batchSize to 1
			if flushInterval == 0 {
				bm.batchSize = 1
			}
			bm.flushInterval = flushInterval
		}
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

// NewBatchEventManager returns a new instance of BatchEventManager with options
func NewBatchEventManager(options ...EMOptionFunc) *BatchEventManager {
	// Setting default values
	bm := &BatchEventManager{
		processing:    semaphore.NewWeighted(int64(maxFlushWorkers)),
		maxQueueSize:  utils.DefaultEventQueueSize,
		flushInterval: utils.DefaultEventFlushInterval,
		batchSize:     utils.DefaultBatchSize,
	}

	for _, opt := range options {
		opt(bm)
	}

	bm.logger = logging.GetLogger(bm.sdkKey, "BatchEventManager")

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
// odpConfig is required here since it can be updated anytime and ticker needs to be aware of latest changes
func (bm *BatchEventManager) Start(ctx context.Context, odpConfig config.Config) {
	if !bm.IsOdpServiceIntegrated(odpConfig.GetAPIKey(), odpConfig.GetAPIHost()) {
		return
	}
	bm.startTicker(ctx, odpConfig)
}

// IdentifyUser associates a full-stack userid with an established VUID
func (bm *BatchEventManager) IdentifyUser(apiKey, apiHost, userID string) {
	if !bm.IsOdpServiceIntegrated(apiKey, apiHost) {
		bm.logger.Debug(utils.IdentityOdpNotIntegrated)
		return
	}
	identifiers := map[string]string{utils.OdpFSUserIDKey: userID}
	odpEvent := Event{
		Type:        utils.OdpEventType,
		Action:      utils.OdpActionIdentified,
		Identifiers: identifiers,
	}
	_ = bm.ProcessEvent(apiKey, apiHost, odpEvent)
}

// ProcessEvent takes the given odp event and queues it up to be dispatched.
// A dispatch happens when we flush the events, which can happen on a set interval or
// when the specified batch size is reached.
func (bm *BatchEventManager) ProcessEvent(apiKey, apiHost string, odpEvent Event) (err error) {
	if !bm.IsOdpServiceIntegrated(apiKey, apiHost) {
		bm.logger.Debug(utils.OdpNotIntegrated)
		return errors.New(utils.OdpNotIntegrated)
	}

	if odpEvent.Action == "" {
		bm.logger.Error(utils.OdpInvalidAction, errors.New("invalid event action"))
		return errors.New(utils.OdpInvalidAction)
	}

	if !utils.IsValidOdpData(odpEvent.Data) {
		bm.logger.Error(utils.OdpInvalidData, errors.New("invalid event data"))
		return errors.New(utils.OdpInvalidData)
	}

	if bm.eventQueue.Size() >= bm.maxQueueSize {
		err = errors.New("ODP EventQueue is full")
		bm.logger.Error("maxQueueSize has been met. Discarding event", err)
		return err
	}

	bm.addCommonData(&odpEvent)
	bm.convertIdentifiers(&odpEvent)
	bm.eventQueue.Add(odpEvent)

	if bm.eventQueue.Size() < bm.batchSize {
		return nil
	}

	if bm.processing.TryAcquire(1) {
		// it doesn't matter if the timer has kicked in here.
		// we just want to start one go routine when the batch size is met.
		bm.logger.Debug("batch size reached.  Flushing routine being called")
		go func() {
			bm.FlushEvents(apiKey, apiHost)
			bm.processing.Release(1)
		}()
	}

	return nil
}

// StartTicker starts new ticker for flushing events
func (bm *BatchEventManager) startTicker(ctx context.Context, odpConfig config.Config) {
	// Do not start ticker if flushInterval is 0
	if bm.flushInterval <= 0 {
		return
	}
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
			bm.FlushEvents(odpConfig.GetAPIKey(), odpConfig.GetAPIHost())
		case <-ctx.Done():
			bm.logger.Debug("BatchEventManager stopped, flushing events.")
			bm.FlushEvents(odpConfig.GetAPIKey(), odpConfig.GetAPIHost())
			bm.flushLock.Lock()
			bm.ticker.Stop()
			bm.flushLock.Unlock()
			return
		}
	}
}

// FlushEvents flushes events in queue
func (bm *BatchEventManager) FlushEvents(apiKey, apiHost string) {
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
				shouldRetry, err := bm.apiManager.SendOdpEvents(apiKey, apiHost, batchEvent)
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
func (bm *BatchEventManager) IsOdpServiceIntegrated(apiKey, apiHost string) bool {
	if apiKey == "" || apiHost == "" {
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

// Convert incorrect case/separator of identifier key `fs_user_id`
// (ie. `fs-user-id`, `FS_USER_ID`).
func (bm *BatchEventManager) convertIdentifiers(odpEvent *Event) {
	if odpEvent.Identifiers[utils.OdpFSUserIDKey] != "" {
		return
	}
	keysToCheck := map[string]bool{"fs-user-id": true, utils.OdpFSUserIDKey: true}
	for k, v := range odpEvent.Identifiers {
		if keysToCheck[strings.ToLower(k)] {
			odpEvent.Identifiers[utils.OdpFSUserIDKey] = v
			delete(odpEvent.Identifiers, k)
			break
		}
	}
}
