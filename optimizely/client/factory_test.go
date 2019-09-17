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
	"errors"
	"testing"

	"github.com/optimizely/go-sdk/optimizely/config"
	"github.com/optimizely/go-sdk/optimizely/config/datafileprojectconfig"
	"github.com/optimizely/go-sdk/optimizely/event"
	"github.com/stretchr/testify/assert"
)

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

	clientOptions := Options{}

	client, err := factory.ClientWithOptions(clientOptions)
	assert.NoError(t, err)
	assert.NotNil(t, client.configManager)
	assert.NotNil(t, client.decisionService)
	assert.NotNil(t, client.eventProcessor)
}

func TestClientWithProjectConfigManagerInOptions(t *testing.T) {
	factory := OptimizelyFactory{}
	projectConfig := datafileprojectconfig.DatafileProjectConfig{}
	configManager := config.NewStaticProjectConfigManager(projectConfig)

	clientOptions := Options{ProjectConfigManager: configManager}

	client, err := factory.ClientWithOptions(clientOptions)
	assert.NoError(t, err)
	assert.NotNil(t, client.configManager)
	assert.NotNil(t, client.decisionService)
	assert.NotNil(t, client.eventProcessor)
}

func TestClientWithNoDecisionServiceAndEventProcessorInOptions(t *testing.T) {
	factory := OptimizelyFactory{}
	projectConfig := datafileprojectconfig.DatafileProjectConfig{}
	configManager := config.NewStaticProjectConfigManager(projectConfig)

	clientOptions := Options{ProjectConfigManager: configManager}

	client, err := factory.ClientWithOptions(clientOptions)
	assert.NoError(t, err)
	assert.NotNil(t, client.configManager)
	assert.NotNil(t, client.decisionService)
	assert.NotNil(t, client.eventProcessor)
}

func TestClientWithDecisionServiceAndEventProcessorInOptions(t *testing.T) {
	factory := OptimizelyFactory{}
	projectConfig := datafileprojectconfig.DatafileProjectConfig{}
	configManager := config.NewStaticProjectConfigManager(projectConfig)
	decisionService := new(MockDecisionService)
	processor := &event.QueueingEventProcessor{
		MaxQueueSize:    100,
		FlushInterval:   100,
		Q:               event.NewInMemoryQueue(100),
		EventDispatcher: &MockDispatcher{},
	}

	clientOptions := Options{
		ProjectConfigManager: configManager,
		DecisionService:      decisionService,
		EventProcessor:       processor,
	}

	client, err := factory.ClientWithOptions(clientOptions)
	assert.NoError(t, err)
	assert.Equal(t, decisionService, client.decisionService)
	assert.Equal(t, processor, client.eventProcessor)
}

func TestClientWithOptionsErrorCase(t *testing.T) {
	// Error when no config manager, sdk key, or datafile is provided
	factory := OptimizelyFactory{}
	clientOptions := Options{}

	_, err := factory.ClientWithOptions(clientOptions)
	expectedErr := errors.New("unable to instantiate client: no project config manager, SDK key, or a Datafile provided")
	if assert.Error(t, err) {
		assert.Equal(t, err, expectedErr)
	}
}
