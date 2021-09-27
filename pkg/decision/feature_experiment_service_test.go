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
	"testing"

	"github.com/optimizely/go-sdk/pkg/decide"
	"github.com/optimizely/go-sdk/pkg/entities"
	"github.com/optimizely/go-sdk/pkg/logging"

	"github.com/stretchr/testify/suite"
)

type FeatureExperimentServiceTestSuite struct {
	suite.Suite
	mockConfig                 *mockProjectConfig
	testFeatureDecisionContext FeatureDecisionContext
	mockExperimentService      *MockExperimentDecisionService
	options                    *decide.Options
	reasons                    decide.DecisionReasons
}

func (s *FeatureExperimentServiceTestSuite) SetupTest() {
	s.mockConfig = new(mockProjectConfig)
	s.testFeatureDecisionContext = FeatureDecisionContext{
		Feature:               &testFeat3335,
		ProjectConfig:         s.mockConfig,
		Variable:              testVariable,
		ForcedDecisionService: NewForcedDecisionService("test_user"),
	}
	s.mockExperimentService = new(MockExperimentDecisionService)
	s.options = &decide.Options{}
	s.reasons = decide.NewDecisionReasons(s.options)
}

func (s *FeatureExperimentServiceTestSuite) TestGetDecision() {
	testUserContext := entities.UserContext{
		ID: "test_user_1",
	}

	expectedVariation := testExp1113.Variations["2223"]
	returnExperimentDecision := ExperimentDecision{
		Variation: &expectedVariation,
	}
	testExperimentDecisionContext := ExperimentDecisionContext{
		Experiment:    &testExp1113,
		ProjectConfig: s.mockConfig,
	}
	s.mockExperimentService.On("GetDecision", testExperimentDecisionContext, testUserContext, s.options).Return(returnExperimentDecision, s.reasons, nil)

	featureExperimentService := &FeatureExperimentService{
		compositeExperimentService: s.mockExperimentService,
		logger:                     logging.GetLogger("sdkKey", "FeatureExperimentService"),
	}

	expectedFeatureDecision := FeatureDecision{
		Experiment: *testExperimentDecisionContext.Experiment,
		Variation:  &expectedVariation,
		Source:     FeatureTest,
	}
	decision, _, err := featureExperimentService.GetDecision(s.testFeatureDecisionContext, testUserContext, s.options)
	s.Equal(expectedFeatureDecision, decision)
	s.NoError(err)
	s.mockExperimentService.AssertExpectations(s.T())
}

func (s *FeatureExperimentServiceTestSuite) TestGetDecisionWithForcedDecision() {
	testUserContext := entities.UserContext{
		ID: "test_user_1",
	}

	expectedVariation := testExp1113.Variations["2223"]
	flagVariationsMap := map[string][]entities.Variation{
		s.testFeatureDecisionContext.Feature.Key: {
			expectedVariation,
		},
	}
	s.mockConfig.On("GetFlagVariationsMap").Return(flagVariationsMap)
	s.testFeatureDecisionContext.ForcedDecisionService.SetForcedDecision(s.testFeatureDecisionContext.Feature.Key, testExp1113Key, expectedVariation.Key)

	testExperimentDecisionContext := ExperimentDecisionContext{
		Experiment:    &testExp1113,
		ProjectConfig: s.mockConfig,
	}

	featureExperimentService := &FeatureExperimentService{
		compositeExperimentService: s.mockExperimentService,
		logger:                     logging.GetLogger("sdkKey", "FeatureExperimentService"),
	}

	expectedFeatureDecision := FeatureDecision{
		Experiment: *testExperimentDecisionContext.Experiment,
		Variation:  &expectedVariation,
		Source:     FeatureTest,
	}
	decision, _, err := featureExperimentService.GetDecision(s.testFeatureDecisionContext, testUserContext, s.options)
	s.Equal(expectedFeatureDecision, decision)
	s.NoError(err)
	// Makes sure that decision returned was a forcedDecision
	s.mockExperimentService.AssertNotCalled(s.T(), "GetDecision", testExperimentDecisionContext, testUserContext, s.options)
}

func (s *FeatureExperimentServiceTestSuite) TestGetDecisionMutex() {
	testUserContext := entities.UserContext{
		ID: "test_user_1",
	}

	// first experiment returns nil to simulate user not being bucketed into this experiment in the group
	nilDecision := ExperimentDecision{}
	testExperimentDecisionContext1 := ExperimentDecisionContext{
		Experiment:    &testExp1113,
		ProjectConfig: s.mockConfig,
	}
	s.mockExperimentService.On("GetDecision", testExperimentDecisionContext1, testUserContext, s.options).Return(nilDecision, s.reasons, nil)

	// second experiment returns a valid decision to simulate user being bucketed into this experiment in the group
	expectedVariation := testExp1114.Variations["2225"]
	returnExperimentDecision := ExperimentDecision{
		Variation: &expectedVariation,
	}
	testExperimentDecisionContext2 := ExperimentDecisionContext{
		Experiment:    &testExp1114,
		ProjectConfig: s.mockConfig,
	}
	s.mockExperimentService.On("GetDecision", testExperimentDecisionContext2, testUserContext, s.options).Return(returnExperimentDecision, s.reasons, nil)

	expectedFeatureDecision := FeatureDecision{
		Experiment: *testExperimentDecisionContext2.Experiment,
		Variation:  &expectedVariation,
		Source:     FeatureTest,
	}
	featureExperimentService := &FeatureExperimentService{
		compositeExperimentService: s.mockExperimentService,
		logger:                     logging.GetLogger("sdkKey", "FeatureExperimentService"),
	}
	decision, _, err := featureExperimentService.GetDecision(s.testFeatureDecisionContext, testUserContext, s.options)
	s.Equal(expectedFeatureDecision, decision)
	s.NoError(err)
	s.mockExperimentService.AssertExpectations(s.T())
}

func (s *FeatureExperimentServiceTestSuite) TestNewFeatureExperimentService() {
	compositeExperimentService := &CompositeExperimentService{logger: logging.GetLogger("sdkKey", "CompositeExperimentService")}
	featureExperimentService := NewFeatureExperimentService(logging.GetLogger("", ""), compositeExperimentService)
	s.IsType(compositeExperimentService, featureExperimentService.compositeExperimentService)
}

func TestFeatureExperimentServiceTestSuite(t *testing.T) {
	suite.Run(t, new(FeatureExperimentServiceTestSuite))
}
