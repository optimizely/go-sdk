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
	"strings"
	"testing"

	"github.com/optimizely/go-sdk/v2/pkg/cmab"
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
	mockCmabService       *MockCmabService
	mockProjectConfig     *mockProjectConfig
	experimentCmabService *ExperimentCmabService
	testUserContext       entities.UserContext
	options               *decide.Options
	logger                logging.OptimizelyLogProducer
	cmabExperiment        entities.Experiment
	nonCmabExperiment     entities.Experiment
}

func (s *ExperimentCmabTestSuite) SetupTest() {
	s.mockCmabService = new(MockCmabService)
	s.mockProjectConfig = new(mockProjectConfig)
	s.logger = logging.GetLogger("test_sdk_key", "ExperimentCmabService")
	s.options = &decide.Options{
		IncludeReasons: true,
	}

	s.experimentCmabService = NewExperimentCmabService(s.mockCmabService, s.logger)

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
	// Test that nil experiment returns empty decision with appropriate reason
	testUserContext := entities.UserContext{
		ID: "test_user_1",
	}

	testDecisionContext := ExperimentDecisionContext{
		Experiment:    nil,
		ProjectConfig: s.mockProjectConfig,
	}

	// Create options with reasons enabled
	options := &decide.Options{
		IncludeReasons: true,
	}

	decision, reasons, err := s.experimentCmabService.GetDecision(testDecisionContext, testUserContext, options)
	s.NoError(err)
	s.Equal(ExperimentDecision{}, decision)

	// Check that reasons are populated
	s.NotEmpty(reasons.ToReport())

	// Check for specific reason message
	reasonsReport := reasons.ToReport()
	expectedMessage := "experiment is nil"
	found := false
	for _, msg := range reasonsReport {
		if strings.Contains(msg, expectedMessage) {
			found = true
			break
		}
	}
	s.True(found, "Expected message not found in decision reasons")
}

func (s *ExperimentCmabTestSuite) TestGetDecisionWithNonCmabExperiment() {
	// Test that non-CMAB experiment returns empty decision
	testDecisionContext := ExperimentDecisionContext{
		Experiment:    &s.nonCmabExperiment,
		ProjectConfig: s.mockProjectConfig,
	}

	decision, _, err := s.experimentCmabService.GetDecision(testDecisionContext, s.testUserContext, s.options)
	s.NoError(err)
	s.Equal(ExperimentDecision{}, decision)

	// Since we're not adding reasons for non-CMAB experiments to avoid breaking other tests,
	// we'll just check that the decision is empty and there's no error
	s.Empty(decision)
	s.NoError(err)
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
		Return(cmab.Decision{}, errors.New("CMAB service error"))

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
		Return(cmab.Decision{VariationID: "invalid_var_id"}, nil)

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
		Return(cmab.Decision{VariationID: "var1"}, nil)

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

func (m *MockCmabService) GetDecision(projectConfig config.ProjectConfig, userContext entities.UserContext, ruleID string, options *decide.Options) (cmab.Decision, error) {
	args := m.Called(projectConfig, userContext, ruleID, options)
	return args.Get(0).(cmab.Decision), args.Error(1)
}

func TestExperimentCmabTestSuite(t *testing.T) {
	suite.Run(t, new(ExperimentCmabTestSuite))
}
