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

	"github.com/optimizely/go-sdk/optimizely/entities"
	"github.com/stretchr/testify/mock"
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
		"test_user_1": {
			"test_experiment_1111": "2222",
		},
	}
	whitelistService := NewExperimentWhitelistService(whitelist)

	mockExp := entities.Experiment{
		Key: "test_experiment_1111",
		Variations: map[string]entities.Variation{
			"2222": entities.Variation{
				Key: "2222",
			},
		},
	}
	s.mockConfig.On("GetExperimentByKey", mock.Anything).Return(mockExp, nil)

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
	s.Exactly("2222", decision.Variation.Key)
}

func TestExperimentWhitelistTestSuite(t *testing.T) {
	suite.Run(t, new(ExperimentWhitelistServiceTestSuite))
}
