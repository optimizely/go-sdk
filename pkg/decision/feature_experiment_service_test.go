/****************************************************************************
 * Copyright 2019-2025, Optimizely, Inc. and contributors                   *
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

	"github.com/optimizely/go-sdk/v2/pkg/decide"
	"github.com/optimizely/go-sdk/v2/pkg/entities"
	"github.com/optimizely/go-sdk/v2/pkg/logging"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type FeatureExperimentServiceTestSuite struct {
	suite.Suite
	mockConfig                 *mockProjectConfig
	testFeatureDecisionContext FeatureDecisionContext
	mockExperimentService      *MockExperimentDecisionService
	options                    *decide.Options
	reasons                    decide.DecisionReasons
}

func (s *FeatureExperimentServiceTestSuite) SetupTest() {
	s.mockConfig = new(mockProjectConfig)
	s.testFeatureDecisionContext = FeatureDecisionContext{
		Feature:               &testFeat3335,
		ProjectConfig:         s.mockConfig,
		Variable:              testVariable,
		ForcedDecisionService: NewForcedDecisionService("test_user"),
	}
	s.mockExperimentService = new(MockExperimentDecisionService)
	s.options = &decide.Options{}
	s.reasons = decide.NewDecisionReasons(s.options)
}

func (s *FeatureExperimentServiceTestSuite) TestGetDecision() {
	testUserContext := entities.UserContext{
		ID: "test_user_1",
	}

	expectedVariation := testExp1113.Variations["2223"]
	returnExperimentDecision := ExperimentDecision{
		Variation: &expectedVariation,
	}
	testExperimentDecisionContext := ExperimentDecisionContext{
		Experiment:    &testExp1113,
		ProjectConfig: s.mockConfig,
	}
	s.mockExperimentService.On("GetDecision", testExperimentDecisionContext, testUserContext, s.options).Return(returnExperimentDecision, s.reasons, nil)

	featureExperimentService := &FeatureExperimentService{
		compositeExperimentService: s.mockExperimentService,
		logger:                     logging.GetLogger("sdkKey", "FeatureExperimentService"),
	}

	expectedFeatureDecision := FeatureDecision{
		Experiment: *testExperimentDecisionContext.Experiment,
		Variation:  &expectedVariation,
		Source:     FeatureTest,
	}
	decision, _, err := featureExperimentService.GetDecision(s.testFeatureDecisionContext, testUserContext, s.options)
	s.Equal(expectedFeatureDecision, decision)
	s.NoError(err)
	s.mockExperimentService.AssertExpectations(s.T())
}

func (s *FeatureExperimentServiceTestSuite) TestGetDecisionWithForcedDecision() {
	testUserContext := entities.UserContext{
		ID: "test_user_1",
	}

	expectedVariation := testExp1113.Variations["2223"]
	flagVariationsMap := map[string][]entities.Variation{
		s.testFeatureDecisionContext.Feature.Key: {
			expectedVariation,
		},
	}
	s.mockConfig.On("GetFlagVariationsMap").Return(flagVariationsMap)
	s.testFeatureDecisionContext.ForcedDecisionService.SetForcedDecision(OptimizelyDecisionContext{FlagKey: s.testFeatureDecisionContext.Feature.Key, RuleKey: testExp1113Key}, OptimizelyForcedDecision{VariationKey: expectedVariation.Key})

	testExperimentDecisionContext := ExperimentDecisionContext{
		Experiment:    &testExp1113,
		ProjectConfig: s.mockConfig,
	}

	featureExperimentService := &FeatureExperimentService{
		compositeExperimentService: s.mockExperimentService,
		logger:                     logging.GetLogger("sdkKey", "FeatureExperimentService"),
	}

	expectedFeatureDecision := FeatureDecision{
		Experiment: *testExperimentDecisionContext.Experiment,
		Variation:  &expectedVariation,
		Source:     FeatureTest,
	}
	options := &decide.Options{IncludeReasons: true}
	decision, reasons, err := featureExperimentService.GetDecision(s.testFeatureDecisionContext, testUserContext, options)
	s.Equal(expectedFeatureDecision, decision)
	s.Equal(expectedFeatureDecision, decision)
	s.Equal("Variation (2223) is mapped to flag (test_feature_3335_key), rule (test_experiment_1113) and user (test_user) in the forced decision map.", reasons.ToReport()[0])
	s.NoError(err)
	// Makes sure that decision returned was a forcedDecision
	s.mockExperimentService.AssertNotCalled(s.T(), "GetDecision", testExperimentDecisionContext, testUserContext, options)

	// invalid forced decision
	s.testFeatureDecisionContext.ForcedDecisionService.SetForcedDecision(OptimizelyDecisionContext{FlagKey: s.testFeatureDecisionContext.Feature.Key, RuleKey: testExp1113Key}, OptimizelyForcedDecision{VariationKey: "invalid"})

	expectedVariation = testExp1113.Variations["2223"]
	returnExperimentDecision := ExperimentDecision{
		Variation: &expectedVariation,
	}
	s.mockExperimentService.On("GetDecision", testExperimentDecisionContext, testUserContext, options).Return(returnExperimentDecision, s.reasons, nil)
	decision, reasons, err = featureExperimentService.GetDecision(s.testFeatureDecisionContext, testUserContext, options)
	s.Equal(expectedFeatureDecision, decision)
	s.Equal("Invalid variation is mapped to flag (test_feature_3335_key), rule (test_experiment_1113) and user (test_user) in the forced decision map.", reasons.ToReport()[0])
	s.NoError(err)
	s.mockExperimentService.AssertExpectations(s.T())
}

func (s *FeatureExperimentServiceTestSuite) TestGetDecisionMutex() {
	testUserContext := entities.UserContext{
		ID: "test_user_1",
	}

	// first experiment returns nil to simulate user not being bucketed into this experiment in the group
	nilDecision := ExperimentDecision{}
	testExperimentDecisionContext1 := ExperimentDecisionContext{
		Experiment:    &testExp1113,
		ProjectConfig: s.mockConfig,
	}
	s.mockExperimentService.On("GetDecision", testExperimentDecisionContext1, testUserContext, s.options).Return(nilDecision, s.reasons, nil)

	// second experiment returns a valid decision to simulate user being bucketed into this experiment in the group
	expectedVariation := testExp1114.Variations["2225"]
	returnExperimentDecision := ExperimentDecision{
		Variation: &expectedVariation,
	}
	testExperimentDecisionContext2 := ExperimentDecisionContext{
		Experiment:    &testExp1114,
		ProjectConfig: s.mockConfig,
	}
	s.mockExperimentService.On("GetDecision", testExperimentDecisionContext2, testUserContext, s.options).Return(returnExperimentDecision, s.reasons, nil)

	expectedFeatureDecision := FeatureDecision{
		Experiment: *testExperimentDecisionContext2.Experiment,
		Variation:  &expectedVariation,
		Source:     FeatureTest,
	}
	featureExperimentService := &FeatureExperimentService{
		compositeExperimentService: s.mockExperimentService,
		logger:                     logging.GetLogger("sdkKey", "FeatureExperimentService"),
	}
	decision, _, err := featureExperimentService.GetDecision(s.testFeatureDecisionContext, testUserContext, s.options)
	s.Equal(expectedFeatureDecision, decision)
	s.NoError(err)
	s.mockExperimentService.AssertExpectations(s.T())
}

func (s *FeatureExperimentServiceTestSuite) TestNewFeatureExperimentService() {
	compositeExperimentService := &CompositeExperimentService{logger: logging.GetLogger("sdkKey", "CompositeExperimentService")}
	featureExperimentService := NewFeatureExperimentService(logging.GetLogger("", ""), compositeExperimentService)
	s.IsType(compositeExperimentService, featureExperimentService.compositeExperimentService)
}

func (s *FeatureExperimentServiceTestSuite) TestGetDecisionWithCmabUUID() {
	testUserContext := entities.UserContext{
		ID: "test_user_1",
	}

	// Create test UUID
	testUUID := "test-cmab-uuid-12345"

	// Create experiment decision with UUID
	expectedVariation := testExp1113.Variations["2223"]
	returnExperimentDecision := ExperimentDecision{
		Variation: &expectedVariation,
		CmabUUID:  &testUUID,
	}

	// Setup experiment decision context
	testExperimentDecisionContext := ExperimentDecisionContext{
		Experiment:    &testExp1113,
		ProjectConfig: s.mockConfig,
	}

	// Setup mock to return experiment decision with UUID
	s.mockExperimentService.On("GetDecision", testExperimentDecisionContext, testUserContext, s.options).
		Return(returnExperimentDecision, s.reasons, nil)

	// Create service under test
	featureExperimentService := &FeatureExperimentService{
		compositeExperimentService: s.mockExperimentService,
		logger:                     logging.GetLogger("sdkKey", "FeatureExperimentService"),
	}

	// Create expected feature decision with propagated UUID
	expectedFeatureDecision := FeatureDecision{
		Experiment: *testExperimentDecisionContext.Experiment,
		Variation:  &expectedVariation,
		Source:     FeatureTest,
		CmabUUID:   &testUUID, // UUID should be propagated
	}

	// Call GetDecision
	actualFeatureDecision, _, err := featureExperimentService.GetDecision(s.testFeatureDecisionContext, testUserContext, s.options)

	// Verify results
	s.NoError(err)
	s.Equal(expectedFeatureDecision, actualFeatureDecision)

	// Verify CMAB UUID specifically
	s.NotNil(actualFeatureDecision.CmabUUID, "CmabUUID should not be nil")
	s.Equal(testUUID, *actualFeatureDecision.CmabUUID, "CmabUUID should match the expected value")

	s.mockExperimentService.AssertExpectations(s.T())
}

func (s *FeatureExperimentServiceTestSuite) TestGetDecisionWithCmabError() {
	testUserContext := entities.UserContext{
		ID: "test_user_1",
	}

	// Create a NEW CMAB experiment (don't modify existing testExp1113)
	cmabExperiment := entities.Experiment{
		ID:  "cmab_experiment_id",
		Key: "cmab_experiment_key",
		Cmab: &entities.Cmab{
			AttributeIds:      []string{"attr1", "attr2"},
			TrafficAllocation: 5000, // 50%
		},
		Variations: testExp1113.Variations, // Reuse variations for simplicity
	}

	// Setup experiment decision context for CMAB experiment
	testExperimentDecisionContext := ExperimentDecisionContext{
		Experiment:    &cmabExperiment,
		ProjectConfig: s.mockConfig,
	}

	// Mock the experiment service to return a CMAB error
	cmabError := errors.New("Failed to fetch CMAB data for experiment cmab_experiment_key.")
	s.mockExperimentService.On("GetDecision", testExperimentDecisionContext, testUserContext, s.options).
		Return(ExperimentDecision{}, s.reasons, cmabError)

	// Create a test feature that uses our CMAB experiment
	testFeatureWithCmab := entities.Feature{
		ID:  "test_feature_cmab",
		Key: "test_feature_cmab_key",
		FeatureExperiments: []entities.Experiment{
			cmabExperiment, // Only our CMAB experiment
		},
	}

	// Create feature decision context with our CMAB feature
	testFeatureDecisionContextWithCmab := FeatureDecisionContext{
		Feature:               &testFeatureWithCmab,
		ProjectConfig:         s.mockConfig,
		Variable:              testVariable,
		ForcedDecisionService: NewForcedDecisionService("test_user"),
	}

	// Create service under test
	featureExperimentService := &FeatureExperimentService{
		compositeExperimentService: s.mockExperimentService,
		logger:                     logging.GetLogger("sdkKey", "FeatureExperimentService"),
	}

	// Call GetDecision
	actualFeatureDecision, actualReasons, err := featureExperimentService.GetDecision(testFeatureDecisionContextWithCmab, testUserContext, s.options)

	// Verify that CMAB error is propagated (UPDATE THIS)
	s.Error(err, "CMAB errors should be propagated to prevent rollout fallback")
	s.Contains(err.Error(), "Failed to fetch CMAB data for experiment cmab_experiment_key")
	s.Equal(FeatureDecision{}, actualFeatureDecision, "Should return empty FeatureDecision when CMAB fails")

	// Verify that reasons include the CMAB error
	s.NotNil(actualReasons, "Decision reasons should not be nil")

	s.mockExperimentService.AssertExpectations(s.T())
}

func (s *FeatureExperimentServiceTestSuite) TestGetDecisionWithLocalHoldout() {
	// Create a custom mock that returns holdouts for the specific rule
	customMock := &mockProjectConfigWithHoldouts{
		mockProjectConfig: s.mockConfig,
		holdoutsForRule: map[string][]entities.Holdout{
			testExp1113.ID: {
				{
					ID:     "holdout_1",
					Key:    "test_holdout",
					Status: entities.HoldoutStatusRunning,
					Variations: map[string]entities.Variation{
						"var_holdout": {ID: "var_holdout", Key: "holdout_variation"},
					},
					TrafficAllocation: []entities.Range{
						{EntityID: "var_holdout", EndOfRange: 10000},
					},
					IncludedRules: []string{testExp1113.ID},
				},
			},
		},
		audienceMap: map[string]entities.Audience{},
	}

	testUserContext := entities.UserContext{ID: "test_user_1"}
	customDecisionContext := FeatureDecisionContext{
		Feature:       &testFeat3335,
		ProjectConfig: customMock,
	}

	featureExperimentService := &FeatureExperimentService{
		compositeExperimentService: s.mockExperimentService,
		logger:                     logging.GetLogger("sdkKey", "FeatureExperimentService"),
	}

	decision, _, err := featureExperimentService.GetDecision(customDecisionContext, testUserContext, s.options)

	s.NoError(err)
	s.Equal(Holdout, decision.Source, "Decision source should be Holdout")
	s.NotNil(decision.Variation, "Should have a variation from holdout")
}

func (s *FeatureExperimentServiceTestSuite) TestGetDecisionWithLocalHoldoutNotBucketed() {
	// Create a holdout with 0 traffic so user won't be bucketed
	customMock := &mockProjectConfigWithHoldouts{
		mockProjectConfig: s.mockConfig,
		holdoutsForRule: map[string][]entities.Holdout{
			testExp1113.ID: {
				{
					ID:     "holdout_zero_traffic",
					Key:    "test_holdout_no_bucket",
					Status: entities.HoldoutStatusRunning,
					Variations: map[string]entities.Variation{
						"var_holdout": {ID: "var_holdout", Key: "holdout_variation"},
					},
					TrafficAllocation: []entities.Range{
						{EntityID: "var_holdout", EndOfRange: 0}, // 0% traffic
					},
					IncludedRules: []string{testExp1113.ID},
				},
			},
		},
		audienceMap: map[string]entities.Audience{},
	}

	testUserContext := entities.UserContext{ID: "test_user_1"}
	customDecisionContext := FeatureDecisionContext{
		Feature:       &testFeat3335,
		ProjectConfig: customMock,
	}

	// Mock experiment service to return a variation (simulating normal bucketing after holdout check)
	expectedVariation := &testExp1113Var2223
	expectedDecision := ExperimentDecision{
		Variation: expectedVariation,
		Decision:  Decision{Reason: "bucketed"},
	}
	s.mockExperimentService.On("GetDecision", mock.Anything, testUserContext, s.options).
		Return(expectedDecision, decide.NewDecisionReasons(s.options), nil).Once()

	featureExperimentService := &FeatureExperimentService{
		compositeExperimentService: s.mockExperimentService,
		logger:                     logging.GetLogger("sdkKey", "FeatureExperimentService"),
	}

	decision, _, err := featureExperimentService.GetDecision(customDecisionContext, testUserContext, s.options)

	s.NoError(err)
	// Should fall through to normal experiment bucketing since holdout didn't bucket user
	s.Equal(FeatureTest, decision.Source, "Decision source should be FeatureTest (not Holdout)")
	s.Equal(expectedVariation, decision.Variation, "Should have variation from experiment bucketing")
	s.mockExperimentService.AssertExpectations(s.T())
}

func (s *FeatureExperimentServiceTestSuite) TestGetDecisionWithMultipleLocalHoldouts() {
	// Create a scenario with multiple holdouts where first doesn't bucket but second does
	customMock := &mockProjectConfigWithHoldouts{
		mockProjectConfig: s.mockConfig,
		holdoutsForRule: map[string][]entities.Holdout{
			testExp1113.ID: {
				// First holdout - 0% traffic, won't bucket user
				{
					ID:     "holdout_first",
					Key:    "test_holdout_first",
					Status: entities.HoldoutStatusRunning,
					Variations: map[string]entities.Variation{
						"var_first": {ID: "var_first", Key: "first_variation"},
					},
					TrafficAllocation: []entities.Range{
						{EntityID: "var_first", EndOfRange: 0}, // 0% traffic
					},
					IncludedRules: []string{testExp1113.ID},
				},
				// Second holdout - 100% traffic, will bucket user
				{
					ID:     "holdout_second",
					Key:    "test_holdout_second",
					Status: entities.HoldoutStatusRunning,
					Variations: map[string]entities.Variation{
						"var_second": {ID: "var_second", Key: "second_variation"},
					},
					TrafficAllocation: []entities.Range{
						{EntityID: "var_second", EndOfRange: 10000}, // 100% traffic
					},
					IncludedRules: []string{testExp1113.ID},
				},
			},
		},
		audienceMap: map[string]entities.Audience{},
	}

	testUserContext := entities.UserContext{ID: "test_user_1"}
	customDecisionContext := FeatureDecisionContext{
		Feature:       &testFeat3335,
		ProjectConfig: customMock,
	}

	featureExperimentService := &FeatureExperimentService{
		compositeExperimentService: s.mockExperimentService,
		logger:                     logging.GetLogger("sdkKey", "FeatureExperimentService"),
	}

	decision, _, err := featureExperimentService.GetDecision(customDecisionContext, testUserContext, s.options)

	s.NoError(err)
	// Should bucket into second holdout after first one returns nil
	s.Equal(Holdout, decision.Source, "Decision source should be Holdout")
	s.NotNil(decision.Variation, "Should have a variation from second holdout")
	s.Equal("second_variation", decision.Variation.Key, "Should be from second holdout")
}

// Custom mock that overrides GetHoldoutsForRule and GetAudienceMap
type mockProjectConfigWithHoldouts struct {
	*mockProjectConfig
	holdoutsForRule map[string][]entities.Holdout
	audienceMap     map[string]entities.Audience
}

func (m *mockProjectConfigWithHoldouts) GetHoldoutsForRule(ruleID string) []entities.Holdout {
	if holdouts, ok := m.holdoutsForRule[ruleID]; ok {
		return holdouts
	}
	return []entities.Holdout{}
}

func (m *mockProjectConfigWithHoldouts) GetAudienceMap() map[string]entities.Audience {
	return m.audienceMap
}

func TestFeatureExperimentServiceTestSuite(t *testing.T) {
	suite.Run(t, new(FeatureExperimentServiceTestSuite))
}
