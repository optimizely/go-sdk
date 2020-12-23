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
	"sync"
	"testing"

	"github.com/optimizely/go-sdk/pkg/decide"
	"github.com/optimizely/go-sdk/pkg/decision/reasons"
	"github.com/optimizely/go-sdk/pkg/entities"
	"github.com/optimizely/go-sdk/pkg/logging"

	"github.com/stretchr/testify/suite"
)

type ExperimentOverrideServiceTestSuite struct {
	suite.Suite
	mockConfig      *mockProjectConfig
	overrides       *MapExperimentOverridesStore
	overrideService *ExperimentOverrideService
	options         *decide.Options
}

func (s *ExperimentOverrideServiceTestSuite) SetupTest() {
	s.mockConfig = new(mockProjectConfig)
	s.overrides = NewMapExperimentOverridesStore()
	s.overrideService = NewExperimentOverrideService(s.overrides, logging.GetLogger("", ""))
	s.options = &decide.Options{}
}

func (s *ExperimentOverrideServiceTestSuite) TestOverridesIncludeVariation() {
	testDecisionContext := ExperimentDecisionContext{
		Experiment:    &testExp1111,
		ProjectConfig: s.mockConfig,
	}
	testUserContext := entities.UserContext{
		ID: "test_user_1",
	}
	s.overrides.SetVariation(ExperimentOverrideKey{ExperimentKey: testExp1111.Key, UserID: "test_user_1"}, testExp1111Var2222.Key)
	decision, _, err := s.overrideService.GetDecision(testDecisionContext, testUserContext, s.options)
	s.NoError(err)
	s.NotNil(decision.Variation)
	s.Exactly(testExp1111Var2222.Key, decision.Variation.Key)
	s.Exactly(reasons.OverrideVariationAssignmentFound, decision.Reason)
}

func (s *ExperimentOverrideServiceTestSuite) TestNilDecisionContextExperiment() {
	testDecisionContext := ExperimentDecisionContext{
		ProjectConfig: s.mockConfig,
	}
	testUserContext := entities.UserContext{
		ID: "test_user_1",
	}
	decision, _, err := s.overrideService.GetDecision(testDecisionContext, testUserContext, s.options)
	s.Error(err)
	s.Nil(decision.Variation)
}

func (s *ExperimentOverrideServiceTestSuite) TestNoOverrideForExperiment() {
	testDecisionContext := ExperimentDecisionContext{
		Experiment:    &testExp1111,
		ProjectConfig: s.mockConfig,
	}
	testUserContext := entities.UserContext{
		ID: "test_user_1",
	}
	// The decision context refers to testExp1111, but this override is for another experiment
	s.overrides.SetVariation(ExperimentOverrideKey{ExperimentKey: testExp1113.Key, UserID: "test_user_1"}, testExp1113Var2224.Key)
	decision, _, err := s.overrideService.GetDecision(testDecisionContext, testUserContext, s.options)
	s.NoError(err)
	s.Nil(decision.Variation)
	s.Exactly(reasons.NoOverrideVariationAssignment, decision.Reason)
}

func (s *ExperimentOverrideServiceTestSuite) TestNoOverrideForUser() {
	testDecisionContext := ExperimentDecisionContext{
		Experiment:    &testExp1111,
		ProjectConfig: s.mockConfig,
	}
	testUserContext := entities.UserContext{
		ID: "test_user_1",
	}
	// The user context refers to "test_user_1", but this override is for another user
	s.overrides.SetVariation(ExperimentOverrideKey{ExperimentKey: testExp1111.Key, UserID: "test_user_2"}, testExp1111Var2222.Key)
	decision, _, err := s.overrideService.GetDecision(testDecisionContext, testUserContext, s.options)
	s.NoError(err)
	s.Nil(decision.Variation)
	s.Exactly(reasons.NoOverrideVariationAssignment, decision.Reason)
}

func (s *ExperimentOverrideServiceTestSuite) TestNoOverrideForUserOrExperiment() {
	testDecisionContext := ExperimentDecisionContext{
		Experiment:    &testExp1111,
		ProjectConfig: s.mockConfig,
	}
	testUserContext := entities.UserContext{
		ID: "test_user_1",
	}
	// This override is for both a different user and a different experiment than the ones in the contexts above
	s.overrides.SetVariation(ExperimentOverrideKey{ExperimentKey: testExp1113.Key, UserID: "test_user_3"}, testExp1111Var2222.Key)
	decision, _, err := s.overrideService.GetDecision(testDecisionContext, testUserContext, s.options)
	s.NoError(err)
	s.Nil(decision.Variation)
	s.Exactly(reasons.NoOverrideVariationAssignment, decision.Reason)
}

func (s *ExperimentOverrideServiceTestSuite) TestInvalidVariationInOverride() {
	testDecisionContext := ExperimentDecisionContext{
		Experiment:    &testExp1111,
		ProjectConfig: s.mockConfig,
	}
	testUserContext := entities.UserContext{
		ID: "test_user_1",
	}
	// This override variation key does not exist in the experiment
	s.overrides.SetVariation(ExperimentOverrideKey{ExperimentKey: testExp1111.Key, UserID: "test_user_1"}, "invalid_variation_key")
	decision, _, err := s.overrideService.GetDecision(testDecisionContext, testUserContext, s.options)
	s.NoError(err)
	s.Nil(decision.Variation)
	s.Exactly(reasons.InvalidOverrideVariationAssignment, decision.Reason)
}

// Test concurrent use of the MapExperimentOverrideStore
// Create 3 goroutines that set and get variations, and assert that all their sets take effect at the end
func (s *ExperimentOverrideServiceTestSuite) TestMapExperimentOverridesStoreConcurrent() {
	testDecisionContext := ExperimentDecisionContext{
		Experiment:    &testExp1111,
		ProjectConfig: s.mockConfig,
	}
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		s.overrides.SetVariation(ExperimentOverrideKey{ExperimentKey: testExp1111.Key, UserID: "test_user_1"}, testExp1111Var2222.Key)
		user1Decision, _, _ := s.overrideService.GetDecision(testDecisionContext, entities.UserContext{
			ID: "test_user_1",
		}, s.options)
		s.NotNil(user1Decision.Variation)
		s.Exactly(testExp1111Var2222.Key, user1Decision.Variation.Key)
		wg.Done()
	}()
	wg.Add(1)
	go func() {
		s.overrides.SetVariation(ExperimentOverrideKey{ExperimentKey: testExp1111.Key, UserID: "test_user_2"}, testExp1111Var2222.Key)
		user2Decision, _, _ := s.overrideService.GetDecision(testDecisionContext, entities.UserContext{
			ID: "test_user_2",
		}, s.options)
		s.NotNil(user2Decision.Variation)
		s.Exactly(testExp1111Var2222.Key, user2Decision.Variation.Key)
		wg.Done()
	}()
	wg.Add(1)
	go func() {
		s.overrides.SetVariation(ExperimentOverrideKey{ExperimentKey: testExp1111.Key, UserID: "test_user_3"}, testExp1111Var2222.Key)
		user3Decision, _, _ := s.overrideService.GetDecision(testDecisionContext, entities.UserContext{
			ID: "test_user_3",
		}, s.options)
		s.NotNil(user3Decision.Variation)
		s.Exactly(testExp1111Var2222.Key, user3Decision.Variation.Key)
		wg.Done()
	}()
	wg.Wait()
	user1Decision, _, _ := s.overrideService.GetDecision(testDecisionContext, entities.UserContext{
		ID: "test_user_1",
	}, s.options)
	user2Decision, _, _ := s.overrideService.GetDecision(testDecisionContext, entities.UserContext{
		ID: "test_user_2",
	}, s.options)
	user3Decision, _, _ := s.overrideService.GetDecision(testDecisionContext, entities.UserContext{
		ID: "test_user_3",
	}, s.options)
	s.NotNil(user1Decision.Variation)
	s.Exactly(testExp1111Var2222.Key, user1Decision.Variation.Key)
	s.NotNil(user2Decision.Variation)
	s.Exactly(testExp1111Var2222.Key, user2Decision.Variation.Key)
	s.NotNil(user3Decision.Variation)
	s.Exactly(testExp1111Var2222.Key, user3Decision.Variation.Key)
}

func (s *ExperimentOverrideServiceTestSuite) TestRemovePreviouslySetVariation() {
	testDecisionContext := ExperimentDecisionContext{
		Experiment:    &testExp1111,
		ProjectConfig: s.mockConfig,
	}
	testUserContext := entities.UserContext{
		ID: "test_user_1",
	}
	overrideKey := ExperimentOverrideKey{ExperimentKey: testExp1111.Key, UserID: "test_user_1"}
	s.overrides.SetVariation(overrideKey, testExp1111Var2222.Key)
	s.overrides.RemoveVariation(overrideKey)
	decision, _, err := s.overrideService.GetDecision(testDecisionContext, testUserContext, s.options)
	s.NoError(err)
	s.Nil(decision.Variation)
}

func TestExperimentOverridesTestSuite(t *testing.T) {
	suite.Run(t, new(ExperimentOverrideServiceTestSuite))
}
