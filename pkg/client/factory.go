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

// Package client has client facing factories
package client

import (
	"errors"
	"time"

	"github.com/optimizely/go-sdk/pkg"
	"github.com/optimizely/go-sdk/pkg/config"
	"github.com/optimizely/go-sdk/pkg/decision"
	"github.com/optimizely/go-sdk/pkg/event"
	"github.com/optimizely/go-sdk/pkg/utils"
)

// OptimizelyFactory is used to construct an instance of the OptimizelyClient
type OptimizelyFactory struct {
	SDKKey   string
	Datafile []byte

	configManager      pkg.ProjectConfigManager
	decisionService    decision.Service
	eventProcessor     event.Processor
	executionCtx       utils.ExecutionCtx
	userProfileService decision.UserProfileService
	overrideStore      decision.ExperimentOverrideStore
	onTrack            OnTrack
}

// OptionFunc is used to provide custom client configuration to the OptimizelyFactory
type OptionFunc func(*OptimizelyFactory)

// Client instantiates a new OptimizelyClient with the given options
func (f OptimizelyFactory) Client(clientOptions ...OptionFunc) (*OptimizelyClient, error) {
	// extract options
	for _, opt := range clientOptions {
		opt(&f)
	}

	if f.SDKKey == "" && f.Datafile == nil && f.configManager == nil {
		return nil, errors.New("unable to instantiate client: no project config manager, SDK key, or a Datafile provided")
	}

	var executionCtx utils.ExecutionCtx
	if f.executionCtx != nil {
		executionCtx = f.executionCtx
	} else {
		executionCtx = utils.NewCancelableExecutionCtx()
	}

	appClient := &OptimizelyClient{executionCtx: executionCtx}

	if f.configManager != nil {
		appClient.ConfigManager = f.configManager
	} else {
		appClient.ConfigManager = config.NewPollingProjectConfigManager(
			f.SDKKey,
			config.InitialDatafile(f.Datafile),
			config.PollingInterval(config.DefaultPollingInterval),
		)
	}

	if f.eventProcessor != nil {
		appClient.EventProcessor = f.eventProcessor
	} else {
		appClient.EventProcessor = event.NewBatchEventProcessor(
			event.WithBatchSize(event.DefaultBatchSize),
			event.WithQueueSize(event.DefaultEventQueueSize),
			event.WithFlushInterval(event.DefaultEventFlushInterval),
			event.WithSDKKey(f.SDKKey),
		)
	}

	if f.decisionService != nil {
		appClient.DecisionService = f.decisionService
	} else {
		experimentServiceOptions := []decision.CESOptionFunc{}
		if f.userProfileService != nil {
			experimentServiceOptions = append(experimentServiceOptions, decision.WithUserProfileService(f.userProfileService))
		}
		if f.overrideStore != nil {
			experimentServiceOptions = append(experimentServiceOptions, decision.WithOverrideStore(f.overrideStore))
		}
		compositeExperimentService := decision.NewCompositeExperimentService(experimentServiceOptions...)
		compositeService := decision.NewCompositeService(f.SDKKey, decision.WithCompositeExperimentService(compositeExperimentService))
		appClient.DecisionService = compositeService
	}

	// Initialize the default services with the execution context
	if pollingConfigManager, ok := appClient.ConfigManager.(*config.PollingProjectConfigManager); ok {
		pollingConfigManager.Start(f.SDKKey, appClient.executionCtx)
	}

	if batchProcessor, ok := appClient.EventProcessor.(*event.BatchEventProcessor); ok {
		batchProcessor.Start(appClient.executionCtx)
	}

	appClient.onTrack = f.onTrack

	return appClient, nil
}

// WithPollingConfigManager sets polling config manager on a client
func WithPollingConfigManager(sdkKey string, pollingInterval time.Duration, initDataFile []byte) OptionFunc {
	return func(f *OptimizelyFactory) {
		f.configManager = config.NewPollingProjectConfigManager(sdkKey, config.InitialDatafile(initDataFile),
			config.PollingInterval(pollingInterval))
	}
}

// WithPollingConfigManagerRequester sets polling config manager on a client
func WithPollingConfigManagerRequester(requester utils.Requester, pollingInterval time.Duration, initDataFile []byte) OptionFunc {
	return func(f *OptimizelyFactory) {
		f.configManager = config.NewPollingProjectConfigManager("", config.InitialDatafile(initDataFile),
			config.PollingInterval(pollingInterval), config.Requester(requester))
	}
}

// WithConfigManager sets polling config manager on a client
func WithConfigManager(configManager pkg.ProjectConfigManager) OptionFunc {
	return func(f *OptimizelyFactory) {
		f.configManager = configManager
	}
}

// WithOnTrack sets callback which is called when Track is called.
func WithOnTrack(onTrack OnTrack) OptionFunc {
	return func(f *OptimizelyFactory) {
		f.onTrack = onTrack
	}
}

// WithCompositeDecisionService sets decision service on a client
func WithCompositeDecisionService(sdkKey string) OptionFunc {
	return func(f *OptimizelyFactory) {
		f.decisionService = decision.NewCompositeService(sdkKey)
	}
}

// WithDecisionService sets decision service on a client
func WithDecisionService(decisionService decision.Service) OptionFunc {
	return func(f *OptimizelyFactory) {
		f.decisionService = decisionService
	}
}

// WithUserProfileService sets the user profile service on the decision service
func WithUserProfileService(userProfileService decision.UserProfileService) OptionFunc {
	return func(f *OptimizelyFactory) {
		f.userProfileService = userProfileService
	}
}

// WithExperimentOverrides sets the experiment override store on the decision service
func WithExperimentOverrides(overrideStore decision.ExperimentOverrideStore) OptionFunc {
	return func(f *OptimizelyFactory) {
		f.overrideStore = overrideStore
	}
}

// WithBatchEventProcessor sets event processor on a client
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

// WithExecutionContext allows user to pass in their own execution context to override the default one in the client
func WithExecutionContext(executionContext utils.ExecutionCtx) OptionFunc {
	return func(f *OptimizelyFactory) {
		f.executionCtx = executionContext
	}
}

// StaticClient returns a client initialized with a static project config
func (f OptimizelyFactory) StaticClient() (*OptimizelyClient, error) {
	var configManager pkg.ProjectConfigManager

	if f.SDKKey != "" {
		staticConfigManager, err := config.NewStaticProjectConfigManagerFromURL(f.SDKKey)

		if err != nil {
			return nil, err
		}

		configManager = staticConfigManager

	} else if f.Datafile != nil {
		staticConfigManager, err := config.NewStaticProjectConfigManagerFromPayload(f.Datafile)

		if err != nil {
			return nil, err
		}

		configManager = staticConfigManager
	}

	optlyClient, e := f.Client(
		WithConfigManager(configManager),
		WithCompositeDecisionService(f.SDKKey),
		WithBatchEventProcessor(event.DefaultBatchSize, event.DefaultEventQueueSize, event.DefaultEventFlushInterval),
	)

	return optlyClient, e
}
