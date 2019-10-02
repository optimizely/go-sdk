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
	"errors"
	"testing"

	"github.com/optimizely/go-sdk/optimizely/decision/reasons"
	"github.com/optimizely/go-sdk/optimizely/entities"
	"github.com/stretchr/testify/suite"
)

type ExperimentWhitelistServiceTestSuite struct {
	suite.Suite
	mockConfig *mockProjectConfig
}

func (s *ExperimentWhitelistServiceTestSuite) SetupTest() {
	s.mockConfig = new(mockProjectConfig)
}

func (s *ExperimentWhitelistServiceTestSuite) TestWhitelistIncludesDecision() {
	whitelist := map[string]map[string]string{
		"test_experiment_1111": {
			"test_user_1": "2222",
		},
	}
	whitelistService := NewExperimentWhitelistService(whitelist)

	s.mockConfig.On("GetExperimentByKey", "test_experiment_1111").Return(testExp1111, nil)

	testDecisionContext := ExperimentDecisionContext{
		Experiment:    &testExp1111,
		ProjectConfig: s.mockConfig,
	}

	testUserContext := entities.UserContext{
		ID: "test_user_1",
	}

	decision, err := whitelistService.GetDecision(testDecisionContext, testUserContext)

	s.NoError(err)
	s.NotNil(decision.Variation)
}

func (s *ExperimentWhitelistServiceTestSuite) TestNoUserEntryInWhitelist() {
	whitelist := map[string]map[string]string{
		"test_experiment_1111": {
			"test_user_2": "2222",
		},
	}
	whitelistService := NewExperimentWhitelistService(whitelist)

	s.mockConfig.On("GetExperimentByKey", "test_experiment_1111").Return(testExp1111, nil)

	testDecisionContext := ExperimentDecisionContext{
		Experiment:    &testExp1111,
		ProjectConfig: s.mockConfig,
	}

	// user context has test_user_1, but there's only a whitelist entry for test_user_2
	testUserContext := entities.UserContext{
		ID: "test_user_1",
	}

	decision, err := whitelistService.GetDecision(testDecisionContext, testUserContext)

	s.NoError(err)
	s.Nil(decision.Variation)
	s.Exactly(decision.Reason, reasons.NoWhitelistVariationAssignment)
}

func (s *ExperimentWhitelistServiceTestSuite) TestNoVariationInUserEntry() {
	whitelist := map[string]map[string]string{
		"test_experiment_1113": {
			"test_user_1": "2223",
		},
	}
	whitelistService := NewExperimentWhitelistService(whitelist)

	s.mockConfig.On("GetExperimentByKey", "test_experiment_1111").Return(testExp1111, nil)

	// decision context has testExp1111, but there's only a whitelist entry for testExp1113
	testDecisionContext := ExperimentDecisionContext{
		Experiment:    &testExp1111,
		ProjectConfig: s.mockConfig,
	}

	testUserContext := entities.UserContext{
		ID: "test_user_1",
	}

	decision, err := whitelistService.GetDecision(testDecisionContext, testUserContext)

	s.NoError(err)
	s.Nil(decision.Variation)
	s.Exactly(decision.Reason, reasons.NoWhitelistVariationAssignment)
}

func (s *ExperimentWhitelistServiceTestSuite) TestInvalidVariationInUserEntry() {
	whitelist := map[string]map[string]string{
		"test_experiment_1111": {
			// Whitelist has assigned 222222222 for this user, but no such variation exists
			"test_user_1": "222222222",
		},
	}
	whitelistService := NewExperimentWhitelistService(whitelist)

	s.mockConfig.On("GetExperimentByKey", "test_experiment_1111").Return(testExp1111, nil)

	testDecisionContext := ExperimentDecisionContext{
		Experiment:    &testExp1111,
		ProjectConfig: s.mockConfig,
	}

	testUserContext := entities.UserContext{
		ID: "test_user_1",
	}

	decision, err := whitelistService.GetDecision(testDecisionContext, testUserContext)

	s.NoError(err)
	s.Nil(decision.Variation)
	s.Exactly(decision.Reason, reasons.InvalidWhitelistVariationAssignment)
}

func (s *ExperimentWhitelistServiceTestSuite) TestEmptyWhitelist() {
	whitelist := map[string]map[string]string{}
	whitelistService := NewExperimentWhitelistService(whitelist)

	s.mockConfig.On("GetExperimentByKey", "test_experiment_1111").Return(testExp1111, nil)

	testDecisionContext := ExperimentDecisionContext{
		Experiment:    &testExp1111,
		ProjectConfig: s.mockConfig,
	}

	testUserContext := entities.UserContext{
		ID: "test_user_1",
	}

	decision, err := whitelistService.GetDecision(testDecisionContext, testUserContext)

	s.NoError(err)
	s.Nil(decision.Variation)
	s.Exactly(decision.Reason, reasons.NoWhitelistVariationAssignment)
}

func (s *ExperimentWhitelistServiceTestSuite) TestNoExperimentInDecisionContext() {
	whitelist := map[string]map[string]string{
		"test_experiment_1111": {
			"test_user_1": "2222",
		},
	}
	whitelistService := NewExperimentWhitelistService(whitelist)

	s.mockConfig.On("GetExperimentByKey", "test_experiment_1111").Return(testExp1111, nil)

	testDecisionContext := ExperimentDecisionContext{
		Experiment:    nil,
		ProjectConfig: s.mockConfig,
	}

	testUserContext := entities.UserContext{
		ID: "test_user_1",
	}

	decision, err := whitelistService.GetDecision(testDecisionContext, testUserContext)

	s.Error(err)
	s.Nil(decision.Variation)
}

func (s *ExperimentWhitelistServiceTestSuite) TestNoExperimentInProjectConfig() {
	whitelist := map[string]map[string]string{
		"test_experiment_1111": {
			"test_user_1": "2222",
		},
	}
	whitelistService := NewExperimentWhitelistService(whitelist)

	s.mockConfig.On("GetExperimentByKey", "test_experiment_1111").Return(entities.Experiment{}, errors.New("Experiment not found"))

	testDecisionContext := ExperimentDecisionContext{
		Experiment:    &testExp1111,
		ProjectConfig: s.mockConfig,
	}

	testUserContext := entities.UserContext{
		ID: "test_user_1",
	}

	decision, err := whitelistService.GetDecision(testDecisionContext, testUserContext)

	s.Error(err)
	s.Nil(decision.Variation)
}

func TestExperimentWhitelistTestSuite(t *testing.T) {
	suite.Run(t, new(ExperimentWhitelistServiceTestSuite))
}
