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

// Package client //
package client

import (
	"github.com/optimizely/go-sdk/optimizely"
	"github.com/optimizely/go-sdk/optimizely/decision"
	"github.com/optimizely/go-sdk/optimizely/entities"
	"github.com/optimizely/go-sdk/optimizely/event"
	"github.com/optimizely/go-sdk/optimizely/notification"
	"github.com/stretchr/testify/mock"
)

/**
 * This file provides mocks and other test fixtures to facilitate our test scenarios
 */

type MockProjectConfig struct {
	optimizely.ProjectConfig
	mock.Mock
}

func (c *MockProjectConfig) GetEventByKey(string) (entities.Event, error) {
	args := c.Called()
	return args.Get(0).(entities.Event), args.Error(1)
}

func (c *MockProjectConfig) GetExperimentByKey(experimentKey string) (entities.Experiment, error) {
	args := c.Called(experimentKey)
	return args.Get(0).(entities.Experiment), args.Error(1)
}

func (c *MockProjectConfig) GetFeatureByKey(featureKey string) (entities.Feature, error) {
	args := c.Called(featureKey)
	return args.Get(0).(entities.Feature), args.Error(1)
}

func (c *MockProjectConfig) GetFeatureList() []entities.Feature {
	args := c.Called()
	return args.Get(0).([]entities.Feature)
}

func (c *MockProjectConfig) GetVariableByKey(featureKey string, variableKey string) (entities.Variable, error) {
	args := c.Called(featureKey, variableKey)
	return args.Get(0).(entities.Variable), args.Error(1)
}

func (c *MockProjectConfig) GetProjectID() string {
	return "15389410617"
}
func (c *MockProjectConfig) GetRevision() string {
	return "7"
}
func (c *MockProjectConfig) GetAccountID() string {
	return "8362480420"
}
func (c *MockProjectConfig) GetAnonymizeIP() bool {
	return true
}
func (c *MockProjectConfig) GetBotFiltering() bool {
	return false
}

type MockDecisionService struct {
	decision.Service
	mock.Mock
}

func (m *MockDecisionService) GetFeatureDecision(decisionContext decision.FeatureDecisionContext, userContext entities.UserContext) (decision.FeatureDecision, error) {
	args := m.Called(decisionContext, userContext)
	return args.Get(0).(decision.FeatureDecision), args.Error(1)
}

func (m *MockDecisionService) GetExperimentDecision(decisionContext decision.ExperimentDecisionContext, userContext entities.UserContext) (decision.ExperimentDecision, error) {
	args := m.Called(decisionContext, userContext)
	return args.Get(0).(decision.ExperimentDecision), args.Error(1)
}

type MockEventProcessor struct {
	event.Processor
	mock.Mock
}

func (m *MockEventProcessor) ProcessEvent(userEvent event.UserEvent) {
	m.Called(userEvent)
}

type PanickingConfigManager struct {
}

func (m *PanickingConfigManager) GetConfig() (optimizely.ProjectConfig, error) {
	panic("I'm panicking")
}

type PanickingDecisionService struct {
}

func (m *PanickingDecisionService) GetFeatureDecision(decisionContext decision.FeatureDecisionContext, userContext entities.UserContext) (decision.FeatureDecision, error) {
	panic("I'm panicking")
}

func (m *PanickingDecisionService) GetExperimentDecision(decisionContext decision.ExperimentDecisionContext, userContext entities.UserContext) (decision.ExperimentDecision, error) {
	panic("I'm panicking")
}

func (m *PanickingDecisionService) OnDecision(callback func(notification.DecisionNotification)) (int, error) {
	panic("I'm panicking")
}

func (m *PanickingDecisionService) RemoveOnDecision(id int) error {
	panic("I'm panicking")
}

// Helper methods for creating test entities
func getTestExperiment(experimentKey string) entities.Experiment {
	return entities.Experiment{
		Key: experimentKey,
		Variations: map[string]entities.Variation{
			"v1": entities.Variation{Key: "v1"},
			"v2": entities.Variation{Key: "v2"},
		},
	}
}
