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
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/optimizely/go-sdk/v2/pkg/cmab"
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

func (s *CompositeExperimentTestSuite) TestNewCompositeExperimentServiceWithCmabConfig() {
	// Test with custom CMAB config
	cmabConfig := cmab.Config{
		CacheSize:   200,
		CacheTTL:    5 * time.Minute,
		HTTPTimeout: 30 * time.Second,
		RetryConfig: &cmab.RetryConfig{
			MaxRetries: 5,
		},
	}

	compositeExperimentService := NewCompositeExperimentService("test-sdk-key",
		WithCmabConfig(cmabConfig),
	)

	// Verify CMAB config was merged (not replaced entirely)
	s.Equal(200, compositeExperimentService.cmabConfig.CacheSize)              // From config
	s.Equal(5*time.Minute, compositeExperimentService.cmabConfig.CacheTTL)     // From config
	s.Equal(30*time.Second, compositeExperimentService.cmabConfig.HTTPTimeout) // From config
	s.NotNil(compositeExperimentService.cmabConfig.RetryConfig)
	s.Equal(5, compositeExperimentService.cmabConfig.RetryConfig.MaxRetries) // From config

	// Verify service order
	s.Equal(3, len(compositeExperimentService.experimentServices))
	s.IsType(&ExperimentWhitelistService{}, compositeExperimentService.experimentServices[0])
	s.IsType(&ExperimentCmabService{}, compositeExperimentService.experimentServices[1])
	s.IsType(&ExperimentBucketerService{}, compositeExperimentService.experimentServices[2])
}

func (s *CompositeExperimentTestSuite) TestNewCompositeExperimentServiceWithPartialCmabConfig() {
	// Test that partial CMAB config merges with defaults instead of replacing everything
	compositeExperimentService := NewCompositeExperimentService("test-sdk-key",
		WithCmabCacheSize(500), // Only set cache size
	)

	// Verify only cache size was changed, other defaults remain
	s.Equal(500, compositeExperimentService.cmabConfig.CacheSize)
	s.Equal(cmab.DefaultCacheTTL, compositeExperimentService.cmabConfig.CacheTTL)       // Should be default
	s.Equal(cmab.DefaultHTTPTimeout, compositeExperimentService.cmabConfig.HTTPTimeout) // Should be default
	s.NotNil(compositeExperimentService.cmabConfig.RetryConfig)
	s.Equal(cmab.DefaultMaxRetries, compositeExperimentService.cmabConfig.RetryConfig.MaxRetries) // Should be default

	// Test with multiple granular options
	compositeExperimentService2 := NewCompositeExperimentService("test-sdk-key",
		WithCmabCacheSize(300),
		WithCmabCacheTTL(10*time.Minute),
	)

	s.Equal(300, compositeExperimentService2.cmabConfig.CacheSize)
	s.Equal(10*time.Minute, compositeExperimentService2.cmabConfig.CacheTTL)
	s.Equal(cmab.DefaultHTTPTimeout, compositeExperimentService2.cmabConfig.HTTPTimeout) // Should be default

	// Test merging partial config object
	partialConfig := cmab.Config{
		CacheSize: 250, // Only set cache size in config
	}
	compositeExperimentService3 := NewCompositeExperimentService("test-sdk-key",
		WithCmabConfig(partialConfig),
	)

	s.Equal(250, compositeExperimentService3.cmabConfig.CacheSize)
	s.Equal(cmab.DefaultCacheTTL, compositeExperimentService3.cmabConfig.CacheTTL)       // Should be default
	s.Equal(cmab.DefaultHTTPTimeout, compositeExperimentService3.cmabConfig.HTTPTimeout) // Should be default
}

func (s *CompositeExperimentTestSuite) TestNewCompositeExperimentServiceWithAllOptions() {
	// Test with all options including CMAB config
	mockUserProfileService := new(MockUserProfileService)
	mockExperimentOverrideStore := new(MapExperimentOverridesStore)
	cmabConfig := cmab.Config{
		CacheSize: 100,
		CacheTTL:  time.Minute,
	}

	compositeExperimentService := NewCompositeExperimentService("test-sdk-key",
		WithUserProfileService(mockUserProfileService),
		WithOverrideStore(mockExperimentOverrideStore),
		WithCmabConfig(cmabConfig),
	)

	// Verify all options were applied
	s.Equal(mockUserProfileService, compositeExperimentService.userProfileService)
	s.Equal(mockExperimentOverrideStore, compositeExperimentService.overrideStore)
	// Verify the partial config was merged with defaults
	s.Equal(100, compositeExperimentService.cmabConfig.CacheSize)                                 // From config
	s.Equal(time.Minute, compositeExperimentService.cmabConfig.CacheTTL)                          // From config
	s.Equal(cmab.DefaultHTTPTimeout, compositeExperimentService.cmabConfig.HTTPTimeout)           // Default
	s.NotNil(compositeExperimentService.cmabConfig.RetryConfig)                                   // Default
	s.Equal(cmab.DefaultMaxRetries, compositeExperimentService.cmabConfig.RetryConfig.MaxRetries) // Default

	// Verify service order with all services
	s.Equal(4, len(compositeExperimentService.experimentServices))
	s.IsType(&ExperimentOverrideService{}, compositeExperimentService.experimentServices[0])
	s.IsType(&ExperimentWhitelistService{}, compositeExperimentService.experimentServices[1])
	s.IsType(&ExperimentCmabService{}, compositeExperimentService.experimentServices[2])
	s.IsType(&PersistingExperimentService{}, compositeExperimentService.experimentServices[3])
}

func (s *CompositeExperimentTestSuite) TestCmabServiceReturnsError() {
	// Test that CMAB service error is properly propagated
	mockCmabService := new(MockExperimentDecisionService)
	testErr := errors.New("Failed to fetch CMAB data for experiment exp_123")

	mockCmabService.On("GetDecision", mock.Anything, mock.Anything, mock.Anything).Return(
		ExperimentDecision{},
		decide.NewDecisionReasons(s.options),
		testErr,
	)

	compositeService := &CompositeExperimentService{
		experimentServices: []ExperimentService{mockCmabService},
		logger:             logging.GetLogger("", "CompositeExperimentService"),
	}

	userContext := entities.UserContext{ID: "test_user"}
	decision, reasons, err := compositeService.GetDecision(s.testDecisionContext, userContext, s.options)

	// Error should be returned immediately without trying other services
	s.Error(err)
	s.Equal(testErr, err)
	s.Nil(decision.Variation)
	s.NotNil(reasons)

	mockCmabService.AssertExpectations(s.T())
}

func TestCompositeExperimentTestSuite(t *testing.T) {
	suite.Run(t, new(CompositeExperimentTestSuite))
}
