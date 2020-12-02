/****************************************************************************
 * Copyright 2019-2020, Optimizely, Inc. and contributors                   *
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

package client

import (
	"context"
	"errors"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/optimizely/go-sdk/pkg/config"
	"github.com/optimizely/go-sdk/pkg/decide"
	"github.com/optimizely/go-sdk/pkg/decision"
	"github.com/optimizely/go-sdk/pkg/event"
	"github.com/optimizely/go-sdk/pkg/metrics"
	"github.com/optimizely/go-sdk/pkg/utils"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockRequester struct {
	utils.Requester
	mock.Mock
}

func (m *MockRequester) Get(uri string, headers ...utils.Header) (response []byte, responseHeaders http.Header, code int, err error) {
	args := m.Called(headers)
	return args.Get(0).([]byte), args.Get(1).(http.Header), args.Int(2), args.Error(3)
}

type MockDispatcher struct {
	Events []event.LogEvent
}

func (f *MockDispatcher) DispatchEvent(event event.LogEvent) (bool, error) {
	f.Events = append(f.Events, event)
	return true, nil
}

func TestFactoryClientReturnsDefaultClient(t *testing.T) {
	factory := OptimizelyFactory{}

	_, err := factory.Client()
	expectedErr := errors.New("unable to instantiate client: no project config manager, SDK key, or a Datafile provided")
	if assert.Error(t, err) {
		assert.Equal(t, err, expectedErr)
	}
}

func TestClientWithSDKKey(t *testing.T) {
	factory := OptimizelyFactory{SDKKey: "1212"}

	optimizelyClient, err := factory.Client()
	assert.NoError(t, err)
	assert.NotNil(t, optimizelyClient.ConfigManager)
	assert.NotNil(t, optimizelyClient.DecisionService)
	assert.NotNil(t, optimizelyClient.EventProcessor)
}

func TestClientWithPollingConfigManager(t *testing.T) {
	factory := OptimizelyFactory{}

	optimizelyClient, err := factory.Client(WithPollingConfigManager(time.Hour, nil))
	assert.NoError(t, err)
	assert.NotNil(t, optimizelyClient.ConfigManager)
	assert.NotNil(t, optimizelyClient.DecisionService)
	assert.NotNil(t, optimizelyClient.EventProcessor)
}

func TestClientWithPollingConfigManagerDatafileAccessToken(t *testing.T) {
	factory := OptimizelyFactory{}

	optimizelyClient, err := factory.Client(WithPollingConfigManagerDatafileAccessToken(time.Hour, nil, "some_token"))
	assert.NoError(t, err)
	assert.NotNil(t, optimizelyClient.ConfigManager)
	assert.NotNil(t, optimizelyClient.DecisionService)
	assert.NotNil(t, optimizelyClient.EventProcessor)
}
func TestClientWithProjectConfigManagerInOptions(t *testing.T) {
	factory := OptimizelyFactory{}
	mockDatafile := []byte(`{"version":"4"}`)
	configManager := config.NewStaticProjectConfigManagerWithOptions("", config.WithInitialDatafile(mockDatafile))

	optimizelyClient, err := factory.Client(WithConfigManager(configManager))
	assert.NoError(t, err)
	assert.NotNil(t, optimizelyClient.ConfigManager)
	assert.NotNil(t, optimizelyClient.DecisionService)
	assert.NotNil(t, optimizelyClient.EventProcessor)
}

func TestClientWithDecisionServiceAndEventProcessorInOptions(t *testing.T) {
	factory := OptimizelyFactory{}
	mockDatafile := []byte(`{"version":"4"}`)
	configManager := config.NewStaticProjectConfigManagerWithOptions("", config.WithInitialDatafile(mockDatafile))
	decisionService := new(MockDecisionService)
	processor := event.NewBatchEventProcessor(event.WithQueueSize(100), event.WithFlushInterval(100),
		event.WithQueue(event.NewInMemoryQueue(100)), event.WithEventDispatcher(&MockDispatcher{Events: []event.LogEvent{}}))

	optimizelyClient, err := factory.Client(WithConfigManager(configManager), WithDecisionService(decisionService), WithEventProcessor(processor))
	assert.NoError(t, err)
	assert.Equal(t, decisionService, optimizelyClient.DecisionService)
	assert.Equal(t, processor, optimizelyClient.EventProcessor)
}

func TestClientWithCustomCtx(t *testing.T) {
	factory := OptimizelyFactory{}
	ctx, cancel := context.WithCancel(context.Background())
	mockConfigManager := new(MockProjectConfigManager)
	client, err := factory.Client(
		WithConfigManager(mockConfigManager),
		WithContext(ctx),
	)
	assert.NoError(t, err)

	wg := &sync.WaitGroup{}
	wg.Add(1)
	client.execGroup.Go(func(ctx context.Context) {
		<-ctx.Done()
		wg.Done()
	})

	cancel()
	wg.Wait()
}

func TestStaticClient(t *testing.T) {
	factory := OptimizelyFactory{Datafile: []byte(`{"revision": "42", "version": "4"}`)}
	optlyClient, err := factory.StaticClient()
	assert.NoError(t, err)

	parsedConfig, _ := optlyClient.ConfigManager.GetConfig()
	assert.Equal(t, "42", parsedConfig.GetRevision())

	factory = OptimizelyFactory{SDKKey: "key_does_not_exist", Datafile: nil}
	optlyClient, err = factory.StaticClient()
	assert.Error(t, err)
	assert.Nil(t, optlyClient)
}

func TestClientWithCustomDecisionServiceOptions(t *testing.T) {
	factory := OptimizelyFactory{SDKKey: "1212"}

	mockUserProfileService := new(MockUserProfileService)
	mockOverrideStore := new(decision.MapExperimentOverridesStore)
	optimizelyClient, err := factory.Client(
		WithUserProfileService(mockUserProfileService),
		WithExperimentOverrides(mockOverrideStore),
	)
	assert.NoError(t, err)
	assert.NotNil(t, optimizelyClient.DecisionService)
}

func TestClientWithEventDispatcher(t *testing.T) {
	factory := OptimizelyFactory{SDKKey: "1212"}

	mockEventDispatcher := new(MockDispatcher)
	optimizelyClient, err := factory.Client(WithEventDispatcher(mockEventDispatcher))
	assert.NoError(t, err)

	dispatcher := optimizelyClient.EventProcessor.(*event.BatchEventProcessor).EventDispatcher
	assert.Equal(t, dispatcher, mockEventDispatcher)
}

func TestClientMetrics(t *testing.T) {
	factory := OptimizelyFactory{SDKKey: "1212"}

	metricsRegistry := metrics.NewNoopRegistry()

	mockEventDispatcher := new(MockDispatcher)
	optimizelyClient, err := factory.Client(WithEventDispatcher(mockEventDispatcher), WithMetricsRegistry(metricsRegistry))
	assert.NoError(t, err)

	eventProcessor := optimizelyClient.EventProcessor.(*event.BatchEventProcessor)
	assert.NotNil(t, eventProcessor)
}

func TestClientWithDatafileAccessToken(t *testing.T) {
	factory := OptimizelyFactory{SDKKey: "1212"}
	accessToken := "some_token"
	optimizelyClient, err := factory.Client(WithDatafileAccessToken(accessToken))
	assert.NoError(t, err)
	assert.NotNil(t, optimizelyClient.ConfigManager)
	assert.NotNil(t, optimizelyClient.DecisionService)
	assert.NotNil(t, optimizelyClient.EventProcessor)

	assert.Equal(t, accessToken, factory.DatafileAccessToken)
}

func TestClientWithDefaultDecideOptions(t *testing.T) {
	decideOptions := []decide.OptimizelyDecideOptions{
		decide.DisableDecisionEvent,
		decide.EnabledFlagsOnly,
	}
	factory := OptimizelyFactory{SDKKey: "1212"}
	optimizelyClient, err := factory.Client(WithDefaultDecideOptions(decideOptions))
	assert.NoError(t, err)
	assert.Equal(t, convertDecideOptions(decideOptions), optimizelyClient.defaultDecideOptions)

	// Verify that defaultDecideOptions are initialized as empty by default
	factory = OptimizelyFactory{SDKKey: "1212"}
	optimizelyClient, err = factory.Client()
	assert.NoError(t, err)
	assert.Equal(t, &decide.Options{}, optimizelyClient.defaultDecideOptions)
}
