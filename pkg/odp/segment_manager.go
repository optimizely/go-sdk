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
	"fmt"
)

// SMOptionConfig are the SegmentManager options that give you the ability to add one more more options before the segment manager is initialized.
type SMOptionConfig func(em *DefaultSegmentManager)

// SegmentManager represents the odp segment manager.
type SegmentManager interface {
	FetchQualifiedSegments(userKey, userValue string, options []OptimizelySegmentOption) (segments []string, err error)
	Reset()
}

// DefaultSegmentManager represents default implementation of odp segment manager
type DefaultSegmentManager struct {
	odpConfig         Config
	segmentsCache     Cache
	segmentAPIManager SegmentAPIManager
}

// WithSegmentsCache sets cache option to be passed into the NewSegmentManager method
func WithSegmentsCache(cache Cache) SMOptionConfig {
	return func(sm *DefaultSegmentManager) {
		sm.segmentsCache = cache
	}
}

// WithODPConfig sets odpConfig option to be passed into the NewSegmentManager method
func WithODPConfig(odpConfig Config) SMOptionConfig {
	return func(sm *DefaultSegmentManager) {
		sm.odpConfig = odpConfig
	}
}

// WithAPIManager sets segmentAPIManager as a config option to be passed into the NewSegmentManager method
func WithAPIManager(segmentAPIManager SegmentAPIManager) SMOptionConfig {
	return func(sm *DefaultSegmentManager) {
		sm.segmentAPIManager = segmentAPIManager
	}
}

// NewSegmentManager creates and returns a new instance of DefaultSegmentManager.
func NewSegmentManager(sdkKey string, cacheSize int, cacheTimeoutInSecs int64, options ...SMOptionConfig) *DefaultSegmentManager {
	segmentManager := &DefaultSegmentManager{}
	for _, opt := range options {
		opt(segmentManager)
	}

	if segmentManager.segmentsCache == nil {
		segmentManager.segmentsCache = NewLRUCache(cacheSize, cacheTimeoutInSecs)
	}

	if segmentManager.segmentAPIManager == nil {
		segmentManager.segmentAPIManager = NewSegmentAPIManager(sdkKey, nil)
	}
	return segmentManager
}

// FetchQualifiedSegments fetches and returns qualified segments
func (s *DefaultSegmentManager) FetchQualifiedSegments(userKey, userValue string, options []OptimizelySegmentOption) (segments []string, err error) {
	if s.odpConfig == nil || !s.odpConfig.IsOdpServiceIntegrated() {
		return nil, fmt.Errorf(fetchSegmentsFailedError, "apiKey/apiHost not defined")
	}

	// empty segmentsToCheck (no ODP audiences found in datafile) is not an error. return immediately without checking with the ODP server.
	if len(s.odpConfig.GetSegmentsToCheck()) == 0 {
		return []string{}, nil
	}

	cacheKey := MakeCacheKey(userKey, userValue)
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

	segments, err = s.segmentAPIManager.FetchQualifiedSegments(s.odpConfig, userKey, userValue)
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
func MakeCacheKey(userKey, userValue string) string {
	return userKey + "-$-" + userValue
}
