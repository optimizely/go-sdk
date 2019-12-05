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

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewDefaultMetrics(t *testing.T) {

	metric := NewDefaultMetrics()
	assert.Equal(t, 0, metric.QueueSize)
	assert.Equal(t, int64(0), metric.SuccessFlushCount)
	assert.Equal(t, int64(0), metric.FailFlushCount)
	assert.Equal(t, int64(0), metric.RetryFlushCount)

}

func TestSetQueueSize(t *testing.T) {

	metric := NewDefaultMetrics()
	metric.SetQueueSize(24)
	assert.Equal(t, 24, metric.QueueSize)
	assert.Equal(t, int64(0), metric.SuccessFlushCount)
	assert.Equal(t, int64(0), metric.FailFlushCount)
	assert.Equal(t, int64(0), metric.RetryFlushCount)

}

func TestIncrSuccessFlushCount(t *testing.T) {

	metric := NewDefaultMetrics()
	metric.IncrSuccessFlushCount()
	assert.Equal(t, 0, metric.QueueSize)
	assert.Equal(t, int64(1), metric.SuccessFlushCount)
	assert.Equal(t, int64(0), metric.FailFlushCount)
	assert.Equal(t, int64(0), metric.RetryFlushCount)

}

func TestIncrFailFlushCount(t *testing.T) {

	metric := NewDefaultMetrics()
	metric.IncrFailFlushCount()
	assert.Equal(t, 0, metric.QueueSize)
	assert.Equal(t, int64(0), metric.SuccessFlushCount)
	assert.Equal(t, int64(1), metric.FailFlushCount)
	assert.Equal(t, int64(0), metric.RetryFlushCount)

}

func TestIncrRetryFlushCount(t *testing.T) {

	metric := NewDefaultMetrics()
	metric.IncrRetryFlushCount()
	assert.Equal(t, 0, metric.QueueSize)
	assert.Equal(t, int64(0), metric.SuccessFlushCount)
	assert.Equal(t, int64(0), metric.FailFlushCount)
	assert.Equal(t, int64(1), metric.RetryFlushCount)

}
