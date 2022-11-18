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

// Package config //
package config

import (
	"sync"

	"github.com/optimizely/go-sdk/pkg/utils"
)

// Config is used to represent odp config
type Config interface {
	Update(apiKey, apiHost string, segmentsToCheck []string) bool
	GetAPIKey() string
	GetAPIHost() string
	GetSegmentsToCheck() []string
	IsOdpServiceIntegrated() bool
}

// DefaultConfig represents default implementation of odp config
type DefaultConfig struct {
	apiKey, apiHost string
	segmentsToCheck []string
	lock            sync.RWMutex
}

// NewConfig creates and returns a new instance of DefaultConfig.
func NewConfig(apiKey, apiHost string, segmentsToCheck []string) *DefaultConfig {
	return &DefaultConfig{
		apiKey:          apiKey,
		apiHost:         apiHost,
		segmentsToCheck: segmentsToCheck,
	}
}

// Update updates config.
func (s *DefaultConfig) Update(apiKey, apiHost string, segmentsToCheck []string) bool {
	s.lock.Lock()
	defer s.lock.Unlock()

	if s.apiKey == apiKey && s.apiHost == apiHost && utils.CompareSlices(s.segmentsToCheck, segmentsToCheck) {
		return false
	}
	s.apiKey = apiKey
	s.apiHost = apiHost
	s.segmentsToCheck = segmentsToCheck
	return true
}

// GetAPIKey returns value for APIKey.
func (s *DefaultConfig) GetAPIKey() string {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return s.apiKey
}

// GetAPIHost returns value for APIHost.
func (s *DefaultConfig) GetAPIHost() string {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return s.apiHost
}

// GetSegmentsToCheck returns an array of all ODP segments used in the current datafile (associated with apiHost/apiKey).
func (s *DefaultConfig) GetSegmentsToCheck() []string {
	s.lock.RLock()
	defer s.lock.RUnlock()
	if s.segmentsToCheck == nil {
		return nil
	}
	segmentsToCheck := make([]string, len(s.segmentsToCheck))
	copy(segmentsToCheck, s.segmentsToCheck)
	return segmentsToCheck
}

// IsOdpServiceIntegrated returns true if odp service is integrated
func (s *DefaultConfig) IsOdpServiceIntegrated() bool {
	return s.GetAPIHost() != "" && s.GetAPIKey() != ""
}
