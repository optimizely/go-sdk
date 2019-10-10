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
}

// OptionFunc is a type to a proper func
type OptionFunc func(*OptimizelyClient)

// Client gets client and sets some parameters
func (f OptimizelyFactory) Client(clientOptions ...OptionFunc) (*OptimizelyClient, error) {

	executionCtx := utils.NewCancelableExecutionCtx()

	appClient := &OptimizelyClient{
		executionCtx:    executionCtx,
		DecisionService: decision.NewCompositeService(f.SDKKey),
		EventProcessor: event.NewEventProcessor(event.WithBatchSize(event.DefaultBatchSize),
			event.WithQueueSize(event.DefaultEventQueueSize), event.WithFlushInterval(event.DefaultEventFlushInterval),
			event.WithSDKKey(f.SDKKey)),
	}

	for _, opt := range clientOptions {
		opt(appClient)
	}

	if f.SDKKey == "" && f.Datafile == nil && appClient.ConfigManager == nil {
		return nil, errors.New("unable to instantiate client: no project config manager, SDK key, or a Datafile provided")
	}

	if appClient.ConfigManager == nil { // if it was not passed then assign here
		appClient.ConfigManager = config.NewPollingProjectConfigManager(f.SDKKey,
			config.InitialDatafile(f.Datafile), config.PollingInterval(config.DefaultPollingInterval))
	}

	// Initialize the default services with the execution context
	if pollingConfigManager, ok := appClient.ConfigManager.(*config.PollingProjectConfigManager); ok {
		pollingConfigManager.Start(appClient.executionCtx)
	}

	if batchProcessor, ok := appClient.EventProcessor.(*event.BatchEventProcessor); ok {
		batchProcessor.Start(appClient.executionCtx)
	}

	return appClient, nil
}

// WithPollingConfigManager sets polling config manager on a client
func WithPollingConfigManager(sdkKey string, pollingInterval time.Duration, initDataFile []byte) OptionFunc {
	return func(f *OptimizelyClient) {
		f.ConfigManager = config.NewPollingProjectConfigManager(sdkKey, config.InitialDatafile(initDataFile),
			config.PollingInterval(pollingInterval))
	}
}

// WithPollingConfigManagerRequester sets polling config manager on a client
func WithPollingConfigManagerRequester(requester utils.Requester, pollingInterval time.Duration, initDataFile []byte) OptionFunc {
	return func(f *OptimizelyClient) {
		f.ConfigManager = config.NewPollingProjectConfigManager("", config.InitialDatafile(initDataFile),
			config.PollingInterval(pollingInterval), config.Requester(requester))
	}
}

// WithConfigManager sets polling config manager on a client
func WithConfigManager(configManager pkg.ProjectConfigManager) OptionFunc {
	return func(f *OptimizelyClient) {
		f.ConfigManager = configManager
	}
}

// WithCompositeDecisionService sets decision service on a client
func WithCompositeDecisionService(sdkKey string) OptionFunc {
	return func(f *OptimizelyClient) {
		f.DecisionService = decision.NewCompositeService(sdkKey)
	}
}

// WithDecisionService sets decision service on a client
func WithDecisionService(decisionService decision.Service) OptionFunc {
	return func(f *OptimizelyClient) {
		f.DecisionService = decisionService
	}
}

// WithBatchEventProcessor sets event processor on a client
func WithBatchEventProcessor(batchSize, queueSize int, flushInterval time.Duration) OptionFunc {
	return func(f *OptimizelyClient) {
		f.EventProcessor = event.NewEventProcessor(event.WithBatchSize(batchSize),
			event.WithQueueSize(queueSize), event.WithFlushInterval(flushInterval))
	}
}

// WithEventProcessor sets event processor on a client
func WithEventProcessor(eventProcessor event.Processor) OptionFunc {
	return func(f *OptimizelyClient) {
		f.EventProcessor = eventProcessor
	}
}

// WithExecutionContext allows user to pass in their own execution context to override the default one in the client
func WithExecutionContext(executionContext utils.ExecutionCtx) OptionFunc {
	return func(f *OptimizelyClient) {
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
