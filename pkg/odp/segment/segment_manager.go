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

// Package segment //
package segment

import (
	"fmt"
	"time"

	"github.com/optimizely/go-sdk/v2/pkg/odp/cache"
	"github.com/optimizely/go-sdk/v2/pkg/odp/utils"
)

// SMOptionFunc are the SegmentManager options that give you the ability to add one more more options before the segment manager is initialized.
type SMOptionFunc func(em *DefaultSegmentManager)

// Manager represents the odp segment manager.
type Manager interface {
	FetchQualifiedSegments(apiKey, apiHost, userID string, segmentsToCheck []string, options []OptimizelySegmentOption) (segments []string, err error)
	Reset()
}

// DefaultSegmentManager represents default implementation of odp segment manager
type DefaultSegmentManager struct {
	segmentsCacheSize    int
	segmentsCacheTimeout time.Duration
	segmentsCache        cache.Cache
	apiManager           APIManager
}

// WithSegmentsCacheSize sets segmentsCacheSize option to be passed into the NewSegmentManager method.
// default value is 10000
func WithSegmentsCacheSize(segmentsCacheSize int) SMOptionFunc {
	return func(om *DefaultSegmentManager) {
		om.segmentsCacheSize = segmentsCacheSize
	}
}

// WithSegmentsCacheTimeout sets segmentsCacheTimeout option to be passed into the NewSegmentManager method
// default value is 600s
func WithSegmentsCacheTimeout(segmentsCacheTimeout time.Duration) SMOptionFunc {
	return func(om *DefaultSegmentManager) {
		om.segmentsCacheTimeout = segmentsCacheTimeout
	}
}

// WithSegmentsCache sets cache option to be passed into the NewSegmentManager method
func WithSegmentsCache(segmentsCache cache.Cache) SMOptionFunc {
	return func(sm *DefaultSegmentManager) {
		sm.segmentsCache = segmentsCache
	}
}

// WithAPIManager sets segmentAPIManager as a config option to be passed into the NewSegmentManager method
func WithAPIManager(segmentAPIManager APIManager) SMOptionFunc {
	return func(sm *DefaultSegmentManager) {
		sm.apiManager = segmentAPIManager
	}
}

// NewSegmentManager creates and returns a new instance of DefaultSegmentManager.
func NewSegmentManager(sdkKey string, options ...SMOptionFunc) *DefaultSegmentManager {
	// Setting default values
	segmentManager := &DefaultSegmentManager{
		segmentsCacheSize:    utils.DefaultSegmentsCacheSize,
		segmentsCacheTimeout: utils.DefaultSegmentsCacheTimeout,
	}

	for _, opt := range options {
		opt(segmentManager)
	}

	if segmentManager.segmentsCache == nil {
		segmentManager.segmentsCache = cache.NewLRUCache(segmentManager.segmentsCacheSize, segmentManager.segmentsCacheTimeout)
	}

	if segmentManager.apiManager == nil {
		segmentManager.apiManager = NewSegmentAPIManager(sdkKey, nil)
	}
	return segmentManager
}

// FetchQualifiedSegments fetches and returns qualified segments
func (s *DefaultSegmentManager) FetchQualifiedSegments(apiKey, apiHost, userID string, segmentsToCheck []string, options []OptimizelySegmentOption) (segments []string, err error) {
	if !s.isOdpServiceIntegrated(apiKey, apiHost) {
		return nil, fmt.Errorf(utils.FetchSegmentsFailedError, "apiKey/apiHost not defined")
	}

	// empty segmentsToCheck (no ODP audiences found in datafile) is not an error. return immediately without checking with the ODP server.
	if len(segmentsToCheck) == 0 {
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

	segments, err = s.apiManager.FetchQualifiedSegments(apiKey, apiHost, userID, segmentsToCheck)
	if err == nil && len(segments) > 0 && !ignoreCache {
		s.segmentsCache.Save(cacheKey, segments)
	}
	return segments, err
}

// Reset resets segmentsCache.
func (s *DefaultSegmentManager) Reset() {
	s.segmentsCache.Reset()
}

// isOdpServiceIntegrated returns true if odp service is integrated
func (s *DefaultSegmentManager) isOdpServiceIntegrated(apiKey, apiHost string) bool {
	return apiKey != "" && apiHost != ""
}

// MakeCacheKey creates and returns cacheKey
func MakeCacheKey(userID string) string {
	return utils.OdpFSUserIDKey + "-$-" + userID
}
