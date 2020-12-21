/****************************************************************************
 * Copyright 2019-2020, Optimizely, Inc. and contributors                   *
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
	"fmt"

	"github.com/optimizely/go-sdk/pkg/config"
	"github.com/optimizely/go-sdk/pkg/decide"
	"github.com/optimizely/go-sdk/pkg/decision"
	"github.com/optimizely/go-sdk/pkg/entities"
	"github.com/optimizely/go-sdk/pkg/event"
	"github.com/optimizely/go-sdk/pkg/notification"
	"github.com/stretchr/testify/mock"
)

/**
 * This file provides mocks and other test fixtures to facilitate our test scenarios
 */

type MockProjectConfig struct {
	config.ProjectConfig
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
func (c *MockProjectConfig) SendFlagDecisions() bool {
	return false
}

type MockProjectConfigManager struct {
	projectConfig config.ProjectConfig
	mock.Mock
}

func (p *MockProjectConfigManager) GetConfig() (config.ProjectConfig, error) {
	if p.projectConfig != nil {
		return p.projectConfig, nil
	}

	args := p.Called()
	return args.Get(0).(config.ProjectConfig), args.Error(1)
}

func (p *MockProjectConfigManager) OnProjectConfigUpdate(callback func(notification.ProjectConfigUpdateNotification)) (int, error) {
	return 0, nil
}

func (p *MockProjectConfigManager) RemoveOnProjectConfigUpdate(id int) error {
	return nil
}

func (p *MockProjectConfigManager) GetOptimizelyConfig() *config.OptimizelyConfig {

	optimizelyConfig := &config.OptimizelyConfig{}
	optimizelyConfig.Revision = "232"
	return optimizelyConfig
}

type MockDecisionService struct {
	decision.Service
	notificationCenter notification.Center
	mock.Mock
}

func (m *MockDecisionService) GetFeatureDecision(decisionContext decision.FeatureDecisionContext, userContext entities.UserContext, options *decide.Options) (decision.FeatureDecision, decide.DecisionReasons, error) {
	args := m.Called(decisionContext, userContext, options)
	return args.Get(0).(decision.FeatureDecision), args.Get(1).(decide.DecisionReasons), args.Error(2)
}

func (m *MockDecisionService) GetExperimentDecision(decisionContext decision.ExperimentDecisionContext, userContext entities.UserContext, options *decide.Options) (decision.ExperimentDecision, decide.DecisionReasons, error) {
	args := m.Called(decisionContext, userContext, options)
	return args.Get(0).(decision.ExperimentDecision), args.Get(1).(decide.DecisionReasons), args.Error(2)
}

func (m *MockDecisionService) OnDecision(callback func(note notification.DecisionNotification)) (int, error) {
	m.Called(callback)
	handler := func(payload interface{}) {
		if decisionNotification, ok := payload.(notification.DecisionNotification); ok {
			callback(decisionNotification)
		}
	}
	id, err := m.notificationCenter.AddHandler(notification.Decision, handler)

	return id, err
}

type MockEventProcessor struct {
	event.Processor
	mock.Mock
}

func (m *MockEventProcessor) ProcessEvent(event event.UserEvent) bool {
	m.Called(event)
	return false
}

type PanickingConfigManager struct {
	config.ProjectConfigManager
}

func (m *PanickingConfigManager) GetConfig() (config.ProjectConfig, error) {
	panic("I'm panicking")
}

type PanickingDecisionService struct {
}

func (m *PanickingDecisionService) GetFeatureDecision(decisionContext decision.FeatureDecisionContext, userContext entities.UserContext, options *decide.Options) (decision.FeatureDecision, decide.DecisionReasons, error) {
	panic("I'm panicking")
}

func (m *PanickingDecisionService) GetExperimentDecision(decisionContext decision.ExperimentDecisionContext, userContext entities.UserContext, options *decide.Options) (decision.ExperimentDecision, decide.DecisionReasons, error) {
	panic("I'm panicking")
}

func (m *PanickingDecisionService) OnDecision(callback func(notification.DecisionNotification)) (int, error) {
	panic("I'm panicking")
}

func (m *PanickingDecisionService) RemoveOnDecision(id int) error {
	panic("I'm panicking")
}

type MockUserProfileService struct {
	decision.UserProfileService
	mock.Mock
}

func (m *MockUserProfileService) Lookup(userID string) decision.UserProfile {
	args := m.Called(userID)
	return args.Get(0).(decision.UserProfile)
}

func (m *MockUserProfileService) Save(userProfile decision.UserProfile) {
	m.Called(userProfile)
}

// Helper methods for creating test entities
func makeTestExperiment(experimentKey string) entities.Experiment {
	return entities.Experiment{
		Key: experimentKey,
		Variations: map[string]entities.Variation{
			"v1": entities.Variation{Key: "v1"},
			"v2": entities.Variation{Key: "v2"},
		},
	}
}

func makeTestVariation(variationKey string, featureEnabled bool) entities.Variation {
	return entities.Variation{
		ID:             fmt.Sprintf("test_variation_%s", variationKey),
		Key:            variationKey,
		FeatureEnabled: featureEnabled,
	}
}

func makeTestExperimentWithVariations(experimentKey string, variations []entities.Variation) entities.Experiment {
	variationsMap := make(map[string]entities.Variation)
	for _, variation := range variations {
		variationsMap[variation.ID] = variation
	}
	return entities.Experiment{
		Key:        experimentKey,
		ID:         fmt.Sprintf("test_experiment_%s", experimentKey),
		Variations: variationsMap,
	}
}

func makeTestFeatureWithExperiment(featureKey string, experiment entities.Experiment) entities.Feature {
	testFeature := entities.Feature{
		ID:                 fmt.Sprintf("test_feature_%s", featureKey),
		Key:                featureKey,
		FeatureExperiments: []entities.Experiment{experiment},
	}

	return testFeature
}
