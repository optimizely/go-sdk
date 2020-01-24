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

package utils

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestIsChannelClosed(t *testing.T) {
	var ch = make(chan struct{})
	assert.Equal(t, false, IsChannelClosed(ch))
	close(ch)
	assert.Equal(t, true, IsChannelClosed(ch))
}

func TestWaitForChannelToCloseOrTimeoutStopsWaitingWhenChannelIsClosed(t *testing.T) {
	var ch = make(chan struct{})
	var finishedWaiting bool
	go func() {
		WaitForChannelToCloseOrTimeout(ch, 10*time.Second)
		finishedWaiting = true
	}()
	close(ch)
	time.Sleep(5 * time.Millisecond)
	// Check WaitForChannelToCloseOrTimeout finishes waiting when channel is closed
	assert.Equal(t, true, finishedWaiting)
}

func TestWaitForChannelToCloseOrTimeoutWaitsForTimeoutWhenChannelIsNotClosed(t *testing.T) {
	var ch = make(chan struct{})
	blockingTimeOut := 1 * time.Second
	var expectedExecutionTime time.Time
	var actualExecutionTime time.Time
	go func() {
		// Keep track of execution time
		start := time.Now()
		expectedExecutionTime = start.Add(blockingTimeOut)
		WaitForChannelToCloseOrTimeout(ch, blockingTimeOut)
		actualExecutionTime = start.Add(time.Since(start))
	}()
	// Wait for execution to finish
	time.Sleep(1005 * time.Millisecond)
	// Verify method timed out
	assert.WithinDuration(t, expectedExecutionTime, actualExecutionTime, 10*time.Millisecond)
}
