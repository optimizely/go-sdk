/****************************************************************************
 * Copyright 2022-2025, Optimizely, Inc. and contributors                   *
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

// Package odp provides compatibility with the previously located cache package.
// This file exists to maintain backward compatibility with code that imports
// cache from the odp package. New code should import from pkg/cache directly.
package odp

import (
	"time"

	"github.com/optimizely/go-sdk/v2/pkg/cache"
)

// LRUCache wraps the cache.LRUCache to maintain backward compatibility
type LRUCache struct {
	*cache.LRUCache
}

// NewLRUCache returns a new instance of Least Recently Used in-memory cache
// Deprecated: For new code, use pkg/cache directly instead.
// This function exists for backward compatibility with code that imports from pkg/odp
func NewLRUCache(size int, timeout time.Duration) *LRUCache {
	return &LRUCache{
		LRUCache: cache.NewLRUCache(size, timeout),
	}
}
