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
	"github.com/optimizely/go-sdk/optimizely/utils"

	"github.com/optimizely/go-sdk/optimizely/event"

	"github.com/optimizely/go-sdk/optimizely/notification"

	"github.com/optimizely/go-sdk/optimizely"
	"github.com/optimizely/go-sdk/optimizely/config"
	"github.com/optimizely/go-sdk/optimizely/decision"
)

// Options are used to create an instance of the OptimizelyClient with custom configuration
type Options struct {
	ProjectConfigManager optimizely.ProjectConfigManager
	DecisionService      decision.Service
	EventProcessor       event.Processor
}

// OptimizelyFactory is used to construct an instance of the OptimizelyClient
type OptimizelyFactory struct {
	SDKKey   string
	Datafile []byte
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

	clientOptions := Options{
		ProjectConfigManager: configManager,
	}
	client, err := f.ClientWithOptions(clientOptions)
	return client, err
}

// ClientWithOptions returns a client initialized with the given configuration options
func (f OptimizelyFactory) ClientWithOptions(clientOptions Options) (*OptimizelyClient, error) {

	executionCtx := utils.NewCancelableExecutionCtx()
	client := &OptimizelyClient{
		isValid:      false,
		executionCtx: executionCtx,
	}

	notificationCenter := notification.NewNotificationCenter()

	switch {
	case clientOptions.ProjectConfigManager != nil:
		client.configManager = clientOptions.ProjectConfigManager
	case f.SDKKey != "":
		options := config.PollingProjectConfigManagerOptions{
			Datafile: f.Datafile,
		}
		client.configManager = config.NewPollingProjectConfigManagerWithOptions(executionCtx, f.SDKKey, options)
	case f.Datafile != nil:
		staticConfigManager, _ := config.NewStaticProjectConfigManagerFromPayload(f.Datafile)
		client.configManager = staticConfigManager
	default:
		return client, errors.New("unable to instantiate client: no project config manager, SDK key, or a Datafile provided")
	}

	if clientOptions.DecisionService != nil {
		client.decisionService = clientOptions.DecisionService
	} else {
		client.decisionService = decision.NewCompositeService(notificationCenter)
	}

	if clientOptions.EventProcessor != nil {
		client.eventProcessor = clientOptions.EventProcessor
	} else {
		client.eventProcessor = event.NewEventProcessor(executionCtx, event.DefaultBatchSize, event.DefaultEventQueueSize, event.DefaultEventFlushInterval)
	}

	client.isValid = true
	return client, nil
}

// Client returns a client initialized with the defaults
func (f OptimizelyFactory) Client() (*OptimizelyClient, error) {
	// Creates a default, canceleable context
	clientOptions := Options{}
	client, err := f.ClientWithOptions(clientOptions)
	return client, err
}
