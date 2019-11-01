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

	"github.com/optimizely/go-sdk/pkg/entities"
	"github.com/stretchr/testify/suite"
)

type ExperimentOverrideServiceTestSuite struct {
	suite.Suite
	mockConfig      *mockProjectConfig
	overrides       map[OverrideKey]string
	overrideService *ExperimentOverrideService
}

func (s *ExperimentOverrideServiceTestSuite) SetupTest() {
	s.mockConfig = new(mockProjectConfig)
	mapOverrides := &mapOverridesStore{
		overridesMap: map[OverrideKey]string{
			OverrideKey{Experiment: testExp1111.Key, UserID: "test_user_1"}: testExp1111Var2222.Key,
		},
	}
	s.overrideService = NewExperimentOverrideService(mapOverrides)
}

func (s *ExperimentOverrideServiceTestSuite) TestOverridesIncludeVariation() {
	testDecisionContext := ExperimentDecisionContext{
		Experiment:    &testExp1111,
		ProjectConfig: s.mockConfig,
	}

	testUserContext := entities.UserContext{
		ID: "test_user_1",
	}

	decision, err := s.overrideService.GetDecision(testDecisionContext, testUserContext)

	s.NoError(err)
	s.NotNil(decision.Variation)
}

func TestExperimentOverridesTestSuite(t *testing.T) {
	suite.Run(t, new(ExperimentOverrideServiceTestSuite))
}
