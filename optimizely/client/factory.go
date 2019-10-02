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

	"github.com/optimizely/go-sdk/optimizely"
	"github.com/optimizely/go-sdk/optimizely/config"
	"github.com/optimizely/go-sdk/optimizely/decision"
	"github.com/optimizely/go-sdk/optimizely/event"
	"github.com/optimizely/go-sdk/optimizely/utils"
)

// OptimizelyFactory is used to construct an instance of the OptimizelyClient
type OptimizelyFactory struct {
	SDKKey   string
	Datafile []byte
}

// OptionFunc is a type to a proper func
type OptionFunc func(*OptimizelyClient, utils.ExecutionCtx)

// Client gets client and sets some parameters
func (f OptimizelyFactory) Client(clientOptions ...OptionFunc) (*OptimizelyClient, error) {

	executionCtx := utils.NewCancelableExecutionCtx()

	appClient := &OptimizelyClient{
		executionCtx:    executionCtx,
		DecisionService: decision.NewCompositeService(f.SDKKey),
		EventProcessor: event.NewEventProcessor(executionCtx, event.BatchSize(event.DefaultBatchSize),
			event.QueueSize(event.DefaultEventQueueSize), event.FlushInterval(event.DefaultEventFlushInterval)),
	}

	for _, opt := range clientOptions {
		opt(appClient, executionCtx)
	}

	if f.SDKKey == "" && f.Datafile == nil && appClient.ConfigManager == nil {
		return nil, errors.New("unable to instantiate client: no project config manager, SDK key, or a Datafile provided")
	}

	if appClient.ConfigManager == nil { // if it was not passed then assign here

		appClient.ConfigManager = config.NewPollingProjectConfigManager(executionCtx, f.SDKKey,
			config.InitialDatafile(f.Datafile), config.PollingInterval(config.DefaultPollingInterval))
	}

	return appClient, nil
}

// PollingConfigManager sets polling config manager on a client
func PollingConfigManager(sdkKey string, pollingInterval time.Duration, initDataFile []byte) OptionFunc {
	return func(f *OptimizelyClient, executionCtx utils.ExecutionCtx) {
		f.ConfigManager = config.NewPollingProjectConfigManager(f.executionCtx, sdkKey, config.InitialDatafile(initDataFile),
			config.PollingInterval(pollingInterval))
	}
}

// PollingConfigManagerRequester sets polling config manager on a client
func PollingConfigManagerRequester(requester utils.Requester, pollingInterval time.Duration, initDataFile []byte) OptionFunc {
	return func(f *OptimizelyClient, executionCtx utils.ExecutionCtx) {
		f.ConfigManager = config.NewPollingProjectConfigManager(f.executionCtx, "", config.InitialDatafile(initDataFile),
			config.PollingInterval(pollingInterval), config.Requester(requester))
	}
}

// ConfigManager sets polling config manager on a client
func ConfigManager(configManager optimizely.ProjectConfigManager) OptionFunc {
	return func(f *OptimizelyClient, executionCtx utils.ExecutionCtx) {
		f.ConfigManager = configManager
	}
}

// CompositeDecisionService sets decision service on a client
func CompositeDecisionService(sdkKey string) OptionFunc {
	return func(f *OptimizelyClient, executionCtx utils.ExecutionCtx) {
		f.DecisionService = decision.NewCompositeService(sdkKey)
	}
}

// DecisionService sets decision service on a client
func DecisionService(decisionService decision.Service) OptionFunc {
	return func(f *OptimizelyClient, executionCtx utils.ExecutionCtx) {
		f.DecisionService = decisionService
	}
}

// BatchEventProcessor sets event processor on a client
func BatchEventProcessor(batchSize, queueSize int, flushInterval time.Duration) OptionFunc {
	return func(f *OptimizelyClient, executionCtx utils.ExecutionCtx) {
		f.EventProcessor = event.NewEventProcessor(executionCtx, event.BatchSize(batchSize),
			event.QueueSize(queueSize), event.FlushInterval(flushInterval))
	}
}

// EventProcessor sets event processor on a client
func EventProcessor(eventProcessor event.Processor) OptionFunc {
	return func(f *OptimizelyClient, executionCtx utils.ExecutionCtx) {
		f.EventProcessor = eventProcessor
	}
}

// StaticClient returns a client initialized with a static project config
func (f OptimizelyFactory) StaticClient() (*OptimizelyClient, error) {
	var configManager optimizely.ProjectConfigManager

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
		ConfigManager(configManager),
		CompositeDecisionService(f.SDKKey),
		BatchEventProcessor(event.DefaultBatchSize, event.DefaultEventQueueSize, event.DefaultEventFlushInterval),
	)
	return optlyClient, e
}
