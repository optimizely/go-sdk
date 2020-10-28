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

// Package decision //
package decision

import (
	"testing"

	"github.com/optimizely/go-sdk/pkg/decide"
	"github.com/optimizely/go-sdk/pkg/entities"
	"github.com/optimizely/go-sdk/pkg/logging"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

var testUserContext entities.UserContext = entities.UserContext{
	ID: "test_user_1",
}

type PersistingExperimentServiceTestSuite struct {
	suite.Suite
	mockProjectConfig      *mockProjectConfig
	mockExperimentService  *MockExperimentDecisionService
	mockUserProfileService *MockUserProfileService
	testComputedDecision   ExperimentDecision
	testDecisionContext    ExperimentDecisionContext
	options                decide.OptimizelyDecideOptions
	reasons                *decide.DecisionReasons
}

func (s *PersistingExperimentServiceTestSuite) SetupTest() {
	s.mockProjectConfig = new(mockProjectConfig)
	s.mockExperimentService = new(MockExperimentDecisionService)
	s.mockUserProfileService = new(MockUserProfileService)
	s.testDecisionContext = ExperimentDecisionContext{
		Experiment:    &testExp1113,
		ProjectConfig: s.mockProjectConfig,
	}

	computedVariation := testExp1113.Variations["2223"]
	s.testComputedDecision = ExperimentDecision{
		Variation: &computedVariation,
	}
	s.options = decide.OptimizelyDecideOptions{}
	s.reasons = decide.NewDecisionReasons(s.options)
	s.mockExperimentService.On("GetDecision", s.testDecisionContext, testUserContext, s.options, s.reasons).Return(s.testComputedDecision, nil)
}

func (s *PersistingExperimentServiceTestSuite) TestNilUserProfileService() {
	persistingExperimentService := NewPersistingExperimentService(nil, s.mockExperimentService, logging.GetLogger("", "NewPersistingExperimentService"))
	decision, err := persistingExperimentService.GetDecision(s.testDecisionContext, testUserContext, s.options, s.reasons)
	s.Equal(s.testComputedDecision, decision)
	s.NoError(err)
	s.mockExperimentService.AssertExpectations(s.T())
}

func (s *PersistingExperimentServiceTestSuite) TestSavedVariationFound() {
	decisionKey := NewUserDecisionKey(s.testDecisionContext.Experiment.ID)
	savedUserProfile := UserProfile{
		ID:                  testUserContext.ID,
		ExperimentBucketMap: map[UserDecisionKey]string{decisionKey: testExp1113Var2224.ID},
	}
	s.mockUserProfileService.On("Lookup", testUserContext.ID).Return(savedUserProfile)
	s.mockUserProfileService.On("Save", mock.Anything)

	persistingExperimentService := NewPersistingExperimentService(s.mockUserProfileService, s.mockExperimentService, logging.GetLogger("", "NewPersistingExperimentService"))
	decision, err := persistingExperimentService.GetDecision(s.testDecisionContext, testUserContext, s.options, s.reasons)
	savedDecision := ExperimentDecision{
		Variation: &testExp1113Var2224,
	}
	s.Equal(savedDecision, decision)
	s.NoError(err)
	s.mockExperimentService.AssertNotCalled(s.T(), "GetDecision", s.testDecisionContext, testUserContext)
	s.mockUserProfileService.AssertNotCalled(s.T(), "Save", mock.Anything)
}

func (s *PersistingExperimentServiceTestSuite) TestNoSavedVariation() {
	s.mockUserProfileService.On("Lookup", testUserContext.ID).Return(UserProfile{ID: testUserContext.ID}) // empty user profile
	decisionKey := NewUserDecisionKey(s.testDecisionContext.Experiment.ID)
	updatedUserProfile := UserProfile{
		ID:                  testUserContext.ID,
		ExperimentBucketMap: map[UserDecisionKey]string{decisionKey: s.testComputedDecision.Variation.ID},
	}

	s.mockUserProfileService.On("Save", updatedUserProfile)
	persistingExperimentService := NewPersistingExperimentService(s.mockUserProfileService, s.mockExperimentService, logging.GetLogger("", "NewPersistingExperimentService"))
	decision, err := persistingExperimentService.GetDecision(s.testDecisionContext, testUserContext, s.options, s.reasons)
	s.Equal(s.testComputedDecision, decision)
	s.NoError(err)
	s.mockExperimentService.AssertExpectations(s.T())
	s.mockUserProfileService.AssertExpectations(s.T())
}

func (s *PersistingExperimentServiceTestSuite) TestSavedVariationNoLongerValid() {
	decisionKey := NewUserDecisionKey(s.testDecisionContext.Experiment.ID)
	savedUserProfile := UserProfile{
		ID:                  testUserContext.ID,
		ExperimentBucketMap: map[UserDecisionKey]string{decisionKey: "forgotten_variation"},
	}
	s.mockUserProfileService.On("Lookup", testUserContext.ID).Return(savedUserProfile) // empty user profile

	updatedUserProfile := UserProfile{
		ID:                  testUserContext.ID,
		ExperimentBucketMap: map[UserDecisionKey]string{decisionKey: s.testComputedDecision.Variation.ID},
	}
	s.mockUserProfileService.On("Save", updatedUserProfile)
	persistingExperimentService := NewPersistingExperimentService(s.mockUserProfileService, s.mockExperimentService, logging.GetLogger("", "NewPersistingExperimentService"))
	decision, err := persistingExperimentService.GetDecision(s.testDecisionContext, testUserContext, s.options, s.reasons)
	s.Equal(s.testComputedDecision, decision)
	s.NoError(err)
	s.mockExperimentService.AssertExpectations(s.T())
	s.mockUserProfileService.AssertExpectations(s.T())
}

func TestPersistingExperimentServiceTestSuite(t *testing.T) {
	suite.Run(t, new(PersistingExperimentServiceTestSuite))
}
