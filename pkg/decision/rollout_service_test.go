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

package decision

import (
	"testing"

	"github.com/optimizely/go-sdk/pkg/decision/evaluator"

	"github.com/optimizely/go-sdk/pkg/decision/reasons"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/optimizely/go-sdk/pkg/entities"
)

type RolloutServiceTestSuite struct {
	suite.Suite
	mockConfig                    *mockProjectConfig
	mockAudienceTreeEvaluator     *MockAudienceTreeEvaluator
	mockExperimentService         *MockExperimentDecisionService
	testExperimentDecisionContext ExperimentDecisionContext
	testFeatureDecisionContext    FeatureDecisionContext
	testConditionTreeParams       *entities.TreeParameters
	testUserContext               entities.UserContext
}

func (s *RolloutServiceTestSuite) SetupTest() {
	s.mockConfig = new(mockProjectConfig)
	s.mockAudienceTreeEvaluator = new(MockAudienceTreeEvaluator)
	s.mockExperimentService = new(MockExperimentDecisionService)
	s.testExperimentDecisionContext = ExperimentDecisionContext{
		Experiment:    &testExp1112,
		ProjectConfig: s.mockConfig,
	}
	s.testFeatureDecisionContext = FeatureDecisionContext{
		Feature:       &testFeatRollout3334,
		ProjectConfig: s.mockConfig,
	}

	testAudienceMap := map[string]entities.Audience{
		"5555": testAudience5555,
	}
	s.testUserContext = entities.UserContext{
		ID: "test_user",
	}
	s.testConditionTreeParams = entities.NewTreeParameters(&s.testUserContext, testAudienceMap)
	s.mockConfig.On("GetAudienceMap").Return(testAudienceMap)
}

func (s *RolloutServiceTestSuite) TestGetDecisionHappyPath() {
	// Test experiment passes targeting and bucketing
	testExperimentBucketerDecision := ExperimentDecision{
		Variation: &testExp1112Var2222,
		Decision:  Decision{Reason: reasons.BucketedIntoVariation},
	}
	s.mockAudienceTreeEvaluator.On("Evaluate", testExp1112.AudienceConditionTree, s.testConditionTreeParams).Return(true, true)
	s.mockExperimentService.On("GetDecision", s.testExperimentDecisionContext, s.testUserContext).Return(testExperimentBucketerDecision, nil)

	testRolloutService := RolloutService{
		audienceTreeEvaluator:     s.mockAudienceTreeEvaluator,
		experimentBucketerService: s.mockExperimentService,
	}
	expectedFeatureDecision := FeatureDecision{
		Experiment: testExp1112,
		Variation:  &testExp1112Var2222,
		Source:     Rollout,
		Decision:   Decision{Reason: reasons.BucketedIntoRollout},
	}
	decision, _ := testRolloutService.GetDecision(s.testFeatureDecisionContext, s.testUserContext)
	s.Equal(expectedFeatureDecision, decision)
	s.mockAudienceTreeEvaluator.AssertExpectations(s.T())
	s.mockExperimentService.AssertExpectations(s.T())
}

func (s *RolloutServiceTestSuite) TestGetDecisionFailsBucketing() {
	// Test experiment passes targeting but not bucketing
	testExperimentBucketerDecision := ExperimentDecision{
		Decision: Decision{
			Reason: reasons.NotBucketedIntoVariation,
		},
	}

	s.mockAudienceTreeEvaluator.On("Evaluate", testExp1112.AudienceConditionTree, s.testConditionTreeParams).Return(true, true)
	s.mockExperimentService.On("GetDecision", s.testExperimentDecisionContext, s.testUserContext).Return(testExperimentBucketerDecision, nil)
	testRolloutService := RolloutService{
		audienceTreeEvaluator:     s.mockAudienceTreeEvaluator,
		experimentBucketerService: s.mockExperimentService,
	}
	expectedFeatureDecision := FeatureDecision{
		Decision: Decision{
			Reason: reasons.FailedRolloutBucketing,
		},
		Experiment: testExp1112,
		Source:     Rollout,
	}
	decision, _ := testRolloutService.GetDecision(s.testFeatureDecisionContext, s.testUserContext)
	s.Equal(expectedFeatureDecision, decision)
	s.mockAudienceTreeEvaluator.AssertExpectations(s.T())
	s.mockExperimentService.AssertExpectations(s.T())
}

func (s *RolloutServiceTestSuite) TestGetDecisionFailsTargeting() {
	s.mockAudienceTreeEvaluator.On("Evaluate", testExp1112.AudienceConditionTree, s.testConditionTreeParams).Return(false, true)
	testRolloutService := RolloutService{
		audienceTreeEvaluator:     s.mockAudienceTreeEvaluator,
		experimentBucketerService: s.mockExperimentService,
	}
	expectedFeatureDecision := FeatureDecision{
		Decision: Decision{
			Reason: reasons.FailedRolloutTargeting,
		},
		Source: Rollout,
	}
	decision, _ := testRolloutService.GetDecision(s.testFeatureDecisionContext, s.testUserContext)
	s.Equal(expectedFeatureDecision, decision)
	s.mockAudienceTreeEvaluator.AssertExpectations(s.T())
	s.mockExperimentService.AssertExpectations(s.T())
}

func TestNewRolloutService(t *testing.T) {
	rolloutService := NewRolloutService("")
	assert.IsType(t, &evaluator.MixedTreeEvaluator{}, rolloutService.audienceTreeEvaluator)
	assert.IsType(t, &ExperimentBucketerService{}, rolloutService.experimentBucketerService)
}

func TestRolloutServiceTestSuite(t *testing.T) {
	suite.Run(t, new(RolloutServiceTestSuite))
}
