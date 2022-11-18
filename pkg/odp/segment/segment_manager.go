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

// Package segment //
package segment

import (
	"fmt"

	"github.com/optimizely/go-sdk/pkg/odp/cache"
	"github.com/optimizely/go-sdk/pkg/odp/config"
	"github.com/optimizely/go-sdk/pkg/odp/utils"
)

// SMOptionConfig are the SegmentManager options that give you the ability to add one more more options before the segment manager is initialized.
type SMOptionConfig func(em *DefaultSegmentManager)

// Manager represents the odp segment manager.
type Manager interface {
	FetchQualifiedSegments(userID string, options []OptimizelySegmentOption) (segments []string, err error)
	Reset()
}

// DefaultSegmentManager represents default implementation of odp segment manager
type DefaultSegmentManager struct {
	odpConfig     config.Config
	segmentsCache cache.Cache
	apiManager    APIManager
}

// WithSegmentsCache sets cache option to be passed into the NewSegmentManager method
func WithSegmentsCache(segmentsCache cache.Cache) SMOptionConfig {
	return func(sm *DefaultSegmentManager) {
		sm.segmentsCache = segmentsCache
	}
}

// WithOdpConfig sets odpConfig option to be passed into the NewSegmentManager method
func WithOdpConfig(odpConfig config.Config) SMOptionConfig {
	return func(sm *DefaultSegmentManager) {
		sm.odpConfig = odpConfig
	}
}

// WithAPIManager sets segmentAPIManager as a config option to be passed into the NewSegmentManager method
func WithAPIManager(segmentAPIManager APIManager) SMOptionConfig {
	return func(sm *DefaultSegmentManager) {
		sm.apiManager = segmentAPIManager
	}
}

// NewSegmentManager creates and returns a new instance of DefaultSegmentManager.
func NewSegmentManager(sdkKey string, cacheSize int, cacheTimeoutInSecs int64, options ...SMOptionConfig) *DefaultSegmentManager {
	segmentManager := &DefaultSegmentManager{}
	for _, opt := range options {
		opt(segmentManager)
	}

	if segmentManager.segmentsCache == nil {
		segmentManager.segmentsCache = cache.NewLRUCache(cacheSize, cacheTimeoutInSecs)
	}

	if segmentManager.apiManager == nil {
		segmentManager.apiManager = NewSegmentAPIManager(sdkKey, nil)
	}
	return segmentManager
}

// FetchQualifiedSegments fetches and returns qualified segments
func (s *DefaultSegmentManager) FetchQualifiedSegments(userID string, options []OptimizelySegmentOption) (segments []string, err error) {
	if s.odpConfig == nil || !s.odpConfig.IsOdpServiceIntegrated() {
		return nil, fmt.Errorf(utils.FetchSegmentsFailedError, "apiKey/apiHost not defined")
	}

	// empty segmentsToCheck (no ODP audiences found in datafile) is not an error. return immediately without checking with the ODP server.
	if len(s.odpConfig.GetSegmentsToCheck()) == 0 {
		return []string{}, nil
	}

	cacheKey := MakeCacheKey(userID)
	var ignoreCache = false
	var resetCache = false
	for _, v := range options {
		switch v {
		case IgnoreCache:
			ignoreCache = true
		case ResetCache:
			resetCache = true
		default:
		}
	}

	if resetCache {
		s.Reset()
	}

	if !ignoreCache {
		if fSegments, ok := s.segmentsCache.Lookup(cacheKey).([]string); ok {
			return fSegments, nil
		}
	}

	segments, err = s.apiManager.FetchQualifiedSegments(s.odpConfig, userID)
	if err == nil && len(segments) > 0 && !ignoreCache {
		s.segmentsCache.Save(cacheKey, segments)
	}
	return segments, err
}

// Reset resets segmentsCache.
func (s *DefaultSegmentManager) Reset() {
	s.segmentsCache.Reset()
}

// MakeCacheKey creates and returns cacheKey
func MakeCacheKey(userID string) string {
	return utils.OdpFSUserIDKey + "-$-" + userID
}
