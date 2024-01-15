/****************************************************************************
 * Copyright 2020-2022, Optimizely, Inc. and contributors                   *
 *                                                                          *
 * Licensed under the Apache License, Version 2.0 (the "License");          *
 * you may not use this file except in compliance with the License.         *
 * You may obtain a copy of the License at                                  *
 *                                                                          *
 *    https://www.apache.org/licenses/LICENSE-2.0                           *
 *                                                                          *
 * Unless required by applicable law or agreed to in writing, software      *
 * distributed under the License is distributed on an "AS IS" BASIS,        *
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. *
 * See the License for the specific language governing permissions and      *
 * limitations under the License.                                           *
 ***************************************************************************/

package client

import (
	"context"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/optimizely/go-sdk/pkg/decide"
	"github.com/optimizely/go-sdk/pkg/decision"
	"github.com/optimizely/go-sdk/pkg/entities"
	"github.com/optimizely/go-sdk/pkg/event"
	"github.com/optimizely/go-sdk/pkg/notification"
	"github.com/optimizely/go-sdk/pkg/optimizelyjson"
)

var doOnce sync.Once // required since we only need to read datafile once
var datafile []byte

type OptimizelyUserContextTestSuite struct {
	suite.Suite
	*OptimizelyClient
	userID         string
	factory        OptimizelyFactory
	eventProcessor *MockProcessor
}

func (s *OptimizelyUserContextTestSuite) SetupTest() {
	doOnce.Do(func() {
		absPath, _ := filepath.Abs("../../test-data/decide-test-datafile.json")
		datafile, _ = os.ReadFile(absPath)
	})
	s.eventProcessor = new(MockProcessor)
	s.eventProcessor.On("ProcessEvent", mock.AnythingOfType("event.UserEvent")).Return(true)
	s.factory = OptimizelyFactory{Datafile: datafile}
	s.OptimizelyClient, _ = s.factory.Client(WithEventProcessor(s.eventProcessor))
	s.userID = "tester"
}

func (s *OptimizelyUserContextTestSuite) TestOptimizelyUserContextWithAttributesAndSegments() {
	attributes := map[string]interface{}{"key1": 1212, "key2": 1213}
	segments := []string{"123"}
	optimizelyUserContext := newOptimizelyUserContext(s.OptimizelyClient, s.userID, attributes, nil, segments)

	s.Equal(s.OptimizelyClient, optimizelyUserContext.GetOptimizely())
	s.Equal(s.userID, optimizelyUserContext.GetUserID())
	s.Equal(attributes, optimizelyUserContext.GetUserAttributes())
	s.Equal(segments, optimizelyUserContext.GetQualifiedSegments())
	s.Nil(optimizelyUserContext.forcedDecisionService)
}

func (s *OptimizelyUserContextTestSuite) TestOptimizelyUserContextNoAttributesAndNilSegments() {
	attributes := map[string]interface{}{}
	optimizelyUserContext := newOptimizelyUserContext(s.OptimizelyClient, s.userID, attributes, nil, nil)

	s.Equal(s.OptimizelyClient, optimizelyUserContext.GetOptimizely())
	s.Equal(s.userID, optimizelyUserContext.GetUserID())
	s.Equal(attributes, optimizelyUserContext.GetUserAttributes())
	s.Nil(optimizelyUserContext.GetQualifiedSegments())
}

func (s *OptimizelyUserContextTestSuite) TestUpatingProvidedUserContextHasNoImpactOnOptimizelyUserContext() {
	attributes := map[string]interface{}{"k1": "v1", "k2": false}
	segments := []string{"123"}
	optimizelyUserContext := newOptimizelyUserContext(s.OptimizelyClient, s.userID, attributes, nil, segments)

	s.Equal(s.OptimizelyClient, optimizelyUserContext.GetOptimizely())
	s.Equal(s.userID, optimizelyUserContext.GetUserID())
	s.Equal(attributes, optimizelyUserContext.GetUserAttributes())
	s.Equal(segments, optimizelyUserContext.GetQualifiedSegments())

	attributes["k1"] = "v2"
	attributes["k2"] = true
	segments[0] = "456"

	s.Equal("v1", optimizelyUserContext.GetUserAttributes()["k1"])
	s.Equal(false, optimizelyUserContext.GetUserAttributes()["k2"])
	s.Equal([]string{"123"}, optimizelyUserContext.GetQualifiedSegments())

	attributes = optimizelyUserContext.GetUserAttributes()
	segments = optimizelyUserContext.GetQualifiedSegments()
	attributes["k1"] = "v2"
	attributes["k2"] = true
	segments[0] = "456"

	s.Equal("v1", optimizelyUserContext.GetUserAttributes()["k1"])
	s.Equal(false, optimizelyUserContext.GetUserAttributes()["k2"])
	s.Equal([]string{"123"}, optimizelyUserContext.GetQualifiedSegments())
}

func (s *OptimizelyUserContextTestSuite) TestSetAndGetUserAttributesRaceCondition() {
	userID := "1212121"
	var attributes map[string]interface{}

	optimizelyUserContext := newOptimizelyUserContext(s.OptimizelyClient, userID, attributes, nil, nil)
	s.Equal(s.OptimizelyClient, optimizelyUserContext.GetOptimizely())

	var wg sync.WaitGroup
	wg.Add(8)
	addInsideGoRoutine := func(key string, value interface{}, wg *sync.WaitGroup) {
		optimizelyUserContext.SetAttribute(key, value)
		wg.Done()
	}
	getInsideGoRoutine := func(wg *sync.WaitGroup) {
		optimizelyUserContext.GetUserAttributes()
		wg.Done()
	}

	go addInsideGoRoutine("k1", "v1", &wg)
	go addInsideGoRoutine("k2", true, &wg)
	go addInsideGoRoutine("k3", 100, &wg)
	go addInsideGoRoutine("k4", 3.5, &wg)
	go getInsideGoRoutine(&wg)
	go getInsideGoRoutine(&wg)
	go getInsideGoRoutine(&wg)
	go getInsideGoRoutine(&wg)
	wg.Wait()

	s.Equal(userID, optimizelyUserContext.GetUserID())
	s.Equal("v1", optimizelyUserContext.GetUserAttributes()["k1"])
	s.Equal(true, optimizelyUserContext.GetUserAttributes()["k2"])
	s.Equal(100, optimizelyUserContext.GetUserAttributes()["k3"])
	s.Equal(3.5, optimizelyUserContext.GetUserAttributes()["k4"])
}

func (s *OptimizelyUserContextTestSuite) TestSetAttributeOverride() {
	userID := "1212121"
	attributes := map[string]interface{}{"k1": "v1", "k2": false}
	optimizelyUserContext := newOptimizelyUserContext(s.OptimizelyClient, userID, attributes, nil, nil)

	s.Equal(s.OptimizelyClient, optimizelyUserContext.GetOptimizely())
	s.Equal(userID, optimizelyUserContext.GetUserID())
	s.Equal(attributes, optimizelyUserContext.GetUserAttributes())

	optimizelyUserContext.SetAttribute("k1", "v2")
	optimizelyUserContext.SetAttribute("k2", true)

	s.Equal("v2", optimizelyUserContext.GetUserAttributes()["k1"])
	s.Equal(true, optimizelyUserContext.GetUserAttributes()["k2"])
}

func (s *OptimizelyUserContextTestSuite) TestSetAttributeNullValue() {
	userID := "1212121"
	attributes := map[string]interface{}{"k1": nil}
	optimizelyUserContext := newOptimizelyUserContext(s.OptimizelyClient, userID, attributes, nil, nil)

	s.Equal(s.OptimizelyClient, optimizelyUserContext.GetOptimizely())
	s.Equal(userID, optimizelyUserContext.GetUserID())
	s.Equal(attributes, optimizelyUserContext.GetUserAttributes())

	optimizelyUserContext.SetAttribute("k1", true)
	s.Equal(true, optimizelyUserContext.GetUserAttributes()["k1"])

	optimizelyUserContext.SetAttribute("k1", nil)
	s.Equal(nil, optimizelyUserContext.GetUserAttributes()["k1"])
}

func (s *OptimizelyUserContextTestSuite) TestSetAndGetQualifiedSegments() {
	userID := "1212121"
	var attributes map[string]interface{}
	qualifiedSegments := []string{"1", "2", "3"}
	optimizelyUserContext := newOptimizelyUserContext(s.OptimizelyClient, userID, attributes, nil, []string{})
	s.Len(optimizelyUserContext.GetQualifiedSegments(), 0)

	optimizelyUserContext.SetQualifiedSegments(nil)
	s.Nil(optimizelyUserContext.GetQualifiedSegments())

	optimizelyUserContext.SetQualifiedSegments(qualifiedSegments)
	s.Equal(qualifiedSegments, optimizelyUserContext.GetQualifiedSegments())
}

func (s *OptimizelyUserContextTestSuite) TestQualifiedSegmentsRaceCondition() {
	userID := "1212121"
	qualifiedSegments := []string{"1", "2", "3"}
	segment := "1"
	var attributes map[string]interface{}

	optimizelyUserContext := newOptimizelyUserContext(s.OptimizelyClient, userID, attributes, nil, nil)
	s.Nil(optimizelyUserContext.GetQualifiedSegments())
	var wg sync.WaitGroup
	wg.Add(9)

	setQualifiedSegments := func(value []string, wg *sync.WaitGroup) {
		optimizelyUserContext.SetQualifiedSegments(value)
		wg.Done()
	}
	getQualifiedSegments := func(wg *sync.WaitGroup) {
		optimizelyUserContext.GetQualifiedSegments()
		wg.Done()
	}

	IsQualifiedFor := func(segment string, wg *sync.WaitGroup) {
		optimizelyUserContext.IsQualifiedFor(segment)
		wg.Done()
	}

	go setQualifiedSegments(qualifiedSegments, &wg)
	go setQualifiedSegments(qualifiedSegments, &wg)
	go setQualifiedSegments(qualifiedSegments, &wg)
	go getQualifiedSegments(&wg)
	go getQualifiedSegments(&wg)
	go getQualifiedSegments(&wg)
	go IsQualifiedFor(segment, &wg)
	go IsQualifiedFor(segment, &wg)
	go IsQualifiedFor(segment, &wg)

	wg.Wait()

	s.Equal(qualifiedSegments, optimizelyUserContext.GetQualifiedSegments())
	s.Equal(true, optimizelyUserContext.IsQualifiedFor(segment))
}

func (s *OptimizelyUserContextTestSuite) TestIsQualifiedFor() {
	userID := "1212121"
	qualifiedSegments := []string{"1", "2", "3"}
	var attributes map[string]interface{}

	optimizelyUserContext := newOptimizelyUserContext(s.OptimizelyClient, userID, attributes, nil, nil)
	s.False(optimizelyUserContext.IsQualifiedFor("1"))
	optimizelyUserContext.SetQualifiedSegments(qualifiedSegments)

	var wg sync.WaitGroup
	wg.Add(6)
	testInsideGoRoutine := func(value string, result bool, wg *sync.WaitGroup) {
		s.Equal(result, optimizelyUserContext.IsQualifiedFor(value))
		wg.Done()
	}

	go testInsideGoRoutine("1", true, &wg)
	go testInsideGoRoutine("2", true, &wg)
	go testInsideGoRoutine("3", true, &wg)
	go testInsideGoRoutine("4", false, &wg)
	go testInsideGoRoutine("5", false, &wg)
	go testInsideGoRoutine("6", false, &wg)

	wg.Wait()
}

func (s *OptimizelyUserContextTestSuite) TestDecideResponseContainsUserContextCopy() {
	flagKey := "feature_2"
	userContext := s.OptimizelyClient.CreateUserContext(s.userID, nil)
	decision := userContext.Decide(context.Background(), flagKey, nil)
	decisionUserContext := decision.UserContext

	// Change attributes for user context
	userContext.SetAttribute("test", 123)
	// Change qualifiedSegments for user context
	userContext.SetQualifiedSegments([]string{"123"})
	// Attributes and qualifiedSegments should not update for the userContext returned inside decision
	s.Nil(decisionUserContext.Attributes["test"])
	s.Len(decisionUserContext.qualifiedSegments, 0)
}

func (s *OptimizelyUserContextTestSuite) TestDecideFeatureTest() {
	flagKey := "feature_2"
	ruleKey := "exp_no_audience"
	variationKey := "variation_with_traffic"
	variablesExpected, err := s.OptimizelyClient.GetAllFeatureVariables(context.Background(), flagKey, entities.UserContext{ID: s.userID})
	s.Nil(err)

	user := s.OptimizelyClient.CreateUserContext(s.userID, nil)
	decision := user.Decide(context.Background(), flagKey, []decide.OptimizelyDecideOptions{decide.IncludeReasons})

	s.Equal(variationKey, decision.VariationKey)
	s.Equal(true, decision.Enabled)
	s.Equal(variablesExpected.ToMap(), decision.Variables.ToMap())
	s.Equal(ruleKey, decision.RuleKey)
	s.Equal(flagKey, decision.FlagKey)
	s.Equal(user, decision.UserContext)
	reasons := decision.Reasons
	s.Len(reasons, 1)
	s.Equal(`Audiences for experiment exp_no_audience collectively evaluated to true.`, reasons[0])

	s.True(len(s.eventProcessor.Events) == 1)
	s.Equal(s.userID, s.eventProcessor.Events[0].VisitorID)

	impressionEvent := s.eventProcessor.Events[0].Impression
	s.Equal(flagKey, impressionEvent.Metadata.FlagKey)
	s.Equal(ruleKey, impressionEvent.Metadata.RuleKey)
	s.Equal("feature-test", impressionEvent.Metadata.RuleType)
	s.Equal(variationKey, impressionEvent.Metadata.VariationKey)
	s.Equal(true, impressionEvent.Metadata.Enabled)
	s.Equal("10420810910", impressionEvent.ExperimentID)
	s.Equal("10418551353", impressionEvent.VariationID)
}

func (s *OptimizelyUserContextTestSuite) TestDecideFeatureTestWithForcedDecision() {
	numberOfNotifications := 0
	testForcedDecision := func(flagKey, ruleKey, experimentID, variationKey, reason string, expectedEventCount int) {
		variablesExpected, err := s.OptimizelyClient.GetAllFeatureVariables(context.Background(), flagKey, entities.UserContext{ID: s.userID})
		s.Nil(err)

		note := notification.DecisionNotification{}
		callback := func(notification notification.DecisionNotification) {
			note = notification
			numberOfNotifications++
		}
		notificationID, err := s.OptimizelyClient.DecisionService.OnDecision(callback)
		s.NoError(err)

		user := s.OptimizelyClient.CreateUserContext(s.userID, nil)
		user.SetForcedDecision(decision.OptimizelyDecisionContext{FlagKey: flagKey, RuleKey: ruleKey}, decision.OptimizelyForcedDecision{VariationKey: variationKey})
		decision := user.Decide(context.Background(), flagKey, []decide.OptimizelyDecideOptions{decide.IncludeReasons})
		s.OptimizelyClient.DecisionService.RemoveOnDecision(notificationID)

		s.Equal(variationKey, decision.VariationKey)
		s.Equal(false, decision.Enabled)
		s.Equal(variablesExpected.ToMap(), decision.Variables.ToMap())
		s.Equal(ruleKey, decision.RuleKey)
		s.Equal(flagKey, decision.FlagKey)
		s.Equal(user, decision.UserContext)
		reasons := decision.Reasons
		s.Len(reasons, 1)
		s.Equal(reason, reasons[0])

		s.True(len(s.eventProcessor.Events) == expectedEventCount)
		s.Equal(s.userID, s.eventProcessor.Events[expectedEventCount-1].VisitorID)

		impressionEvent := s.eventProcessor.Events[expectedEventCount-1].Impression
		s.Equal(flagKey, impressionEvent.Metadata.FlagKey)
		s.Equal(ruleKey, impressionEvent.Metadata.RuleKey)
		s.Equal("feature-test", impressionEvent.Metadata.RuleType)
		s.Equal(variationKey, impressionEvent.Metadata.VariationKey)
		s.Equal(false, impressionEvent.Metadata.Enabled)
		s.Equal(experimentID, impressionEvent.ExperimentID)
		s.Equal("10416523121", impressionEvent.VariationID)

		// Checking notification data
		s.Equal(note.DecisionInfo["flagKey"], impressionEvent.Metadata.FlagKey)
		s.Equal(note.DecisionInfo["ruleKey"], impressionEvent.Metadata.RuleKey)
		s.Equal(note.DecisionInfo["enabled"], impressionEvent.Metadata.Enabled)
		s.Equal(note.DecisionInfo["variationKey"], impressionEvent.Metadata.VariationKey)
	}

	// valid rule key
	expectedEventCount := 1
	flagKey := "feature_1"
	ruleKey := "exp_with_audience"
	experimentID := "10390977673"
	variationKey := "b"
	reason := `Variation (b) is mapped to flag (feature_1), rule (exp_with_audience) and user (tester) in the forced decision map.`
	testForcedDecision(flagKey, ruleKey, experimentID, variationKey, reason, expectedEventCount)

	// empty rule key
	expectedEventCount = 2
	ruleKey = ""
	experimentID = ""
	reason = `Variation (b) is mapped to flag (feature_1) and user (tester) in the forced decision map.`
	testForcedDecision(flagKey, ruleKey, experimentID, variationKey, reason, expectedEventCount)

	s.Equal(2, numberOfNotifications)
}

func (s *OptimizelyUserContextTestSuite) TestDecideFeatureTestWithForcedDecisionEmptyRuleKey() {
	flagKey := "feature_1"
	ruleKey := ""
	variationKey := "b"
	variablesExpected, err := s.OptimizelyClient.GetAllFeatureVariables(context.Background(), flagKey, entities.UserContext{ID: s.userID})
	s.Nil(err)

	user := s.OptimizelyClient.CreateUserContext(s.userID, nil)
	user.SetForcedDecision(decision.OptimizelyDecisionContext{FlagKey: flagKey, RuleKey: ruleKey}, decision.OptimizelyForcedDecision{VariationKey: variationKey})
	decision := user.Decide(context.Background(), flagKey, []decide.OptimizelyDecideOptions{decide.IncludeReasons})

	s.Equal(variationKey, decision.VariationKey)
	s.Equal(false, decision.Enabled)
	s.Equal(variablesExpected.ToMap(), decision.Variables.ToMap())
	s.Equal(ruleKey, decision.RuleKey)
	s.Equal(flagKey, decision.FlagKey)
	s.Equal(user, decision.UserContext)
	reasons := decision.Reasons
	s.Len(reasons, 1)
	s.Equal(`Variation (b) is mapped to flag (feature_1) and user (tester) in the forced decision map.`, reasons[0])

	s.True(len(s.eventProcessor.Events) == 1)
	s.Equal(s.userID, s.eventProcessor.Events[0].VisitorID)

	impressionEvent := s.eventProcessor.Events[0].Impression
	s.Equal(flagKey, impressionEvent.Metadata.FlagKey)
	s.Equal(ruleKey, impressionEvent.Metadata.RuleKey)
	s.Equal("feature-test", impressionEvent.Metadata.RuleType)
	s.Equal(variationKey, impressionEvent.Metadata.VariationKey)
	s.Equal(false, impressionEvent.Metadata.Enabled)
	s.Equal("", impressionEvent.ExperimentID)
	s.Equal("10416523121", impressionEvent.VariationID)
}

func (s *OptimizelyUserContextTestSuite) TestDecideRollout() {
	flagKey := "feature_1"
	ruleKey := "18322080788"
	variationKey := "18257766532"
	variablesExpected, err := s.OptimizelyClient.GetAllFeatureVariables(context.Background(), flagKey, entities.UserContext{ID: s.userID})
	s.Nil(err)

	user := s.OptimizelyClient.CreateUserContext(s.userID, nil)
	decision := user.Decide(context.Background(), flagKey, []decide.OptimizelyDecideOptions{decide.IncludeReasons})

	s.Equal(variationKey, decision.VariationKey)
	s.Equal(true, decision.Enabled)
	s.Equal(variablesExpected.ToMap(), decision.Variables.ToMap())
	s.Equal(ruleKey, decision.RuleKey)
	s.Equal(flagKey, decision.FlagKey)
	s.Equal(user, decision.UserContext)
	reasons := decision.Reasons
	s.Len(reasons, 9)

	expectedLogs := []string{
		`an error occurred while evaluating nested tree for audience ID "13389141123"`,
		`Audiences for experiment exp_with_audience collectively evaluated to false.`,
		`User "tester" does not meet conditions to be in experiment "exp_with_audience".`,
		`an error occurred while evaluating nested tree for audience ID "13389130056"`,
		`User "tester" does not meet conditions for targeting rule 1.`,
		`an error occurred while evaluating nested tree for audience ID "12208130097"`,
		`User "tester" does not meet conditions for targeting rule 2.`,
		`Audiences for experiment 18322080788 collectively evaluated to true.`,
		`User "tester" meets conditions for targeting rule "Everyone Else".`,
	}

	for index, log := range expectedLogs {
		s.Equal(log, reasons[index])
	}

	s.True(len(s.eventProcessor.Events) == 1)
	s.Equal(s.userID, s.eventProcessor.Events[0].VisitorID)

	impressionEvent := s.eventProcessor.Events[0].Impression
	s.Equal(flagKey, impressionEvent.Metadata.FlagKey)
	s.Equal(ruleKey, impressionEvent.Metadata.RuleKey)
	s.Equal("rollout", impressionEvent.Metadata.RuleType)
	s.Equal(variationKey, impressionEvent.Metadata.VariationKey)
	s.Equal(true, impressionEvent.Metadata.Enabled)
	s.Equal("18322080788", impressionEvent.ExperimentID)
	s.Equal("18257766532", impressionEvent.VariationID)
}

func (s *OptimizelyUserContextTestSuite) TestDecideRolloutWithForcedDecision() {
	flagKey := "feature_1"
	ruleKey := "3332020515"
	variationKey := "3324490633"
	variablesExpected, err := s.OptimizelyClient.GetAllFeatureVariables(context.Background(), flagKey, entities.UserContext{ID: s.userID})
	s.Nil(err)

	user := s.OptimizelyClient.CreateUserContext(s.userID, nil)
	user.SetForcedDecision(decision.OptimizelyDecisionContext{FlagKey: flagKey, RuleKey: ruleKey}, decision.OptimizelyForcedDecision{VariationKey: variationKey})
	decision := user.Decide(context.Background(), flagKey, []decide.OptimizelyDecideOptions{decide.IncludeReasons})

	s.Equal(variationKey, decision.VariationKey)
	s.Equal(true, decision.Enabled)
	s.Equal(variablesExpected.ToMap(), decision.Variables.ToMap())
	s.Equal(ruleKey, decision.RuleKey)
	s.Equal(flagKey, decision.FlagKey)
	s.Equal(user, decision.UserContext)
	reasons := decision.Reasons
	s.Len(reasons, 4)

	expectedLogs := []string{
		`an error occurred while evaluating nested tree for audience ID "13389141123"`,
		`Audiences for experiment exp_with_audience collectively evaluated to false.`,
		`User "tester" does not meet conditions to be in experiment "exp_with_audience".`,
		`Variation (3324490633) is mapped to flag (feature_1), rule (3332020515) and user (tester) in the forced decision map.`,
	}

	for index, log := range expectedLogs {
		s.Equal(log, reasons[index])
	}

	s.True(len(s.eventProcessor.Events) == 1)
	s.Equal(s.userID, s.eventProcessor.Events[0].VisitorID)

	impressionEvent := s.eventProcessor.Events[0].Impression
	s.Equal(flagKey, impressionEvent.Metadata.FlagKey)
	s.Equal(ruleKey, impressionEvent.Metadata.RuleKey)
	s.Equal("rollout", impressionEvent.Metadata.RuleType)
	s.Equal(variationKey, impressionEvent.Metadata.VariationKey)
	s.Equal(true, impressionEvent.Metadata.Enabled)
	s.Equal("3332020515", impressionEvent.ExperimentID)
	s.Equal("3324490633", impressionEvent.VariationID)
}

func (s *OptimizelyUserContextTestSuite) TestDecideNullVariation() {
	flagKey := "feature_3"
	variablesExpected := optimizelyjson.NewOptimizelyJSONfromMap(map[string]interface{}{})

	user := s.OptimizelyClient.CreateUserContext(s.userID, nil)
	decision := user.Decide(context.Background(), flagKey, []decide.OptimizelyDecideOptions{decide.IncludeReasons})

	s.Equal("", decision.VariationKey)
	s.Equal(false, decision.Enabled)
	s.Equal(variablesExpected.ToMap(), decision.Variables.ToMap())
	s.Equal("", decision.RuleKey)
	s.Equal("feature_3", decision.FlagKey)
	s.Equal(user, decision.UserContext)
	reasons := decision.Reasons
	s.Len(reasons, 1)
	s.Equal(`Rollout with ID "" is not in the datafile.`, reasons[0])

	s.True(len(s.eventProcessor.Events) == 1)
	s.Equal(s.userID, s.eventProcessor.Events[0].VisitorID)

	impressionEvent := s.eventProcessor.Events[0].Impression
	s.Equal(flagKey, impressionEvent.Metadata.FlagKey)
	s.Equal("", impressionEvent.Metadata.RuleKey)
	s.Equal("rollout", impressionEvent.Metadata.RuleType)
	s.Equal("", impressionEvent.Metadata.VariationKey)
	s.Equal(false, impressionEvent.Metadata.Enabled)
	s.Equal("", impressionEvent.ExperimentID)
	s.Equal("", impressionEvent.VariationID)
}

func (s *OptimizelyUserContextTestSuite) TestDecideForKeysOneFlag() {
	flagKey := "feature_2"
	flagKeys := []string{flagKey}
	ruleKey := "exp_no_audience"
	variationKey := "variation_with_traffic"
	variablesExpected, err := s.OptimizelyClient.GetAllFeatureVariables(context.Background(), flagKey, entities.UserContext{ID: s.userID})
	s.Nil(err)

	user := s.OptimizelyClient.CreateUserContext(s.userID, nil)
	decisions := user.DecideForKeys(context.Background(), flagKeys, []decide.OptimizelyDecideOptions{decide.IncludeReasons})
	s.Len(decisions, 1)

	decision1 := decisions[flagKey]

	s.Equal(variationKey, decision1.VariationKey)
	s.Equal(true, decision1.Enabled)
	s.Equal(variablesExpected.ToMap(), decision1.Variables.ToMap())
	s.Equal(ruleKey, decision1.RuleKey)
	s.Equal(flagKey, decision1.FlagKey)
	s.Equal(user, decision1.UserContext)

	reasons := decision1.Reasons
	s.Len(reasons, 1)
	s.Equal(`Audiences for experiment exp_no_audience collectively evaluated to true.`, reasons[0])

	s.True(len(s.eventProcessor.Events) == 1)
	s.Equal(s.userID, s.eventProcessor.Events[0].VisitorID)

	impressionEvent := s.eventProcessor.Events[0].Impression
	s.Equal(flagKey, impressionEvent.Metadata.FlagKey)
	s.Equal(ruleKey, impressionEvent.Metadata.RuleKey)
	s.Equal("feature-test", impressionEvent.Metadata.RuleType)
	s.Equal(variationKey, impressionEvent.Metadata.VariationKey)
	s.Equal(true, impressionEvent.Metadata.Enabled)
	s.Equal("10420810910", impressionEvent.ExperimentID)
	s.Equal("10418551353", impressionEvent.VariationID)
}

func (s *OptimizelyUserContextTestSuite) TestDecideForKeysWithMultipleFlags() {
	flagKey1 := "feature_1"
	flagKey2 := "feature_2"
	ruleKey1 := "exp_with_audience"
	ruleKey2 := "exp_no_audience"
	variationKey1 := "a"
	variationKey2 := "variation_with_traffic"
	flagKeys := []string{flagKey1, flagKey2}
	variablesExpected1, err := s.OptimizelyClient.GetAllFeatureVariables(context.Background(), flagKey1, entities.UserContext{ID: s.userID})
	s.Nil(err)
	variablesExpected2, err := s.OptimizelyClient.GetAllFeatureVariables(context.Background(), flagKey2, entities.UserContext{ID: s.userID})
	s.Nil(err)

	user := s.OptimizelyClient.CreateUserContext(s.userID, map[string]interface{}{"gender": "f"})
	decisions := user.DecideForKeys(context.Background(), flagKeys, []decide.OptimizelyDecideOptions{decide.IncludeReasons})
	s.Len(decisions, 2)

	decision1 := decisions[flagKey1]
	s.Equal(variationKey1, decision1.VariationKey)
	s.Equal(true, decision1.Enabled)
	s.Equal(variablesExpected1.ToMap(), decision1.Variables.ToMap())
	s.Equal(ruleKey1, decision1.RuleKey)
	s.Equal(flagKey1, decision1.FlagKey)
	s.Equal(user, decision1.UserContext)
	reasons := decision1.Reasons
	s.Len(reasons, 1)
	s.Equal(`Audiences for experiment exp_with_audience collectively evaluated to true.`, reasons[0])

	decision2 := decisions[flagKey2]
	s.Equal(variationKey2, decision2.VariationKey)
	s.Equal(true, decision2.Enabled)
	s.Equal(variablesExpected2.ToMap(), decision2.Variables.ToMap())
	s.Equal(ruleKey2, decision2.RuleKey)
	s.Equal(flagKey2, decision2.FlagKey)
	s.Equal(user, decision2.UserContext)
	reasons = decision2.Reasons
	s.Len(reasons, 1)
	s.Equal(`Audiences for experiment exp_no_audience collectively evaluated to true.`, reasons[0])

	s.True(len(s.eventProcessor.Events) == 2)
	s.Equal(s.userID, s.eventProcessor.Events[0].VisitorID)

	eventMapping := map[int]event.UserEvent{}
	for _, event := range s.eventProcessor.Events {
		switch event.Impression.Metadata.FlagKey {
		case flagKey1:
			eventMapping[0] = event
		case flagKey2:
			eventMapping[1] = event
		}
	}

	impressionEvent1 := eventMapping[0].Impression
	s.Equal(flagKey1, impressionEvent1.Metadata.FlagKey)
	s.Equal(ruleKey1, impressionEvent1.Metadata.RuleKey)
	s.Equal("feature-test", impressionEvent1.Metadata.RuleType)
	s.Equal(variationKey1, impressionEvent1.Metadata.VariationKey)
	s.Equal(true, impressionEvent1.Metadata.Enabled)
	s.Equal("10390977673", impressionEvent1.ExperimentID)
	s.Equal("10389729780", impressionEvent1.VariationID)

	impressionEvent2 := eventMapping[1].Impression
	s.Equal(flagKey2, impressionEvent2.Metadata.FlagKey)
	s.Equal(ruleKey2, impressionEvent2.Metadata.RuleKey)
	s.Equal("feature-test", impressionEvent2.Metadata.RuleType)
	s.Equal(variationKey2, impressionEvent2.Metadata.VariationKey)
	s.Equal(true, impressionEvent2.Metadata.Enabled)
	s.Equal("10420810910", impressionEvent2.ExperimentID)
	s.Equal("10418551353", impressionEvent2.VariationID)
}

func (s *OptimizelyUserContextTestSuite) TestDecideAllFlags() {
	flagKey1 := "feature_1"
	flagKey2 := "feature_2"
	flagKey3 := "feature_3"
	variationKey1 := "a"
	variationKey2 := "variation_with_traffic"
	variationKey3 := ""
	ruleKey1 := "exp_with_audience"
	ruleKey2 := "exp_no_audience"
	ruleKey3 := ""

	variablesExpected1, err := s.OptimizelyClient.GetAllFeatureVariables(context.Background(), flagKey1, entities.UserContext{ID: s.userID})
	s.Nil(err)
	variablesExpected2, err := s.OptimizelyClient.GetAllFeatureVariables(context.Background(), flagKey2, entities.UserContext{ID: s.userID})
	s.Nil(err)
	variablesExpected3 := optimizelyjson.NewOptimizelyJSONfromMap(map[string]interface{}{})

	user := s.OptimizelyClient.CreateUserContext(s.userID, map[string]interface{}{"gender": "f"})
	decisions := user.DecideAll(context.Background(), nil)
	s.Len(decisions, 3)

	decision1 := decisions[flagKey1]
	s.Equal(variationKey1, decision1.VariationKey)
	s.Equal(true, decision1.Enabled)
	s.Equal(variablesExpected1.ToMap(), decision1.Variables.ToMap())
	s.Equal(ruleKey1, decision1.RuleKey)
	s.Equal(flagKey1, decision1.FlagKey)
	s.Equal(user, decision1.UserContext)
	s.Len(decision1.Reasons, 0)

	decision2 := decisions[flagKey2]
	s.Equal(variationKey2, decision2.VariationKey)
	s.Equal(true, decision2.Enabled)
	s.Equal(variablesExpected2.ToMap(), decision2.Variables.ToMap())
	s.Equal(ruleKey2, decision2.RuleKey)
	s.Equal(flagKey2, decision2.FlagKey)
	s.Equal(user, decision2.UserContext)
	s.Len(decision2.Reasons, 0)

	decision3 := decisions[flagKey3]
	s.Equal(variationKey3, decision3.VariationKey)
	s.Equal(false, decision3.Enabled)
	s.Equal(variablesExpected3.ToMap(), decision3.Variables.ToMap())
	s.Equal(ruleKey3, decision3.RuleKey)
	s.Equal(flagKey3, decision3.FlagKey)
	s.Equal(user, decision3.UserContext)
	s.Len(decision3.Reasons, 0)

	s.True(len(s.eventProcessor.Events) == 3)
	s.Equal(s.userID, s.eventProcessor.Events[0].VisitorID)

	eventMapping := map[int]event.UserEvent{}
	for _, event := range s.eventProcessor.Events {
		switch event.Impression.Metadata.FlagKey {
		case flagKey1:
			eventMapping[0] = event
		case flagKey2:
			eventMapping[1] = event
		case flagKey3:
			eventMapping[2] = event
		}
	}

	impressionEvent1 := eventMapping[0].Impression
	s.Equal(flagKey1, impressionEvent1.Metadata.FlagKey)
	s.Equal(ruleKey1, impressionEvent1.Metadata.RuleKey)
	s.Equal("feature-test", impressionEvent1.Metadata.RuleType)
	s.Equal(variationKey1, impressionEvent1.Metadata.VariationKey)
	s.Equal(true, impressionEvent1.Metadata.Enabled)
	s.Equal("10390977673", impressionEvent1.ExperimentID)
	s.Equal("10389729780", impressionEvent1.VariationID)

	impressionEvent2 := eventMapping[1].Impression
	s.Equal(flagKey2, impressionEvent2.Metadata.FlagKey)
	s.Equal(ruleKey2, impressionEvent2.Metadata.RuleKey)
	s.Equal("feature-test", impressionEvent2.Metadata.RuleType)
	s.Equal(variationKey2, impressionEvent2.Metadata.VariationKey)
	s.Equal(true, impressionEvent2.Metadata.Enabled)
	s.Equal("10420810910", impressionEvent2.ExperimentID)
	s.Equal("10418551353", impressionEvent2.VariationID)

	impressionEvent3 := eventMapping[2].Impression
	s.Equal(flagKey3, impressionEvent3.Metadata.FlagKey)
	s.Equal(ruleKey3, impressionEvent3.Metadata.RuleKey)
	s.Equal("rollout", impressionEvent3.Metadata.RuleType)
	s.Equal("", impressionEvent3.Metadata.VariationKey)
	s.Equal(false, impressionEvent3.Metadata.Enabled)
	s.Equal("", impressionEvent3.ExperimentID)
	s.Equal("", impressionEvent3.VariationID)
}

func (s *OptimizelyUserContextTestSuite) TestDecideAllEnabledFlagsOnly() {
	flagKey1 := "feature_1"
	variablesExpected1, err := s.OptimizelyClient.GetAllFeatureVariables(context.Background(), flagKey1, entities.UserContext{ID: s.userID})
	s.Nil(err)

	user := s.OptimizelyClient.CreateUserContext(s.userID, map[string]interface{}{"gender": "f"})
	decisions := user.DecideAll(context.Background(), []decide.OptimizelyDecideOptions{decide.EnabledFlagsOnly, decide.IncludeReasons})
	s.Len(decisions, 2)

	decision1 := decisions[flagKey1]
	s.Equal("a", decision1.VariationKey)
	s.Equal(true, decision1.Enabled)
	s.Equal(variablesExpected1.ToMap(), decision1.Variables.ToMap())
	s.Equal("exp_with_audience", decision1.RuleKey)
	s.Equal(flagKey1, decision1.FlagKey)
	s.Equal(user, decision1.UserContext)
	reasons := decision1.Reasons
	s.Len(reasons, 1)
	s.Equal(`Audiences for experiment exp_with_audience collectively evaluated to true.`, reasons[0])
}

func (s *OptimizelyUserContextTestSuite) TestTrackEvent() {
	eventKey := "event1"
	eventTags := map[string]interface{}{"name": "carrot"}
	attributes := map[string]interface{}{"gender": "f"}
	user := s.OptimizelyClient.CreateUserContext(s.userID, attributes)
	err := user.TrackEvent(eventKey, eventTags)
	s.Nil(err)
	s.True(len(s.eventProcessor.Events) == 1)
	s.Equal(s.userID, s.eventProcessor.Events[0].VisitorID)
	s.Equal(eventKey, s.eventProcessor.Events[0].Conversion.Key)
	s.Equal(eventTags, s.eventProcessor.Events[0].Conversion.Tags)
	s.Equal("gender", s.eventProcessor.Events[0].Conversion.Attributes[0].Key)
	s.Equal("f", s.eventProcessor.Events[0].Conversion.Attributes[0].Value)
}

func (s *OptimizelyUserContextTestSuite) TestTrackEventWithoutEventTags() {
	eventKey := "event1"
	attributes := map[string]interface{}{"gender": "f"}
	user := s.OptimizelyClient.CreateUserContext(s.userID, attributes)
	err := user.TrackEvent(eventKey, nil)
	s.Nil(err)
	s.True(len(s.eventProcessor.Events) == 1)
	s.Equal(s.userID, s.eventProcessor.Events[0].VisitorID)
	s.Equal(eventKey, s.eventProcessor.Events[0].Conversion.Key)
	s.Equal("gender", s.eventProcessor.Events[0].Conversion.Attributes[0].Key)
	s.Equal("f", s.eventProcessor.Events[0].Conversion.Attributes[0].Value)
}

func (s *OptimizelyUserContextTestSuite) TestTrackEventWithEmptyAttributes() {
	eventKey := "event1"
	eventTags := map[string]interface{}{"name": "carrot"}
	attributes := map[string]interface{}{}
	user := s.OptimizelyClient.CreateUserContext(s.userID, attributes)
	err := user.TrackEvent(eventKey, eventTags)
	s.Nil(err)
	s.True(len(s.eventProcessor.Events) == 1)
	s.Equal(s.userID, s.eventProcessor.Events[0].VisitorID)
	s.Equal(eventKey, s.eventProcessor.Events[0].Conversion.Key)
	s.Equal(eventTags, s.eventProcessor.Events[0].Conversion.Tags)
	s.Len(s.eventProcessor.Events[0].Conversion.Attributes, 1)
	s.Equal("$opt_bot_filtering", s.eventProcessor.Events[0].Conversion.Attributes[0].Key)
}

func (s *OptimizelyUserContextTestSuite) TestDecideSendEvent() {
	flagKey := "feature_2"
	experimentID := "10420810910"
	variationID := "10418551353"

	user := s.OptimizelyClient.CreateUserContext(s.userID, nil)
	decision := user.Decide(context.Background(), flagKey, nil)

	s.Equal("variation_with_traffic", decision.VariationKey)
	s.True(len(s.eventProcessor.Events) == 1)
	s.Equal(s.userID, s.eventProcessor.Events[0].VisitorID)
	s.Equal(experimentID, s.eventProcessor.Events[0].Impression.ExperimentID)
	s.Equal(variationID, s.eventProcessor.Events[0].Impression.VariationID)
}

func (s *OptimizelyUserContextTestSuite) TestDecideDoNotSendEvent() {
	flagKey := "feature_2"

	user := s.OptimizelyClient.CreateUserContext(s.userID, nil)
	decision := user.Decide(context.Background(), flagKey, []decide.OptimizelyDecideOptions{decide.DisableDecisionEvent})

	s.Equal("variation_with_traffic", decision.VariationKey)
	s.True(len(s.eventProcessor.Events) == 0)
}

func (s *OptimizelyUserContextTestSuite) TestDecisionNotification() {
	flagKey := "feature_2"
	variationKey := "variation_with_traffic"
	enabled := true
	variablesExpected, err := s.OptimizelyClient.GetAllFeatureVariables(context.Background(), flagKey, entities.UserContext{ID: s.userID})
	s.Nil(err)

	ruleKey := "exp_no_audience"
	reasons := []string{}
	attributes := map[string]interface{}{"gender": "f"}
	user := s.OptimizelyClient.CreateUserContext(s.userID, attributes)
	var receivedNotification notification.DecisionNotification
	callback := func(notification notification.DecisionNotification) {
		receivedNotification = notification
	}

	expectedDecisionInfo := map[string]interface{}{
		"flagKey":                 flagKey,
		"variationKey":            variationKey,
		"enabled":                 enabled,
		"variables":               variablesExpected.ToMap(),
		"ruleKey":                 ruleKey,
		"reasons":                 reasons,
		"decisionEventDispatched": true,
	}
	s.OptimizelyClient.DecisionService.OnDecision(callback)
	_ = user.Decide(context.Background(), flagKey, nil)

	s.Equal(notification.Flag, receivedNotification.Type)
	s.Equal(s.userID, receivedNotification.UserContext.ID)
	s.Equal(attributes, receivedNotification.UserContext.Attributes)
	s.Equal(expectedDecisionInfo, receivedNotification.DecisionInfo)

	receivedNotification = notification.DecisionNotification{}
	expectedDecisionInfo["decisionEventDispatched"] = false
	_ = user.Decide(context.Background(), flagKey, []decide.OptimizelyDecideOptions{decide.DisableDecisionEvent})
	s.Equal(expectedDecisionInfo, receivedNotification.DecisionInfo)
}

func (s *OptimizelyUserContextTestSuite) TestDecideOptionsBypassUps() {
	flagKey := "feature_2" // embedding experiment: "exp_no_audience"
	experimentID := "10420810910"
	variationID2 := "10418510624"
	variationKey1 := "variation_with_traffic"
	variationKey2 := "variation_no_traffic"
	options := []decide.OptimizelyDecideOptions{decide.IncludeReasons}

	userProfileService := new(MockUserProfileService)
	s.OptimizelyClient, _ = s.factory.Client(
		WithEventProcessor(s.eventProcessor),
		WithUserProfileService(userProfileService),
	)

	decisionKey := decision.NewUserDecisionKey(experimentID)
	savedUserProfile := decision.UserProfile{
		ID:                  s.userID,
		ExperimentBucketMap: map[decision.UserDecisionKey]string{decisionKey: variationID2},
	}
	userProfileService.On("Lookup", s.userID).Return(savedUserProfile)
	userProfileService.On("Save", mock.Anything)

	userContext := s.OptimizelyClient.CreateUserContext(s.userID, map[string]interface{}{})
	decision := userContext.Decide(context.Background(), flagKey, options)
	reasons := decision.Reasons
	s.Len(reasons, 1)
	s.Equal(`User "tester" was previously bucketed into variation "variation_no_traffic" of experiment "exp_no_audience".`, reasons[0])
	// should return variationId2 set by UPS
	s.Equal(variationKey2, decision.VariationKey)
	userProfileService.AssertCalled(s.T(), "Lookup", s.userID)
	userProfileService.AssertNotCalled(s.T(), "Save", mock.Anything)

	options = append(options, decide.IgnoreUserProfileService)
	decision = userContext.Decide(context.Background(), flagKey, options)
	reasons = decision.Reasons
	s.Len(reasons, 1)
	s.Equal(`Audiences for experiment exp_no_audience collectively evaluated to true.`, reasons[0])
	// should not lookup, ignore variationId2 set by UPS and return variationId1
	s.Equal(variationKey1, decision.VariationKey)
	userProfileService.AssertNumberOfCalls(s.T(), "Lookup", 1)
	// also should not save either
	userProfileService.AssertNotCalled(s.T(), "Save", mock.Anything)
}

func (s *OptimizelyUserContextTestSuite) TestDecideOptionsExcludeVariables() {
	flagKey := "feature_1"
	options := []decide.OptimizelyDecideOptions{}
	user := s.OptimizelyClient.CreateUserContext(s.userID, nil)

	decision := user.Decide(context.Background(), flagKey, options)
	s.True(len(decision.Variables.ToMap()) > 0)

	options = append(options, decide.ExcludeVariables)
	decision = user.Decide(context.Background(), flagKey, options)
	s.Len(decision.Variables.ToMap(), 0)
}

func (s *OptimizelyUserContextTestSuite) TestDecideOptionsIncludeReasons() {
	flagKey := "invalid_key"
	var options []decide.OptimizelyDecideOptions
	user := s.OptimizelyClient.CreateUserContext(s.userID, nil)

	// invalid flag key
	decision := user.Decide(context.Background(), flagKey, options)
	s.Len(decision.Reasons, 1)
	s.Equal(decide.GetDecideMessage(decide.FlagKeyInvalid, flagKey), decision.Reasons[0])

	// invalid flag key with includeReasons
	options = append(options, decide.IncludeReasons)
	decision = user.Decide(context.Background(), flagKey, options)
	s.Len(decision.Reasons, 1)
	s.Equal(decide.GetDecideMessage(decide.FlagKeyInvalid, flagKey), decision.Reasons[0])

	// valid flag key
	flagKey = "feature_1"
	decision = user.Decide(context.Background(), flagKey, options)
	reasons := decision.Reasons
	s.Len(reasons, 9)

	expectedLogs := []string{
		`an error occurred while evaluating nested tree for audience ID "13389141123"`,
		`Audiences for experiment exp_with_audience collectively evaluated to false.`,
		`User "tester" does not meet conditions to be in experiment "exp_with_audience".`,
		`an error occurred while evaluating nested tree for audience ID "13389130056"`,
		`User "tester" does not meet conditions for targeting rule 1.`,
		`an error occurred while evaluating nested tree for audience ID "12208130097"`,
		`User "tester" does not meet conditions for targeting rule 2.`,
		`Audiences for experiment 18322080788 collectively evaluated to true.`,
		`User "tester" meets conditions for targeting rule "Everyone Else".`,
	}

	for index, log := range expectedLogs {
		s.Equal(log, reasons[index])
	}
}

func (s *OptimizelyUserContextTestSuite) TestDefaultDecideOptionsExcludeVariables() {
	flagKey := "feature_1"
	options := []decide.OptimizelyDecideOptions{decide.ExcludeVariables}
	client, _ := s.factory.Client(WithEventProcessor(s.eventProcessor), WithDefaultDecideOptions(options))
	userContext := client.CreateUserContext(s.userID, nil)

	// should be excluded by DefaultDecideOption
	decision := userContext.Decide(context.Background(), flagKey, nil)
	s.Len(decision.Variables.ToMap(), 0)
	reasons := decision.Reasons
	s.Len(reasons, 0)

	options = append(options, decide.IncludeReasons)
	client, _ = s.factory.Client(WithEventProcessor(s.eventProcessor), WithDefaultDecideOptions(options))
	userContext = client.CreateUserContext(s.userID, nil)

	decision = userContext.Decide(context.Background(), flagKey, nil)
	reasons = decision.Reasons
	s.Len(reasons, 9)

	expectedLogs := []string{
		`an error occurred while evaluating nested tree for audience ID "13389141123"`,
		`Audiences for experiment exp_with_audience collectively evaluated to false.`,
		`User "tester" does not meet conditions to be in experiment "exp_with_audience".`,
		`an error occurred while evaluating nested tree for audience ID "13389130056"`,
		`User "tester" does not meet conditions for targeting rule 1.`,
		`an error occurred while evaluating nested tree for audience ID "12208130097"`,
		`User "tester" does not meet conditions for targeting rule 2.`,
		`Audiences for experiment 18322080788 collectively evaluated to true.`,
		`User "tester" meets conditions for targeting rule "Everyone Else".`,
	}

	for index, log := range expectedLogs {
		s.Equal(log, reasons[index])
	}
}

func (s *OptimizelyUserContextTestSuite) TestDefaultDecideOptionsEnabledFlagsOnly() {
	flagKey := "feature_1"
	variablesExpected, _ := s.OptimizelyClient.GetAllFeatureVariables(context.Background(), flagKey, entities.UserContext{ID: s.userID})
	options := []decide.OptimizelyDecideOptions{decide.EnabledFlagsOnly, decide.IncludeReasons}
	client, _ := s.factory.Client(WithEventProcessor(s.eventProcessor), WithDefaultDecideOptions(options))
	user := client.CreateUserContext(s.userID, map[string]interface{}{"gender": "f"})

	// should get EnabledFlagsOnly by DefaultDecideOption
	decisions := user.DecideAll(context.Background(), nil)
	s.Len(decisions, 2)

	decision1 := decisions[flagKey]
	s.Equal("a", decision1.VariationKey)
	s.Equal(true, decision1.Enabled)
	s.Equal(variablesExpected.ToMap(), decision1.Variables.ToMap())
	s.Equal("exp_with_audience", decision1.RuleKey)
	s.Equal(flagKey, decision1.FlagKey)
	s.Equal(user, decision1.UserContext)
	reasons := decision1.Reasons
	s.Len(reasons, 1)
	s.Equal("Audiences for experiment exp_with_audience collectively evaluated to true.", reasons[0])
}

func (s *OptimizelyUserContextTestSuite) TestDefaultDecideOptionsIncludeReasons() {
	flagKey := "invalid_key"
	options := []decide.OptimizelyDecideOptions{decide.IncludeReasons}
	client, _ := s.factory.Client(WithEventProcessor(s.eventProcessor), WithDefaultDecideOptions(options))
	user := client.CreateUserContext(s.userID, nil)

	// should get IncludeReasons by DefaultDecideOption
	decision := user.Decide(context.Background(), flagKey, options)
	s.Len(decision.Reasons, 1)
	s.Equal(decide.GetDecideMessage(decide.FlagKeyInvalid, flagKey), decision.Reasons[0])
}

func (s *OptimizelyUserContextTestSuite) TestDefaultDecideOptionsBypassUps() {
	flagKey := "feature_2" // embedding experiment: "exp_no_audience"
	experimentID := "10420810910"
	variationID2 := "10418510624"
	variationKey1 := "variation_with_traffic"

	userProfileService := new(MockUserProfileService)
	s.OptimizelyClient, _ = s.factory.Client(
		WithEventProcessor(s.eventProcessor),
		WithUserProfileService(userProfileService),
	)

	decisionKey := decision.NewUserDecisionKey(experimentID)
	savedUserProfile := decision.UserProfile{
		ID:                  s.userID,
		ExperimentBucketMap: map[decision.UserDecisionKey]string{decisionKey: variationID2},
	}
	userProfileService.On("Lookup", s.userID).Return(savedUserProfile)
	userProfileService.On("Save", mock.Anything)

	options := []decide.OptimizelyDecideOptions{decide.IgnoreUserProfileService}
	client, _ := s.factory.Client(WithEventProcessor(s.eventProcessor), WithDefaultDecideOptions(options))
	user := client.CreateUserContext(s.userID, nil)
	decision := user.Decide(context.Background(), flagKey, []decide.OptimizelyDecideOptions{decide.IncludeReasons})
	s.Len(decision.Reasons, 1)

	// should get IgnoreUserProfileService by DefaultDecideOption
	// should not lookup, ignore variationId2 set by UPS and return variationId1
	s.Equal(variationKey1, decision.VariationKey)
	userProfileService.AssertNotCalled(s.T(), "Lookup", s.userID)
	userProfileService.AssertNotCalled(s.T(), "Save", mock.Anything)
}

func (s *OptimizelyUserContextTestSuite) TestGetAllOptionsUsesOrOperator() {
	options1 := []decide.OptimizelyDecideOptions{
		decide.DisableDecisionEvent,
		decide.EnabledFlagsOnly,
		decide.IgnoreUserProfileService,
		decide.IncludeReasons,
		decide.ExcludeVariables,
	}
	client, _ := s.factory.Client(WithDefaultDecideOptions(options1))
	// Pass all false options
	options2 := client.getAllOptions(&decide.Options{})

	s.Equal(decide.Options{
		DisableDecisionEvent:     true,
		EnabledFlagsOnly:         true,
		IgnoreUserProfileService: true,
		IncludeReasons:           true,
		ExcludeVariables:         true,
	}, options2)
}

func (s *OptimizelyUserContextTestSuite) TestDecideSDKNotReady() {
	flagKey := "feature_1"
	factory := OptimizelyFactory{SDKKey: "121"}
	client, _ := factory.Client()
	userContext := client.CreateUserContext(s.userID, nil)
	decision := userContext.Decide(context.Background(), flagKey, nil)

	s.Equal("", decision.VariationKey)
	s.False(decision.Enabled)
	s.Len(decision.Variables.ToMap(), 0)
	s.Equal(decision.FlagKey, flagKey)
	s.Equal(decision.UserContext, userContext)
	s.Len(decision.Reasons, 1)
	s.Equal(decide.GetDecideMessage(decide.SDKNotReady), decision.Reasons[0])
}

func (s *OptimizelyUserContextTestSuite) TestDecideInvalidFeatureKey() {
	flagKey := "invalid_key"
	userContext := s.OptimizelyClient.CreateUserContext(s.userID, nil)
	decision := userContext.Decide(context.Background(), flagKey, []decide.OptimizelyDecideOptions{decide.IncludeReasons})

	s.Equal("", decision.VariationKey)
	s.False(decision.Enabled)
	s.Len(decision.Variables.ToMap(), 0)
	s.Equal(decision.FlagKey, flagKey)
	s.Equal(decision.UserContext, userContext)
	s.Len(decision.Reasons, 1)
	s.Equal(decide.GetDecideMessage(decide.FlagKeyInvalid, flagKey), decision.Reasons[0])
}

func (s *OptimizelyUserContextTestSuite) TestDecideForKeySDKNotReady() {
	flagKeys := []string{"feature_1"}
	factory := OptimizelyFactory{SDKKey: "121"}
	client, _ := factory.Client()
	userContext := client.CreateUserContext(s.userID, nil)
	decisions := userContext.DecideForKeys(context.Background(), flagKeys, []decide.OptimizelyDecideOptions{decide.IncludeReasons})

	s.Len(decisions, 0)
}

func (s *OptimizelyUserContextTestSuite) TestDecideAllSDKNotReady() {
	factory := OptimizelyFactory{SDKKey: "121"}
	client, _ := factory.Client()
	userContext := client.CreateUserContext(s.userID, nil)
	decisions := userContext.DecideAll(context.Background(), nil)

	s.Len(decisions, 0)
}

func (s *OptimizelyUserContextTestSuite) TestDecideForKeysErrorDecisionIncluded() {
	flagKey1 := "feature_2"
	flagKey2 := "invalid_key"
	flagKeys := []string{flagKey1, flagKey2}
	variablesExpected1, err := s.OptimizelyClient.GetAllFeatureVariables(context.Background(), flagKey1, entities.UserContext{ID: s.userID})
	s.Nil(err)

	user := s.OptimizelyClient.CreateUserContext(s.userID, nil)
	decisions := user.DecideForKeys(context.Background(), flagKeys, []decide.OptimizelyDecideOptions{decide.IncludeReasons})
	s.Len(decisions, 2)

	decision := decisions[flagKey1]
	s.Equal("variation_with_traffic", decision.VariationKey)
	s.Equal(true, decision.Enabled)
	s.Equal(variablesExpected1.ToMap(), decision.Variables.ToMap())
	s.Equal("exp_no_audience", decision.RuleKey)
	s.Equal(flagKey1, decision.FlagKey)
	s.Equal(user, decision.UserContext)
	reasons := decision.Reasons
	s.Len(reasons, 1)
	s.Equal(`Audiences for experiment exp_no_audience collectively evaluated to true.`, reasons[0])

	decision = decisions[flagKey2]
	s.Equal(flagKey2, decision.FlagKey)
	s.Equal(user, decision.UserContext)
	reasons = decision.Reasons
	s.Len(reasons, 1)
	s.Equal(decide.GetDecideMessage(decide.FlagKeyInvalid, flagKey2), reasons[0])
}

func (s *OptimizelyUserContextTestSuite) TestForcedDecisionWithNilConfig() {
	s.OptimizelyClient.ConfigManager = nil

	flagKeyA := "feature_1"
	ruleKey := ""
	variationKeyA := "a"

	// checking with nil forcedDecisionService
	user := s.OptimizelyClient.CreateUserContext(s.userID, nil)
	s.Nil(user.forcedDecisionService)

	s.True(user.SetForcedDecision(decision.OptimizelyDecisionContext{FlagKey: flagKeyA, RuleKey: ruleKey}, decision.OptimizelyForcedDecision{VariationKey: variationKeyA}))
	s.NotNil(user.forcedDecisionService)

	forcedDecision, err := user.GetForcedDecision(decision.OptimizelyDecisionContext{FlagKey: flagKeyA, RuleKey: ruleKey})
	s.Equal(variationKeyA, forcedDecision.VariationKey)
	s.NoError(err)
	s.True(user.RemoveForcedDecision(decision.OptimizelyDecisionContext{FlagKey: flagKeyA, RuleKey: ruleKey}))
	s.True(user.RemoveAllForcedDecisions())
}

func (s *OptimizelyUserContextTestSuite) TestForcedDecision() {
	flagKeyA := "feature_1"
	flagKeyB := "feature_2"
	ruleKey := ""
	variationKeyA := "a"
	variationKeyB := "b"

	// checking with nil forcedDecisionService
	user := s.OptimizelyClient.CreateUserContext(s.userID, nil)
	s.Nil(user.forcedDecisionService)
	forcedDecision, err := user.GetForcedDecision(decision.OptimizelyDecisionContext{FlagKey: flagKeyA, RuleKey: ruleKey})
	s.Equal("", forcedDecision.VariationKey)
	s.Error(err)
	s.False(user.RemoveForcedDecision(decision.OptimizelyDecisionContext{FlagKey: flagKeyA, RuleKey: ruleKey}))
	s.True(user.RemoveAllForcedDecisions())

	// checking if forcedDecisionService was created using SetForcedDecision
	s.True(user.SetForcedDecision(decision.OptimizelyDecisionContext{FlagKey: flagKeyA, RuleKey: ruleKey}, decision.OptimizelyForcedDecision{VariationKey: variationKeyA}))
	s.NotNil(user.forcedDecisionService)

	s.True(user.SetForcedDecision(decision.OptimizelyDecisionContext{FlagKey: flagKeyB, RuleKey: ruleKey}, decision.OptimizelyForcedDecision{VariationKey: variationKeyB}))
	forcedDecision, err = user.GetForcedDecision(decision.OptimizelyDecisionContext{FlagKey: flagKeyA, RuleKey: ruleKey})
	s.Equal(variationKeyA, forcedDecision.VariationKey)
	s.NoError(err)
	forcedDecision, err = user.GetForcedDecision(decision.OptimizelyDecisionContext{FlagKey: flagKeyB, RuleKey: ruleKey})
	s.Equal(variationKeyB, forcedDecision.VariationKey)
	s.NoError(err)

	s.True(user.RemoveForcedDecision(decision.OptimizelyDecisionContext{FlagKey: flagKeyA, RuleKey: ruleKey}))
	forcedDecision, err = user.GetForcedDecision(decision.OptimizelyDecisionContext{FlagKey: flagKeyA, RuleKey: ruleKey})
	s.Equal("", forcedDecision.VariationKey)
	s.NoError(err)
	forcedDecision, err = user.GetForcedDecision(decision.OptimizelyDecisionContext{FlagKey: flagKeyB, RuleKey: ruleKey})
	s.Equal(variationKeyB, forcedDecision.VariationKey)
	s.NoError(err)

	s.True(user.RemoveAllForcedDecisions())
	forcedDecision, err = user.GetForcedDecision(decision.OptimizelyDecisionContext{FlagKey: flagKeyA, RuleKey: ruleKey})
	s.Equal("", forcedDecision.VariationKey)
	s.Error(err)
	forcedDecision, err = user.GetForcedDecision(decision.OptimizelyDecisionContext{FlagKey: flagKeyB, RuleKey: ruleKey})
	s.Equal("", forcedDecision.VariationKey)
	s.Error(err)
}

func TestOptimizelyUserContextTestSuite(t *testing.T) {
	suite.Run(t, new(OptimizelyUserContextTestSuite))
}
