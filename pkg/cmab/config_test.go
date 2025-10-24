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

package cmab

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewDefaultConfig(t *testing.T) {
	config := NewDefaultConfig()

	assert.Equal(t, DefaultCacheSize, config.CacheSize)
	assert.Equal(t, DefaultCacheTTL, config.CacheTTL)
	assert.Equal(t, DefaultHTTPTimeout, config.HTTPTimeout)
	assert.NotNil(t, config.RetryConfig)
	assert.Equal(t, DefaultMaxRetries, config.RetryConfig.MaxRetries)
	assert.Nil(t, config.Cache) // Should be nil by default
}
