/****************************************************************************
 * Copyright 2025, Optimizely, Inc. and contributors                        *
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

package decision

import (
	"testing"

	"github.com/optimizely/go-sdk/v2/pkg/decide"
	"github.com/optimizely/go-sdk/v2/pkg/decision/bucketer"
	"github.com/optimizely/go-sdk/v2/pkg/decision/evaluator"
	"github.com/optimizely/go-sdk/v2/pkg/decision/reasons"
	"github.com/optimizely/go-sdk/v2/pkg/entities"
	"github.com/optimizely/go-sdk/v2/pkg/logging"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

// Test holdout data
var testHoldoutVar1 = entities.Variation{ID: "holdout_var_1", Key: "holdout_variation_1"}
var testHoldoutVar2 = entities.Variation{ID: "holdout_var_2", Key: "holdout_variation_2"}

var testAudience7777 = entities.Audience{ID: "7777"}
var testAudience7778 = entities.Audience{ID: "7778"}

var testHoldout1 = entities.Holdout{
	ID:     "holdout_1",
	Key:    "test_holdout_1",
	Status: entities.HoldoutStatusRunning,
	AudienceConditionTree: &entities.TreeNode{
		Operator: "and",
		Nodes: []*entities.TreeNode{
			{Item: "7777"},
		},
	},
	Variations: map[string]entities.Variation{
		"holdout_var_1": testHoldoutVar1,
		"holdout_var_2": testHoldoutVar2,
	},
	TrafficAllocation: []entities.Range{
		{EntityID: "holdout_var_1", EndOfRange: 5000},
		{EntityID: "holdout_var_2", EndOfRange: 10000},
	},
}

var testHoldout2NoAudience = entities.Holdout{
	ID:                    "holdout_2",
	Key:                   "test_holdout_2_no_audience",
	Status:                entities.HoldoutStatusRunning,
	AudienceConditionTree: nil, // No audience targeting
	Variations: map[string]entities.Variation{
		"holdout_var_1": testHoldoutVar1,
	},
	TrafficAllocation: []entities.Range{
		{EntityID: "holdout_var_1", EndOfRange: 10000},
	},
}

var testHoldout3NotRunning = entities.Holdout{
	ID:     "holdout_3",
	Key:    "test_holdout_3_not_running",
	Status: entities.HoldoutStatus("Paused"),
}

type HoldoutServiceTestSuite struct {
	suite.Suite
	mockConfig                 *mockProjectConfig
	mockAudienceTreeEvaluator  *MockAudienceTreeEvaluator
	mockBucketer               *MockExperimentBucketer
	testFeatureDecisionContext FeatureDecisionContext
	testUserContext            entities.UserContext
	mockLogger                 *MockLogger
	options                    *decide.Options
	decisionReasons            decide.DecisionReasons
}

func (s *HoldoutServiceTestSuite) SetupTest() {
	s.mockConfig = new(mockProjectConfig)
	s.mockAudienceTreeEvaluator = new(MockAudienceTreeEvaluator)
	s.mockBucketer = new(MockExperimentBucketer)

	testAudienceMap := map[string]entities.Audience{
		"7777": testAudience7777,
		"7778": testAudience7778,
	}

	s.testUserContext = entities.UserContext{
		ID: "test_user_holdout",
	}

	testFeature := entities.Feature{
		ID:  "feature_1",
		Key: "test_feature_with_holdout",
	}

	s.testFeatureDecisionContext = FeatureDecisionContext{
		Feature:       &testFeature,
		ProjectConfig: s.mockConfig,
	}

	s.mockConfig.On("GetAudienceMap").Return(testAudienceMap)
	s.mockLogger = new(MockLogger)
	s.options = &decide.Options{}
	s.decisionReasons = decide.NewDecisionReasons(s.options)
}

func (s *HoldoutServiceTestSuite) TestGetDecisionWithNoHoldouts() {
	// Setup: No holdouts for the feature
	s.mockConfig.On("GetHoldoutsForFlag", "test_feature_with_holdout").Return([]entities.Holdout{})

	testHoldoutService := HoldoutService{
		audienceTreeEvaluator: s.mockAudienceTreeEvaluator,
		bucketer:              s.mockBucketer,
		logger:                s.mockLogger,
	}

	decision, _, err := testHoldoutService.GetDecision(s.testFeatureDecisionContext, s.testUserContext, s.options)

	s.NoError(err)
	s.Equal(FeatureDecision{}, decision)
}

func (s *HoldoutServiceTestSuite) TestGetDecisionWithHoldoutNotRunning() {
	// Setup: Holdout exists but is not running
	s.mockConfig.On("GetHoldoutsForFlag", "test_feature_with_holdout").Return([]entities.Holdout{testHoldout3NotRunning})
	s.mockLogger.On("Debug", mock.Anything).Return()
	s.mockLogger.On("Info", mock.MatchedBy(func(msg string) bool {
		return true // Accept any info log message
	})).Return()

	testHoldoutService := HoldoutService{
		audienceTreeEvaluator: s.mockAudienceTreeEvaluator,
		bucketer:              s.mockBucketer,
		logger:                s.mockLogger,
	}

	decision, _, err := testHoldoutService.GetDecision(s.testFeatureDecisionContext, s.testUserContext, s.options)

	s.NoError(err)
	s.Equal(FeatureDecision{}, decision)
}

func (s *HoldoutServiceTestSuite) TestGetDecisionUserNotInAudience() {
	// Setup: User doesn't meet audience conditions
	s.mockConfig.On("GetHoldoutsForFlag", "test_feature_with_holdout").Return([]entities.Holdout{testHoldout1})
	s.mockAudienceTreeEvaluator.On("Evaluate", testHoldout1.AudienceConditionTree, mock.Anything, s.options).Return(false, true, s.decisionReasons)
	s.mockLogger.On("Debug", mock.Anything).Return()
	s.mockLogger.On("Info", mock.Anything).Return()

	testHoldoutService := HoldoutService{
		audienceTreeEvaluator: s.mockAudienceTreeEvaluator,
		bucketer:              s.mockBucketer,
		logger:                s.mockLogger,
	}

	decision, _, err := testHoldoutService.GetDecision(s.testFeatureDecisionContext, s.testUserContext, s.options)

	s.NoError(err)
	s.Equal(FeatureDecision{}, decision)
	s.mockAudienceTreeEvaluator.AssertExpectations(s.T())
}

func (s *HoldoutServiceTestSuite) TestGetDecisionUserInAudienceButNotBucketed() {
	// Setup: User meets audience conditions but doesn't get bucketed into a variation
	s.mockConfig.On("GetHoldoutsForFlag", "test_feature_with_holdout").Return([]entities.Holdout{testHoldout1})
	s.mockAudienceTreeEvaluator.On("Evaluate", testHoldout1.AudienceConditionTree, mock.Anything, s.options).Return(true, true, s.decisionReasons)
	s.mockBucketer.On("Bucket", "test_user_holdout", mock.AnythingOfType("entities.Experiment"), entities.Group{}).Return(nil, reasons.Reason(""), nil)
	s.mockLogger.On("Debug", mock.Anything).Return()
	s.mockLogger.On("Info", mock.Anything).Return()

	testHoldoutService := HoldoutService{
		audienceTreeEvaluator: s.mockAudienceTreeEvaluator,
		bucketer:              s.mockBucketer,
		logger:                s.mockLogger,
	}

	decision, _, err := testHoldoutService.GetDecision(s.testFeatureDecisionContext, s.testUserContext, s.options)

	s.NoError(err)
	s.Equal(FeatureDecision{}, decision)
	s.mockAudienceTreeEvaluator.AssertExpectations(s.T())
	s.mockBucketer.AssertExpectations(s.T())
}

func (s *HoldoutServiceTestSuite) TestGetDecisionHappyPath() {
	// Setup: User meets audience conditions and gets bucketed into a variation
	s.mockConfig.On("GetHoldoutsForFlag", "test_feature_with_holdout").Return([]entities.Holdout{testHoldout1})
	s.mockAudienceTreeEvaluator.On("Evaluate", testHoldout1.AudienceConditionTree, mock.Anything, s.options).Return(true, true, s.decisionReasons)
	s.mockBucketer.On("Bucket", "test_user_holdout", mock.AnythingOfType("entities.Experiment"), entities.Group{}).Return(&testHoldoutVar1, reasons.Reason(""), nil)
	s.mockLogger.On("Debug", mock.Anything).Return()
	s.mockLogger.On("Info", mock.Anything).Return()

	testHoldoutService := HoldoutService{
		audienceTreeEvaluator: s.mockAudienceTreeEvaluator,
		bucketer:              s.mockBucketer,
		logger:                s.mockLogger,
	}

	decision, _, err := testHoldoutService.GetDecision(s.testFeatureDecisionContext, s.testUserContext, s.options)

	s.NoError(err)
	s.NotNil(decision.Variation)
	s.Equal(testHoldoutVar1.ID, decision.Variation.ID)
	s.Equal(Holdout, decision.Source)
	s.mockAudienceTreeEvaluator.AssertExpectations(s.T())
	s.mockBucketer.AssertExpectations(s.T())
}

func (s *HoldoutServiceTestSuite) TestGetDecisionNoAudienceTargeting() {
	// Setup: Holdout with no audience targeting (applies to everyone)
	s.mockConfig.On("GetHoldoutsForFlag", "test_feature_with_holdout").Return([]entities.Holdout{testHoldout2NoAudience})
	s.mockBucketer.On("Bucket", "test_user_holdout", mock.AnythingOfType("entities.Experiment"), entities.Group{}).Return(&testHoldoutVar1, reasons.Reason(""), nil)
	s.mockLogger.On("Debug", mock.Anything).Return()
	s.mockLogger.On("Info", mock.Anything).Return()

	testHoldoutService := HoldoutService{
		audienceTreeEvaluator: s.mockAudienceTreeEvaluator,
		bucketer:              s.mockBucketer,
		logger:                s.mockLogger,
	}

	decision, _, err := testHoldoutService.GetDecision(s.testFeatureDecisionContext, s.testUserContext, s.options)

	s.NoError(err)
	s.NotNil(decision.Variation)
	s.Equal(testHoldoutVar1.ID, decision.Variation.ID)
	s.Equal(Holdout, decision.Source)
	s.mockBucketer.AssertExpectations(s.T())
}

func (s *HoldoutServiceTestSuite) TestGetDecisionMultipleHoldoutsFirstMatches() {
	// Setup: Multiple holdouts, first one matches
	holdouts := []entities.Holdout{testHoldout1, testHoldout2NoAudience}
	s.mockConfig.On("GetHoldoutsForFlag", "test_feature_with_holdout").Return(holdouts)
	s.mockAudienceTreeEvaluator.On("Evaluate", testHoldout1.AudienceConditionTree, mock.Anything, s.options).Return(true, true, s.decisionReasons)
	s.mockBucketer.On("Bucket", "test_user_holdout", mock.AnythingOfType("entities.Experiment"), entities.Group{}).Return(&testHoldoutVar1, reasons.Reason(""), nil)
	s.mockLogger.On("Debug", mock.Anything).Return()
	s.mockLogger.On("Info", mock.Anything).Return()

	testHoldoutService := HoldoutService{
		audienceTreeEvaluator: s.mockAudienceTreeEvaluator,
		bucketer:              s.mockBucketer,
		logger:                s.mockLogger,
	}

	decision, _, err := testHoldoutService.GetDecision(s.testFeatureDecisionContext, s.testUserContext, s.options)

	s.NoError(err)
	s.NotNil(decision.Variation)
	s.Equal(testHoldout1.ID, decision.Experiment.ID)
	s.Equal(Holdout, decision.Source)
	s.mockAudienceTreeEvaluator.AssertExpectations(s.T())
	s.mockBucketer.AssertExpectations(s.T())
}

func (s *HoldoutServiceTestSuite) TestGetDecisionMultipleHoldoutsSecondMatches() {
	// Setup: Multiple holdouts, first doesn't match, second does
	holdouts := []entities.Holdout{testHoldout1, testHoldout2NoAudience}
	s.mockConfig.On("GetHoldoutsForFlag", "test_feature_with_holdout").Return(holdouts)
	// First holdout: user not in audience
	s.mockAudienceTreeEvaluator.On("Evaluate", testHoldout1.AudienceConditionTree, mock.Anything, s.options).Return(false, true, s.decisionReasons)
	// Second holdout: no audience, user gets bucketed
	s.mockBucketer.On("Bucket", "test_user_holdout", mock.AnythingOfType("entities.Experiment"), entities.Group{}).Return(&testHoldoutVar1, reasons.Reason(""), nil)
	s.mockLogger.On("Debug", mock.Anything).Return()
	s.mockLogger.On("Info", mock.Anything).Return()

	testHoldoutService := HoldoutService{
		audienceTreeEvaluator: s.mockAudienceTreeEvaluator,
		bucketer:              s.mockBucketer,
		logger:                s.mockLogger,
	}

	decision, _, err := testHoldoutService.GetDecision(s.testFeatureDecisionContext, s.testUserContext, s.options)

	s.NoError(err)
	s.NotNil(decision.Variation)
	s.Equal(testHoldout2NoAudience.ID, decision.Experiment.ID)
	s.Equal(Holdout, decision.Source)
	s.mockAudienceTreeEvaluator.AssertExpectations(s.T())
	s.mockBucketer.AssertExpectations(s.T())
}

func (s *HoldoutServiceTestSuite) TestCheckIfUserInHoldoutAudienceNilHoldout() {
	testHoldoutService := HoldoutService{
		audienceTreeEvaluator: s.mockAudienceTreeEvaluator,
		logger:                s.mockLogger,
	}
	s.mockLogger.On("Debug", mock.Anything).Return()

	result := testHoldoutService.checkIfUserInHoldoutAudience(nil, s.testUserContext, s.mockConfig, s.options)

	s.False(result.result)
}

func (s *HoldoutServiceTestSuite) TestCheckIfUserInHoldoutAudienceNoConditionTree() {
	holdout := testHoldout2NoAudience
	testHoldoutService := HoldoutService{
		audienceTreeEvaluator: s.mockAudienceTreeEvaluator,
		logger:                s.mockLogger,
	}
	s.mockLogger.On("Debug", mock.Anything).Return()

	result := testHoldoutService.checkIfUserInHoldoutAudience(&holdout, s.testUserContext, s.mockConfig, s.options)

	s.True(result.result)
}

func (s *HoldoutServiceTestSuite) TestCheckIfUserInHoldoutAudienceWithConditionTree() {
	holdout := testHoldout1
	testHoldoutService := HoldoutService{
		audienceTreeEvaluator: s.mockAudienceTreeEvaluator,
		logger:                s.mockLogger,
	}
	s.mockAudienceTreeEvaluator.On("Evaluate", holdout.AudienceConditionTree, mock.Anything, s.options).Return(true, true, s.decisionReasons)
	s.mockLogger.On("Debug", mock.Anything).Return()

	result := testHoldoutService.checkIfUserInHoldoutAudience(&holdout, s.testUserContext, s.mockConfig, s.options)

	s.True(result.result)
	s.mockAudienceTreeEvaluator.AssertExpectations(s.T())
}

func TestHoldoutServiceTestSuite(t *testing.T) {
	suite.Run(t, new(HoldoutServiceTestSuite))
}

func TestNewHoldoutService(t *testing.T) {
	sdkKey := "test_sdk_key"
	service := NewHoldoutService(sdkKey)

	assert.NotNil(t, service)
	assert.NotNil(t, service.audienceTreeEvaluator)
	assert.NotNil(t, service.bucketer)
	assert.NotNil(t, service.logger)
}

// Integration test with real bucketer and evaluator
func TestHoldoutServiceIntegration(t *testing.T) {
	logger := logging.GetLogger("", "HoldoutService")

	// Create real dependencies
	audienceEvaluator := evaluator.NewMixedTreeEvaluator(logger)
	bucketer := bucketer.NewMurmurhashExperimentBucketer(logger, bucketer.DefaultHashSeed)

	service := &HoldoutService{
		audienceTreeEvaluator: audienceEvaluator,
		bucketer:              *bucketer,
		logger:                logger,
	}

	// Create a simple holdout with no audience targeting
	holdoutVar := entities.Variation{ID: "var_1", Key: "variation_1"}
	holdout := entities.Holdout{
		ID:                    "holdout_integration",
		Key:                   "test_holdout_integration",
		Status:                entities.HoldoutStatusRunning,
		AudienceConditionTree: nil,
		Variations: map[string]entities.Variation{
			"var_1": holdoutVar,
		},
		TrafficAllocation: []entities.Range{
			{EntityID: "var_1", EndOfRange: 10000}, // 100% traffic
		},
	}

	// Create mock config
	mockConfig := new(mockProjectConfig)
	mockConfig.On("GetHoldoutsForFlag", "test_feature").Return([]entities.Holdout{holdout})
	mockConfig.On("GetAudienceMap").Return(map[string]entities.Audience{})

	feature := entities.Feature{
		ID:  "feature_integration",
		Key: "test_feature",
	}

	decisionContext := FeatureDecisionContext{
		Feature:       &feature,
		ProjectConfig: mockConfig,
	}

	userContext := entities.UserContext{
		ID: "user_123",
	}

	options := &decide.Options{}

	// Execute decision
	decision, _, err := service.GetDecision(decisionContext, userContext, options)

	// Verify
	assert.NoError(t, err)
	assert.NotNil(t, decision.Variation)
	assert.Equal(t, holdoutVar.ID, decision.Variation.ID)
	assert.Equal(t, Holdout, decision.Source)
}

// Level 2 — GetDecisionForRule (local holdout decision service) tests (FSSDK-12369)

// TestGetDecisionForRuleNoLocalHoldouts verifies that when there are no local holdouts for a rule,
// the function returns an empty decision (the rule is evaluated normally).
func TestGetDecisionForRuleNoLocalHoldouts(t *testing.T) {
	mockConfig := new(mockProjectConfig)
	mockBucketer := new(MockExperimentBucketer)
	mockAudienceEval := new(MockAudienceTreeEvaluator)
	mockLogger := new(MockLogger)

	mockConfig.On("GetHoldoutsForRule", "rule_id_1").Return([]entities.Holdout{})

	service := HoldoutService{
		audienceTreeEvaluator: mockAudienceEval,
		bucketer:              mockBucketer,
		logger:                mockLogger,
	}

	userCtx := entities.UserContext{ID: "test_user"}
	options := &decide.Options{}

	decision, _, err := service.GetDecisionForRule("rule_id_1", mockConfig, userCtx, options)

	assert.NoError(t, err)
	assert.Nil(t, decision.Variation, "No local holdouts means no holdout decision")
	mockConfig.AssertExpectations(t)
}

// TestGetDecisionForRuleUserBucketedIntoLocalHoldout verifies that when a user is bucketed into a
// local holdout for a specific rule, the holdout variation is returned and rule evaluation is skipped.
func TestGetDecisionForRuleUserBucketedIntoLocalHoldout(t *testing.T) {
	mockConfig := new(mockProjectConfig)
	mockBucketer := new(MockExperimentBucketer)
	mockAudienceEval := new(MockAudienceTreeEvaluator)
	mockLogger := new(MockLogger)

	localHoldoutVar := entities.Variation{ID: "local_var_1", Key: "local_variation_1"}
	localHoldout := entities.Holdout{
		ID:     "local_holdout_1",
		Key:    "test_local_holdout",
		Status: entities.HoldoutStatusRunning,
		Variations: map[string]entities.Variation{
			"local_var_1": localHoldoutVar,
		},
		TrafficAllocation: []entities.Range{
			{EntityID: "local_var_1", EndOfRange: 10000}, // 100% traffic
		},
	}

	mockConfig.On("GetHoldoutsForRule", "rule_x").Return([]entities.Holdout{localHoldout})
	mockBucketer.On("Bucket", "test_user", mock.AnythingOfType("entities.Experiment"), entities.Group{}).Return(&localHoldoutVar, reasons.Reason(""), nil)
	mockLogger.On("Debug", mock.Anything).Return()
	mockLogger.On("Info", mock.Anything).Return()

	service := HoldoutService{
		audienceTreeEvaluator: mockAudienceEval,
		bucketer:              mockBucketer,
		logger:                mockLogger,
	}

	userCtx := entities.UserContext{ID: "test_user"}
	options := &decide.Options{}

	decision, _, err := service.GetDecisionForRule("rule_x", mockConfig, userCtx, options)

	assert.NoError(t, err)
	assert.NotNil(t, decision.Variation, "User bucketed into local holdout should return holdout variation")
	assert.Equal(t, localHoldoutVar.ID, decision.Variation.ID)
	assert.Equal(t, Holdout, decision.Source)
	mockConfig.AssertExpectations(t)
	mockBucketer.AssertExpectations(t)
}

// TestGetDecisionForRuleUserMissesLocalHoldout verifies that when a user is NOT bucketed into a
// local holdout, an empty decision is returned so that regular rule evaluation proceeds.
func TestGetDecisionForRuleUserMissesLocalHoldout(t *testing.T) {
	mockConfig := new(mockProjectConfig)
	mockBucketer := new(MockExperimentBucketer)
	mockAudienceEval := new(MockAudienceTreeEvaluator)
	mockLogger := new(MockLogger)

	localHoldout := entities.Holdout{
		ID:     "local_holdout_miss",
		Key:    "test_local_holdout_miss",
		Status: entities.HoldoutStatusRunning,
		Variations: map[string]entities.Variation{
			"local_var_1": {ID: "local_var_1", Key: "local_variation_1"},
		},
		TrafficAllocation: []entities.Range{
			{EntityID: "local_var_1", EndOfRange: 0}, // 0% traffic — no user will bucket in
		},
	}

	mockConfig.On("GetHoldoutsForRule", "rule_y").Return([]entities.Holdout{localHoldout})
	// Bucketer returns nil (user not bucketed)
	mockBucketer.On("Bucket", "test_user_miss", mock.AnythingOfType("entities.Experiment"), entities.Group{}).Return(nil, reasons.Reason(""), nil)
	mockLogger.On("Debug", mock.Anything).Return()
	mockLogger.On("Info", mock.Anything).Return()

	service := HoldoutService{
		audienceTreeEvaluator: mockAudienceEval,
		bucketer:              mockBucketer,
		logger:                mockLogger,
	}

	userCtx := entities.UserContext{ID: "test_user_miss"}
	options := &decide.Options{}

	decision, _, err := service.GetDecisionForRule("rule_y", mockConfig, userCtx, options)

	assert.NoError(t, err)
	assert.Nil(t, decision.Variation, "User not bucketed into local holdout should fall through to regular evaluation")
	mockConfig.AssertExpectations(t)
	mockBucketer.AssertExpectations(t)
}

// TestGetDecisionForRuleRuleSpecificity verifies that a local holdout targeting rule X
// does NOT affect rule Y.
func TestGetDecisionForRuleRuleSpecificity(t *testing.T) {
	mockConfig := new(mockProjectConfig)
	mockBucketer := new(MockExperimentBucketer)
	mockAudienceEval := new(MockAudienceTreeEvaluator)
	mockLogger := new(MockLogger)

	localHoldoutVar := entities.Variation{ID: "local_var_1", Key: "local_variation_1"}
	localHoldout := entities.Holdout{
		ID:     "local_holdout_for_rule_x",
		Key:    "holdout_for_rule_x",
		Status: entities.HoldoutStatusRunning,
		Variations: map[string]entities.Variation{
			"local_var_1": localHoldoutVar,
		},
		TrafficAllocation: []entities.Range{
			{EntityID: "local_var_1", EndOfRange: 10000}, // 100% traffic
		},
	}

	// rule_x has the holdout; rule_y has none
	mockConfig.On("GetHoldoutsForRule", "rule_x").Return([]entities.Holdout{localHoldout})
	mockConfig.On("GetHoldoutsForRule", "rule_y").Return([]entities.Holdout{})
	mockConfig.On("GetAudienceMap").Return(map[string]entities.Audience{})
	mockBucketer.On("Bucket", "test_user", mock.AnythingOfType("entities.Experiment"), entities.Group{}).Return(&localHoldoutVar, reasons.Reason(""), nil)
	mockLogger.On("Debug", mock.Anything).Return()
	mockLogger.On("Info", mock.Anything).Return()

	service := HoldoutService{
		audienceTreeEvaluator: mockAudienceEval,
		bucketer:              mockBucketer,
		logger:                mockLogger,
	}

	userCtx := entities.UserContext{ID: "test_user"}
	options := &decide.Options{}

	// rule_x: user should be in holdout
	decisionX, _, errX := service.GetDecisionForRule("rule_x", mockConfig, userCtx, options)
	assert.NoError(t, errX)
	assert.NotNil(t, decisionX.Variation, "Local holdout for rule_x must apply to rule_x")

	// rule_y: user should NOT be in any holdout (holdout doesn't target rule_y)
	decisionY, _, errY := service.GetDecisionForRule("rule_y", mockConfig, userCtx, options)
	assert.NoError(t, errY)
	assert.Nil(t, decisionY.Variation, "Local holdout for rule_x must NOT apply to rule_y")
}

// TestGetDecisionForRuleLocalHoldoutSkippedIfNotRunning verifies that a non-running local holdout
// is skipped and no holdout decision is returned.
func TestGetDecisionForRuleLocalHoldoutSkippedIfNotRunning(t *testing.T) {
	mockConfig := new(mockProjectConfig)
	mockBucketer := new(MockExperimentBucketer)
	mockAudienceEval := new(MockAudienceTreeEvaluator)
	mockLogger := new(MockLogger)

	pausedLocalHoldout := entities.Holdout{
		ID:     "paused_local_holdout",
		Key:    "paused_holdout",
		Status: entities.HoldoutStatus("Paused"), // not running
		Variations: map[string]entities.Variation{
			"var_1": {ID: "var_1", Key: "variation_1"},
		},
		TrafficAllocation: []entities.Range{
			{EntityID: "var_1", EndOfRange: 10000},
		},
	}

	mockConfig.On("GetHoldoutsForRule", "rule_z").Return([]entities.Holdout{pausedLocalHoldout})
	mockConfig.On("GetAudienceMap").Return(map[string]entities.Audience{})
	mockLogger.On("Debug", mock.Anything).Return()
	mockLogger.On("Info", mock.Anything).Return()

	service := HoldoutService{
		audienceTreeEvaluator: mockAudienceEval,
		bucketer:              mockBucketer,
		logger:                mockLogger,
	}

	userCtx := entities.UserContext{ID: "test_user"}
	options := &decide.Options{}

	decision, _, err := service.GetDecisionForRule("rule_z", mockConfig, userCtx, options)

	assert.NoError(t, err)
	assert.Nil(t, decision.Variation, "Non-running local holdout should be skipped")
	// Bucketer should never be called for a non-running holdout
	mockBucketer.AssertNotCalled(t, "Bucket")
}
