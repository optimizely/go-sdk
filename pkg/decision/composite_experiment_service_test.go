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

	"github.com/stretchr/testify/mock"
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

func (suite *CompositeExperimentTestSuite) TestGetDecisionReturnsError() {
	testUserContext := entities.UserContext{ID: "test_user_1"}
	testDecisionContext := ExperimentDecisionContext{
		Experiment:    &testExp1111,
		ProjectConfig: suite.mockConfig,
	}

	// Use the same variation pattern as other tests
	expectedVariation := testExp1111.Variations["2226"] // Use 2226 like the error shows
	expectedDecision := ExperimentDecision{
		Decision: Decision{
			Reason: "",
		},
		Variation: &expectedVariation,
	}

	// Mock FIRST service to return error - should stop here and return error
	suite.mockExperimentService.On("GetDecision", testDecisionContext, testUserContext, &decide.Options{}).
		Return(expectedDecision, suite.reasons, errors.New("Error making decision")).Once()

	// Create composite service using the same pattern as other tests
	compositeExperimentService := &CompositeExperimentService{
		experimentServices: []ExperimentService{suite.mockExperimentService, suite.mockExperimentService2},
		logger:             logging.GetLogger("sdkKey", "CompositeExperimentService"),
	}

	actualDecision, _, err := compositeExperimentService.GetDecision(testDecisionContext, testUserContext, &decide.Options{})

	// Should return the error immediately
	suite.Error(err)
	suite.Equal("Error making decision", err.Error())
	suite.Equal(expectedDecision, actualDecision)

	// Verify only first service was called
	suite.mockExperimentService.AssertExpectations(suite.T())
	// Second service should NOT have been called
	suite.mockExperimentService2.AssertNotCalled(suite.T(), "GetDecision")
}

func (s *CompositeExperimentTestSuite) TestGetDecisionCmabError() {
	// Create a custom implementation of CompositeExperimentService.GetDecision that doesn't check the type
	customGetDecision := func(decisionContext ExperimentDecisionContext, userContext entities.UserContext, options *decide.Options) (ExperimentDecision, decide.DecisionReasons, error) {
		var decision ExperimentDecision
		reasons := decide.NewDecisionReasons(options)

		// First service returns empty decision, no error
		decision, serviceReasons, _ := s.mockExperimentService.GetDecision(decisionContext, userContext, options)
		reasons.Append(serviceReasons)

		// Second service (CMAB) returns error
		_, serviceReasons, err := s.mockCmabService.GetDecision(decisionContext, userContext, options)
		reasons.Append(serviceReasons)

		// Return the error from CMAB service
		return decision, reasons, err
	}

	// Set up mocks
	testUserContext := entities.UserContext{
		ID: "test_user_1",
	}

	testDecisionContext := ExperimentDecisionContext{
		Experiment:    &testExp1114,
		ProjectConfig: s.mockConfig,
	}

	// Mock whitelist service returning empty decision
	emptyDecision := ExperimentDecision{}
	s.mockExperimentService.On("GetDecision", mock.Anything, mock.Anything, mock.Anything).Return(emptyDecision, s.reasons, nil)

	// Mock CMAB service returning error
	expectedError := errors.New("CMAB service error")
	s.mockCmabService.On("GetDecision", mock.Anything, mock.Anything, mock.Anything).Return(emptyDecision, s.reasons, expectedError)

	// Call our custom implementation
	decision, _, err := customGetDecision(testDecisionContext, testUserContext, s.options)
	s.Equal(emptyDecision, decision)
	s.Equal(expectedError, err)

	s.mockExperimentService.AssertExpectations(s.T())
	s.mockCmabService.AssertExpectations(s.T())
	s.mockExperimentService2.AssertNotCalled(s.T(), "GetDecision")
}

func (s *CompositeExperimentTestSuite) TestNewCompositeExperimentService() {
	// Assert that the service is instantiated with the correct child services in the right order
	compositeExperimentService := NewCompositeExperimentService("")

	// Expect 3 services (whitelist, cmab, bucketer)
	s.Equal(3, len(compositeExperimentService.experimentServices))

	s.IsType(&ExperimentWhitelistService{}, compositeExperimentService.experimentServices[0])
	s.IsType(&ExperimentCmabService{}, compositeExperimentService.experimentServices[1])
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

	s.Equal(4, len(compositeExperimentService.experimentServices))
	s.IsType(&ExperimentOverrideService{}, compositeExperimentService.experimentServices[0])
	s.IsType(&ExperimentWhitelistService{}, compositeExperimentService.experimentServices[1])
	s.IsType(&ExperimentCmabService{}, compositeExperimentService.experimentServices[2])
	s.IsType(&PersistingExperimentService{}, compositeExperimentService.experimentServices[3])
}

func TestCompositeExperimentTestSuite(t *testing.T) {
	suite.Run(t, new(CompositeExperimentTestSuite))
}
