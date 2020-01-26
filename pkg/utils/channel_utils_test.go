/****************************************************************************
 * Copyright 2020, Optimizely, Inc. and contributors                        *
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
	"sync/atomic"
	"time"

	"github.com/stretchr/testify/suite"
)

type ChannelUtilsTestSuite struct {
	suite.Suite
	channel chan struct{}
}

func (s *ChannelUtilsTestSuite) SetupTest() {
	s.channel = make(chan struct{})
}

func (s *ChannelUtilsTestSuite) TestIsChannelClosed() {
	s.Equal(false, IsChannelClosed(s.channel))
	close(s.channel)
	s.Equal(true, IsChannelClosed(s.channel))
}

func (s *ChannelUtilsTestSuite) TestWaitForChannelToCloseOrTimeoutStopsWaitingWhenChannelIsClosed() {
	var finishedWaiting int32
	go func() {
		WaitForChannelToCloseOrTimeout(s.channel, 10*time.Second)
		atomic.StoreInt32(&finishedWaiting, 1)
	}()
	close(s.channel)
	time.Sleep(5 * time.Millisecond)
	// Check WaitForChannelToCloseOrTimeout finishes waiting when channel is closed
	s.Equal(true, atomic.LoadInt32(&finishedWaiting) == 1)
}

func (s *ChannelUtilsTestSuite) TestWaitForChannelToCloseOrTimeoutWaitsForTimeoutWhenChannelIsNotClosed() {
	blockingTimeOut := 1 * time.Second
	var expectedExecutionTime time.Time
	var actualExecutionTime time.Time
	go func() {
		// Keep track of execution time
		start := time.Now()
		expectedExecutionTime = start.Add(blockingTimeOut)
		WaitForChannelToCloseOrTimeout(s.channel, blockingTimeOut)
		actualExecutionTime = start.Add(time.Since(start))
	}()
	// Wait for execution to finish
	time.Sleep(1050 * time.Millisecond)
	// Verify method timed out
	s.WithinDuration(expectedExecutionTime, actualExecutionTime, 10*time.Millisecond)
}
