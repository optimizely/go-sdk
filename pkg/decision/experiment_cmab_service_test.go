/****************************************************************************
 * Copyright 2025, Optimizely, Inc. and contributors                        *
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

	"github.com/optimizely/go-sdk/v2/pkg/config"
	"github.com/optimizely/go-sdk/v2/pkg/decide"
	"github.com/optimizely/go-sdk/v2/pkg/decision/reasons"
	"github.com/optimizely/go-sdk/v2/pkg/entities"
	"github.com/optimizely/go-sdk/v2/pkg/logging"
)

type ExperimentCmabTestSuite struct {
	suite.Suite
	mockCmabService   *MockCmabService
	mockProjectConfig *mockProjectConfig
	testUserContext   entities.UserContext
	options           *decide.Options
	logger            logging.OptimizelyLogProducer
	cmabExperiment    entities.Experiment
	nonCmabExperiment entities.Experiment
}

func (s *ExperimentCmabTestSuite) SetupTest() {
	s.mockCmabService = new(MockCmabService)
	s.mockProjectConfig = new(mockProjectConfig)
	s.logger = logging.GetLogger("test_sdk_key", "ExperimentCmabService")
	s.options = &decide.Options{
		IncludeReasons: true, // Enable reasons
	}

	// Setup test user context
	s.testUserContext = entities.UserContext{
		ID: "test_user_1",
		Attributes: map[string]interface{}{
			"attr1": "value1",
		},
	}

	// Setup CMAB experiment
	s.cmabExperiment = entities.Experiment{
		ID:  "cmab_exp_1",
		Key: "cmab_experiment",
		Cmab: &entities.Cmab{
			AttributeIds: []string{"attr1", "attr2"},
		},
		Variations: map[string]entities.Variation{
			"var1": {
				ID:  "var1",
				Key: "variation_1",
			},
			"var2": {
				ID:  "var2",
				Key: "variation_2",
			},
		},
	}

	// Setup non-CMAB experiment
	s.nonCmabExperiment = entities.Experiment{
		ID:  "non_cmab_exp_1",
		Key: "non_cmab_experiment",
		Variations: map[string]entities.Variation{
			"var1": {
				ID:  "var1",
				Key: "variation_1",
			},
			"var2": {
				ID:  "var2",
				Key: "variation_2",
			},
		},
	}
}

func (s *ExperimentCmabTestSuite) TestIsCmab() {
	// Test with CMAB experiment
	s.True(isCmab(s.cmabExperiment))

	// Test with non-CMAB experiment
	s.False(isCmab(s.nonCmabExperiment))
}

func (s *ExperimentCmabTestSuite) TestGetDecisionWithNilExperiment() {
	// Create decision context with nil experiment
	decisionContext := ExperimentDecisionContext{
		Experiment:    nil,
		ProjectConfig: s.mockProjectConfig,
	}

	// Create CMAB service
	cmabService := NewExperimentCmabService(s.mockCmabService, s.logger)

	// Get decision
	decision, decisionReasons, err := cmabService.GetDecision(decisionContext, s.testUserContext, s.options)

	// Verify results
	s.Nil(decision.Variation)
	s.NoError(err)

	// Check for the message in the reasons
	report := decisionReasons.ToReport()
	s.NotEmpty(report, "Decision reasons report should not be empty")
	found := false
	for _, msg := range report {
		if msg == "Not a CMAB experiment, skipping CMAB decision service" {
			found = true
			break
		}
	}
	s.True(found, "Expected message not found in decision reasons")

	// Verify mock expectations
	s.mockCmabService.AssertNotCalled(s.T(), "GetDecision")
}

func (s *ExperimentCmabTestSuite) TestGetDecisionWithNonCmabExperiment() {
	// Create decision context with non-CMAB experiment
	decisionContext := ExperimentDecisionContext{
		Experiment:    &s.nonCmabExperiment,
		ProjectConfig: s.mockProjectConfig,
	}

	// Create CMAB service
	cmabService := NewExperimentCmabService(s.mockCmabService, s.logger)

	// Get decision
	decision, decisionReasons, err := cmabService.GetDecision(decisionContext, s.testUserContext, s.options)

	// Verify results
	s.Nil(decision.Variation)
	s.NoError(err)

	// Check for the message in the reasons
	report := decisionReasons.ToReport()
	s.NotEmpty(report, "Decision reasons report should not be empty")
	found := false
	for _, msg := range report {
		if msg == "Not a CMAB experiment, skipping CMAB decision service" {
			found = true
			break
		}
	}
	s.True(found, "Expected message not found in decision reasons")

	// Verify mock expectations
	s.mockCmabService.AssertNotCalled(s.T(), "GetDecision")
}

func (s *ExperimentCmabTestSuite) TestGetDecisionWithNilCmabService() {
	// Create decision context with CMAB experiment
	decisionContext := ExperimentDecisionContext{
		Experiment:    &s.cmabExperiment,
		ProjectConfig: s.mockProjectConfig,
	}

	// Create CMAB service with nil CMAB service
	cmabService := NewExperimentCmabService(nil, s.logger)

	// Get decision
	decision, decisionReasons, err := cmabService.GetDecision(decisionContext, s.testUserContext, s.options)

	// Verify results
	s.Nil(decision.Variation)
	s.Error(err)
	s.Equal("CMAB service is not available", err.Error())

	// Check for the message in the reasons
	report := decisionReasons.ToReport()
	s.NotEmpty(report, "Decision reasons report should not be empty")
	found := false
	for _, msg := range report {
		if msg == "CMAB service is not available" {
			found = true
			break
		}
	}
	s.True(found, "Expected message not found in decision reasons")
}

func (s *ExperimentCmabTestSuite) TestGetDecisionWithCmabServiceError() {
	// Create decision context with CMAB experiment
	decisionContext := ExperimentDecisionContext{
		Experiment:    &s.cmabExperiment,
		ProjectConfig: s.mockProjectConfig,
	}

	// Setup mock CMAB service to return error
	s.mockCmabService.On("GetDecision", s.mockProjectConfig, s.testUserContext, "cmab_exp_1", s.options).
		Return(CmabDecision{}, errors.New("CMAB service error"))

	// Create CMAB service
	cmabService := NewExperimentCmabService(s.mockCmabService, s.logger)

	// Get decision
	decision, decisionReasons, err := cmabService.GetDecision(decisionContext, s.testUserContext, s.options)

	// Verify results
	s.Nil(decision.Variation)
	s.Error(err)
	s.Equal("failed to get CMAB decision: CMAB service error", err.Error())

	// Check for the message in the reasons
	report := decisionReasons.ToReport()
	s.NotEmpty(report, "Decision reasons report should not be empty")
	found := false
	for _, msg := range report {
		if msg == "Failed to get CMAB decision: CMAB service error" {
			found = true
			break
		}
	}
	s.True(found, "Expected message not found in decision reasons")

	// Verify mock expectations
	s.mockCmabService.AssertExpectations(s.T())
}

func (s *ExperimentCmabTestSuite) TestGetDecisionWithInvalidVariationID() {
	// Create decision context with CMAB experiment
	decisionContext := ExperimentDecisionContext{
		Experiment:    &s.cmabExperiment,
		ProjectConfig: s.mockProjectConfig,
	}

	// Setup mock CMAB service to return invalid variation ID
	s.mockCmabService.On("GetDecision", s.mockProjectConfig, s.testUserContext, "cmab_exp_1", s.options).
		Return(CmabDecision{VariationID: "invalid_var_id"}, nil)

	// Create CMAB service
	cmabService := NewExperimentCmabService(s.mockCmabService, s.logger)

	// Get decision
	decision, decisionReasons, err := cmabService.GetDecision(decisionContext, s.testUserContext, s.options)

	// Verify results
	s.Nil(decision.Variation)
	s.Error(err)
	s.Equal("variation with ID invalid_var_id not found in experiment cmab_exp_1", err.Error())

	// Check for the message in the reasons
	report := decisionReasons.ToReport()
	s.NotEmpty(report, "Decision reasons report should not be empty")
	found := false
	for _, msg := range report {
		if msg == "variation with ID invalid_var_id not found in experiment cmab_exp_1" {
			found = true
			break
		}
	}
	s.True(found, "Expected message not found in decision reasons")

	// Verify mock expectations
	s.mockCmabService.AssertExpectations(s.T())
}

func (s *ExperimentCmabTestSuite) TestGetDecisionSuccess() {
	// Create decision context with CMAB experiment
	decisionContext := ExperimentDecisionContext{
		Experiment:    &s.cmabExperiment,
		ProjectConfig: s.mockProjectConfig,
	}

	// Setup mock CMAB service to return valid variation ID
	s.mockCmabService.On("GetDecision", s.mockProjectConfig, s.testUserContext, "cmab_exp_1", s.options).
		Return(CmabDecision{VariationID: "var1"}, nil)

	// Create CMAB service
	cmabService := NewExperimentCmabService(s.mockCmabService, s.logger)

	// Get decision
	decision, decisionReasons, err := cmabService.GetDecision(decisionContext, s.testUserContext, s.options)

	// Verify results
	s.NotNil(decision.Variation)
	s.Equal("var1", decision.Variation.ID)
	s.Equal("variation_1", decision.Variation.Key)
	s.Equal(reasons.CmabVariationAssigned, decision.Reason)
	s.NoError(err)

	// Check for the message in the reasons
	report := decisionReasons.ToReport()
	s.NotEmpty(report, "Decision reasons report should not be empty")
	found := false
	for _, msg := range report {
		if msg == "User bucketed into variation variation_1 by CMAB service" {
			found = true
			break
		}
	}
	s.True(found, "Expected message not found in decision reasons")

	// Verify mock expectations
	s.mockCmabService.AssertExpectations(s.T())
}

// Mock CMAB service for testing
type MockCmabService struct {
	mock.Mock
}

func (m *MockCmabService) GetDecision(projectConfig config.ProjectConfig, userContext entities.UserContext, ruleID string, options *decide.Options) (CmabDecision, error) {
	args := m.Called(projectConfig, userContext, ruleID, options)
	return args.Get(0).(CmabDecision), args.Error(1)
}

func TestExperimentCmabTestSuite(t *testing.T) {
	suite.Run(t, new(ExperimentCmabTestSuite))
}
