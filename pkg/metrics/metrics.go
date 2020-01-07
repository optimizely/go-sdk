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

// Counter interface
type Counter interface {
	Add(delta float64)
}

// Gauge interface
type Gauge interface {
	Set(delta float64)
}

// Registry provides the interface for the metric registry
type Registry interface {
	GetCounter(name string) Counter
	GetGauge(name string) Gauge
}

// BasicCounter implements Counter interface, provides minimal implementation
type BasicCounter struct{}

// Add implements the method from Counter interface
func (m BasicCounter) Add(value float64) {}

// BasicGauge implements Gauge interface, provides minimal implementation
type BasicGauge struct{}

// Set implements the method from Gauge interface
func (m BasicGauge) Set(value float64) {}

// BasicRegistry contains default metrics registry, provides minimal implementation
type BasicRegistry struct{}

// NewRegistry returns basic registry
func NewRegistry() *BasicRegistry {
	return &BasicRegistry{}
}

// GetCounter gets the Counter
func (m *BasicRegistry) GetCounter(key string) Counter {
	return &BasicCounter{}
}

// GetGauge gets the Gauge
func (m *BasicRegistry) GetGauge(key string) Gauge {
	return &BasicGauge{}
}
