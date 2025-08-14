/****************************************************************************
 * Copyright 2019-2025, Optimizely, Inc. and contributors                   *
 *                                                                          *
 * Licensed under the Apache License, Version 2.0 (the "License");          *
 * you may not use this file except in compliance with the License.         *
 * You may obtain a copy of the License at                                  *
 *                                                                          *
 *    https://www.apache.org/licenses/LICENSE-2.0                           *
 *                                                                          *
 * Unless required by applicable law or agreed to in writing, software      *
 * distributed under the License is distributed on an "AS IS" BASIS,        *
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. *
 * See the License for the specific language governing permissions and      *
 * limitations under the License.                                           *
 ***************************************************************************/

package evaluator

import (
	"testing"

	"github.com/optimizely/go-sdk/v2/pkg/decide"
	"github.com/optimizely/go-sdk/v2/pkg/entities"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

// MockTreeEvaluator is a mock implementation of TreeEvaluator
type MockTreeEvaluator struct {
	mock.Mock
}

func (m *MockTreeEvaluator) Evaluate(conditionTree *entities.TreeNode, condTreeParams *entities.TreeParameters, options *decide.Options) (bool, bool, decide.DecisionReasons) {
	args := m.Called(conditionTree, condTreeParams, options)
	return args.Bool(0), args.Bool(1), args.Get(2).(decide.DecisionReasons)
}

// MockProjectConfig is a mock implementation of ProjectConfig
type MockProjectConfig struct {
	mock.Mock
}

func (m *MockProjectConfig) GetProjectID() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockProjectConfig) GetRevision() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockProjectConfig) GetAccountID() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockProjectConfig) GetAnonymizeIP() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockProjectConfig) GetAttributeID(key string) string {
	args := m.Called(key)
	return args.String(0)
}

func (m *MockProjectConfig) GetAttributes() []entities.Attribute {
	args := m.Called()
	return args.Get(0).([]entities.Attribute)
}

func (m *MockProjectConfig) GetAttributeByKey(key string) (entities.Attribute, error) {
	args := m.Called(key)
	return args.Get(0).(entities.Attribute), args.Error(1)
}

func (m *MockProjectConfig) GetAttributeKeyByID(id string) (string, error) {
	args := m.Called(id)
	return args.String(0), args.Error(1)
}

func (m *MockProjectConfig) GetAudienceByID(id string) (entities.Audience, error) {
	args := m.Called(id)
	return args.Get(0).(entities.Audience), args.Error(1)
}

func (m *MockProjectConfig) GetEventByKey(key string) (entities.Event, error) {
	args := m.Called(key)
	return args.Get(0).(entities.Event), args.Error(1)
}

func (m *MockProjectConfig) GetEvents() []entities.Event {
	args := m.Called()
	return args.Get(0).([]entities.Event)
}

func (m *MockProjectConfig) GetFeatureByKey(featureKey string) (entities.Feature, error) {
	args := m.Called(featureKey)
	return args.Get(0).(entities.Feature), args.Error(1)
}

func (m *MockProjectConfig) GetExperimentByKey(experimentKey string) (entities.Experiment, error) {
	args := m.Called(experimentKey)
	return args.Get(0).(entities.Experiment), args.Error(1)
}

func (m *MockProjectConfig) GetExperimentByID(id string) (entities.Experiment, error) {
	args := m.Called(id)
	return args.Get(0).(entities.Experiment), args.Error(1)
}

func (m *MockProjectConfig) GetExperimentList() []entities.Experiment {
	args := m.Called()
	return args.Get(0).([]entities.Experiment)
}

func (m *MockProjectConfig) GetPublicKeyForODP() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockProjectConfig) GetHostForODP() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockProjectConfig) GetSegmentList() []string {
	args := m.Called()
	return args.Get(0).([]string)
}

func (m *MockProjectConfig) GetBotFiltering() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockProjectConfig) GetSdkKey() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockProjectConfig) GetEnvironmentKey() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockProjectConfig) GetVariableByKey(featureKey, variableKey string) (entities.Variable, error) {
	args := m.Called(featureKey, variableKey)
	return args.Get(0).(entities.Variable), args.Error(1)
}

func (m *MockProjectConfig) GetFeatureList() []entities.Feature {
	args := m.Called()
	return args.Get(0).([]entities.Feature)
}

func (m *MockProjectConfig) GetIntegrationList() []entities.Integration {
	args := m.Called()
	return args.Get(0).([]entities.Integration)
}

func (m *MockProjectConfig) GetRolloutList() []entities.Rollout {
	args := m.Called()
	return args.Get(0).([]entities.Rollout)
}

func (m *MockProjectConfig) GetAudienceList() []entities.Audience {
	args := m.Called()
	return args.Get(0).([]entities.Audience)
}

func (m *MockProjectConfig) GetAudienceMap() map[string]entities.Audience {
	args := m.Called()
	return args.Get(0).(map[string]entities.Audience)
}

func (m *MockProjectConfig) GetGroupByID(groupID string) (entities.Group, error) {
	args := m.Called(groupID)
	return args.Get(0).(entities.Group), args.Error(1)
}

func (m *MockProjectConfig) SendFlagDecisions() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockProjectConfig) GetFlagVariationsMap() map[string][]entities.Variation {
	args := m.Called()
	return args.Get(0).(map[string][]entities.Variation)
}

func (m *MockProjectConfig) GetDatafile() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockProjectConfig) GetRegion() string {
	args := m.Called()
	return args.String(0)
}

// MockLogger is a mock implementation of OptimizelyLogProducer
// (This declaration has been removed to resolve the redeclaration error)

type AudienceEvaluatorTestSuite struct {
	suite.Suite
	mockLogger        *MockLogger
	mockTreeEvaluator *MockTreeEvaluator
	mockProjectConfig *MockProjectConfig
	options           decide.Options
	userContext       entities.UserContext
}

func (s *AudienceEvaluatorTestSuite) SetupTest() {
	s.mockLogger = new(MockLogger)
	s.mockTreeEvaluator = new(MockTreeEvaluator)
	s.mockProjectConfig = new(MockProjectConfig)
	s.options = decide.Options{IncludeReasons: true}
	s.userContext = entities.UserContext{
		ID: "test_user",
		Attributes: map[string]interface{}{
			"age":     25,
			"country": "US",
		},
	}
}

func (s *AudienceEvaluatorTestSuite) TestCheckIfUserInAudienceWithNoAudienceConditionTree() {
	experiment := &entities.Experiment{
		ID:                    "exp1",
		Key:                   "test_experiment",
		AudienceConditionTree: nil,
	}

	s.mockLogger.On("Debug", mock.AnythingOfType("string")).Return()

	result, reasons := CheckIfUserInAudience(experiment, s.userContext, s.mockProjectConfig, s.mockTreeEvaluator, &s.options, s.mockLogger)

	s.True(result)
	s.NotNil(reasons)
	messages := reasons.ToReport()
	s.Len(messages, 1)
	s.Contains(messages[0], "Audiences for experiment test_experiment collectively evaluated to true.")

	s.mockLogger.AssertExpectations(s.T())
}

func (s *AudienceEvaluatorTestSuite) TestCheckIfUserInAudienceWithAudienceConditionTreeUserMatches() {
	experiment := &entities.Experiment{
		ID:  "exp1",
		Key: "test_experiment",
		AudienceConditionTree: &entities.TreeNode{
			Operator: "and",
			Nodes: []*entities.TreeNode{
				{Item: "audience1"},
			},
		},
	}

	audienceMap := map[string]entities.Audience{
		"audience1": {
			ID:   "audience1",
			Name: "Test Audience",
		},
	}

	audienceReasons := decide.NewDecisionReasons(&s.options)
	audienceReasons.AddInfo("User matches audience conditions")

	s.mockProjectConfig.On("GetAudienceMap").Return(audienceMap)
	s.mockLogger.On("Debug", mock.AnythingOfType("string")).Return()
	s.mockTreeEvaluator.On("Evaluate",
		experiment.AudienceConditionTree,
		mock.AnythingOfType("*entities.TreeParameters"),
		&s.options,
	).Return(true, true, audienceReasons)

	result, reasons := CheckIfUserInAudience(experiment, s.userContext, s.mockProjectConfig, s.mockTreeEvaluator, &s.options, s.mockLogger)

	s.True(result)
	s.NotNil(reasons)
	messages := reasons.ToReport()
	s.GreaterOrEqual(len(messages), 1)
	s.Contains(messages[len(messages)-1], "Audiences for experiment test_experiment collectively evaluated to true.")

	s.mockProjectConfig.AssertExpectations(s.T())
	s.mockLogger.AssertExpectations(s.T())
	s.mockTreeEvaluator.AssertExpectations(s.T())
}

func (s *AudienceEvaluatorTestSuite) TestCheckIfUserInAudienceWithAudienceConditionTreeUserDoesNotMatch() {
	experiment := &entities.Experiment{
		ID:  "exp1",
		Key: "test_experiment",
		AudienceConditionTree: &entities.TreeNode{
			Operator: "and",
			Nodes: []*entities.TreeNode{
				{Item: "audience1"},
			},
		},
	}

	audienceMap := map[string]entities.Audience{
		"audience1": {
			ID:   "audience1",
			Name: "Test Audience",
		},
	}

	audienceReasons := decide.NewDecisionReasons(&s.options)
	audienceReasons.AddInfo("User does not match audience conditions")

	s.mockProjectConfig.On("GetAudienceMap").Return(audienceMap)
	s.mockLogger.On("Debug", mock.AnythingOfType("string")).Return()
	s.mockTreeEvaluator.On("Evaluate",
		experiment.AudienceConditionTree,
		mock.AnythingOfType("*entities.TreeParameters"),
		&s.options,
	).Return(false, false, audienceReasons)

	result, reasons := CheckIfUserInAudience(experiment, s.userContext, s.mockProjectConfig, s.mockTreeEvaluator, &s.options, s.mockLogger)

	s.False(result)
	s.NotNil(reasons)
	messages := reasons.ToReport()
	s.GreaterOrEqual(len(messages), 1)
	s.Contains(messages[len(messages)-1], "Audiences for experiment test_experiment collectively evaluated to false.")

	s.mockProjectConfig.AssertExpectations(s.T())
	s.mockLogger.AssertExpectations(s.T())
	s.mockTreeEvaluator.AssertExpectations(s.T())
}

func (s *AudienceEvaluatorTestSuite) TestCheckIfUserInAudienceWithNilOptions() {
	experiment := &entities.Experiment{
		ID:                    "exp1",
		Key:                   "test_experiment",
		AudienceConditionTree: nil,
	}

	s.mockLogger.On("Debug", mock.AnythingOfType("string")).Return()

	result, reasons := CheckIfUserInAudience(experiment, s.userContext, s.mockProjectConfig, s.mockTreeEvaluator, nil, s.mockLogger)

	s.True(result)
	s.NotNil(reasons)

	s.mockLogger.AssertExpectations(s.T())
}

func (s *AudienceEvaluatorTestSuite) TestCheckIfUserInAudienceWithIncludeReasonsFalse() {
	experiment := &entities.Experiment{
		ID:                    "exp1",
		Key:                   "test_experiment",
		AudienceConditionTree: nil,
	}

	optionsWithoutReasons := decide.Options{IncludeReasons: false}

	s.mockLogger.On("Debug", mock.AnythingOfType("string")).Return()

	result, reasons := CheckIfUserInAudience(experiment, s.userContext, s.mockProjectConfig, s.mockTreeEvaluator, &optionsWithoutReasons, s.mockLogger)

	s.True(result)
	s.NotNil(reasons)
	messages := reasons.ToReport()
	s.Equal(0, len(messages))

	s.mockLogger.AssertExpectations(s.T())
}

func (s *AudienceEvaluatorTestSuite) TestCheckIfUserInAudienceWithComplexAudienceTree() {
	experiment := &entities.Experiment{
		ID:  "exp1",
		Key: "complex_experiment",
		AudienceConditionTree: &entities.TreeNode{
			Operator: "or",
			Nodes: []*entities.TreeNode{
				{
					Operator: "and",
					Nodes: []*entities.TreeNode{
						{Item: "audience1"},
						{Item: "audience2"},
					},
				},
				{Item: "audience3"},
			},
		},
	}

	audienceMap := map[string]entities.Audience{
		"audience1": {ID: "audience1", Name: "Age Audience"},
		"audience2": {ID: "audience2", Name: "Country Audience"},
		"audience3": {ID: "audience3", Name: "Premium Audience"},
	}

	audienceReasons := decide.NewDecisionReasons(&s.options)
	audienceReasons.AddInfo("Complex audience evaluation completed")

	s.mockProjectConfig.On("GetAudienceMap").Return(audienceMap)
	s.mockLogger.On("Debug", mock.AnythingOfType("string")).Return()
	s.mockTreeEvaluator.On("Evaluate",
		experiment.AudienceConditionTree,
		mock.AnythingOfType("*entities.TreeParameters"),
		&s.options,
	).Return(true, true, audienceReasons)

	result, reasons := CheckIfUserInAudience(experiment, s.userContext, s.mockProjectConfig, s.mockTreeEvaluator, &s.options, s.mockLogger)

	s.True(result)
	s.NotNil(reasons)
	messages := reasons.ToReport()
	s.GreaterOrEqual(len(messages), 1)

	s.mockProjectConfig.AssertExpectations(s.T())
	s.mockLogger.AssertExpectations(s.T())
	s.mockTreeEvaluator.AssertExpectations(s.T())
}

func (s *AudienceEvaluatorTestSuite) TestCheckIfUserInAudienceWithNilExperiment() {
	s.mockLogger.On("Debug", "Experiment is nil, defaulting to false").Return()

	s.NotPanics(func() {
		CheckIfUserInAudience(nil, s.userContext, s.mockProjectConfig, s.mockTreeEvaluator, &s.options, s.mockLogger)
	})
	// Assert expectations AFTER the function call
	s.mockLogger.AssertExpectations(s.T())
}

func (s *AudienceEvaluatorTestSuite) TestCheckIfUserInAudienceWithEmptyUserContext() {
	experiment := &entities.Experiment{
		ID:                    "exp1",
		Key:                   "test_experiment",
		AudienceConditionTree: nil,
	}

	emptyUserContext := entities.UserContext{}

	s.mockLogger.On("Debug", mock.AnythingOfType("string")).Return()

	result, reasons := CheckIfUserInAudience(experiment, emptyUserContext, s.mockProjectConfig, s.mockTreeEvaluator, &s.options, s.mockLogger)

	s.True(result)
	s.NotNil(reasons)

	s.mockLogger.AssertExpectations(s.T())
}

func (s *AudienceEvaluatorTestSuite) TestCheckIfUserInAudienceTreeParametersCreation() {
	experiment := &entities.Experiment{
		ID:  "exp1",
		Key: "test_experiment",
		AudienceConditionTree: &entities.TreeNode{
			Operator: "and",
			Nodes: []*entities.TreeNode{
				{Item: "audience1"},
			},
		},
	}

	audienceMap := map[string]entities.Audience{
		"audience1": {
			ID:   "audience1",
			Name: "Test Audience",
		},
	}

	audienceReasons := decide.NewDecisionReasons(&s.options)

	s.mockProjectConfig.On("GetAudienceMap").Return(audienceMap)
	s.mockLogger.On("Debug", mock.AnythingOfType("string")).Return()

	s.mockTreeEvaluator.On("Evaluate",
		experiment.AudienceConditionTree,
		mock.MatchedBy(func(params *entities.TreeParameters) bool {
			return params.User != nil && params.User.ID == s.userContext.ID
		}),
		&s.options,
	).Return(true, true, audienceReasons)

	result, _ := CheckIfUserInAudience(experiment, s.userContext, s.mockProjectConfig, s.mockTreeEvaluator, &s.options, s.mockLogger)

	s.True(result)

	s.mockProjectConfig.AssertExpectations(s.T())
	s.mockLogger.AssertExpectations(s.T())
	s.mockTreeEvaluator.AssertExpectations(s.T())
}

func TestAudienceEvaluatorTestSuite(t *testing.T) {
	suite.Run(t, new(AudienceEvaluatorTestSuite))
}
