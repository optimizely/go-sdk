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

	"github.com/stretchr/testify/suite"

	"github.com/optimizely/go-sdk/v2/pkg/decide"
	"github.com/optimizely/go-sdk/v2/pkg/entities"
	"github.com/optimizely/go-sdk/v2/pkg/logging"
)

type CompositeExperimentTestSuite struct {
	suite.Suite
	mockConfig             *mockProjectConfig
	mockExperimentService  *MockExperimentDecisionService
	mockExperimentService2 *MockExperimentDecisionService
	mockCmabService        *MockExperimentDecisionService
	testDecisionContext    ExperimentDecisionContext
	options                *decide.Options
	reasons                decide.DecisionReasons
}

func (s *CompositeExperimentTestSuite) SetupTest() {
	s.mockConfig = new(mockProjectConfig)
	s.mockExperimentService = new(MockExperimentDecisionService)
	s.mockExperimentService2 = new(MockExperimentDecisionService)
	s.mockCmabService = new(MockExperimentDecisionService)
	s.options = &decide.Options{}
	s.reasons = decide.NewDecisionReasons(s.options)

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
	s.mockExperimentService.On("GetDecision", s.testDecisionContext, testUserContext, s.options).Return(expectedExperimentDecision, s.reasons, nil)

	compositeExperimentService := &CompositeExperimentService{
		experimentServices: []ExperimentService{s.mockExperimentService, s.mockCmabService, s.mockExperimentService2},
		logger:             logging.GetLogger("sdkKey", "ExperimentService"),
	}
	decision, _, err := compositeExperimentService.GetDecision(s.testDecisionContext, testUserContext, s.options)
	s.Equal(expectedExperimentDecision, decision)
	s.NoError(err)
	s.mockExperimentService.AssertExpectations(s.T())
	s.mockCmabService.AssertNotCalled(s.T(), "GetDecision")
	s.mockExperimentService2.AssertNotCalled(s.T(), "GetDecision")
}

func (s *CompositeExperimentTestSuite) TestGetDecisionFallthrough() {
	// test that we move onto the next decision service if no decision is made
	testUserContext := entities.UserContext{
		ID: "test_user_1",
	}

	expectedVariation := testExp1111.Variations["2222"]
	emptyExperimentDecision := ExperimentDecision{}
	s.mockExperimentService.On("GetDecision", s.testDecisionContext, testUserContext, s.options).Return(emptyExperimentDecision, s.reasons, nil)
	s.mockCmabService.On("GetDecision", s.testDecisionContext, testUserContext, s.options).Return(emptyExperimentDecision, s.reasons, nil)

	expectedExperimentDecision2 := ExperimentDecision{
		Variation: &expectedVariation,
	}
	s.mockExperimentService2.On("GetDecision", s.testDecisionContext, testUserContext, s.options).Return(expectedExperimentDecision2, s.reasons, nil)

	compositeExperimentService := &CompositeExperimentService{
		experimentServices: []ExperimentService{s.mockExperimentService, s.mockCmabService, s.mockExperimentService2},
		logger:             logging.GetLogger("sdkKey", "CompositeExperimentService"),
	}
	decision, _, err := compositeExperimentService.GetDecision(s.testDecisionContext, testUserContext, s.options)

	s.NoError(err)
	s.Equal(expectedExperimentDecision2, decision)
	s.mockExperimentService.AssertExpectations(s.T())
	s.mockCmabService.AssertExpectations(s.T())
	s.mockExperimentService2.AssertExpectations(s.T())
}

func (s *CompositeExperimentTestSuite) TestGetDecisionNoDecisionsMade() {
	// test when no decisions are made
	testUserContext := entities.UserContext{
		ID: "test_user_1",
	}
	emptyExperimentDecision := ExperimentDecision{}
	s.mockExperimentService.On("GetDecision", s.testDecisionContext, testUserContext, s.options).Return(emptyExperimentDecision, s.reasons, nil)
	s.mockCmabService.On("GetDecision", s.testDecisionContext, testUserContext, s.options).Return(emptyExperimentDecision, s.reasons, nil)
	s.mockExperimentService2.On("GetDecision", s.testDecisionContext, testUserContext, s.options).Return(emptyExperimentDecision, s.reasons, nil)

	compositeExperimentService := &CompositeExperimentService{
		experimentServices: []ExperimentService{s.mockExperimentService, s.mockCmabService, s.mockExperimentService2},
		logger:             logging.GetLogger("sdkKey", "CompositeExperimentService"),
	}
	decision, _, err := compositeExperimentService.GetDecision(s.testDecisionContext, testUserContext, s.options)

	s.NoError(err)
	s.Equal(emptyExperimentDecision, decision)
	s.mockExperimentService.AssertExpectations(s.T())
	s.mockCmabService.AssertExpectations(s.T())
	s.mockExperimentService2.AssertExpectations(s.T())
}

func (s *CompositeExperimentTestSuite) TestGetDecisionReturnsError() {
	// Assert that we continue to the next inner service when a non-CMAB service GetDecision returns an error
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
	s.mockExperimentService.On("GetDecision", testDecisionContext, testUserContext, s.options).Return(shouldBeIgnoredDecision, s.reasons, errors.New("Error making decision"))

	emptyDecision := ExperimentDecision{}
	s.mockCmabService.On("GetDecision", testDecisionContext, testUserContext, s.options).Return(emptyDecision, s.reasons, nil)

	expectedDecision := ExperimentDecision{
		Variation: &testExp1114Var2226,
	}
	s.mockExperimentService2.On("GetDecision", testDecisionContext, testUserContext, s.options).Return(expectedDecision, s.reasons, nil)

	compositeExperimentService := &CompositeExperimentService{
		experimentServices: []ExperimentService{
			s.mockExperimentService,
			s.mockCmabService,
			s.mockExperimentService2,
		},
		logger: logging.GetLogger("sdkKey", "CompositeExperimentService"),
	}
	decision, _, err := compositeExperimentService.GetDecision(testDecisionContext, testUserContext, s.options)
	s.Equal(expectedDecision, decision)
	s.NoError(err)
	s.mockExperimentService.AssertExpectations(s.T())
	s.mockCmabService.AssertExpectations(s.T())
	s.mockExperimentService2.AssertExpectations(s.T())
}

func (s *CompositeExperimentTestSuite) TestGetDecisionCmabError() {
	// Test that CMAB service errors are propagated up
	testUserContext := entities.UserContext{
		ID: "test_user_1",
	}

	testDecisionContext := ExperimentDecisionContext{
		Experiment:    &testExp1114,
		ProjectConfig: s.mockConfig,
	}

	// Mock whitelist service returning empty decision
	emptyDecision := ExperimentDecision{}
	s.mockExperimentService.On("GetDecision", testDecisionContext, testUserContext, s.options).Return(emptyDecision, s.reasons, nil)

	// Mock CMAB service returning error
	expectedError := errors.New("CMAB service error")
	s.mockCmabService.On("GetDecision", testDecisionContext, testUserContext, s.options).Return(emptyDecision, s.reasons, expectedError)

	compositeExperimentService := &CompositeExperimentService{
		experimentServices: []ExperimentService{
			s.mockExperimentService,
			s.mockCmabService,
			s.mockExperimentService2,
		},
		logger: logging.GetLogger("sdkKey", "CompositeExperimentService"),
	}

	// The error from CMAB service should be propagated
	decision, _, err := compositeExperimentService.GetDecision(testDecisionContext, testUserContext, s.options)
	s.Equal(emptyDecision, decision)
	s.Equal(expectedError, err)

	s.mockExperimentService.AssertExpectations(s.T())
	s.mockCmabService.AssertExpectations(s.T())
	s.mockExperimentService2.AssertNotCalled(s.T(), "GetDecision")
}

func (s *CompositeExperimentTestSuite) TestNewCompositeExperimentService() {
	// Assert that the service is instantiated with the correct child services in the right order
	compositeExperimentService := NewCompositeExperimentService("")

	// Update from 2 to 3 services (now includes CMAB service)
	s.Equal(3, len(compositeExperimentService.experimentServices))

	s.IsType(&ExperimentWhitelistService{}, compositeExperimentService.experimentServices[0])

	// Add assertion for the new CMAB service
	s.IsType(&ExperimentCmabService{}, compositeExperimentService.experimentServices[1])

	// Update index from 1 to 2 for the bucketer service
	s.IsType(&ExperimentBucketerService{}, compositeExperimentService.experimentServices[2])
}

func (s *CompositeExperimentTestSuite) TestNewCompositeExperimentServiceWithCustomOptions() {
	mockUserProfileService := new(MockUserProfileService)
	mockExperimentOverrideStore := new(MapExperimentOverridesStore)
	compositeExperimentService := NewCompositeExperimentService("",
		WithUserProfileService(mockUserProfileService),
		WithOverrideStore(mockExperimentOverrideStore),
	)
	s.Equal(mockUserProfileService, compositeExperimentService.userProfileService)
	s.Equal(mockExperimentOverrideStore, compositeExperimentService.overrideStore)

	// Verify that CMAB service is still created with custom options
	s.Equal(3, len(compositeExperimentService.experimentServices))
	s.IsType(&ExperimentCmabService{}, compositeExperimentService.experimentServices[1])
}

func TestCompositeExperimentTestSuite(t *testing.T) {
	suite.Run(t, new(CompositeExperimentTestSuite))
}
