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

package decision

import (
	"fmt"
	"testing"

	"github.com/optimizely/go-sdk/pkg/decision/evaluator"
	"github.com/optimizely/go-sdk/pkg/decision/reasons"
	"github.com/optimizely/go-sdk/pkg/entities"
	"github.com/optimizely/go-sdk/pkg/logging"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type RolloutServiceTestSuite struct {
	suite.Suite
	mockConfig                        *mockProjectConfig
	mockAudienceTreeEvaluator         *MockAudienceTreeEvaluator
	mockExperimentService             *MockExperimentDecisionService
	testExperiment1112DecisionContext ExperimentDecisionContext
	testFeatureDecisionContext        FeatureDecisionContext
	testConditionTreeParams           *entities.TreeParameters
	testUserContext                   entities.UserContext
	mockLogger                        *MockLogger
}

func (s *RolloutServiceTestSuite) SetupTest() {
	s.mockConfig = new(mockProjectConfig)
	s.mockAudienceTreeEvaluator = new(MockAudienceTreeEvaluator)
	s.mockExperimentService = new(MockExperimentDecisionService)
	s.testExperiment1112DecisionContext = ExperimentDecisionContext{
		Experiment:    &testExp1112,
		ProjectConfig: s.mockConfig,
	}
	s.testFeatureDecisionContext = FeatureDecisionContext{
		Feature:       &testFeatRollout3334,
		ProjectConfig: s.mockConfig,
	}

	testAudienceMap := map[string]entities.Audience{
		"5555": testAudience5555,
		"5556": testAudience5556,
		"5557": testAudience5557,
	}
	s.testUserContext = entities.UserContext{
		ID: "test_user",
	}
	s.testConditionTreeParams = entities.NewTreeParameters(&s.testUserContext, testAudienceMap)
	s.mockConfig.On("GetAudienceMap").Return(testAudienceMap)
	s.mockLogger = new(MockLogger)
}

func (s *RolloutServiceTestSuite) TestGetDecisionWithEmptyRolloutID() {

	testRolloutService := RolloutService{
		logger: s.mockLogger,
	}
	feature := testFeatRollout3334
	feature.Rollout.ID = ""
	featureDecisionContext := FeatureDecisionContext{
		Feature:       &feature,
		ProjectConfig: s.mockConfig,
	}
	expectedFeatureDecision := FeatureDecision{
		Source:   Rollout,
		Decision: Decision{Reason: reasons.NoRolloutForFeature},
	}
	decision, _ := testRolloutService.GetDecision(featureDecisionContext, s.testUserContext)
	s.Equal(expectedFeatureDecision, decision)
}

func (s *RolloutServiceTestSuite) TestGetDecisionWithNoExperiments() {

	testRolloutService := RolloutService{
		logger: s.mockLogger,
	}
	feature := testFeatRollout3334
	feature.Rollout.Experiments = []entities.Experiment{}
	featureDecisionContext := FeatureDecisionContext{
		Feature:       &feature,
		ProjectConfig: s.mockConfig,
	}
	expectedFeatureDecision := FeatureDecision{
		Source:   Rollout,
		Decision: Decision{Reason: reasons.RolloutHasNoExperiments},
	}
	decision, _ := testRolloutService.GetDecision(featureDecisionContext, s.testUserContext)
	s.Equal(expectedFeatureDecision, decision)
}

func (s *RolloutServiceTestSuite) TestGetDecisionHappyPath() {
	// Test experiment passes targeting and bucketing
	testExperimentBucketerDecision := ExperimentDecision{
		Variation: &testExp1112Var2222,
		Decision:  Decision{Reason: reasons.BucketedIntoVariation},
	}
	s.mockAudienceTreeEvaluator.On("Evaluate", testExp1112.AudienceConditionTree, s.testConditionTreeParams).Return(true, true)
	s.mockExperimentService.On("GetDecision", s.testExperiment1112DecisionContext, s.testUserContext).Return(testExperimentBucketerDecision, nil)

	testRolloutService := RolloutService{
		audienceTreeEvaluator:     s.mockAudienceTreeEvaluator,
		experimentBucketerService: s.mockExperimentService,
		logger:                    s.mockLogger,
	}
	expectedFeatureDecision := FeatureDecision{
		Experiment: testExp1112,
		Variation:  &testExp1112Var2222,
		Source:     Rollout,
		Decision:   Decision{Reason: reasons.BucketedIntoRollout},
	}
	s.mockLogger.On("Debug", fmt.Sprintf(logging.EvaluatingAudiencesForRollout.String(), "1"))
	s.mockLogger.On("Debug", fmt.Sprintf(logging.RolloutAudiencesEvaluatedTo.String(), "1", true))
	s.mockLogger.On("Debug", `Decision made for user "test_user" for feature rollout with key "test_feature_rollout_3334_key": Bucketed into feature rollout.`)
	decision, _ := testRolloutService.GetDecision(s.testFeatureDecisionContext, s.testUserContext)
	s.Equal(expectedFeatureDecision, decision)
	s.mockAudienceTreeEvaluator.AssertExpectations(s.T())
	s.mockExperimentService.AssertExpectations(s.T())
	s.mockLogger.AssertExpectations(s.T())
}

func (s *RolloutServiceTestSuite) TestGetDecisionFallbacksToLastWhenFailsBucketing() {
	testExperiment1112BucketerDecision := ExperimentDecision{
		Decision: Decision{
			Reason: reasons.NotBucketedIntoVariation,
		},
	}
	testExperiment1118BucketerDecision := ExperimentDecision{
		Variation: &testExp1118Var2224,
		Decision:  Decision{Reason: reasons.BucketedIntoVariation},
	}
	experiment1118DecisionContext := ExperimentDecisionContext{
		Experiment:    &testExp1118,
		ProjectConfig: s.mockConfig,
	}
	s.mockAudienceTreeEvaluator.On("Evaluate", testExp1112.AudienceConditionTree, s.testConditionTreeParams).Return(true, true)
	s.mockAudienceTreeEvaluator.On("Evaluate", testExp1118.AudienceConditionTree, s.testConditionTreeParams).Return(true, true)
	s.mockExperimentService.On("GetDecision", s.testExperiment1112DecisionContext, s.testUserContext).Return(testExperiment1112BucketerDecision, nil)
	s.mockExperimentService.On("GetDecision", experiment1118DecisionContext, s.testUserContext).Return(testExperiment1118BucketerDecision, nil)

	testRolloutService := RolloutService{
		audienceTreeEvaluator:     s.mockAudienceTreeEvaluator,
		experimentBucketerService: s.mockExperimentService,
		logger:                    s.mockLogger,
	}
	expectedFeatureDecision := FeatureDecision{
		Experiment: testExp1118,
		Variation:  &testExp1118Var2224,
		Source:     Rollout,
		Decision:   Decision{Reason: reasons.BucketedIntoRollout},
	}
	s.mockLogger.On("Debug", fmt.Sprintf(logging.EvaluatingAudiencesForRollout.String(), "1"))
	s.mockLogger.On("Debug", fmt.Sprintf(logging.RolloutAudiencesEvaluatedTo.String(), "1", true))
	s.mockLogger.On("Debug", fmt.Sprintf(logging.EvaluatingAudiencesForRollout.String(), "Everyone Else"))
	s.mockLogger.On("Debug", fmt.Sprintf(logging.RolloutAudiencesEvaluatedTo.String(), "Everyone Else", true))
	s.mockLogger.On("Debug", fmt.Sprintf(logging.UserInEveryoneElse.String(), "test_user"))
	s.mockLogger.On("Debug", `Decision made for user "test_user" for feature rollout with key "test_feature_rollout_3334_key": Bucketed into feature rollout.`)
	decision, _ := testRolloutService.GetDecision(s.testFeatureDecisionContext, s.testUserContext)
	s.Equal(expectedFeatureDecision, decision)
	s.mockAudienceTreeEvaluator.AssertExpectations(s.T())
	s.mockExperimentService.AssertExpectations(s.T())
	s.mockLogger.AssertExpectations(s.T())
}

func (s *RolloutServiceTestSuite) TestGetDecisionWhenFallbackBucketingFails() {
	testExperiment1112BucketerDecision := ExperimentDecision{
		Decision: Decision{
			Reason: reasons.NotBucketedIntoVariation,
		},
	}
	testExperiment1118DecisionContext := ExperimentDecisionContext{
		Experiment:    &testExp1118,
		ProjectConfig: s.mockConfig,
	}
	s.mockAudienceTreeEvaluator.On("Evaluate", testExp1112.AudienceConditionTree, s.testConditionTreeParams).Return(true, true)
	s.mockAudienceTreeEvaluator.On("Evaluate", testExp1118.AudienceConditionTree, s.testConditionTreeParams).Return(true, true)
	s.mockExperimentService.On("GetDecision", s.testExperiment1112DecisionContext, s.testUserContext).Return(testExperiment1112BucketerDecision, nil)
	s.mockExperimentService.On("GetDecision", testExperiment1118DecisionContext, s.testUserContext).Return(testExperiment1112BucketerDecision, nil)

	testRolloutService := RolloutService{
		audienceTreeEvaluator:     s.mockAudienceTreeEvaluator,
		experimentBucketerService: s.mockExperimentService,
		logger:                    s.mockLogger,
	}
	expectedFeatureDecision := FeatureDecision{
		Experiment: testExp1118,
		Source:     Rollout,
		Decision:   Decision{Reason: reasons.FailedRolloutBucketing},
	}
	s.mockLogger.On("Debug", fmt.Sprintf(logging.EvaluatingAudiencesForRollout.String(), "1"))
	s.mockLogger.On("Debug", fmt.Sprintf(logging.RolloutAudiencesEvaluatedTo.String(), "1", true))
	s.mockLogger.On("Debug", fmt.Sprintf(logging.EvaluatingAudiencesForRollout.String(), "Everyone Else"))
	s.mockLogger.On("Debug", fmt.Sprintf(logging.RolloutAudiencesEvaluatedTo.String(), "Everyone Else", true))
	s.mockLogger.On("Debug", fmt.Sprintf(logging.UserInEveryoneElse.String(), "test_user"))
	s.mockLogger.On("Debug", `Decision made for user "test_user" for feature rollout with key "test_feature_rollout_3334_key": Not bucketed into rollout.`)
	decision, _ := testRolloutService.GetDecision(s.testFeatureDecisionContext, s.testUserContext)
	s.Equal(expectedFeatureDecision, decision)
	s.mockAudienceTreeEvaluator.AssertExpectations(s.T())
	s.mockExperimentService.AssertExpectations(s.T())
	s.mockLogger.AssertExpectations(s.T())
}

func (s *RolloutServiceTestSuite) TestEvaluatesNextIfPreviousTargetingFails() {
	s.mockAudienceTreeEvaluator.On("Evaluate", testExp1112.AudienceConditionTree, s.testConditionTreeParams).Return(false, true)
	s.mockAudienceTreeEvaluator.On("Evaluate", testExp1117.AudienceConditionTree, s.testConditionTreeParams).Return(true, true)
	experiment1117DecisionContext := ExperimentDecisionContext{
		Experiment:    &testExp1117,
		ProjectConfig: s.mockConfig,
	}
	testExperimentBucketerDecision := ExperimentDecision{
		Variation: &testExp1117Var2223,
		Decision:  Decision{Reason: reasons.BucketedIntoVariation},
	}
	s.mockExperimentService.On("GetDecision", experiment1117DecisionContext, s.testUserContext).Return(testExperimentBucketerDecision, nil)

	testRolloutService := RolloutService{
		audienceTreeEvaluator:     s.mockAudienceTreeEvaluator,
		experimentBucketerService: s.mockExperimentService,
		logger:                    s.mockLogger,
	}
	expectedFeatureDecision := FeatureDecision{
		Experiment: testExp1117,
		Variation:  &testExp1117Var2223,
		Source:     Rollout,
		Decision:   Decision{Reason: reasons.BucketedIntoRollout},
	}
	s.mockLogger.On("Debug", fmt.Sprintf(logging.EvaluatingAudiencesForRollout.String(), "1"))
	s.mockLogger.On("Debug", fmt.Sprintf(logging.RolloutAudiencesEvaluatedTo.String(), "1", false))
	s.mockLogger.On("Debug", fmt.Sprintf(logging.UserNotInRollout.String(), "test_user", "1"))
	s.mockLogger.On("Debug", fmt.Sprintf(logging.EvaluatingAudiencesForRollout.String(), "2"))
	s.mockLogger.On("Debug", fmt.Sprintf(logging.RolloutAudiencesEvaluatedTo.String(), "2", true))
	s.mockLogger.On("Debug", `Decision made for user "test_user" for feature rollout with key "test_feature_rollout_3334_key": Bucketed into feature rollout.`)
	decision, _ := testRolloutService.GetDecision(s.testFeatureDecisionContext, s.testUserContext)
	s.Equal(expectedFeatureDecision, decision)
	s.mockAudienceTreeEvaluator.AssertExpectations(s.T())
	s.mockExperimentService.AssertExpectations(s.T())
	s.mockLogger.AssertExpectations(s.T())
}

func (s *RolloutServiceTestSuite) TestGetDecisionFailsTargeting() {
	s.mockAudienceTreeEvaluator.On("Evaluate", testExp1112.AudienceConditionTree, s.testConditionTreeParams).Return(false, true)
	s.mockAudienceTreeEvaluator.On("Evaluate", testExp1117.AudienceConditionTree, s.testConditionTreeParams).Return(false, true)
	s.mockAudienceTreeEvaluator.On("Evaluate", testExp1118.AudienceConditionTree, s.testConditionTreeParams).Return(false, true)
	testRolloutService := RolloutService{
		audienceTreeEvaluator:     s.mockAudienceTreeEvaluator,
		experimentBucketerService: s.mockExperimentService,
		logger:                    s.mockLogger,
	}
	expectedFeatureDecision := FeatureDecision{
		Decision: Decision{
			Reason: reasons.FailedRolloutTargeting,
		},
		Source: Rollout,
	}
	s.mockLogger.On("Debug", fmt.Sprintf(logging.EvaluatingAudiencesForRollout.String(), "1"))
	s.mockLogger.On("Debug", fmt.Sprintf(logging.RolloutAudiencesEvaluatedTo.String(), "1", false))
	s.mockLogger.On("Debug", fmt.Sprintf(logging.UserNotInRollout.String(), "test_user", "1"))
	s.mockLogger.On("Debug", fmt.Sprintf(logging.EvaluatingAudiencesForRollout.String(), "2"))
	s.mockLogger.On("Debug", fmt.Sprintf(logging.RolloutAudiencesEvaluatedTo.String(), "2", false))
	s.mockLogger.On("Debug", fmt.Sprintf(logging.UserNotInRollout.String(), "test_user", "2"))
	s.mockLogger.On("Debug", fmt.Sprintf(logging.EvaluatingAudiencesForRollout.String(), "Everyone Else"))
	s.mockLogger.On("Debug", fmt.Sprintf(logging.RolloutAudiencesEvaluatedTo.String(), "Everyone Else", false))
	decision, _ := testRolloutService.GetDecision(s.testFeatureDecisionContext, s.testUserContext)
	s.Equal(expectedFeatureDecision, decision)
	s.mockAudienceTreeEvaluator.AssertExpectations(s.T())
	s.mockExperimentService.AssertExpectations(s.T())
	s.mockLogger.AssertExpectations(s.T())
}

func TestNewRolloutService(t *testing.T) {
	rolloutService := NewRolloutService("")
	assert.IsType(t, &evaluator.MixedTreeEvaluator{}, rolloutService.audienceTreeEvaluator)
	assert.IsType(t, &ExperimentBucketerService{logger: logging.GetLogger("sdkKey", "ExperimentBucketerService")}, rolloutService.experimentBucketerService)
}

func TestRolloutServiceTestSuite(t *testing.T) {
	suite.Run(t, new(RolloutServiceTestSuite))
}
