/****************************************************************************
 * Copyright 2019-2020,2022-2025 Optimizely, Inc. and contributors          *
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

// Package client has client facing factories
package client

import (
	"context"
	"errors"
	"time"

	"github.com/optimizely/go-sdk/v2/pkg/cmab"
	"github.com/optimizely/go-sdk/v2/pkg/config"
	"github.com/optimizely/go-sdk/v2/pkg/decide"
	"github.com/optimizely/go-sdk/v2/pkg/decision"
	"github.com/optimizely/go-sdk/v2/pkg/event"
	"github.com/optimizely/go-sdk/v2/pkg/logging"
	"github.com/optimizely/go-sdk/v2/pkg/metrics"
	"github.com/optimizely/go-sdk/v2/pkg/notification"
	"github.com/optimizely/go-sdk/v2/pkg/odp"
	pkgUtils "github.com/optimizely/go-sdk/v2/pkg/odp/utils"
	"github.com/optimizely/go-sdk/v2/pkg/registry"
	"github.com/optimizely/go-sdk/v2/pkg/tracing"
	"github.com/optimizely/go-sdk/v2/pkg/utils"
)

// OptimizelyFactory is used to customize and construct an instance of the OptimizelyClient.
type OptimizelyFactory struct {
	SDKKey              string
	Datafile            []byte
	DatafileAccessToken string

	configManager        config.ProjectConfigManager
	ctx                  context.Context
	decisionService      decision.Service
	defaultDecideOptions *decide.Options
	eventDispatcher      event.Dispatcher
	eventProcessor       event.Processor
	metricsRegistry      metrics.Registry
	tracer               tracing.Tracer
	overrideStore        decision.ExperimentOverrideStore
	userProfileService   decision.UserProfileService
	notificationCenter   notification.Center
	cmabService          cmab.Service

	// ODP
	segmentsCacheSize    int
	segmentsCacheTimeout time.Duration
	odpDisabled          bool
	odpManager           odp.Manager
}

// OptionFunc is used to provide custom client configuration to the OptimizelyFactory.
type OptionFunc func(*OptimizelyFactory)

// Client instantiates a new OptimizelyClient with the given options.
//
//nolint:gocyclo // exceeds gocyclo cyclomatic complexity
func (f *OptimizelyFactory) Client(clientOptions ...OptionFunc) (*OptimizelyClient, error) {
	// Default values for odp cache
	f.segmentsCacheSize = pkgUtils.DefaultSegmentsCacheSize
	f.segmentsCacheTimeout = pkgUtils.DefaultSegmentsCacheTimeout

	// extract options
	for _, opt := range clientOptions {
		opt(f)
	}

	if f.SDKKey == "" && f.Datafile == nil && f.configManager == nil {
		return nil, errors.New("unable to instantiate client: no project config manager, SDK key, or a Datafile provided")
	}

	var metricsRegistry metrics.Registry
	if f.metricsRegistry != nil {
		metricsRegistry = f.metricsRegistry
	} else {
		metricsRegistry = metrics.NewNoopRegistry()
	}

	var ctx context.Context
	if f.ctx != nil {
		ctx = f.ctx
	} else {
		ctx = context.Background()
	}

	var decideOptions *decide.Options
	if f.defaultDecideOptions != nil {
		decideOptions = f.defaultDecideOptions
	} else {
		decideOptions = &decide.Options{}
	}

	eg := utils.NewExecGroup(ctx, logging.GetLogger(f.SDKKey, "ExecGroup"))
	appClient := &OptimizelyClient{
		defaultDecideOptions: decideOptions,
		execGroup:            eg,
		logger:               logging.GetLogger(f.SDKKey, "OptimizelyClient"),
		ctx:                  ctx,
	}

	if f.notificationCenter != nil {
		appClient.notificationCenter = f.notificationCenter
	} else {
		appClient.notificationCenter = registry.GetNotificationCenter(f.SDKKey)
	}

	if f.tracer != nil {
		appClient.tracer = f.tracer
	} else {
		appClient.tracer = &tracing.NoopTracer{}
	}

	if f.configManager != nil {
		appClient.ConfigManager = f.configManager
	} else {
		appClient.ConfigManager = config.NewPollingProjectConfigManager(
			f.SDKKey,
			config.WithInitialDatafile(f.Datafile),
			config.WithDatafileAccessToken(f.DatafileAccessToken),
		)
	}

	if f.eventProcessor != nil {
		appClient.EventProcessor = f.eventProcessor
	} else {
		var eventProcessorOptions = []event.BPOptionConfig{
			event.WithSDKKey(f.SDKKey),
		}
		if f.eventDispatcher != nil {
			eventProcessorOptions = append(eventProcessorOptions, event.WithEventDispatcher(f.eventDispatcher))
		}
		eventProcessorOptions = append(eventProcessorOptions, event.WithEventDispatcherMetrics(metricsRegistry))
		appClient.EventProcessor = event.NewBatchEventProcessor(eventProcessorOptions...)
	}

	if f.userProfileService != nil {
		appClient.UserProfileService = f.userProfileService
	}

	if f.decisionService != nil {
		appClient.DecisionService = f.decisionService
	} else {
		var experimentServiceOptions []decision.CESOptionFunc
		if f.userProfileService != nil {
			experimentServiceOptions = append(experimentServiceOptions, decision.WithUserProfileService(f.userProfileService))
		}
		if f.overrideStore != nil {
			experimentServiceOptions = append(experimentServiceOptions, decision.WithOverrideStore(f.overrideStore))
		}
		compositeExperimentService := decision.NewCompositeExperimentService(f.SDKKey, experimentServiceOptions...)
		compositeService := decision.NewCompositeService(f.SDKKey, decision.WithCompositeExperimentService(compositeExperimentService))
		appClient.DecisionService = compositeService
	}

	// Initialize the default services with the execution context
	if pollingConfigManager, ok := appClient.ConfigManager.(*config.PollingProjectConfigManager); ok {
		eg.Go(pollingConfigManager.Start)
	}

	if batchProcessor, ok := appClient.EventProcessor.(*event.BatchEventProcessor); ok {
		eg.Go(batchProcessor.Start)
	}

	if f.cmabService != nil {
		appClient.cmabService = f.cmabService
	}

	// Initialize and Start odp manager if possible
	// Needed a separate functions for this to avoid cyclo-complexity warning
	f.initializeOdpManager(appClient)
	f.startOdpManager(eg, appClient)

	return appClient, nil
}

// WithDatafileAccessToken sets authenticated datafile token
func WithDatafileAccessToken(datafileAccessToken string) OptionFunc {
	return func(f *OptimizelyFactory) {
		f.DatafileAccessToken = datafileAccessToken
	}
}

// WithPollingConfigManager sets polling config manager on a client.
func WithPollingConfigManager(pollingInterval time.Duration, initDataFile []byte) OptionFunc {
	return func(f *OptimizelyFactory) {
		f.configManager = config.NewPollingProjectConfigManager(f.SDKKey, config.WithInitialDatafile(initDataFile),
			config.WithPollingInterval(pollingInterval))
	}
}

// WithPollingConfigManagerDatafileAccessToken sets polling config manager with auth datafile token on a client
func WithPollingConfigManagerDatafileAccessToken(pollingInterval time.Duration, initDataFile []byte, datafileAccessToken string) OptionFunc {
	return func(f *OptimizelyFactory) {
		f.configManager = config.NewPollingProjectConfigManager(f.SDKKey, config.WithInitialDatafile(initDataFile),
			config.WithPollingInterval(pollingInterval), config.WithDatafileAccessToken(datafileAccessToken))
	}
}

// WithSegmentsCacheSize sets SegmentsCacheSize for odp manager.
// Default value is 10000
func WithSegmentsCacheSize(segmentsCacheSize int) OptionFunc {
	return func(f *OptimizelyFactory) {
		f.segmentsCacheSize = segmentsCacheSize
	}
}

// WithSegmentsCacheTimeout sets SegmentsCacheTimeoutInSecs for odp manager.
// Default value is 600s
func WithSegmentsCacheTimeout(segmentsCacheTimeout time.Duration) OptionFunc {
	return func(f *OptimizelyFactory) {
		f.segmentsCacheTimeout = segmentsCacheTimeout
	}
}

// WithOdpDisabled disables odp for the client.
// Default value is false
func WithOdpDisabled(disable bool) OptionFunc {
	return func(f *OptimizelyFactory) {
		f.odpDisabled = disable
	}
}

// WithOdpManager sets odp manager on a client.
func WithOdpManager(odpManager odp.Manager) OptionFunc {
	return func(f *OptimizelyFactory) {
		f.odpManager = odpManager
	}
}

// WithConfigManager sets polling config manager on a client.
func WithConfigManager(configManager config.ProjectConfigManager) OptionFunc {
	return func(f *OptimizelyFactory) {
		f.configManager = configManager
	}
}

// WithDecisionService sets decision service on a client.
func WithDecisionService(decisionService decision.Service) OptionFunc {
	return func(f *OptimizelyFactory) {
		f.decisionService = decisionService
	}
}

// WithDefaultDecideOptions sets default decide options on a client.
func WithDefaultDecideOptions(decideOptions []decide.OptimizelyDecideOptions) OptionFunc {
	return func(f *OptimizelyFactory) {
		f.defaultDecideOptions = convertDecideOptions(decideOptions)
	}
}

// WithUserProfileService sets the user profile service on the decision service.
func WithUserProfileService(userProfileService decision.UserProfileService) OptionFunc {
	return func(f *OptimizelyFactory) {
		f.userProfileService = userProfileService
	}
}

// WithExperimentOverrides sets the experiment override store on the decision service.
func WithExperimentOverrides(overrideStore decision.ExperimentOverrideStore) OptionFunc {
	return func(f *OptimizelyFactory) {
		f.overrideStore = overrideStore
	}
}

// WithBatchEventProcessor sets event processor on a client.
func WithBatchEventProcessor(batchSize, queueSize int, flushInterval time.Duration) OptionFunc {
	return func(f *OptimizelyFactory) {
		f.eventProcessor = event.NewBatchEventProcessor(event.WithBatchSize(batchSize),
			event.WithQueueSize(queueSize), event.WithFlushInterval(flushInterval))
	}
}

// WithEventProcessor sets event processor on a client
func WithEventProcessor(eventProcessor event.Processor) OptionFunc {
	return func(f *OptimizelyFactory) {
		f.eventProcessor = eventProcessor
	}
}

// WithEventDispatcher sets event dispatcher on the factory.
func WithEventDispatcher(eventDispatcher event.Dispatcher) OptionFunc {
	return func(f *OptimizelyFactory) {
		f.eventDispatcher = eventDispatcher
	}
}

// WithContext allows user to pass in their own context to override the default one in the client.
func WithContext(ctx context.Context) OptionFunc {
	return func(f *OptimizelyFactory) {
		f.ctx = ctx
	}
}

// WithMetricsRegistry allows user to pass in their own implementation of a metrics collector
func WithMetricsRegistry(metricsRegistry metrics.Registry) OptionFunc {
	return func(f *OptimizelyFactory) {
		f.metricsRegistry = metricsRegistry
	}
}

// WithNotificationCenter allows user to pass in their own implementation of the notification.Center interface
func WithNotificationCenter(nc notification.Center) OptionFunc {
	return func(f *OptimizelyFactory) {
		f.notificationCenter = nc
	}
}

// WithTracer allows user to pass in their own implementation of the Tracer interface
func WithTracer(tracer tracing.Tracer) OptionFunc {
	return func(f *OptimizelyFactory) {
		f.tracer = tracer
	}
}

// WithCmabService sets the CMAB service on the client
func WithCmabService(cmabService cmab.Service) OptionFunc {
	return func(f *OptimizelyFactory) {
		f.cmabService = cmabService
	}
}

// StaticClient returns a client initialized with a static project config.
func (f *OptimizelyFactory) StaticClient() (optlyClient *OptimizelyClient, err error) {

	staticManager := config.NewStaticProjectConfigManagerWithOptions(f.SDKKey, config.WithInitialDatafile(f.Datafile), config.WithDatafileAccessToken(f.DatafileAccessToken))

	if staticManager == nil {
		return nil, errors.New("unable to initiate config manager")
	}

	optlyClient, err = f.Client(
		WithConfigManager(staticManager),
		WithBatchEventProcessor(event.DefaultBatchSize, event.DefaultEventQueueSize, event.DefaultEventFlushInterval),
	)

	return optlyClient, err
}

func (f *OptimizelyFactory) initializeOdpManager(appClient *OptimizelyClient) {
	appClient.OdpManager = f.odpManager
	projectConfig, err := appClient.ConfigManager.GetConfig()
	// For cases when project config is not fetched yet
	isProjectConfigAvailable := err == nil && projectConfig != nil

	// Create ODP Manager
	if appClient.OdpManager == nil {
		appClient.OdpManager = odp.NewOdpManager(f.SDKKey, f.odpDisabled, odp.WithSegmentsCacheSize(f.segmentsCacheSize), odp.WithSegmentsCacheTimeout(f.segmentsCacheTimeout))
	}

	// Update odp config with latest config
	if isProjectConfigAvailable {
		appClient.OdpManager.Update(projectConfig.GetPublicKeyForODP(), projectConfig.GetHostForODP(), projectConfig.GetSegmentList())
	}
}

func (f *OptimizelyFactory) startOdpManager(eg *utils.ExecGroup, appClient *OptimizelyClient) {
	// Only start service if odp is enabled
	if f.odpDisabled {
		return
	}

	odpManager, ok := appClient.OdpManager.(*odp.DefaultOdpManager)
	if !ok {
		return
	}

	// Start odp ticker
	eg.Go(func(ctx context.Context) {
		odpManager.EventManager.Start(ctx, odpManager.OdpConfig)
	})

	// Only check for changes if ConfigManager is non static
	if _, ok = appClient.ConfigManager.(*config.StaticProjectConfigManager); ok {
		return
	}

	// listen to ProjectConfigUpdateNotification to update odp config accordingly
	callback := func(notification notification.ProjectConfigUpdateNotification) {
		if conf, err := appClient.ConfigManager.GetConfig(); err == nil && conf != nil {
			// Update odp manager with new changes and start service if not already started
			apiKey := conf.GetPublicKeyForODP()
			apiHost := conf.GetHostForODP()
			segmentList := conf.GetSegmentList()
			eg.Go(func(ctx context.Context) {
				odpManager.Update(apiKey, apiHost, segmentList)
				odpManager.EventManager.Start(ctx, odpManager.OdpConfig)
			})
		}
	}
	// Add callback for config update
	_, _ = appClient.ConfigManager.OnProjectConfigUpdate(callback)
}

func convertDecideOptions(options []decide.OptimizelyDecideOptions) *decide.Options {
	finalOptions := decide.Options{}
	for _, option := range options {
		switch option {
		case decide.DisableDecisionEvent:
			finalOptions.DisableDecisionEvent = true
		case decide.EnabledFlagsOnly:
			finalOptions.EnabledFlagsOnly = true
		case decide.IgnoreUserProfileService:
			finalOptions.IgnoreUserProfileService = true
		case decide.IncludeReasons:
			finalOptions.IncludeReasons = true
		case decide.ExcludeVariables:
			finalOptions.ExcludeVariables = true
		case decide.IgnoreCMABCache:
			finalOptions.IgnoreCMABCache = true
		case decide.ResetCMABCache:
			finalOptions.ResetCMABCache = true
		case decide.InvalidateUserCMABCache:
			finalOptions.InvalidateUserCMABCache = true
		}
	}
	return &finalOptions
}
