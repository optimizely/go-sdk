/****************************************************************************
 * Copyright 2025, Optimizely, Inc. and contributors                        *
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

// Package cmab provides contextual multi-armed bandit functionality
package cmab

import (
	"time"

	"github.com/optimizely/go-sdk/v2/pkg/cache"
)

const (
	// DefaultCacheSize is the default size for CMAB cache (aligned with ODP segments cache)
	DefaultCacheSize = 10000
	// DefaultCacheTTL is the default TTL for CMAB cache (30 minutes to match agent)
	DefaultCacheTTL = 30 * time.Minute

	// DefaultHTTPTimeout is the default HTTP timeout for CMAB requests
	DefaultHTTPTimeout = 10 * time.Second
)

// Config holds CMAB configuration options
type Config struct {
	CacheSize   int
	CacheTTL    time.Duration
	HTTPTimeout time.Duration
	RetryConfig *RetryConfig
	Cache       cache.CacheWithRemove // Custom cache implementation (Redis, etc.)
}

// NewDefaultConfig creates a Config with default values
func NewDefaultConfig() Config {
	return Config{
		CacheSize:   DefaultCacheSize,
		CacheTTL:    DefaultCacheTTL,
		HTTPTimeout: DefaultHTTPTimeout,
		RetryConfig: &RetryConfig{
			MaxRetries: DefaultMaxRetries,
		},
	}
}
