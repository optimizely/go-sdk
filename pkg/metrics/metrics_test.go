/****************************************************************************
 * Copyright 2019, Optimizely, Inc. and contributors                        *
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

// Package metrics //
package metrics

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetCounter(t *testing.T) {
	registry := NewNoopRegistry()

	assert.NotNil(t, registry)

	counter := registry.GetCounter("")
	assert.NotNil(t, counter)
	counter.Add(1)
}

func TestGetGauge(t *testing.T) {
	registry := NewNoopRegistry()

	assert.NotNil(t, registry)

	gauge := registry.GetGauge("")
	assert.NotNil(t, gauge)
	gauge.Set(1)
}
