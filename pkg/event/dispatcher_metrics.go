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

// Package event //
package event

// Metrics is the interface for event processor
type Metrics interface {
	SetQueueSize(queueSize int)
	IncrSuccessFlushCount()
	IncrFailFlushCount()
	IncrRetryFlushCount()
}

// DefaultMetrics stores the actual metrics
type DefaultMetrics struct {
	QueueSize         int
	SuccessFlushCount int64
	FailFlushCount    int64
	RetryFlushCount   int64
}

// SetQueueSize sets the queue size
func (m *DefaultMetrics) SetQueueSize(queueSize int) {
	m.QueueSize = queueSize
}

// IncrSuccessFlushCount increments counter for successful flush
func (m *DefaultMetrics) IncrSuccessFlushCount() {
	m.SuccessFlushCount++
}

// IncrFailFlushCount increments counter for failed flush
func (m *DefaultMetrics) IncrFailFlushCount() {
	m.FailFlushCount++
}

// IncrRetryFlushCount increments counter for retried flush
func (m *DefaultMetrics) IncrRetryFlushCount() {
	m.RetryFlushCount++
}
