/****************************************************************************
 * Copyright 2022-2025, Optimizely, Inc. and contributors                        *
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

// Package odp //
package odp

import (
	"errors"
	"time"

	"github.com/optimizely/go-sdk/v2/pkg/cache"
	"github.com/optimizely/go-sdk/v2/pkg/logging"
	"github.com/optimizely/go-sdk/v2/pkg/odp/config"
	"github.com/optimizely/go-sdk/v2/pkg/odp/event"
	"github.com/optimizely/go-sdk/v2/pkg/odp/segment"
	"github.com/optimizely/go-sdk/v2/pkg/odp/utils"
)

// OMOptionFunc are the ODPManager options that give you the ability to add one more more options before the odp manager is initialized.
type OMOptionFunc func(em *DefaultOdpManager)

// Manager represents the odp manager.
type Manager interface {
	FetchQualifiedSegments(userID string, options []segment.OptimizelySegmentOption) (segments []string, err error)
	IdentifyUser(userID string)
	SendOdpEvent(eventType, action string, identifiers map[string]string, data map[string]interface{}) (err error)
	Update(apiKey, apiHost string, segmentsToCheck []string)
}

// DefaultOdpManager represents default implementation of odp manager
type DefaultOdpManager struct {
	enabled              bool
	segmentsCacheSize    int
	segmentsCacheTimeout time.Duration
	segmentsCache        cache.Cache
	OdpConfig            config.Config
	logger               logging.OptimizelyLogProducer
	SegmentManager       segment.Manager
	EventManager         event.Manager
}

// WithSegmentsCacheSize sets segmentsCacheSize option to be passed into the NewOdpManager method.
func WithSegmentsCacheSize(segmentsCacheSize int) OMOptionFunc {
	return func(om *DefaultOdpManager) {
		om.segmentsCacheSize = segmentsCacheSize
	}
}

// WithSegmentsCacheTimeout sets segmentsCacheTimeout option to be passed into the NewOdpManager method
func WithSegmentsCacheTimeout(segmentsCacheTimeout time.Duration) OMOptionFunc {
	return func(om *DefaultOdpManager) {
		om.segmentsCacheTimeout = segmentsCacheTimeout
	}
}

// WithSegmentsCache sets cache option to be passed into the NewOdpManager method
func WithSegmentsCache(segmentsCache cache.Cache) OMOptionFunc {
	return func(om *DefaultOdpManager) {
		om.segmentsCache = segmentsCache
	}
}

// WithSegmentManager sets segmentManager option to be passed into the NewOdpManager method
func WithSegmentManager(segmentManager segment.Manager) OMOptionFunc {
	return func(om *DefaultOdpManager) {
		om.SegmentManager = segmentManager
	}
}

// WithEventManager sets eventManager option to be passed into the NewOdpManager method
func WithEventManager(eventManager event.Manager) OMOptionFunc {
	return func(om *DefaultOdpManager) {
		om.EventManager = eventManager
	}
}

// NewOdpManager creates and returns a new instance of DefaultOdpManager.
func NewOdpManager(sdkKey string, disable bool, options ...OMOptionFunc) *DefaultOdpManager {
	odpManager := &DefaultOdpManager{enabled: !disable,
		logger:               logging.GetLogger(sdkKey, "ODPManager"),
		segmentsCacheSize:    utils.DefaultSegmentsCacheSize,
		segmentsCacheTimeout: utils.DefaultSegmentsCacheTimeout,
	}

	if disable {
		odpManager.logger.Info(utils.OdpNotEnabled)
		return odpManager
	}

	for _, opt := range options {
		opt(odpManager)
	}

	odpManager.OdpConfig = config.NewConfig("", "", nil)

	if odpManager.SegmentManager == nil {
		segmentOptions := []segment.SMOptionFunc{}
		if odpManager.segmentsCache != nil {
			segmentOptions = append(segmentOptions, segment.WithSegmentsCache(odpManager.segmentsCache))
		} else {
			segmentOptions = append(segmentOptions, segment.WithSegmentsCacheSize(odpManager.segmentsCacheSize), segment.WithSegmentsCacheTimeout(odpManager.segmentsCacheTimeout))
		}
		odpManager.SegmentManager = segment.NewSegmentManager(sdkKey, segmentOptions...)
	}

	// If user has not provided event manager, create a new one and return
	if odpManager.EventManager == nil {
		odpManager.EventManager = event.NewBatchEventManager(event.WithSDKKey(sdkKey))
		return odpManager
	}
	return odpManager
}

// FetchQualifiedSegments fetches and returns qualified segments
func (om *DefaultOdpManager) FetchQualifiedSegments(userID string, options []segment.OptimizelySegmentOption) (segments []string, err error) {
	if !om.enabled {
		return nil, errors.New(utils.OdpNotEnabled)
	}
	apiKey := om.OdpConfig.GetAPIKey()
	apiHost := om.OdpConfig.GetAPIHost()
	segmentsToCheck := om.OdpConfig.GetSegmentsToCheck()
	return om.SegmentManager.FetchQualifiedSegments(apiKey, apiHost, userID, segmentsToCheck, options)
}

// IdentifyUser associates a full-stack userid with an established VUID
func (om *DefaultOdpManager) IdentifyUser(userID string) {
	if !om.enabled {
		om.logger.Debug(utils.IdentityOdpDisabled)
		return
	}
	om.EventManager.IdentifyUser(om.OdpConfig.GetAPIKey(), om.OdpConfig.GetAPIHost(), userID)
}

// SendOdpEvent sends an event to the ODP server.
func (om *DefaultOdpManager) SendOdpEvent(eventType, action string, identifiers map[string]string, data map[string]interface{}) (err error) {
	if !om.enabled {
		om.logger.Debug(utils.OdpNotEnabled)
		return errors.New(utils.OdpNotEnabled)
	}
	if identifiers == nil {
		identifiers = map[string]string{}
	}
	odpEvent := event.Event{
		Type:        eventType,
		Action:      action,
		Identifiers: identifiers,
		Data:        data,
	}
	return om.EventManager.ProcessEvent(om.OdpConfig.GetAPIKey(), om.OdpConfig.GetAPIHost(), odpEvent)
}

// Update updates odp config.
func (om *DefaultOdpManager) Update(apiKey, apiHost string, segmentsToCheck []string) {
	if !om.enabled {
		return
	}

	// flush old events using old odp publicKey (if exists) before updating odp key.
	// NOTE: It should be rare but possible that odp public key is changed for the same datafile (sdkKey).
	//       Try to send all old events with the previous public key.
	//       If it fails to flush all the old events here (network error), remaining events will be discarded.

	om.EventManager.FlushEvents(om.OdpConfig.GetAPIKey(), om.OdpConfig.GetAPIHost())
	if om.OdpConfig.Update(apiKey, apiHost, segmentsToCheck) {
		// reset segments cache when odp integration or segmentsToCheck are changed
		om.SegmentManager.Reset()
	}
}
