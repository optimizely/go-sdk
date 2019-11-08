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
	"errors"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/optimizely/go-sdk/pkg/entities"
)

type CompositeExperimentTestSuite struct {
	suite.Suite
	mockConfig             *mockProjectConfig
	mockExperimentService  *MockExperimentDecisionService
	mockExperimentService2 *MockExperimentDecisionService
	testDecisionContext    ExperimentDecisionContext
}

func (s *CompositeExperimentTestSuite) SetupTest() {
	s.mockConfig = new(mockProjectConfig)
	s.mockExperimentService = new(MockExperimentDecisionService)
	s.mockExperimentService2 = new(MockExperimentDecisionService)

	// Setup test data
	s.testDecisionContext = ExperimentDecisionContext{
		Experiment:    &testExp1111,
		ProjectConfig: s.mockConfig,
	}
}

func (s *CompositeExperimentTestSuite) TestGetDecision() {
	// test that we return out of the decision making and the next one does not get called
	testUserContext := entities.UserContext{
		ID: "test_user_1",
	}

	expectedVariation := testExp1111.Variations["2222"]
	expectedExperimentDecision := ExperimentDecision{
		Variation: &expectedVariation,
	}
	s.mockExperimentService.On("GetDecision", s.testDecisionContext, testUserContext).Return(expectedExperimentDecision, nil)

	compositeExperimentService := &CompositeExperimentService{
		experimentServices: []ExperimentService{s.mockExperimentService, s.mockExperimentService2},
	}
	decision, err := compositeExperimentService.GetDecision(s.testDecisionContext, testUserContext)
	s.Equal(expectedExperimentDecision, decision)
	s.NoError(err)
	s.mockExperimentService.AssertExpectations(s.T())
	s.mockExperimentService2.AssertNotCalled(s.T(), "GetDecision")

}

func (s *CompositeExperimentTestSuite) TestGetDecisionFallthrough() {
	// test that we move onto the next decision service if no decision is made
	testUserContext := entities.UserContext{
		ID: "test_user_1",
	}

	expectedVariation := testExp1111.Variations["2222"]
	expectedExperimentDecision := ExperimentDecision{}
	s.mockExperimentService.On("GetDecision", s.testDecisionContext, testUserContext).Return(expectedExperimentDecision, nil)

	expectedExperimentDecision2 := ExperimentDecision{
		Variation: &expectedVariation,
	}
	s.mockExperimentService2.On("GetDecision", s.testDecisionContext, testUserContext).Return(expectedExperimentDecision2, nil)

	compositeExperimentService := &CompositeExperimentService{
		experimentServices: []ExperimentService{s.mockExperimentService, s.mockExperimentService2},
	}
	decision, err := compositeExperimentService.GetDecision(s.testDecisionContext, testUserContext)

	s.NoError(err)
	s.Equal(expectedExperimentDecision2, decision)
	s.mockExperimentService.AssertExpectations(s.T())
	s.mockExperimentService2.AssertExpectations(s.T())
}

func (s *CompositeExperimentTestSuite) TestGetDecisionNoDecisionsMade() {
	// test when no decisions are made
	testUserContext := entities.UserContext{
		ID: "test_user_1",
	}
	expectedExperimentDecision := ExperimentDecision{}
	s.mockExperimentService.On("GetDecision", s.testDecisionContext, testUserContext).Return(expectedExperimentDecision, nil)

	expectedExperimentDecision2 := ExperimentDecision{}
	s.mockExperimentService2.On("GetDecision", s.testDecisionContext, testUserContext).Return(expectedExperimentDecision2, nil)

	compositeExperimentService := &CompositeExperimentService{
		experimentServices: []ExperimentService{s.mockExperimentService, s.mockExperimentService2},
	}
	decision, err := compositeExperimentService.GetDecision(s.testDecisionContext, testUserContext)

	s.Error(err)
	s.Equal(expectedExperimentDecision2, decision)
	s.mockExperimentService.AssertExpectations(s.T())
	s.mockExperimentService2.AssertExpectations(s.T())
}

func (s *CompositeExperimentTestSuite) TestGetDecisionReturnsError() {
	// Assert that we continue to the next inner service when an inner service GetDecision returns an error
	testUserContext := entities.UserContext{
		ID: "test_user_1",
	}

	testDecisionContext := ExperimentDecisionContext{
		Experiment:    &testExp1114,
		ProjectConfig: s.mockConfig,
	}

	shouldBeIgnoredDecision := ExperimentDecision{
		Variation: &testExp1114Var2225,
	}
	s.mockExperimentService.On("GetDecision", testDecisionContext, testUserContext).Return(shouldBeIgnoredDecision, errors.New("Error making decision"))

	expectedDecision := ExperimentDecision{
		Variation: &testExp1114Var2226,
	}
	s.mockExperimentService2.On("GetDecision", testDecisionContext, testUserContext).Return(expectedDecision, nil)

	compositeExperimentService := &CompositeExperimentService{
		experimentServices: []ExperimentService{
			s.mockExperimentService,
			s.mockExperimentService2,
		},
	}
	decision, err := compositeExperimentService.GetDecision(testDecisionContext, testUserContext)
	s.Equal(expectedDecision, decision)
	s.NoError(err)
	s.mockExperimentService.AssertExpectations(s.T())
	s.mockExperimentService2.AssertExpectations(s.T())
}

func (s *CompositeExperimentTestSuite) TestNewCompositeExperimentService() {
	// Assert that the service is instantiated with the correct child services in the right order
	compositeExperimentService := NewCompositeExperimentService()
	s.Equal(2, len(compositeExperimentService.experimentServices))
	s.IsType(&ExperimentWhitelistService{}, compositeExperimentService.experimentServices[0])
	s.IsType(&ExperimentBucketerService{}, compositeExperimentService.experimentServices[1])
}

func (s *CompositeExperimentTestSuite) TestNewCompositeExperimentServiceWithCustomOptions() {
	mockUserProfileService := new(MockUserProfileService)
	mockExperimentOverrideStore := new(MapOverridesStore)
	compositeExperimentService := NewCompositeExperimentService(
		WithUserProfileService(mockUserProfileService),
		WithOverrideStore(mockExperimentOverrideStore),
	)
	s.Equal(mockUserProfileService, compositeExperimentService.userProfileService)
	s.Equal(mockExperimentOverrideStore, compositeExperimentService.overrideStore)
}

func TestCompositeExperimentTestSuite(t *testing.T) {
	suite.Run(t, new(CompositeExperimentTestSuite))
}
