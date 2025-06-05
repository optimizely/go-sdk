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
	"github.com/optimizely/go-sdk/v2/pkg/decision/evaluator"
	"github.com/optimizely/go-sdk/v2/pkg/decision/reasons"
	"github.com/optimizely/go-sdk/v2/pkg/entities"
	"github.com/optimizely/go-sdk/v2/pkg/logging"
)

// Mock types - MUST be at package level, not inside functions
type MockCmabService struct {
	mock.Mock
}

func (m *MockCmabService) GetDecision(projectConfig config.ProjectConfig, userContext entities.UserContext, ruleID string, options *decide.Options) (cmab.Decision, error) {
	args := m.Called(projectConfig, userContext, ruleID, options)
	return args.Get(0).(cmab.Decision), args.Error(1)
}

type MockExperimentBucketer struct {
	mock.Mock
}

func (m *MockExperimentBucketer) Bucket(bucketingID string, experiment entities.Experiment, group entities.Group) (*entities.Variation, reasons.Reason, error) {
	args := m.Called(bucketingID, experiment, group)

	var variation *entities.Variation
	if args.Get(0) != nil {
		variation = args.Get(0).(*entities.Variation)
	}

	return variation, args.Get(1).(reasons.Reason), args.Error(2)
}

// Add the new method to satisfy the ExperimentBucketer interface
func (m *MockExperimentBucketer) BucketToEntityID(bucketingID string, experiment entities.Experiment, group entities.Group) (string, reasons.Reason, error) {
	args := m.Called(bucketingID, experiment, group)
	return args.String(0), args.Get(1).(reasons.Reason), args.Error(2)
}

type ExperimentCmabTestSuite struct {
	suite.Suite
	mockCmabService        *MockCmabService
	mockProjectConfig      *mockProjectConfig
	mockExperimentBucketer *MockExperimentBucketer
	experimentCmabService  *ExperimentCmabService
	testUserContext        entities.UserContext
	options                *decide.Options
	logger                 logging.OptimizelyLogProducer
	cmabExperiment         entities.Experiment
	nonCmabExperiment      entities.Experiment
}

func (s *ExperimentCmabTestSuite) SetupTest() {
	s.mockCmabService = new(MockCmabService)
	s.mockExperimentBucketer = new(MockExperimentBucketer)
	s.mockProjectConfig = new(mockProjectConfig)
	s.logger = logging.GetLogger("test_sdk_key", "ExperimentCmabService")
	s.options = &decide.Options{
		IncludeReasons: true,
	}

	// Create service with real dependencies first
	s.experimentCmabService = NewExperimentCmabService("test_sdk_key")

	// inject the mocks
	s.experimentCmabService.bucketer = s.mockExperimentBucketer
	s.experimentCmabService.cmabService = s.mockCmabService

	// Initialize the audience tree evaluator w logger
	s.experimentCmabService.audienceTreeEvaluator = evaluator.NewMixedTreeEvaluator(s.logger)

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

func (s *ExperimentCmabTestSuite) TestGetDecisionSuccess() {
	// Create decision context with CMAB experiment
	decisionContext := ExperimentDecisionContext{
		Experiment:    &s.cmabExperiment,
		ProjectConfig: s.mockProjectConfig,
	}

	// Mock bucketer to return CMAB dummy entity ID (so traffic allocation passes)
	s.mockExperimentBucketer.On("BucketToEntityID", mock.Anything, mock.Anything, mock.Anything).
		Return(CmabDummyEntityID, reasons.BucketedIntoVariation, nil)

	// Setup mock CMAB service (only once!)
	s.mockCmabService.On("GetDecision", s.mockProjectConfig, s.testUserContext, "cmab_exp_1", s.options).
		Return(cmab.Decision{VariationID: "var1"}, nil)

	// Create CMAB service with mocked dependencies
	cmabService := &ExperimentCmabService{
		bucketer:    s.mockExperimentBucketer,
		cmabService: s.mockCmabService,
		logger:      s.logger,
	}

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
	s.mockExperimentBucketer.AssertExpectations(s.T())
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

	// Create CMAB service with mocked dependencies
	cmabService := &ExperimentCmabService{
		bucketer:    s.mockExperimentBucketer,
		cmabService: s.mockCmabService,
		logger:      s.logger,
	}

	decision, decisionReasons, err := cmabService.GetDecision(testDecisionContext, testUserContext, options)

	// Should NOT return an error for nil experiment (based on your implementation)
	s.NoError(err)
	s.Equal(ExperimentDecision{}, decision)

	// Check that reasons are populated
	s.NotEmpty(decisionReasons.ToReport())

	// Check for specific reason message
	reasonsReport := decisionReasons.ToReport()
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

	// Create CMAB service with EXPLICITLY nil CMAB service
	cmabService := &ExperimentCmabService{
		audienceTreeEvaluator: evaluator.NewMixedTreeEvaluator(s.logger),
		bucketer:              s.mockExperimentBucketer,
		cmabService:           nil, // ← Explicitly set to nil
		logger:                s.logger,
	}

	// Get decision
	decision, decisionReasons, err := cmabService.GetDecision(decisionContext, s.testUserContext, s.options)

	// Now it should hit the nil check and return an error
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
	testDecisionContext := ExperimentDecisionContext{
		Experiment:    &s.cmabExperiment, // Use s.cmabExperiment from setup
		ProjectConfig: s.mockProjectConfig,
	}

	// Mock bucketer to return CMAB dummy entity ID (traffic allocation passes)
	s.mockExperimentBucketer.On("BucketToEntityID", "test_user_1", mock.AnythingOfType("entities.Experiment"), entities.Group{}).
		Return(CmabDummyEntityID, reasons.BucketedIntoVariation, nil)

	// Mock CMAB service to return error
	s.mockCmabService.On("GetDecision", s.mockProjectConfig, s.testUserContext, "cmab_exp_1", s.options).
		Return(cmab.Decision{}, errors.New("CMAB service error"))

	// Create CMAB service with mocked dependencies (same pattern as TestGetDecisionSuccess)
	cmabService := &ExperimentCmabService{
		bucketer:    s.mockExperimentBucketer,
		cmabService: s.mockCmabService,
		logger:      s.logger,
	}

	decision, _, err := cmabService.GetDecision(testDecisionContext, s.testUserContext, s.options)

	// Should return the CMAB service error
	s.Error(err)
	s.Contains(err.Error(), "CMAB service error")
	s.Nil(decision.Variation) // No variation when error occurs

	s.mockExperimentBucketer.AssertExpectations(s.T())
	s.mockCmabService.AssertExpectations(s.T())
}

func (s *ExperimentCmabTestSuite) TestGetDecisionWithInvalidVariationID() {
	testDecisionContext := ExperimentDecisionContext{
		Experiment:    &s.cmabExperiment, // Use s.cmabExperiment from setup
		ProjectConfig: s.mockProjectConfig,
	}

	// Mock bucketer to return CMAB dummy entity ID (traffic allocation passes)
	s.mockExperimentBucketer.On("BucketToEntityID", "test_user_1", mock.AnythingOfType("entities.Experiment"), entities.Group{}).
		Return(CmabDummyEntityID, reasons.BucketedIntoVariation, nil)

	// Mock CMAB service to return invalid variation ID
	invalidCmabDecision := cmab.Decision{
		VariationID: "invalid_variation_id",
		CmabUUID:    "test-uuid-123",
	}
	s.mockCmabService.On("GetDecision", s.mockProjectConfig, s.testUserContext, "cmab_exp_1", s.options).
		Return(invalidCmabDecision, nil)

	// Create CMAB service with mocked dependencies (same pattern as TestGetDecisionSuccess)
	cmabService := &ExperimentCmabService{
		bucketer:    s.mockExperimentBucketer,
		cmabService: s.mockCmabService,
		logger:      s.logger,
	}

	decision, _, err := cmabService.GetDecision(testDecisionContext, s.testUserContext, s.options)

	// Should return error for invalid variation ID
	s.Error(err)
	s.Contains(err.Error(), "variation with ID invalid_variation_id not found in experiment cmab_exp_1")
	s.Nil(decision.Variation) // No variation when error occurs

	s.mockExperimentBucketer.AssertExpectations(s.T())
	s.mockCmabService.AssertExpectations(s.T())
}

func (s *ExperimentCmabTestSuite) TestGetDecisionCmabExperimentUserNotBucketed() {
	testDecisionContext := ExperimentDecisionContext{
		Experiment:    &s.cmabExperiment,
		ProjectConfig: s.mockProjectConfig,
	}

	// Mock bucketer - expect the MODIFIED experiment with traffic allocation
	s.mockExperimentBucketer.On("BucketToEntityID",
		s.testUserContext.ID, // User ID
		mock.MatchedBy(func(exp entities.Experiment) bool {
			// Check that it's our experiment with the modified traffic allocation
			return exp.ID == "cmab_exp_1" &&
				exp.Key == "cmab_experiment" &&
				len(exp.TrafficAllocation) == 1 &&
				exp.TrafficAllocation[0].EntityID == CmabDummyEntityID
		}),
		entities.Group{}, // Empty group
	).Return("different_entity_id", reasons.NotBucketedIntoVariation, nil) // Return something != CmabDummyEntityID

	decision, _, err := s.experimentCmabService.GetDecision(testDecisionContext, s.testUserContext, s.options)

	// Rest of your assertions...
	s.NoError(err)
	s.Equal(reasons.NotBucketedIntoVariation, decision.Reason)
	s.Nil(decision.Variation)
	s.Nil(decision.CmabUUID)

	s.mockExperimentBucketer.AssertExpectations(s.T())
}

func (s *ExperimentCmabTestSuite) TestGetDecisionCmabExperimentAudienceConditionNotMet() {
	// Create experiment with audience that will actually fail
	cmabExperimentWithAudience := entities.Experiment{
		ID:  "cmab_exp_with_audience",
		Key: "cmab_experiment_with_audience",
		Cmab: &entities.Cmab{
			AttributeIds:      []string{"attr1", "attr2"},
			TrafficAllocation: 10000,
		},
		AudienceIds: []string{"audience_1"},
		// CORRECT AudienceConditionTree structure:
		AudienceConditionTree: &entities.TreeNode{
			Operator: "or",
			Nodes: []*entities.TreeNode{
				{
					Item: "audience_1", // Reference the audience ID
				},
			},
		},
		Variations: map[string]entities.Variation{
			"var1": {ID: "var1", Key: "variation_1"},
		},
		TrafficAllocation: []entities.Range{
			{EntityID: "$", EndOfRange: 10000},
		},
	}

	// User that will NOT match the audience
	userContextNoAudience := entities.UserContext{
		ID: "test_user_no_audience",
		Attributes: map[string]interface{}{
			"country": "US", // This won't match our audience condition
		},
	}

	// Create audience with condition tree that requires Canada
	audienceMap := map[string]entities.Audience{
		"audience_1": {
			ID:   "audience_1",
			Name: "Test Audience",
			ConditionTree: &entities.TreeNode{
				Operator: "or",
				Nodes: []*entities.TreeNode{
					{
						Item: entities.Condition{
							Type:  "custom_attribute",
							Match: "exact",
							Name:  "country",
							Value: "Canada",
						},
					},
				},
			},
		},
	}
	s.mockProjectConfig.On("GetAudienceMap").Return(audienceMap)

	// This mock should NOT be called if audience fails
	s.mockExperimentBucketer.On("BucketToEntityID", mock.Anything, mock.Anything, mock.Anything).Return("", reasons.NotBucketedIntoVariation, nil).Maybe()

	testDecisionContext := ExperimentDecisionContext{
		Experiment:    &cmabExperimentWithAudience,
		ProjectConfig: s.mockProjectConfig,
	}

	decision, _, err := s.experimentCmabService.GetDecision(testDecisionContext, userContextNoAudience, s.options)

	s.NoError(err)
	s.Equal(reasons.FailedAudienceTargeting, decision.Reason)
	s.Nil(decision.Variation)

	s.mockProjectConfig.AssertExpectations(s.T())
}

func TestExperimentCmabTestSuite(t *testing.T) {
	suite.Run(t, new(ExperimentCmabTestSuite))
}
