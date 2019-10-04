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
	"github.com/optimizely/go-sdk/optimizely/utils"

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

	optimizelyClient, err := factory.Client()
	assert.NoError(t, err)
	assert.NotNil(t, optimizelyClient.ConfigManager)
	assert.NotNil(t, optimizelyClient.DecisionService)
	assert.NotNil(t, optimizelyClient.EventProcessor)
}

func TestClientWithProjectConfigManagerInOptions(t *testing.T) {
	factory := OptimizelyFactory{}
	projectConfig := datafileprojectconfig.DatafileProjectConfig{}
	configManager := config.NewStaticProjectConfigManager(projectConfig)

	optimizelyClient, err := factory.Client(WithConfigManager(configManager))
	assert.NoError(t, err)
	assert.NotNil(t, optimizelyClient.ConfigManager)
	assert.NotNil(t, optimizelyClient.DecisionService)
	assert.NotNil(t, optimizelyClient.EventProcessor)
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

	optimizelyClient, err := factory.Client(WithConfigManager(configManager), WithDecisionService(decisionService), WithEventProcessor(processor))
	assert.NoError(t, err)
	assert.Equal(t, decisionService, optimizelyClient.DecisionService)
	assert.Equal(t, processor, optimizelyClient.EventProcessor)
}

func TestClientWithCustomCtx(t *testing.T) {
	factory := OptimizelyFactory{}
	testExecutionCtx := utils.NewCancelableExecutionCtx()
	mockConfigManager := new(MockProjectConfigManager)
	client, err := factory.Client(
		WithConfigManager(mockConfigManager),
		WithExecutionContext(testExecutionCtx),
	)
	assert.NoError(t, err)
	assert.Equal(t, client.executionCtx, testExecutionCtx)
}

func TestStaticClient(t *testing.T) {
	factory := OptimizelyFactory{Datafile: []byte(`{"revision": "42"}`)}
	optlyClient, err := factory.StaticClient()
	assert.NoError(t, err)

	parsedConfig, _ := optlyClient.ConfigManager.GetConfig()
	assert.Equal(t, "42", parsedConfig.GetRevision())
}
