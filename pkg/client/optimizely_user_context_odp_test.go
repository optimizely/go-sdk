/****************************************************************************
 * Copyright 2022, Optimizely, Inc. and contributors                        *
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

package client

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"sync"
	"testing"

	"github.com/optimizely/go-sdk/pkg/config/datafileprojectconfig"
	"github.com/optimizely/go-sdk/pkg/logging"
	"github.com/optimizely/go-sdk/pkg/odp"
	"github.com/optimizely/go-sdk/pkg/odp/config"
	"github.com/optimizely/go-sdk/pkg/odp/event"
	"github.com/optimizely/go-sdk/pkg/odp/segment"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type OptimizelyUserContextODPTestSuite struct {
	suite.Suite
	userID   string
	doOnce   sync.Once // required since we only need to read datafile once
	datafile []byte
}

type MockSegmentManager struct {
	mock.Mock
	segment.Manager
}

func (m *MockSegmentManager) FetchQualifiedSegments(odpConfig config.Config, userID string, options []segment.OptimizelySegmentOption) (segments []string, err error) {
	args := m.Called(odpConfig, userID, options)
	if segArray, ok := args.Get(0).([]string); ok {
		segments = segArray
	}
	return segments, args.Error(1)
}

func (m *MockSegmentManager) Reset() {
	m.Called()
}

type MockEventAPIManager struct {
	wg         sync.WaitGroup
	eventsSent []event.Event // To assert number of events successfully sent
}

func (m *MockEventAPIManager) SendOdpEvents(apiKey, apiHost string, events []event.Event) (canRetry bool, err error) {
	m.eventsSent = append(m.eventsSent, events...)
	m.wg.Done()
	return
}

func (o *OptimizelyUserContextODPTestSuite) SetupTest() {
	o.doOnce.Do(func() {
		absPath, _ := filepath.Abs("../../test-data/odp-test-datafile.json")
		o.datafile, _ = ioutil.ReadFile(absPath)
	})
	o.userID = "tester"
}

func (o *OptimizelyUserContextODPTestSuite) TestIdentifyAndUpdateCalledAutomatically() {
	odpManager := &MockODPManager{}
	projectConfig, _ := datafileprojectconfig.NewDatafileProjectConfig(o.datafile, logging.GetLogger("", ""))
	odpManager.On("IdentifyUser", o.userID)
	odpManager.On("Update", projectConfig.GetPublicKeyForODP(), projectConfig.GetHostForODP(), projectConfig.GetSegmentList())
	factory := OptimizelyFactory{Datafile: o.datafile, odpManager: odpManager}
	optimizelyClient, _ := factory.Client()
	_ = optimizelyClient.CreateUserContext(o.userID, nil)
	odpManager.AssertExpectations(o.T())
}

func (o *OptimizelyUserContextODPTestSuite) TestIsQualifiedFor() {
	factory := OptimizelyFactory{Datafile: o.datafile}
	optimizelyClient, _ := factory.Client()
	userContext := optimizelyClient.CreateUserContext(o.userID, nil)
	o.False(userContext.IsQualifiedFor("a"))
	userContext.SetQualifiedSegments([]string{"a", "b"})
	o.True(userContext.IsQualifiedFor("a"))
	o.False(userContext.IsQualifiedFor("x"))
	userContext.SetQualifiedSegments([]string{})
	o.False(userContext.IsQualifiedFor("a"))
}

func (o *OptimizelyUserContextODPTestSuite) TestFetchQualifiedSegmentsSuccessDefaultUserAsync() {
	segmentManager := &MockSegmentManager{}
	segmentManager.On("Reset")
	segmentManager.On("FetchQualifiedSegments", mock.Anything, o.userID, mock.Anything).Return([]string{"odp-segment-1"}, nil)
	odpManager := odp.NewOdpManager("", false, odp.WithSegmentManager(segmentManager))
	factory := OptimizelyFactory{Datafile: o.datafile, odpManager: odpManager}
	optimizelyClient, _ := factory.Client()
	userContext := optimizelyClient.CreateUserContext(o.userID, nil)
	var wg sync.WaitGroup
	wg.Add(1)
	userContext.FetchQualifiedSegmentsAsync(func(success bool) {
		o.True(success)
		o.Equal([]string{"odp-segment-1"}, userContext.GetQualifiedSegments())
		wg.Done()
	}, nil)
	wg.Wait()
	segmentManager.AssertExpectations(o.T())
}

func (o *OptimizelyUserContextODPTestSuite) TestFetchQualifiedSegmentsSuccessDefaultUserSync() {
	segmentManager := &MockSegmentManager{}
	segmentManager.On("Reset")
	segmentManager.On("FetchQualifiedSegments", mock.Anything, o.userID, mock.Anything).Return([]string{"odp-segment-1"}, nil)
	odpManager := odp.NewOdpManager("", false, odp.WithSegmentManager(segmentManager))
	factory := OptimizelyFactory{Datafile: o.datafile, odpManager: odpManager}
	optimizelyClient, _ := factory.Client()
	userContext := optimizelyClient.CreateUserContext(o.userID, nil)
	o.Nil(userContext.GetQualifiedSegments())
	userContext.FetchQualifiedSegments(nil)
	o.Equal(userContext.GetQualifiedSegments(), []string{"odp-segment-1"})
	segmentManager.AssertExpectations(o.T())
}

func (o *OptimizelyUserContextODPTestSuite) TestFetchQualifiedSegmentsSDKNotReady() {
	factory := OptimizelyFactory{SDKKey: "121"}
	client, _ := factory.Client()
	userContext := client.CreateUserContext(o.userID, nil)
	var wg sync.WaitGroup
	wg.Add(1)
	userContext.FetchQualifiedSegmentsAsync(func(success bool) {
		o.False(success)
		wg.Done()
	}, nil)
	wg.Wait()
}

func (o *OptimizelyUserContextODPTestSuite) TestFetchQualifiedSegmentsFetchFailed() {
	// ODP apiKey is not available
	mockDatafile := []byte(`{"version":"4","integrations": [{"publicKey": "", "host": "www.123.com", "key": "odp"}]}`)
	factory := OptimizelyFactory{Datafile: mockDatafile}
	optimizelyClient, _ := factory.Client()
	userContext := optimizelyClient.CreateUserContext(o.userID, nil)
	userContext.SetQualifiedSegments([]string{"dummy"})
	var wg sync.WaitGroup
	wg.Add(1)
	userContext.FetchQualifiedSegmentsAsync(func(success bool) {
		o.False(success)
		o.Nil(userContext.GetQualifiedSegments())
		wg.Done()
	}, nil)
	wg.Wait()
}

func (o *OptimizelyUserContextODPTestSuite) TestFetchQualifiedSegmentsSegmentsToCheckValidAfterStart() {
	odpManager := odp.NewOdpManager("", false)
	factory := OptimizelyFactory{Datafile: o.datafile, odpManager: odpManager}
	optimizelyClient, _ := factory.Client()
	userContext := optimizelyClient.CreateUserContext(o.userID, nil)
	var wg sync.WaitGroup
	wg.Add(1)
	userContext.FetchQualifiedSegmentsAsync(func(success bool) {
		wg.Done()
	}, nil)
	wg.Wait()
	o.Equal([]string{"odp-segment-1", "odp-segment-2", "odp-segment-3"}, odpManager.OdpConfig.GetSegmentsToCheck())
}

func (o *OptimizelyUserContextODPTestSuite) TestFetchQualifiedSegmentsSegmentsSegmentsNotUsed() {
	mockDatafile := []byte(`{"version":"4","integrations": [{"publicKey": "W4WzcEs-ABgXorzY7h1LCQ", "host": "https://api.zaius.com", "key": "odp"}]}`)
	odpManager := odp.NewOdpManager("", false)
	factory := OptimizelyFactory{Datafile: mockDatafile, odpManager: odpManager}
	optimizelyClient, _ := factory.Client()
	userContext := optimizelyClient.CreateUserContext(o.userID, nil)
	var wg sync.WaitGroup
	wg.Add(1)
	userContext.FetchQualifiedSegmentsAsync(func(success bool) {
		o.True(success)
		o.Equal([]string{}, userContext.GetQualifiedSegments())
		wg.Done()
	}, nil)
	wg.Wait()
}

func (o *OptimizelyUserContextODPTestSuite) TestFetchQualifiedSegmentsParameters() {
	segmentManager := &MockSegmentManager{}
	segmentManager.On("Reset")
	segmentManager.On("FetchQualifiedSegments", mock.Anything, o.userID, []segment.OptimizelySegmentOption{segment.IgnoreCache}).Return([]string{"odp-segment-1"}, nil)
	odpManager := odp.NewOdpManager("", false, odp.WithSegmentManager(segmentManager))
	factory := OptimizelyFactory{Datafile: o.datafile, odpManager: odpManager}
	optimizelyClient, _ := factory.Client()
	userContext := optimizelyClient.CreateUserContext(o.userID, nil)
	userContext.FetchQualifiedSegments([]segment.OptimizelySegmentOption{segment.IgnoreCache})
	o.Equal([]string{"odp-segment-1"}, userContext.GetQualifiedSegments())
	o.Equal([]string{"odp-segment-1", "odp-segment-2", "odp-segment-3"}, odpManager.OdpConfig.GetSegmentsToCheck())
	segmentManager.AssertExpectations(o.T())
}

func (o *OptimizelyUserContextODPTestSuite) TestOdpEventsEarlyEventsDispatched() {
	eventAPIManager := &MockEventAPIManager{}
	odpConfig := config.NewConfig("", "", []string{})
	eventManager := event.NewBatchEventManager(event.WithAPIManager(eventAPIManager), event.WithBatchSize(1), event.WithOdpConfig(odpConfig))
	odpManager := odp.NewOdpManager("", false, odp.WithEventManager(eventManager))
	eventManager.OdpConfig = odpManager.OdpConfig
	factory := OptimizelyFactory{Datafile: o.datafile, odpManager: odpManager}
	optimizelyClient, _ := factory.Client()
	eventAPIManager.wg.Add(1)
	// identified event will be sent
	_ = optimizelyClient.CreateUserContext(o.userID, nil)
	eventAPIManager.wg.Wait()

	o.Equal(1, len(eventAPIManager.eventsSent))

	expectedEvents := 100
	eventAPIManager.wg.Add(expectedEvents)
	for i := 0; i < expectedEvents; i++ {
		_ = optimizelyClient.CreateUserContext(fmt.Sprintf("%d", i), nil)
	}
	eventAPIManager.wg.Wait()
	o.Equal(expectedEvents+1, len(eventAPIManager.eventsSent))
}

// Tests with live ODP server
// func (o *OptimizelyUserContextODPTestSuite) TestLiveOdpGraphQL() {
// 	o.userID = "tester-101"
// 	factory := OptimizelyFactory{Datafile: o.datafile}
// 	optimizelyClient, _ := factory.Client()
// 	userContext := optimizelyClient.CreateUserContext(o.userID, nil)
// 	var wg sync.WaitGroup
// 	wg.Add(1)
// 	userContext.FetchQualifiedSegments(nil, func(segments []string, err error) {
// 		o.NoError(err)
// 		o.Equal([]string{}, segments)
// 		wg.Done()
// 	})
// 	wg.Wait()
// }

func TestOptimizelyUserContextODPTestSuite(t *testing.T) {
	suite.Run(t, new(OptimizelyUserContextODPTestSuite))
}
