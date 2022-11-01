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

	"github.com/optimizely/go-sdk/pkg/odp/utils"
)

// SegmentManager is used to represent odp segment manager
type SegmentManager struct {
	odpConfig         *Config
	segmentsCache     Cache
	segmentAPIManager *SegmentAPIManager
}

// NewSegmentManager creates and returns a new instance of SegmentManager.
func NewSegmentManager(cacheSize int, cacheTimeoutInSecs int64, odpConfig *Config, apiManager *SegmentAPIManager) SegmentManager {
	segmentManager := SegmentManager{
		odpConfig:         odpConfig,
		segmentAPIManager: apiManager,
		segmentsCache:     NewLRUCache(cacheSize, cacheTimeoutInSecs),
	}
	if segmentManager.odpConfig == nil {
		segmentManager.odpConfig = NewConfig("", "", nil)
	}
	if segmentManager.segmentAPIManager == nil {
		segmentManager.segmentAPIManager = NewSegmentAPIManager(nil)
	}
	return segmentManager
}

// FetchQualifiedSegments fetches and returns qualified segments
func (s *SegmentManager) FetchQualifiedSegments(userKey, userValue string, options []SegmentOption) (segments []string, err error) {
	if s.odpConfig == nil || s.odpConfig.apiHost == "" || s.odpConfig.apiKey == "" {
		return nil, fmt.Errorf(fetchSegmentsFailedError, "apiKey/apiHost not defined")
	}

	// empty segmentsToCheck (no ODP audiences found in datafile) is not an error. return immediately without checking with the ODP server.
	if len(s.odpConfig.segmentsToCheck) == 0 {
		return []string{}, nil
	}

	cacheKey := utils.MakeCacheKey(userKey, userValue)
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

	segments, err = s.segmentAPIManager.FetchSegments(s.odpConfig.apiKey, s.odpConfig.apiHost, userKey, userValue, s.odpConfig.segmentsToCheck)
	if err == nil && len(segments) > 0 && !ignoreCache {
		s.segmentsCache.Save(cacheKey, segments)
	}
	return segments, err
}

// Reset resets segmentsCache.
func (s *SegmentManager) Reset() {
	s.segmentsCache.Reset()
}
