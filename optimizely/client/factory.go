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

package client

import (
	"context"
	"errors"
	"fmt"

	"github.com/optimizely/go-sdk/optimizely/event"

	"github.com/optimizely/go-sdk/optimizely/notification"

	"github.com/optimizely/go-sdk/optimizely"
	"github.com/optimizely/go-sdk/optimizely/config"
	"github.com/optimizely/go-sdk/optimizely/decision"
)

// Options are used to create an instance of the OptimizelyClient with custom configuration
type Options struct {
	Context              context.Context
	ProjectConfigManager optimizely.ProjectConfigManager
	DecisionService      decision.DecisionService
}

// OptimizelyFactory is used to construct an instance of the OptimizelyClient
type OptimizelyFactory struct {
	SDKKey   string
	Datafile []byte
}

const defaultEventQueueSize = 10
const defaultEventFlushInterval = 30000

// StaticClient returns a client initialized with a static project config
func (f OptimizelyFactory) StaticClient() (*OptimizelyClient, error) {
	var configManager optimizely.ProjectConfigManager

	if f.SDKKey != "" {
		url := fmt.Sprintf(config.DatafileURLTemplate, f.SDKKey)
		staticConfigManager, err := config.NewStaticProjectConfigManagerFromURL(url)

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
	client := &OptimizelyClient{
		isValid: false,
	}

	var ctx context.Context
	if clientOptions.Context != nil {
		ctx = clientOptions.Context
	} else {
		// if no context is provided, we create our own cancellable context and hand it over to the client so the client can shut down its child processes
		ctx = context.Background()
		var cancel context.CancelFunc
		ctx, cancel = context.WithCancel(ctx)
		client.cancelFunc = cancel
	}

	notificationCenter := notification.NewNotificationCenter()

	if clientOptions.ProjectConfigManager != nil {
		client.configManager = clientOptions.ProjectConfigManager
	} else if f.SDKKey != "" {
		options := config.PollingProjectConfigManagerOptions{
			Datafile: f.Datafile,
		}
		client.configManager = config.NewPollingProjectConfigManagerWithOptions(ctx, f.SDKKey, options)
	} else if f.Datafile != nil {
		staticConfigManager, _ := config.NewStaticProjectConfigManagerFromPayload(f.Datafile)
		client.configManager = staticConfigManager
	} else {
		return client, errors.New("unable to instantiate client: no project config manager, SDK key, or a Datafile provided")
	}

	if clientOptions.DecisionService != nil {
		client.decisionService = clientOptions.DecisionService
	} else {
		client.decisionService = decision.NewCompositeService(notificationCenter)
	}

	// @TODO: allow event processor to be passed in
	// @TODO: pass the context object to the event processor
	client.eventProcessor = event.NewEventProcessor(defaultEventQueueSize, defaultEventFlushInterval)
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
