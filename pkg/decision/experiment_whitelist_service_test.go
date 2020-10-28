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

	"github.com/optimizely/go-sdk/pkg/decide"
	"github.com/optimizely/go-sdk/pkg/decision/reasons"
	"github.com/optimizely/go-sdk/pkg/entities"
	"github.com/stretchr/testify/suite"
)

type ExperimentWhitelistServiceTestSuite struct {
	suite.Suite
	mockConfig       *mockProjectConfig
	whitelistService *ExperimentWhitelistService
	options          decide.OptimizelyDecideOptions
	reasons          *decide.DecisionReasons
}

func (s *ExperimentWhitelistServiceTestSuite) SetupTest() {
	s.mockConfig = new(mockProjectConfig)
	s.whitelistService = NewExperimentWhitelistService()
	s.options = decide.OptimizelyDecideOptions{}
	s.reasons = decide.NewDecisionReasons(s.options)
}

func (s *ExperimentWhitelistServiceTestSuite) TestWhitelistIncludesDecision() {
	testDecisionContext := ExperimentDecisionContext{
		Experiment:    &testExpWhitelist,
		ProjectConfig: s.mockConfig,
	}

	testUserContext := entities.UserContext{
		ID: "test_user_1",
	}

	decision, err := s.whitelistService.GetDecision(testDecisionContext, testUserContext, s.options, s.reasons)

	s.NoError(err)
	s.NotNil(decision.Variation)
}

func (s *ExperimentWhitelistServiceTestSuite) TestNoUserEntryInWhitelist() {
	testDecisionContext := ExperimentDecisionContext{
		Experiment:    &testExpWhitelist,
		ProjectConfig: s.mockConfig,
	}

	// user context has test_user_3, but there's only a whitelist entry for test_user_1 and test_user_2
	testUserContext := entities.UserContext{
		ID: "test_user_3",
	}

	decision, err := s.whitelistService.GetDecision(testDecisionContext, testUserContext, s.options, s.reasons)

	s.NoError(err)
	s.Nil(decision.Variation)
	s.Exactly(decision.Reason, reasons.NoWhitelistVariationAssignment)
}

func (s *ExperimentWhitelistServiceTestSuite) TestEmptyWhitelist() {
	testDecisionContext := ExperimentDecisionContext{
		// testExp1111 has no whitelist
		Experiment:    &testExp1111,
		ProjectConfig: s.mockConfig,
	}

	testUserContext := entities.UserContext{
		ID: "test_user_1",
	}

	decision, err := s.whitelistService.GetDecision(testDecisionContext, testUserContext, s.options, s.reasons)

	s.NoError(err)
	s.Nil(decision.Variation)
	s.Exactly(decision.Reason, reasons.NoWhitelistVariationAssignment)
}

func (s *ExperimentWhitelistServiceTestSuite) TestInvalidVariationInUserEntry() {
	testDecisionContext := ExperimentDecisionContext{
		Experiment:    &testExpWhitelist,
		ProjectConfig: s.mockConfig,
	}

	testUserContext := entities.UserContext{
		// In the whitelist, test_user_2 is mapped to an invalid variation key (no variation with that key exists in the experiment)
		ID: "test_user_2",
	}

	decision, err := s.whitelistService.GetDecision(testDecisionContext, testUserContext, s.options, s.reasons)

	s.NoError(err)
	s.Nil(decision.Variation)
	s.Exactly(decision.Reason, reasons.InvalidWhitelistVariationAssignment)
}

func (s *ExperimentWhitelistServiceTestSuite) TestNoExperimentInDecisionContext() {
	testDecisionContext := ExperimentDecisionContext{
		Experiment:    nil,
		ProjectConfig: s.mockConfig,
	}

	testUserContext := entities.UserContext{
		ID: "test_user_1",
	}

	decision, err := s.whitelistService.GetDecision(testDecisionContext, testUserContext, s.options, s.reasons)

	s.Error(err)
	s.Nil(decision.Variation)
}

func TestExperimentWhitelistTestSuite(t *testing.T) {
	suite.Run(t, new(ExperimentWhitelistServiceTestSuite))
}
