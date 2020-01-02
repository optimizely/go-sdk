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
	"sync"
	"time"
)

// GenericMetrics provides the interface for the metrics
type GenericMetrics interface {
	Inc(key string)
	Set(key string, val int64)
	Get(key string) int64
}

// Metrics contains default metrics
type Metrics struct {
	startTime time.Time

	metricsLock sync.RWMutex
	metricsData map[string]int64
}

// NewMetrics makes thread-safe map to collect any counts/metrics
func NewMetrics() *Metrics {
	return &Metrics{startTime: time.Now(), metricsData: map[string]int64{}}
}

// Add increments value for given key and returns new value
func (m *Metrics) add(key string, delta int64) int64 {

	m.metricsLock.Lock()
	defer m.metricsLock.Unlock()
	m.metricsData[key] += delta
	return m.metricsData[key]
}

// Inc increments value for given key by one
func (m *Metrics) Inc(key string) {
	m.add(key, 1)
}

// Set value for given key
func (m *Metrics) Set(key string, val int64) {
	m.metricsLock.Lock()
	defer m.metricsLock.Unlock()
	m.metricsData[key] = val
}

// Get returns value for given key
func (m *Metrics) Get(key string) int64 {
	m.metricsLock.RLock()
	defer m.metricsLock.RUnlock()

	return m.metricsData[key]
}
