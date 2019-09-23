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

	"github.com/optimizely/go-sdk/optimizely/entities"
	"github.com/stretchr/testify/suite"
)

type FeatureExperimentServiceTestSuite struct {
	suite.Suite
	testFeatureDecisionContext    FeatureDecisionContext
	testExperimentDecisionContext ExperimentDecisionContext
	mockExperimentService         *MockExperimentDecisionService
}

func (s *FeatureExperimentServiceTestSuite) SetupTest() {
	mockProjectConfig := new(mockProjectConfig)
	s.testFeatureDecisionContext = FeatureDecisionContext{
		Feature:       &testFeat3335,
		ProjectConfig: mockProjectConfig,
	}

	s.testExperimentDecisionContext = ExperimentDecisionContext{
		Experiment:    &testExp1113,
		ProjectConfig: mockProjectConfig,
	}
	s.mockExperimentService = new(MockExperimentDecisionService)
}

func (s *FeatureExperimentServiceTestSuite) TestGetDecision() {
	testUserContext := entities.UserContext{
		ID: "test_user_1",
	}

	expectedVariation := testExp1113.Variations["2223"]
	returnExperimentDecision := ExperimentDecision{
		Variation: &expectedVariation,
	}
	s.mockExperimentService.On("GetDecision", s.testExperimentDecisionContext, testUserContext).Return(returnExperimentDecision, nil)
	featureExperimentService := &FeatureExperimentService{
		compositeExperimentService: s.mockExperimentService,
	}

	expectedFeatureDecision := FeatureDecision{
		Experiment: *s.testExperimentDecisionContext.Experiment,
		Variation:  &expectedVariation,
		Source:     FeatureTest,
	}
	decision, err := featureExperimentService.GetDecision(s.testFeatureDecisionContext, testUserContext)
	s.Equal(expectedFeatureDecision, decision)
	s.NoError(err)
	s.mockExperimentService.AssertExpectations(s.T())
}

// func (s *FeatureExperimentServiceTestSuite) TestGetDecisionMutext() {
// 	testUserContext := entities.UserContext{
// 		ID: "test_user_1",
// 	}
// }

func TestFeatureExperimentServiceTestSuite(t *testing.T) {
	suite.Run(t, new(FeatureExperimentServiceTestSuite))
}
