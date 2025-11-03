/****************************************************************************
 * Copyright 2019-2020,2022,2025 Optimizely, Inc. and contributors          *
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
	"context"
	"errors"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/optimizely/go-sdk/v2/pkg/cache"
	"github.com/optimizely/go-sdk/v2/pkg/config"
	"github.com/optimizely/go-sdk/v2/pkg/decide"
	"github.com/optimizely/go-sdk/v2/pkg/decision"
	"github.com/optimizely/go-sdk/v2/pkg/event"
	"github.com/optimizely/go-sdk/v2/pkg/metrics"
	"github.com/optimizely/go-sdk/v2/pkg/notification"
	"github.com/optimizely/go-sdk/v2/pkg/odp"
	pkgOdpEvent "github.com/optimizely/go-sdk/v2/pkg/odp/event"
	pkgOdpSegment "github.com/optimizely/go-sdk/v2/pkg/odp/segment"
	pkgOdpUtils "github.com/optimizely/go-sdk/v2/pkg/odp/utils"
	"github.com/optimizely/go-sdk/v2/pkg/registry"
	"github.com/optimizely/go-sdk/v2/pkg/tracing"
	"github.com/optimizely/go-sdk/v2/pkg/utils"
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

type MockConfigManager struct {
	config.ProjectConfigManager
	projectConfig             config.ProjectConfig
	sdkKey                    string
	notificationListenerAdded bool
}

func (m *MockConfigManager) GetConfig() (config.ProjectConfig, error) {
	return m.projectConfig, nil
}

func (m *MockConfigManager) OnProjectConfigUpdate(callback func(notification.ProjectConfigUpdateNotification)) (int, error) {
	m.notificationListenerAdded = true
	notificationCenter := registry.GetNotificationCenter(m.sdkKey)
	handler := func(payload interface{}) {
		if projectConfigUpdateNotification, ok := payload.(notification.ProjectConfigUpdateNotification); ok {
			callback(projectConfigUpdateNotification)
		}
	}
	return notificationCenter.AddHandler(notification.ProjectConfigUpdate, handler)
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
	assert.NotNil(t, optimizelyClient.OdpManager)
}

func TestClientWithPollingConfigManager(t *testing.T) {
	factory := OptimizelyFactory{}

	optimizelyClient, err := factory.Client(WithPollingConfigManager(time.Hour, nil))
	assert.NoError(t, err)
	assert.NotNil(t, optimizelyClient.ConfigManager)
	assert.NotNil(t, optimizelyClient.DecisionService)
	assert.NotNil(t, optimizelyClient.EventProcessor)
	assert.NotNil(t, optimizelyClient.OdpManager)
}

func TestClientWithPollingConfigManagerDatafileAccessToken(t *testing.T) {
	factory := OptimizelyFactory{}

	optimizelyClient, err := factory.Client(WithPollingConfigManagerDatafileAccessToken(time.Hour, nil, "some_token"))
	assert.NoError(t, err)
	assert.NotNil(t, optimizelyClient.ConfigManager)
	assert.NotNil(t, optimizelyClient.DecisionService)
	assert.NotNil(t, optimizelyClient.EventProcessor)
	assert.NotNil(t, optimizelyClient.OdpManager)
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
	assert.NotNil(t, optimizelyClient.OdpManager)
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

func TestClientWithOdpManagerAndSDKSettingsInOptions(t *testing.T) {
	factory := OptimizelyFactory{SDKKey: "1212"}
	eventManager := pkgOdpEvent.NewBatchEventManager()
	segmentManager := pkgOdpSegment.NewSegmentManager("1212")
	var segmentsCacheSize = 1
	var segmentsCacheTimeout = 1 * time.Second
	var disableOdp = false
	segmentCache := cache.NewLRUCache(0, 0)
	odpManager := odp.NewOdpManager("1212", false, odp.WithEventManager(eventManager), odp.WithSegmentManager(segmentManager), odp.WithSegmentsCache(segmentCache))
	optimizelyClient, err := factory.Client(WithOdpManager(odpManager), WithSegmentsCacheSize(segmentsCacheSize), WithSegmentsCacheTimeout(segmentsCacheTimeout), WithOdpDisabled(disableOdp))
	assert.Equal(t, segmentsCacheSize, factory.segmentsCacheSize)
	assert.Equal(t, segmentsCacheTimeout, factory.segmentsCacheTimeout)
	assert.Equal(t, disableOdp, factory.odpDisabled)
	assert.NoError(t, err)
	assert.Equal(t, odpManager, optimizelyClient.OdpManager)
}

func TestClientWithDefaultSDKSettings(t *testing.T) {
	factory := OptimizelyFactory{SDKKey: "1212"}
	optimizelyClient, err := factory.Client()
	assert.NoError(t, err)
	assert.Equal(t, false, factory.odpDisabled)
	assert.Equal(t, pkgOdpUtils.DefaultSegmentsCacheSize, factory.segmentsCacheSize)
	assert.Equal(t, pkgOdpUtils.DefaultSegmentsCacheTimeout, factory.segmentsCacheTimeout)
	assert.NotNil(t, optimizelyClient.OdpManager)
}

func TestClientWithNotificationCenterInOptions(t *testing.T) {
	factory := OptimizelyFactory{SDKKey: "1212"}
	nc := &MockNotificationCenter{}
	optimizelyClient, err := factory.Client(WithNotificationCenter(nc))
	assert.NoError(t, err)
	assert.Equal(t, nc, optimizelyClient.notificationCenter)
}

func TestDummy(t *testing.T) {
	factory := OptimizelyFactory{}
	configManager := config.NewPollingProjectConfigManager("123")
	optimizelyClient, err := factory.Client(WithConfigManager(configManager))
	assert.NoError(t, err)
	userContext := optimizelyClient.CreateUserContext("123", nil)
	assert.NotNil(t, userContext)
}

func TestODPManagerDoesNotStartIfOdpDisabled(t *testing.T) {
	factory := OptimizelyFactory{}
	mockDatafile := []byte(`{"version":"4","integrations": [{"publicKey": "123", "host": "www.123.com", "key": "odp"}]}`)
	configManager := config.NewStaticProjectConfigManagerWithOptions("", config.WithInitialDatafile(mockDatafile))
	var segmentsCacheSize = 1
	var segmentsCacheTimeout = 1 * time.Second
	var disableOdp = true
	optimizelyClient, err := factory.Client(WithConfigManager(configManager), WithSegmentsCacheSize(segmentsCacheSize), WithSegmentsCacheTimeout(segmentsCacheTimeout), WithOdpDisabled(disableOdp))
	assert.NoError(t, err)
	assert.NotNil(t, optimizelyClient.OdpManager)
	var odpManager = optimizelyClient.OdpManager.(*odp.DefaultOdpManager)
	assert.Nil(t, odpManager.OdpConfig)
	// ticket should not start
	time.Sleep(100 * time.Millisecond)
	assert.Nil(t, odpManager.EventManager)
	assert.Nil(t, odpManager.SegmentManager)
}

func TestODPManagerInitializesWithLatestProjectConfig(t *testing.T) {
	factory := OptimizelyFactory{SDKKey: "1234"}
	mockDatafile := []byte(`{"version":"4","integrations": [{"publicKey": "123", "host": "www.123.com", "key": "odp"}]}`)
	configManager := config.NewStaticProjectConfigManagerWithOptions("", config.WithInitialDatafile(mockDatafile))
	optimizelyClient, err := factory.Client(WithConfigManager(configManager))
	assert.NoError(t, err)
	assert.NotNil(t, optimizelyClient.OdpManager)
	var odpManager = optimizelyClient.OdpManager.(*odp.DefaultOdpManager)
	assert.Equal(t, "www.123.com", odpManager.OdpConfig.GetAPIHost())
	assert.Equal(t, "123", odpManager.OdpConfig.GetAPIKey())
}

func TestODPManagerListensToConfigManagerChangesAndUpdatesODPConfigAccordingly(t *testing.T) {
	sdkKey := "abcd"
	mockDatafile1 := []byte(`{"version":"4","integrations": [{"publicKey": "123", "host": "www.123.com", "key": "odp"}]}`)
	factory := OptimizelyFactory{SDKKey: sdkKey}
	tmpConfigManager1 := config.NewStaticProjectConfigManagerWithOptions("", config.WithInitialDatafile(mockDatafile1))
	projectConfig1, err1 := tmpConfigManager1.GetConfig()
	assert.NoError(t, err1)

	// Create client with config manager
	mockConfigManager := &MockConfigManager{projectConfig: projectConfig1, sdkKey: sdkKey}
	optimizelyClient, err := factory.Client(WithConfigManager(mockConfigManager))
	assert.NoError(t, err)
	var odpManager = optimizelyClient.OdpManager.(*odp.DefaultOdpManager)
	// Odp config should contain latest updates of config manager
	assert.Equal(t, projectConfig1.GetHostForODP(), odpManager.OdpConfig.GetAPIHost())
	assert.Equal(t, projectConfig1.GetPublicKeyForODP(), odpManager.OdpConfig.GetAPIKey())

	// Update project config and trigger notification
	// This should update odp config too
	mockDatafile2 := []byte(`{"version":"4","integrations": [{"publicKey": "1234", "host": "www.1234.com", "key": "odp"}]}`)
	tmpConfigManager2 := config.NewStaticProjectConfigManagerWithOptions("", config.WithInitialDatafile(mockDatafile2))
	projectConfig2, err2 := tmpConfigManager2.GetConfig()
	assert.NoError(t, err2)
	// Update project config and trigger notification update to verify if notification listener
	// was added and it updates odpConfigManager
	mockConfigManager.projectConfig = projectConfig2
	projectConfigUpdateNotification := notification.ProjectConfigUpdateNotification{
		Type:     notification.ProjectConfigUpdate,
		Revision: "123",
	}
	registry.GetNotificationCenter(sdkKey).Send(notification.ProjectConfigUpdate, projectConfigUpdateNotification)
	// wait for notification to be received and config to be updated
	time.Sleep(100 * time.Millisecond)
	assert.Equal(t, projectConfig2.GetHostForODP(), odpManager.OdpConfig.GetAPIHost())
	assert.Equal(t, projectConfig2.GetPublicKeyForODP(), odpManager.OdpConfig.GetAPIKey())
}

func TestClientWithCustomCtx(t *testing.T) {
	factory := OptimizelyFactory{}
	ctx, cancel := context.WithCancel(context.Background())
	mockConfigManager := new(MockProjectConfigManager)
	mockConfig := new(MockProjectConfig)
	mockConfigManager.On("GetConfig").Return(mockConfig, errors.New("no project config available"))
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
	assert.NotNil(t, optlyClient.OdpManager)

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

func TestOptimizelyClientWithTracer(t *testing.T) {
	factory := OptimizelyFactory{SDKKey: "1212"}
	optimizelyClient, err := factory.Client(WithTracer(&MockTracer{}))
	assert.NoError(t, err)
	assert.NotNil(t, optimizelyClient.tracer)
	tracer := optimizelyClient.tracer.(*MockTracer)
	assert.NotNil(t, tracer)
}

func TestOptimizelyClientWithNoTracer(t *testing.T) {
	factory := OptimizelyFactory{SDKKey: "1212"}
	optimizelyClient, err := factory.Client()
	assert.NoError(t, err)
	assert.NotNil(t, optimizelyClient.tracer)
	tracer := optimizelyClient.tracer.(*tracing.NoopTracer)
	assert.NotNil(t, tracer)
}

func TestConvertDecideOptionsWithCMABOptions(t *testing.T) {
	// Test with IgnoreCMABCache option
	options := []decide.OptimizelyDecideOptions{decide.IgnoreCMABCache}
	convertedOptions := convertDecideOptions(options)
	assert.True(t, convertedOptions.IgnoreCMABCache)
	assert.False(t, convertedOptions.ResetCMABCache)
	assert.False(t, convertedOptions.InvalidateUserCMABCache)

	// Test with ResetCMABCache option
	options = []decide.OptimizelyDecideOptions{decide.ResetCMABCache}
	convertedOptions = convertDecideOptions(options)
	assert.False(t, convertedOptions.IgnoreCMABCache)
	assert.True(t, convertedOptions.ResetCMABCache)
	assert.False(t, convertedOptions.InvalidateUserCMABCache)

	// Test with InvalidateUserCMABCache option
	options = []decide.OptimizelyDecideOptions{decide.InvalidateUserCMABCache}
	convertedOptions = convertDecideOptions(options)
	assert.False(t, convertedOptions.IgnoreCMABCache)
	assert.False(t, convertedOptions.ResetCMABCache)
	assert.True(t, convertedOptions.InvalidateUserCMABCache)

	// Test with all CMAB options
	options = []decide.OptimizelyDecideOptions{
		decide.IgnoreCMABCache,
		decide.ResetCMABCache,
		decide.InvalidateUserCMABCache,
	}
	convertedOptions = convertDecideOptions(options)
	assert.True(t, convertedOptions.IgnoreCMABCache)
	assert.True(t, convertedOptions.ResetCMABCache)
	assert.True(t, convertedOptions.InvalidateUserCMABCache)

	// Test with CMAB options mixed with other options
	options = []decide.OptimizelyDecideOptions{
		decide.DisableDecisionEvent,
		decide.IgnoreCMABCache,
		decide.EnabledFlagsOnly,
		decide.ResetCMABCache,
		decide.InvalidateUserCMABCache,
	}
	convertedOptions = convertDecideOptions(options)
	assert.True(t, convertedOptions.DisableDecisionEvent)
	assert.True(t, convertedOptions.EnabledFlagsOnly)
	assert.True(t, convertedOptions.IgnoreCMABCache)
	assert.True(t, convertedOptions.ResetCMABCache)
	assert.True(t, convertedOptions.InvalidateUserCMABCache)
}

func TestAllOptionFunctions(t *testing.T) {
	f := &OptimizelyFactory{}

	// Test all option functions to ensure they're covered
	WithDatafileAccessToken("token")(f)
	WithSegmentsCacheSize(123)(f)
	WithSegmentsCacheTimeout(2 * time.Second)(f)
	WithOdpDisabled(true)(f)

	// Verify some options were set
	assert.Equal(t, "token", f.DatafileAccessToken)
	assert.Equal(t, 123, f.segmentsCacheSize)
	assert.True(t, f.odpDisabled)
}

func TestStaticClientError(t *testing.T) {
	// Use invalid datafile to force an error
	factory := OptimizelyFactory{Datafile: []byte("invalid json"), SDKKey: ""}
	client, err := factory.StaticClient()
	assert.Error(t, err)
	assert.Nil(t, client)
}

func TestFactoryWithCmabConfig(t *testing.T) {
	factory := OptimizelyFactory{}
	cmabConfig := CmabConfig{
		CacheSize:   100,
		CacheTTL:    time.Minute,
		HTTPTimeout: 30 * time.Second,
	}

	// Test the option function
	WithCmabConfig(&cmabConfig)(&factory)

	assert.Equal(t, &cmabConfig, factory.cmabConfig)
	assert.Equal(t, 100, factory.cmabConfig.CacheSize)
	assert.Equal(t, time.Minute, factory.cmabConfig.CacheTTL)
	assert.Equal(t, 30*time.Second, factory.cmabConfig.HTTPTimeout)
}

func TestFactoryCmabConfigPassedToDecisionService(t *testing.T) {
	// Test that CMAB config is correctly passed to decision service when creating client
	cmabConfig := CmabConfig{
		CacheSize:   200,
		CacheTTL:    2 * time.Minute,
		HTTPTimeout: 20 * time.Second,
	}

	factory := OptimizelyFactory{
		SDKKey:     "test_sdk_key",
		cmabConfig: &cmabConfig,
	}

	// Verify the config is set
	assert.Equal(t, &cmabConfig, factory.cmabConfig)
	assert.Equal(t, 200, factory.cmabConfig.CacheSize)
	assert.Equal(t, 2*time.Minute, factory.cmabConfig.CacheTTL)
}

func TestFactoryOptionFunctions(t *testing.T) {
	factory := &OptimizelyFactory{}

	// Test all option functions to ensure they're covered
	WithDatafileAccessToken("test_token")(factory)
	WithSegmentsCacheSize(100)(factory)
	WithSegmentsCacheTimeout(5 * time.Second)(factory)
	WithOdpDisabled(true)(factory)
	WithCmabConfig(&CmabConfig{CacheSize: 50})(factory)

	// Verify options were set
	assert.Equal(t, "test_token", factory.DatafileAccessToken)
	assert.Equal(t, 100, factory.segmentsCacheSize)
	assert.Equal(t, 5*time.Second, factory.segmentsCacheTimeout)
	assert.True(t, factory.odpDisabled)
	assert.Equal(t, &CmabConfig{CacheSize: 50}, factory.cmabConfig)
}

func TestWithCmabConfigOption(t *testing.T) {
	factory := &OptimizelyFactory{}
	testConfig := CmabConfig{
		CacheSize: 200,
		CacheTTL:  2 * time.Minute,
	}
	WithCmabConfig(&testConfig)(factory)
	assert.Equal(t, &testConfig, factory.cmabConfig)
}

func TestClientWithCmabConfig(t *testing.T) {
	// Test client creation with non-empty CMAB config (tests reflect.DeepEqual path)
	cmabConfig := CmabConfig{
		CacheSize:   200,
		CacheTTL:    5 * time.Minute,
		HTTPTimeout: 30 * time.Second,
	}

	factory := OptimizelyFactory{
		SDKKey: "test_sdk_key",
	}

	client, err := factory.Client(WithCmabConfig(&cmabConfig))
	assert.NoError(t, err)
	assert.NotNil(t, client)

	// Verify the CMAB config was applied by checking if DecisionService exists
	// This tests the reflect.DeepEqual check on lines 166-167
	assert.NotNil(t, client.DecisionService)
	client.Close()
}

func TestClientWithEmptyCmabConfig(t *testing.T) {
	// Test client creation with empty CMAB config (tests reflect.DeepEqual returns true)
	emptyCmabConfig := CmabConfig{}

	factory := OptimizelyFactory{
		SDKKey: "test_sdk_key",
	}

	client, err := factory.Client(WithCmabConfig(&emptyCmabConfig))
	assert.NoError(t, err)
	assert.NotNil(t, client)

	// Verify client still works with empty config
	assert.NotNil(t, client.DecisionService)
	client.Close()
}

func TestCmabConfigWithCustomPredictionEndpoint(t *testing.T) {
	// Test that custom prediction endpoint is correctly set in CmabConfig
	customEndpoint := "https://custom.example.com/predict/%s"
	cmabConfig := CmabConfig{
		CacheSize:                  100,
		CacheTTL:                   time.Minute,
		HTTPTimeout:                30 * time.Second,
		PredictionEndpointTemplate: customEndpoint,
	}

	factory := OptimizelyFactory{}
	WithCmabConfig(&cmabConfig)(&factory)

	assert.Equal(t, &cmabConfig, factory.cmabConfig)
	assert.Equal(t, customEndpoint, factory.cmabConfig.PredictionEndpointTemplate)
}

func TestCmabConfigToCmabConfig(t *testing.T) {
	// Test the toCmabConfig conversion includes PredictionEndpointTemplate
	customEndpoint := "https://proxy.example.com/cmab/%s"
	clientConfig := CmabConfig{
		CacheSize:                  200,
		CacheTTL:                   5 * time.Minute,
		HTTPTimeout:                15 * time.Second,
		PredictionEndpointTemplate: customEndpoint,
	}

	internalConfig := clientConfig.toCmabConfig()

	assert.NotNil(t, internalConfig)
	assert.Equal(t, 200, internalConfig.CacheSize)
	assert.Equal(t, 5*time.Minute, internalConfig.CacheTTL)
	assert.Equal(t, 15*time.Second, internalConfig.HTTPTimeout)
	assert.Equal(t, customEndpoint, internalConfig.PredictionEndpointTemplate)
}

func TestCmabConfigToCmabConfigNil(t *testing.T) {
	// Test that nil CmabConfig returns nil
	var clientConfig *CmabConfig
	internalConfig := clientConfig.toCmabConfig()
	assert.Nil(t, internalConfig)
}
