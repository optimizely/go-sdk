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

// Package decision //
package decision

import (
	"testing"

	"github.com/optimizely/go-sdk/pkg/decision/reasons"
	"github.com/optimizely/go-sdk/pkg/entities"
	"github.com/stretchr/testify/suite"
)

type ExperimentOverrideServiceTestSuite struct {
	suite.Suite
	mockConfig      *mockProjectConfig
	overrides       map[ExperimentOverrideKey]string
	overrideService *ExperimentOverrideService
}

func (s *ExperimentOverrideServiceTestSuite) SetupTest() {
	s.mockConfig = new(mockProjectConfig)
	s.overrides = make(map[ExperimentOverrideKey]string)
	s.overrideService = NewExperimentOverrideService(&MapOverridesStore{
		overridesMap: s.overrides,
	})
}

func (s *ExperimentOverrideServiceTestSuite) TestOverridesIncludeVariation() {
	testDecisionContext := ExperimentDecisionContext{
		Experiment:    &testExp1111,
		ProjectConfig: s.mockConfig,
	}
	testUserContext := entities.UserContext{
		ID: "test_user_1",
	}
	s.overrides[ExperimentOverrideKey{ExperimentKey: testExp1111.Key, UserID: "test_user_1"}] = testExp1111Var2222.Key
	decision, err := s.overrideService.GetDecision(testDecisionContext, testUserContext)
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
	decision, err := s.overrideService.GetDecision(testDecisionContext, testUserContext)
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
	s.overrides[ExperimentOverrideKey{ExperimentKey: testExp1113.Key, UserID: "test_user_1"}] = testExp1113Var2224.Key
	decision, err := s.overrideService.GetDecision(testDecisionContext, testUserContext)
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
	s.overrides[ExperimentOverrideKey{ExperimentKey: testExp1111.Key, UserID: "test_user_2"}] = testExp1111Var2222.Key
	decision, err := s.overrideService.GetDecision(testDecisionContext, testUserContext)
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
	s.overrides[ExperimentOverrideKey{ExperimentKey: testExp1113.Key, UserID: "test_user_3"}] = testExp1111Var2222.Key
	decision, err := s.overrideService.GetDecision(testDecisionContext, testUserContext)
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
	s.overrides[ExperimentOverrideKey{ExperimentKey: testExp1111.Key, UserID: "test_user_1"}] = "invalid_variation_key"
	decision, err := s.overrideService.GetDecision(testDecisionContext, testUserContext)
	s.NoError(err)
	s.Nil(decision.Variation)
	s.Exactly(reasons.InvalidOverrideVariationAssignment, decision.Reason)
}

func TestExperimentOverridesTestSuite(t *testing.T) {
	suite.Run(t, new(ExperimentOverrideServiceTestSuite))
}
