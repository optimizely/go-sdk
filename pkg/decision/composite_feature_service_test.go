/****************************************************************************
 * Copyright 2019-2025, Optimizely, Inc. and contributors                   *
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
	"errors"
	"testing"

	"github.com/optimizely/go-sdk/v2/pkg/decide"
	"github.com/optimizely/go-sdk/v2/pkg/decision/reasons"
	"github.com/optimizely/go-sdk/v2/pkg/entities"
	"github.com/optimizely/go-sdk/v2/pkg/logging"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type CompositeFeatureServiceTestSuite struct {
	suite.Suite
	mockFeatureService         *MockFeatureDecisionService
	mockFeatureService2        *MockFeatureDecisionService
	testFeatureDecisionContext FeatureDecisionContext
	options                    *decide.Options
	reasons                    decide.DecisionReasons
}

func (s *CompositeFeatureServiceTestSuite) SetupTest() {
	mockConfig := new(mockProjectConfig)

	s.mockFeatureService = new(MockFeatureDecisionService)
	s.mockFeatureService2 = new(MockFeatureDecisionService)
	s.options = &decide.Options{}
	s.reasons = decide.NewDecisionReasons(s.options)

	// Setup test data
	s.testFeatureDecisionContext = FeatureDecisionContext{
		Feature:       &testFeat3335,
		ProjectConfig: mockConfig,
	}
}

func (s *CompositeFeatureServiceTestSuite) TestGetDecision() {
	// Test that we return the first decision that is made and the next decision service does not get called
	testUserContext := entities.UserContext{
		ID: "test_user_1",
	}

	expectedDecision := FeatureDecision{
		Decision:   Decision{reasons.BucketedIntoVariation},
		Source:     FeatureTest,
		Experiment: testExp1113,
		Variation:  &testExp1113Var2223,
	}
	s.mockFeatureService.On("GetDecision", s.testFeatureDecisionContext, testUserContext, s.options).Return(expectedDecision, s.reasons, nil)

	compositeFeatureService := &CompositeFeatureService{
		featureServices: []FeatureService{
			s.mockFeatureService,
			s.mockFeatureService2,
		},
		logger: logging.GetLogger("sdkKey", "CompositeFeatureService"),
	}
	decision, _, err := compositeFeatureService.GetDecision(s.testFeatureDecisionContext, testUserContext, s.options)
	s.Equal(expectedDecision, decision)
	s.NoError(err)
	s.mockFeatureService.AssertExpectations(s.T())
	s.mockFeatureService2.AssertNotCalled(s.T(), "GetDecision")
}

func (s *CompositeFeatureServiceTestSuite) TestGetDecisionFallthrough() {
	// test that we move onto the next decision service if no decision is made
	testUserContext := entities.UserContext{
		ID: "test_user_1",
	}

	nilDecision := FeatureDecision{}
	s.mockFeatureService.On("GetDecision", s.testFeatureDecisionContext, testUserContext, s.options).Return(nilDecision, s.reasons, nil)

	expectedDecision := FeatureDecision{
		Variation: &testExp1113Var2223,
	}
	s.mockFeatureService2.On("GetDecision", s.testFeatureDecisionContext, testUserContext, s.options).Return(expectedDecision, s.reasons, nil)

	compositeFeatureService := &CompositeFeatureService{
		featureServices: []FeatureService{
			s.mockFeatureService,
			s.mockFeatureService2,
		},
		logger: logging.GetLogger("sdkKey", "CompositeFeatureService"),
	}
	decision, _, err := compositeFeatureService.GetDecision(s.testFeatureDecisionContext, testUserContext, s.options)
	s.Equal(expectedDecision, decision)
	s.NoError(err)
	s.mockFeatureService.AssertExpectations(s.T())
	s.mockFeatureService2.AssertExpectations(s.T())
}

func (s *CompositeFeatureServiceTestSuite) TestGetDecisionReturnsError() {
	// test that errors now propagate up instead of continuing to next service
	testUserContext := entities.UserContext{
		ID: "test_user_1",
	}

	shouldBeIgnoredDecision := FeatureDecision{
		Variation: &testExp1113Var2223,
	}
	// Any error now causes immediate return (no fallthrough)
	s.mockFeatureService.On("GetDecision", s.testFeatureDecisionContext, testUserContext, s.options).Return(shouldBeIgnoredDecision, s.reasons, errors.New("Generic experiment error"))

	compositeFeatureService := &CompositeFeatureService{
		featureServices: []FeatureService{
			s.mockFeatureService,
			s.mockFeatureService2,
		},
		logger: logging.GetLogger("sdkKey", "CompositeFeatureService"),
	}
	decision, _, err := compositeFeatureService.GetDecision(s.testFeatureDecisionContext, testUserContext, s.options)

	// Change: Now we expect error propagation and empty decision
	s.Equal(FeatureDecision{}, decision)
	s.Error(err)
	s.Equal("Generic experiment error", err.Error())
	s.mockFeatureService.AssertExpectations(s.T())
	// Change: Second service should NOT be called when first service returns error
	s.mockFeatureService2.AssertNotCalled(s.T(), "GetDecision")
}
func (s *CompositeFeatureServiceTestSuite) TestGetDecisionReturnsLastDecisionWithError() {
	// This test is now invalid - rename to reflect new behavior
	// Test that first error stops evaluation (no "last decision" concept anymore)
	testUserContext := entities.UserContext{
		ID: "test_user_1",
	}

	expectedDecision := FeatureDecision{
		Variation: &testExp1113Var2223,
	}
	s.mockFeatureService.On("GetDecision", s.testFeatureDecisionContext, testUserContext, s.options).Return(expectedDecision, s.reasons, errors.New("test error"))

	compositeFeatureService := &CompositeFeatureService{
		featureServices: []FeatureService{
			s.mockFeatureService,
			s.mockFeatureService2,
		},
		logger: logging.GetLogger("sdkKey", "CompositeFeatureService"),
	}
	decision, _, err := compositeFeatureService.GetDecision(s.testFeatureDecisionContext, testUserContext, s.options)

	// Change: Now we expect empty decision and error from first service
	s.Equal(FeatureDecision{}, decision)
	s.Error(err)
	s.Equal("test error", err.Error())
	s.mockFeatureService.AssertExpectations(s.T())
	// Change: Second service should NOT be called
	s.mockFeatureService2.AssertNotCalled(s.T(), "GetDecision")
}

func (s *CompositeFeatureServiceTestSuite) TestGetDecisionWithCmabError() {
	// Test that CMAB errors are now propagated as Go errors
	testUserContext := entities.UserContext{
		ID: "test_user_1",
	}

	// Mock the first service (FeatureExperimentService) to return a CMAB error
	cmabError := errors.New("Failed to fetch CMAB data for experiment exp_1.")
	emptyDecision := FeatureDecision{}
	s.mockFeatureService.On("GetDecision", s.testFeatureDecisionContext, testUserContext, s.options).Return(emptyDecision, s.reasons, cmabError)

	compositeFeatureService := &CompositeFeatureService{
		featureServices: []FeatureService{
			s.mockFeatureService,
			s.mockFeatureService2,
		},
		logger: logging.GetLogger("sdkKey", "CompositeFeatureService"),
	}

	decision, _, err := compositeFeatureService.GetDecision(s.testFeatureDecisionContext, testUserContext, s.options)

	// Change: CMAB errors now propagate as Go errors (this is the expected behavior now)
	s.Equal(FeatureDecision{}, decision)
	s.Error(err, "CMAB errors should now propagate as Go errors")
	s.Equal(cmabError.Error(), err.Error())

	s.mockFeatureService.AssertExpectations(s.T())
	// Verify that the rollout service was NOT called
	s.mockFeatureService2.AssertNotCalled(s.T(), "GetDecision")
}

func TestExcludeTDFalseBlocksEverything(t *testing.T) {
	mockConfig := new(mockProjectConfig)
	mockBucketer := new(MockExperimentBucketer)
	mockAudienceEval := new(MockAudienceTreeEvaluator)
	mockLogger := new(MockLogger)

	holdoutVar := entities.Variation{ID: "holdout_var", Key: "holdout_variation"}
	holdout := entities.Holdout{
		ID:                        "holdout_etd_false",
		Key:                       "holdout_exclude_td_false",
		Status:                    entities.HoldoutStatusRunning,
		ExcludeTargetedDeliveries: false,
		Variations:                map[string]entities.Variation{"holdout_var": holdoutVar},
		TrafficAllocation:         []entities.Range{{EntityID: "holdout_var", EndOfRange: 10000}},
	}

	mockConfig.On("GetGlobalHoldouts").Return([]entities.Holdout{holdout})
	mockConfig.On("GetAudienceMap").Return(map[string]entities.Audience{})
	mockBucketer.On("Bucket", "test_user", mock.AnythingOfType("entities.Experiment"), entities.Group{}).Return(&holdoutVar, reasons.Reason(""), nil)
	mockLogger.On("Debug", mock.Anything).Return()
	mockLogger.On("Info", mock.Anything).Return()

	holdoutService := &HoldoutService{
		audienceTreeEvaluator: mockAudienceEval,
		bucketer:              mockBucketer,
		logger:                mockLogger,
	}

	mockFeatureService := new(MockFeatureDecisionService)
	mockRolloutService := new(MockFeatureDecisionService)

	compositeFeatureService := &CompositeFeatureService{
		holdoutService:  holdoutService,
		featureServices: []FeatureService{mockFeatureService, mockRolloutService},
		logger:          logging.GetLogger("", "CompositeFeatureService"),
	}

	feature := entities.Feature{ID: "feat_1", Key: "test_feature"}
	decisionContext := FeatureDecisionContext{
		Feature:       &feature,
		ProjectConfig: mockConfig,
	}
	userContext := entities.UserContext{ID: "test_user"}
	options := &decide.Options{}

	decision, _, err := compositeFeatureService.GetDecision(decisionContext, userContext, options)

	assert.NoError(t, err)
	assert.NotNil(t, decision.Variation)
	assert.Equal(t, holdoutVar.ID, decision.Variation.ID)
	assert.Equal(t, Holdout, decision.Source)
	mockFeatureService.AssertNotCalled(t, "GetDecision")
	mockRolloutService.AssertNotCalled(t, "GetDecision")
}

func TestExcludeTDTrueBlocksABExperiment(t *testing.T) {
	mockConfig := new(mockProjectConfig)
	mockBucketer := new(MockExperimentBucketer)
	mockAudienceEval := new(MockAudienceTreeEvaluator)
	mockLogger := new(MockLogger)

	holdoutVar := entities.Variation{ID: "holdout_var", Key: "holdout_variation"}
	holdout := entities.Holdout{
		ID:                        "holdout_etd_true",
		Key:                       "holdout_exclude_td_true",
		Status:                    entities.HoldoutStatusRunning,
		ExcludeTargetedDeliveries: true,
		Variations:                map[string]entities.Variation{"holdout_var": holdoutVar},
		TrafficAllocation:         []entities.Range{{EntityID: "holdout_var", EndOfRange: 10000}},
	}

	mockConfig.On("GetGlobalHoldouts").Return([]entities.Holdout{holdout})
	mockConfig.On("GetAudienceMap").Return(map[string]entities.Audience{})
	mockBucketer.On("Bucket", "test_user", mock.AnythingOfType("entities.Experiment"), entities.Group{}).Return(&holdoutVar, reasons.Reason(""), nil)
	mockLogger.On("Debug", mock.Anything).Return()
	mockLogger.On("Info", mock.Anything).Return()

	holdoutService := &HoldoutService{
		audienceTreeEvaluator: mockAudienceEval,
		bucketer:              mockBucketer,
		logger:                mockLogger,
	}

	mockFeatureService := new(MockFeatureDecisionService)
	mockRolloutService := new(MockFeatureDecisionService)

	feature := entities.Feature{ID: "feat_1", Key: "test_feature"}
	decisionContext := FeatureDecisionContext{
		Feature:       &feature,
		ProjectConfig: mockConfig,
	}
	userContext := entities.UserContext{ID: "test_user"}
	options := &decide.Options{IncludeReasons: true}
	decisionReasons := decide.NewDecisionReasons(options)

	emptyDecision := FeatureDecision{}
	mockRolloutService.On("GetDecision", decisionContext, userContext, options).Return(emptyDecision, decisionReasons, nil)

	compositeFeatureService := &CompositeFeatureService{
		holdoutService:  holdoutService,
		featureServices: []FeatureService{mockFeatureService, mockRolloutService},
		logger:          logging.GetLogger("", "CompositeFeatureService"),
	}

	decision, decisionReasons, err := compositeFeatureService.GetDecision(decisionContext, userContext, options)

	assert.NoError(t, err)
	assert.Nil(t, decision.Variation)
	assert.NotNil(t, decision.HoldoutExperiment)
	assert.NotNil(t, decision.HoldoutVariation)
	assert.Equal(t, holdoutVar.ID, decision.HoldoutVariation.ID)
	mockFeatureService.AssertNotCalled(t, "GetDecision")
	mockRolloutService.AssertExpectations(t)

	reportedReasons := decisionReasons.ToReport()
	assert.Contains(t, reportedReasons, "Holdout \"holdout_exclude_td_true\" has excludeTargetedDeliveries enabled, continuing to rollout evaluation.")
}

func TestExcludeTDTrueAllowsTDRollout(t *testing.T) {
	mockConfig := new(mockProjectConfig)
	mockBucketer := new(MockExperimentBucketer)
	mockAudienceEval := new(MockAudienceTreeEvaluator)
	mockLogger := new(MockLogger)

	holdoutVar := entities.Variation{ID: "holdout_var", Key: "holdout_variation"}
	holdout := entities.Holdout{
		ID:                        "holdout_etd_true",
		Key:                       "holdout_exclude_td_true",
		Status:                    entities.HoldoutStatusRunning,
		ExcludeTargetedDeliveries: true,
		Variations:                map[string]entities.Variation{"holdout_var": holdoutVar},
		TrafficAllocation:         []entities.Range{{EntityID: "holdout_var", EndOfRange: 10000}},
	}

	mockConfig.On("GetGlobalHoldouts").Return([]entities.Holdout{holdout})
	mockConfig.On("GetAudienceMap").Return(map[string]entities.Audience{})
	mockBucketer.On("Bucket", "test_user", mock.AnythingOfType("entities.Experiment"), entities.Group{}).Return(&holdoutVar, reasons.Reason(""), nil)
	mockLogger.On("Debug", mock.Anything).Return()
	mockLogger.On("Info", mock.Anything).Return()

	holdoutService := &HoldoutService{
		audienceTreeEvaluator: mockAudienceEval,
		bucketer:              mockBucketer,
		logger:                mockLogger,
	}

	rolloutVar := entities.Variation{ID: "rollout_var", Key: "rollout_variation"}
	rolloutDecision := FeatureDecision{
		Variation:  &rolloutVar,
		Source:     Rollout,
		Experiment: entities.Experiment{ID: "rollout_1", Key: "rollout_rule"},
	}

	mockFeatureService := new(MockFeatureDecisionService)
	mockRolloutService := new(MockFeatureDecisionService)

	feature := entities.Feature{ID: "feat_1", Key: "test_feature"}
	decisionContext := FeatureDecisionContext{
		Feature:       &feature,
		ProjectConfig: mockConfig,
	}
	userContext := entities.UserContext{ID: "test_user"}
	options := &decide.Options{IncludeReasons: true}
	decisionReasons := decide.NewDecisionReasons(options)

	mockRolloutService.On("GetDecision", decisionContext, userContext, options).Return(rolloutDecision, decisionReasons, nil)

	compositeFeatureService := &CompositeFeatureService{
		holdoutService:  holdoutService,
		featureServices: []FeatureService{mockFeatureService, mockRolloutService},
		logger:          logging.GetLogger("", "CompositeFeatureService"),
	}

	decision, decisionReasons, err := compositeFeatureService.GetDecision(decisionContext, userContext, options)

	assert.NoError(t, err)
	assert.NotNil(t, decision.Variation)
	assert.Equal(t, rolloutVar.ID, decision.Variation.ID)
	assert.Equal(t, Rollout, decision.Source)
	assert.NotNil(t, decision.HoldoutExperiment)
	assert.NotNil(t, decision.HoldoutVariation)
	assert.Equal(t, holdoutVar.ID, decision.HoldoutVariation.ID)
	mockFeatureService.AssertNotCalled(t, "GetDecision")
	mockRolloutService.AssertExpectations(t)

	reportedReasons := decisionReasons.ToReport()
	assert.Contains(t, reportedReasons, "Holdout \"holdout_exclude_td_true\" has excludeTargetedDeliveries enabled, continuing to rollout evaluation.")
}

func TestExcludeTDTrueNoDownstreamMatchReturnsEmpty(t *testing.T) {
	mockConfig := new(mockProjectConfig)
	mockBucketer := new(MockExperimentBucketer)
	mockAudienceEval := new(MockAudienceTreeEvaluator)
	mockLogger := new(MockLogger)

	holdoutVar := entities.Variation{ID: "holdout_var", Key: "holdout_variation"}
	holdout := entities.Holdout{
		ID:                        "holdout_etd_true",
		Key:                       "holdout_exclude_td_true",
		Status:                    entities.HoldoutStatusRunning,
		ExcludeTargetedDeliveries: true,
		Variations:                map[string]entities.Variation{"holdout_var": holdoutVar},
		TrafficAllocation:         []entities.Range{{EntityID: "holdout_var", EndOfRange: 10000}},
	}

	mockConfig.On("GetGlobalHoldouts").Return([]entities.Holdout{holdout})
	mockConfig.On("GetAudienceMap").Return(map[string]entities.Audience{})
	mockBucketer.On("Bucket", "test_user", mock.AnythingOfType("entities.Experiment"), entities.Group{}).Return(&holdoutVar, reasons.Reason(""), nil)
	mockLogger.On("Debug", mock.Anything).Return()
	mockLogger.On("Info", mock.Anything).Return()

	holdoutService := &HoldoutService{
		audienceTreeEvaluator: mockAudienceEval,
		bucketer:              mockBucketer,
		logger:                mockLogger,
	}

	mockFeatureService := new(MockFeatureDecisionService)
	mockRolloutService := new(MockFeatureDecisionService)

	feature := entities.Feature{ID: "feat_1", Key: "test_feature"}
	decisionContext := FeatureDecisionContext{
		Feature:       &feature,
		ProjectConfig: mockConfig,
	}
	userContext := entities.UserContext{ID: "test_user"}
	options := &decide.Options{IncludeReasons: true}
	decisionReasons := decide.NewDecisionReasons(options)

	emptyDecision := FeatureDecision{}
	mockRolloutService.On("GetDecision", decisionContext, userContext, options).Return(emptyDecision, decisionReasons, nil)

	compositeFeatureService := &CompositeFeatureService{
		holdoutService:  holdoutService,
		featureServices: []FeatureService{mockFeatureService, mockRolloutService},
		logger:          logging.GetLogger("", "CompositeFeatureService"),
	}

	decision, resultReasons, err := compositeFeatureService.GetDecision(decisionContext, userContext, options)

	assert.NoError(t, err)
	assert.Nil(t, decision.Variation)
	assert.Equal(t, "", decision.Source)
	assert.NotNil(t, decision.HoldoutExperiment)
	assert.NotNil(t, decision.HoldoutVariation)
	assert.Equal(t, holdoutVar.ID, decision.HoldoutVariation.ID)
	mockFeatureService.AssertNotCalled(t, "GetDecision")
	mockRolloutService.AssertExpectations(t)

	reportedReasons := resultReasons.ToReport()
	assert.Contains(t, reportedReasons, "Holdout \"holdout_exclude_td_true\" has excludeTargetedDeliveries enabled, continuing to rollout evaluation.")
}

func TestExcludeTDMissingFieldDefaultsFalse(t *testing.T) {
	holdout := entities.Holdout{
		ID:     "holdout_no_etd",
		Key:    "holdout_missing_field",
		Status: entities.HoldoutStatusRunning,
	}

	assert.False(t, holdout.ExcludeTargetedDeliveries)
}

func TestLocalHoldoutIgnoresExcludeTargetedDeliveries(t *testing.T) {
	mockConfig := new(mockProjectConfig)
	mockBucketer := new(MockExperimentBucketer)
	mockAudienceEval := new(MockAudienceTreeEvaluator)
	mockLogger := new(MockLogger)

	holdoutVar := entities.Variation{ID: "local_holdout_var", Key: "local_holdout_variation"}
	localHoldout := entities.Holdout{
		ID:                        "local_holdout_etd",
		Key:                       "local_holdout_with_etd",
		Status:                    entities.HoldoutStatusRunning,
		ExcludeTargetedDeliveries: true,
		Variations:                map[string]entities.Variation{"local_holdout_var": holdoutVar},
		TrafficAllocation:         []entities.Range{{EntityID: "local_holdout_var", EndOfRange: 10000}},
	}

	ruleID := "rule_123"
	mockConfig.On("GetHoldoutsForRule", ruleID).Return([]entities.Holdout{localHoldout})
	mockConfig.On("GetAudienceMap").Return(map[string]entities.Audience{})
	mockBucketer.On("Bucket", "test_user", mock.AnythingOfType("entities.Experiment"), entities.Group{}).Return(&holdoutVar, reasons.Reason(""), nil)
	mockLogger.On("Debug", mock.Anything).Return()
	mockLogger.On("Info", mock.Anything).Return()

	holdoutService := &HoldoutService{
		audienceTreeEvaluator: mockAudienceEval,
		bucketer:              mockBucketer,
		logger:                mockLogger,
	}

	userContext := entities.UserContext{ID: "test_user"}
	options := &decide.Options{}

	decision, _, err := holdoutService.GetLocalDecisionForRule(ruleID, mockConfig, userContext, options)

	assert.NoError(t, err)
	assert.NotNil(t, decision.Variation)
	assert.Equal(t, holdoutVar.ID, decision.Variation.ID)
	assert.Equal(t, Holdout, decision.Source)
}

func (s *CompositeFeatureServiceTestSuite) TestNewCompositeFeatureService() {
	// Assert that the service is instantiated with the correct child services in the right order
	compositeExperimentService := NewCompositeExperimentService("")
	compositeFeatureService := NewCompositeFeatureService("", compositeExperimentService)
	s.NotNil(compositeFeatureService.holdoutService)
	s.Equal(2, len(compositeFeatureService.featureServices))
	s.IsType(&FeatureExperimentService{compositeExperimentService: compositeExperimentService}, compositeFeatureService.featureServices[0])
	s.IsType(&RolloutService{}, compositeFeatureService.featureServices[1])
}

func TestCompositeFeatureTestSuite(t *testing.T) {
	suite.Run(t, new(CompositeFeatureServiceTestSuite))
}
