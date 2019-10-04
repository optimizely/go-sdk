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

package decision

import (
	"testing"

	"github.com/optimizely/go-sdk/pkg/decision/reasons"

	"github.com/optimizely/go-sdk/pkg/entities"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type MockBucketer struct {
	mock.Mock
}

func (m *MockBucketer) Bucket(bucketingID string, experiment entities.Experiment, group entities.Group) (*entities.Variation, reasons.Reason, error) {
	args := m.Called(bucketingID, experiment, group)
	return args.Get(0).(*entities.Variation), args.Get(1).(reasons.Reason), args.Error(2)
}

type ExperimentBucketerTestSuite struct {
	suite.Suite
	mockBucketer *MockBucketer
	mockConfig   *mockProjectConfig
}

func (s *ExperimentBucketerTestSuite) SetupTest() {
	s.mockBucketer = new(MockBucketer)
	s.mockConfig = new(mockProjectConfig)
}

func (s *ExperimentBucketerTestSuite) TestGetDecisionNoTargeting() {
	testUserContext := entities.UserContext{
		ID: "test_user_1",
	}

	expectedDecision := ExperimentDecision{
		Variation: &testExp1111Var2222,
		Decision: Decision{
			Reason: reasons.BucketedIntoVariation,
		},
	}

	testDecisionContext := ExperimentDecisionContext{
		Experiment:    &testExp1111,
		ProjectConfig: s.mockConfig,
	}
	s.mockBucketer.On("Bucket", testUserContext.ID, testExp1111, entities.Group{}).Return(&testExp1111Var2222, reasons.BucketedIntoVariation, nil)

	experimentBucketerService := ExperimentBucketerService{
		bucketer: s.mockBucketer,
	}
	decision, err := experimentBucketerService.GetDecision(testDecisionContext, testUserContext)
	s.Equal(expectedDecision, decision)
	s.NoError(err)
}

func (s *ExperimentBucketerTestSuite) TestGetDecisionWithTargetingPasses() {
	testUserContext := entities.UserContext{
		ID: "test_user_1",
	}

	expectedDecision := ExperimentDecision{
		Variation: &testTargetedExp1116Var2228,
		Decision: Decision{
			Reason: reasons.BucketedIntoVariation,
		},
	}
	s.mockBucketer.On("Bucket", testUserContext.ID, testTargetedExp1116, entities.Group{}).Return(&testTargetedExp1116Var2228, reasons.BucketedIntoVariation, nil)

	mockAudienceTreeEvaluator := new(MockAudienceTreeEvaluator)
	mockAudienceTreeEvaluator.On("Evaluate", mock.Anything, mock.Anything).Return(true)
	experimentBucketerService := ExperimentBucketerService{
		audienceTreeEvaluator: mockAudienceTreeEvaluator,
		bucketer:              s.mockBucketer,
	}
	s.mockConfig.On("GetAudienceMap").Return(map[string]entities.Audience{})

	testDecisionContext := ExperimentDecisionContext{
		Experiment:    &testTargetedExp1116,
		ProjectConfig: s.mockConfig,
	}
	decision, err := experimentBucketerService.GetDecision(testDecisionContext, testUserContext)
	s.Equal(expectedDecision, decision)
	s.NoError(err)
}

func (s *ExperimentBucketerTestSuite) TestGetDecisionWithTargetingFails() {
	testUserContext := entities.UserContext{
		ID: "test_user_1",
	}

	expectedDecision := ExperimentDecision{
		Decision: Decision{
			Reason: reasons.FailedAudienceTargeting,
		},
	}
	mockAudienceTreeEvaluator := new(MockAudienceTreeEvaluator)
	mockAudienceTreeEvaluator.On("Evaluate", mock.Anything, mock.Anything).Return(false)
	experimentBucketerService := ExperimentBucketerService{
		audienceTreeEvaluator: mockAudienceTreeEvaluator,
		bucketer:              s.mockBucketer,
	}
	s.mockConfig.On("GetAudienceMap").Return(map[string]entities.Audience{})

	testDecisionContext := ExperimentDecisionContext{
		Experiment:    &testTargetedExp1116,
		ProjectConfig: s.mockConfig,
	}
	decision, err := experimentBucketerService.GetDecision(testDecisionContext, testUserContext)
	s.Equal(expectedDecision, decision)
	s.NoError(err)
	s.mockBucketer.AssertNotCalled(s.T(), "Bucket")
}

func TestExperimentBucketerTestSuite(t *testing.T) {
	suite.Run(t, new(ExperimentBucketerTestSuite))
}
