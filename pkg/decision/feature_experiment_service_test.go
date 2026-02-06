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

func (s *FeatureExperimentServiceTestSuite) TestGetDecisionSkipsUnsupportedExperimentType() {
	testUserContext := entities.UserContext{
		ID: "test_user_1",
	}

	// Create experiment with unsupported type
	unsupportedExp := testExp1113
	unsupportedExp.Type = "unsupported_type"

	// Create feature with the unsupported experiment
	featureWithUnsupportedExp := testFeat3335
	featureWithUnsupportedExp.FeatureExperiments = []entities.Experiment{unsupportedExp}

	testFeatureDecisionContext := FeatureDecisionContext{
		Feature:       &featureWithUnsupportedExp,
		ProjectConfig: s.mockConfig,
	}

	featureExperimentService := &FeatureExperimentService{
		compositeExperimentService: s.mockExperimentService,
		logger:                     logging.GetLogger("sdkKey", "FeatureExperimentService"),
	}

	decision, _, err := featureExperimentService.GetDecision(testFeatureDecisionContext, testUserContext, s.options)

	// Should return empty decision since experiment was skipped
	s.Equal(FeatureDecision{}, decision)
	s.NoError(err)
}

func (s *FeatureExperimentServiceTestSuite) TestGetDecisionEvaluatesExperimentWithSupportedTypeAB() {
	testUserContext := entities.UserContext{
		ID: "test_user_1",
	}

	// Create experiment with supported type "a/b"
	abExp := testExp1113
	abExp.Type = "a/b"

	expectedVariation := abExp.Variations["2223"]
	returnExperimentDecision := ExperimentDecision{
		Variation: &expectedVariation,
	}
	testExperimentDecisionContext := ExperimentDecisionContext{
		Experiment:    &abExp,
		ProjectConfig: s.mockConfig,
	}
	s.mockExperimentService.On("GetDecision", testExperimentDecisionContext, testUserContext, s.options).Return(returnExperimentDecision, s.reasons, nil)

	// Create feature with the a/b experiment
	featureWithABExp := testFeat3335
	featureWithABExp.FeatureExperiments = []entities.Experiment{abExp}

	testFeatureDecisionContext := FeatureDecisionContext{
		Feature:       &featureWithABExp,
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

	decision, _, err := featureExperimentService.GetDecision(testFeatureDecisionContext, testUserContext, s.options)
	s.Equal(expectedFeatureDecision, decision)
	s.NoError(err)
	s.mockExperimentService.AssertExpectations(s.T())
}

func (s *FeatureExperimentServiceTestSuite) TestGetDecisionEvaluatesExperimentWithSupportedTypeMAB() {
	testUserContext := entities.UserContext{
		ID: "test_user_1",
	}

	// Create experiment with supported type "mab"
	mabExp := testExp1113
	mabExp.Type = "mab"

	expectedVariation := mabExp.Variations["2223"]
	returnExperimentDecision := ExperimentDecision{
		Variation: &expectedVariation,
	}
	testExperimentDecisionContext := ExperimentDecisionContext{
		Experiment:    &mabExp,
		ProjectConfig: s.mockConfig,
	}
	s.mockExperimentService.On("GetDecision", testExperimentDecisionContext, testUserContext, s.options).Return(returnExperimentDecision, s.reasons, nil)

	// Create feature with the mab experiment
	featureWithMABExp := testFeat3335
	featureWithMABExp.FeatureExperiments = []entities.Experiment{mabExp}

	testFeatureDecisionContext := FeatureDecisionContext{
		Feature:       &featureWithMABExp,
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

	decision, _, err := featureExperimentService.GetDecision(testFeatureDecisionContext, testUserContext, s.options)
	s.Equal(expectedFeatureDecision, decision)
	s.NoError(err)
	s.mockExperimentService.AssertExpectations(s.T())
}

func (s *FeatureExperimentServiceTestSuite) TestGetDecisionEvaluatesExperimentWithEmptyType() {
	testUserContext := entities.UserContext{
		ID: "test_user_1",
	}

	// Create experiment with empty type (backward compatible)
	emptyTypeExp := testExp1113
	emptyTypeExp.Type = ""

	expectedVariation := emptyTypeExp.Variations["2223"]
	returnExperimentDecision := ExperimentDecision{
		Variation: &expectedVariation,
	}
	testExperimentDecisionContext := ExperimentDecisionContext{
		Experiment:    &emptyTypeExp,
		ProjectConfig: s.mockConfig,
	}
	s.mockExperimentService.On("GetDecision", testExperimentDecisionContext, testUserContext, s.options).Return(returnExperimentDecision, s.reasons, nil)

	// Create feature with the empty type experiment
	featureWithEmptyTypeExp := testFeat3335
	featureWithEmptyTypeExp.FeatureExperiments = []entities.Experiment{emptyTypeExp}

	testFeatureDecisionContext := FeatureDecisionContext{
		Feature:       &featureWithEmptyTypeExp,
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

	decision, _, err := featureExperimentService.GetDecision(testFeatureDecisionContext, testUserContext, s.options)
	s.Equal(expectedFeatureDecision, decision)
	s.NoError(err)
	s.mockExperimentService.AssertExpectations(s.T())
}

func (s *FeatureExperimentServiceTestSuite) TestGetDecisionSkipsUnsupportedAndEvaluatesSupported() {
	testUserContext := entities.UserContext{
		ID: "test_user_1",
	}

	// Create unsupported experiment
	unsupportedExp := testExp1113
	unsupportedExp.Type = "unsupported_type"
	unsupportedExp.Key = "unsupported_exp"

	// Create supported experiment
	supportedExp := testExp1113
	supportedExp.Type = "a/b"
	supportedExp.Key = "supported_exp"

	expectedVariation := supportedExp.Variations["2223"]
	returnExperimentDecision := ExperimentDecision{
		Variation: &expectedVariation,
	}
	testExperimentDecisionContext := ExperimentDecisionContext{
		Experiment:    &supportedExp,
		ProjectConfig: s.mockConfig,
	}
	s.mockExperimentService.On("GetDecision", testExperimentDecisionContext, testUserContext, s.options).Return(returnExperimentDecision, s.reasons, nil)

	// Create feature with both experiments
	featureWithBothExp := testFeat3335
	featureWithBothExp.FeatureExperiments = []entities.Experiment{unsupportedExp, supportedExp}

	testFeatureDecisionContext := FeatureDecisionContext{
		Feature:       &featureWithBothExp,
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

	decision, _, err := featureExperimentService.GetDecision(testFeatureDecisionContext, testUserContext, s.options)

	// Should skip unsupported and evaluate supported
	s.Equal(expectedFeatureDecision, decision)
	s.NoError(err)
	s.mockExperimentService.AssertExpectations(s.T())
}

func TestFeatureExperimentServiceTestSuite(t *testing.T) {
	suite.Run(t, new(FeatureExperimentServiceTestSuite))
}
