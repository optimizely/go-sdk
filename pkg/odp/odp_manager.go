/****************************************************************************
 * Copyright 2022, Optimizely, Inc. and contributors                        *
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

// Package odp //
package odp

import (
	"errors"

	"github.com/optimizely/go-sdk/pkg/logging"
	"github.com/optimizely/go-sdk/pkg/odp/cache"
	"github.com/optimizely/go-sdk/pkg/odp/config"
	"github.com/optimizely/go-sdk/pkg/odp/event"
	"github.com/optimizely/go-sdk/pkg/odp/segment"
	"github.com/optimizely/go-sdk/pkg/odp/utils"
	"golang.org/x/sync/semaphore"
)

// OMOptionConfig are the ODPManager options that give you the ability to add one more more options before the odp manager is initialized.
type OMOptionConfig func(em *DefaultOdpManager)

const maxUpdateWorkers = 1

// Manager represents the odp manager.
type Manager interface {
	FetchQualifiedSegments(userID string, options []segment.OptimizelySegmentOption) (segments []string, err error)
	IdentifyUser(userID string) bool
	SendOdpEvent(eventType, action string, identifiers map[string]string, data map[string]interface{}) bool
	Update(apiKey, apiHost string, segmentsToCheck []string)
}

// DefaultOdpManager represents default implementation of odp manager
type DefaultOdpManager struct {
	enabled        bool
	segmentsCache  cache.Cache
	odpConfig      config.Config
	logger         logging.OptimizelyLogProducer
	processing     *semaphore.Weighted
	SegmentManager segment.Manager
	EventManager   event.Manager
}

// WithOdpConfig sets odpConfig option to be passed into the NewOdpManager method
func WithOdpConfig(odpConfig config.Config) OMOptionConfig {
	return func(om *DefaultOdpManager) {
		om.odpConfig = odpConfig
	}
}

// WithSegmentsCache sets cache option to be passed into the NewOdpManager method
func WithSegmentsCache(segmentsCache cache.Cache) OMOptionConfig {
	return func(om *DefaultOdpManager) {
		om.segmentsCache = segmentsCache
	}
}

// WithSegmentManager sets segmentManager option to be passed into the NewOdpManager method
func WithSegmentManager(segmentManager segment.Manager) OMOptionConfig {
	return func(om *DefaultOdpManager) {
		om.SegmentManager = segmentManager
	}
}

// WithEventManager sets eventManager option to be passed into the NewOdpManager method
func WithEventManager(eventManager event.Manager) OMOptionConfig {
	return func(om *DefaultOdpManager) {
		om.EventManager = eventManager
	}
}

// NewOdpManager creates and returns a new instance of DefaultOdpManager.
func NewOdpManager(sdkKey string, disable bool, cacheSize int, cacheTimeoutInSecs int64, options ...OMOptionConfig) *DefaultOdpManager {
	odpManager := &DefaultOdpManager{enabled: !disable,
		logger: logging.GetLogger(sdkKey, "ODPManager"),
	}

	if disable {
		odpManager.logger.Info(utils.OdpNotEnabled)
		return odpManager
	}

	for _, opt := range options {
		opt(odpManager)
	}

	odpManager.processing = semaphore.NewWeighted(int64(maxUpdateWorkers))

	if cacheSize < 0 {
		cacheSize = utils.DefaultSegmentsCacheSize
	}
	if cacheTimeoutInSecs < 0 {
		cacheTimeoutInSecs = int64(utils.DefaultSegmentsCacheTimeout)
	}

	if odpManager.odpConfig == nil {
		odpManager.odpConfig = config.NewConfig("", "", nil)
	}

	if odpManager.SegmentManager == nil {
		options := []segment.SMOptionConfig{segment.WithSegmentsCache(odpManager.segmentsCache), segment.WithOdpConfig(odpManager.odpConfig)}
		odpManager.SegmentManager = segment.NewSegmentManager(sdkKey, cacheSize, cacheTimeoutInSecs, options...)
	}

	if odpManager.EventManager == nil {
		odpManager.EventManager = event.NewBatchEventManager(event.WithSDKKey(sdkKey), event.WithOdpConfig(odpManager.odpConfig))
	}
	return odpManager
}

// FetchQualifiedSegments fetches and returns qualified segments
func (om *DefaultOdpManager) FetchQualifiedSegments(userID string, options []segment.OptimizelySegmentOption) (segments []string, err error) {
	if !om.enabled {
		return nil, errors.New(utils.OdpNotEnabled)
	}

	return om.SegmentManager.FetchQualifiedSegments(userID, options)
}

// IdentifyUser associates a full-stack userid with an established VUID
func (om *DefaultOdpManager) IdentifyUser(userID string) bool {
	if !om.enabled {
		om.logger.Debug(utils.IdentityOdpDisabled)
		return false
	}
	return om.EventManager.IdentifyUser(userID)
}

// SendOdpEvent sends an event to the ODP server.
func (om *DefaultOdpManager) SendOdpEvent(eventType, action string, identifiers map[string]string, data map[string]interface{}) bool {
	if !om.enabled {
		om.logger.Debug(utils.OdpNotEnabled)
		return false
	}
	odpEvent := event.Event{
		Type:        eventType,
		Action:      action,
		Identifiers: identifiers,
		Data:        data,
	}
	return om.EventManager.ProcessEvent(odpEvent)
}

// Update updates odp config.
func (om *DefaultOdpManager) Update(apiKey, apiHost string, segmentsToCheck []string) {
	if !om.enabled {
		return
	}

	// flush old events using old odp publicKey (if exists) before updating odp key.
	// NOTE: It should be rare but possible that odp public key is changed for the same datafile (sdkKey).
	//       Try to send all old events with the previous public key.
	//       If it fails to flush all the old events here (network error), remaning events will be discarded.

	// we just want to start one go routine when update is called, subsequent calls will have to wait.
	if om.processing.TryAcquire(1) {
		go func() {
			om.EventManager.FlushEvents()
			if om.odpConfig.Update(apiKey, apiHost, segmentsToCheck) {
				// reset segments cache when odp integration or segmentsToCheck are changed
				om.SegmentManager.Reset()
			}
			om.processing.Release(1)
		}()
	}
}
