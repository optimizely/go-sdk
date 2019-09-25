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
	"github.com/optimizely/go-sdk/optimizely/notification"
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
	notificationCenter := notification.NewNotificationCenter()

	appClient := &OptimizelyClient{
		executionCtx:    executionCtx,
		decisionService: decision.NewCompositeService(notificationCenter),
		eventProcessor:  event.NewEventProcessor(executionCtx, event.DefaultBatchSize, event.DefaultEventQueueSize, event.DefaultEventFlushInterval),
	}

	for _, opt := range clientOptions {
		opt(appClient, executionCtx)
	}

	if f.SDKKey == "" && f.Datafile == nil && appClient.configManager == nil {
		return nil, errors.New("unable to instantiate client: no project config manager, SDK key, or a Datafile provided")
	}

	if appClient.configManager == nil { // if it was not passed then assign here

		appClient.configManager = config.NewPollingProjectConfigManager(executionCtx, f.SDKKey,
			config.SetInitialDatafile(f.Datafile), config.SetPollingInterval(config.DefaultPollingInterval), config.SetNotification(notificationCenter))
	}

	return appClient, nil
}

// PollingConfigManager sets polling config manager on a client
func PollingConfigManager(sdkKey string, pollingInterval time.Duration, initDataFile []byte, notificationCenter notification.Center) OptionFunc {
	return func(f *OptimizelyClient, executionCtx utils.ExecutionCtx) {
		f.configManager = config.NewPollingProjectConfigManager(f.executionCtx, sdkKey, config.SetInitialDatafile(initDataFile),
			config.SetPollingInterval(pollingInterval), config.SetNotification(notificationCenter))
	}
}

// ConfigManager sets polling config manager on a client
func ConfigManager(configManager optimizely.ProjectConfigManager) OptionFunc {
	return func(f *OptimizelyClient, executionCtx utils.ExecutionCtx) {
		f.configManager = configManager
	}
}

// CompositeDecisionService sets decision service on a client
func CompositeDecisionService(notificationCenter notification.Center) OptionFunc {
	return func(f *OptimizelyClient, executionCtx utils.ExecutionCtx) {
		f.decisionService = decision.NewCompositeService(notificationCenter)
	}
}

// DecisionService sets decision service on a client
func DecisionService(decisionService decision.Service) OptionFunc {
	return func(f *OptimizelyClient, executionCtx utils.ExecutionCtx) {
		f.decisionService = decisionService
	}
}

// BatchEventProcessor sets event processor on a client
func BatchEventProcessor(batchSize, queueSize int, flushInterval time.Duration) OptionFunc {
	return func(f *OptimizelyClient, executionCtx utils.ExecutionCtx) {
		f.eventProcessor = event.NewEventProcessor(executionCtx, batchSize, queueSize, flushInterval)
	}
}

// EventProcessor sets event processor on a client
func EventProcessor(eventProcessor event.Processor) OptionFunc {
	return func(f *OptimizelyClient, executionCtx utils.ExecutionCtx) {
		f.eventProcessor = eventProcessor
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

	notificationCenter := notification.NewNotificationCenter()

	optlyClient, e := f.Client(
		ConfigManager(configManager),
		CompositeDecisionService(notificationCenter),
		BatchEventProcessor(event.DefaultBatchSize, event.DefaultEventQueueSize, event.DefaultEventFlushInterval),
	)
	return optlyClient, e
}
