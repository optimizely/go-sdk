/****************************************************************************
 * Copyright 2020, Optimizely, Inc. and contributors                        *
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

package client

import (
	"io/ioutil"
	"path/filepath"
	"sync"
	"testing"

	"github.com/optimizely/go-sdk/pkg/decide"
	"github.com/optimizely/go-sdk/pkg/decision"
	"github.com/optimizely/go-sdk/pkg/entities"
	"github.com/optimizely/go-sdk/pkg/notification"
	"github.com/optimizely/go-sdk/pkg/optimizelyjson"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
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
		datafile, _ = ioutil.ReadFile(absPath)
	})
	s.eventProcessor = new(MockProcessor)
	s.eventProcessor.On("ProcessEvent", mock.AnythingOfType("event.UserEvent")).Return(true)
	s.factory = OptimizelyFactory{Datafile: datafile}
	s.OptimizelyClient, _ = s.factory.Client(WithEventProcessor(s.eventProcessor))
	s.userID = "tester"
}

func (s *OptimizelyUserContextTestSuite) TestOptimizelyUserContextWithAttributes() {
	attributes := map[string]interface{}{"key1": 1212, "key2": 1213}
	optimizelyUserContext := newOptimizelyUserContext(s.OptimizelyClient, s.userID, attributes)

	s.Equal(s.OptimizelyClient, optimizelyUserContext.GetOptimizely())
	s.Equal(s.userID, optimizelyUserContext.GetUserID())
	s.Equal(attributes, optimizelyUserContext.GetUserAttributes())
}

func (s *OptimizelyUserContextTestSuite) TestOptimizelyUserContextNoAttributes() {
	var attributes map[string]interface{}
	optimizelyUserContext := newOptimizelyUserContext(s.OptimizelyClient, s.userID, attributes)

	s.Equal(s.OptimizelyClient, optimizelyUserContext.GetOptimizely())
	s.Equal(s.userID, optimizelyUserContext.GetUserID())
	s.Equal(attributes, optimizelyUserContext.GetUserAttributes())
}

func (s *OptimizelyUserContextTestSuite) TestUpatingProvidedUserContextHasNoImpactOnOptimizelyUserContext() {
	attributes := map[string]interface{}{"k1": "v1", "k2": false}

	userContext := entities.UserContext{ID: s.userID, Attributes: attributes}
	optimizelyUserContext := newOptimizelyUserContext(s.OptimizelyClient, s.userID, attributes)

	s.Equal(s.OptimizelyClient, optimizelyUserContext.GetOptimizely())
	s.Equal(s.userID, optimizelyUserContext.GetUserID())
	s.Equal(attributes, optimizelyUserContext.GetUserAttributes())

	userContext.Attributes["k1"] = "v2"
	userContext.Attributes["k2"] = true

	s.Equal("v1", optimizelyUserContext.GetUserAttributes()["k1"])
	s.Equal(false, optimizelyUserContext.GetUserAttributes()["k2"])

	attributes = optimizelyUserContext.GetUserAttributes()
	attributes["k1"] = "v2"
	attributes["k2"] = true

	s.Equal("v1", optimizelyUserContext.GetUserAttributes()["k1"])
	s.Equal(false, optimizelyUserContext.GetUserAttributes()["k2"])
}

func (s *OptimizelyUserContextTestSuite) TestSetAttribute() {
	userID := "1212121"
	var attributes map[string]interface{}

	optimizelyUserContext := newOptimizelyUserContext(s.OptimizelyClient, userID, attributes)
	s.Equal(s.OptimizelyClient, optimizelyUserContext.GetOptimizely())

	var wg sync.WaitGroup
	wg.Add(4)
	addInsideGoRoutine := func(key string, value interface{}, wg *sync.WaitGroup) {
		optimizelyUserContext.SetAttribute(key, value)
		wg.Done()
	}

	go addInsideGoRoutine("k1", "v1", &wg)
	go addInsideGoRoutine("k2", true, &wg)
	go addInsideGoRoutine("k3", 100, &wg)
	go addInsideGoRoutine("k4", 3.5, &wg)
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
	optimizelyUserContext := newOptimizelyUserContext(s.OptimizelyClient, userID, attributes)

	s.Equal(s.OptimizelyClient, optimizelyUserContext.GetOptimizely())
	s.Equal(userID, optimizelyUserContext.GetUserID())
	s.Equal(attributes, optimizelyUserContext.GetUserAttributes())

	optimizelyUserContext.SetAttribute("k1", "v2")
	optimizelyUserContext.SetAttribute("k2", true)

	s.Equal("v2", optimizelyUserContext.GetUserAttributes()["k1"])
	s.Equal(true, optimizelyUserContext.GetUserAttributes()["k2"])
}

func (s *OptimizelyUserContextTestSuite) TestDecide() {
	flagKey := "feature_2"
	variablesExpected, err := s.OptimizelyClient.GetAllFeatureVariables(flagKey, entities.UserContext{ID: s.userID})
	s.Nil(err)

	var options decide.OptimizelyDecideOptions
	user := s.OptimizelyClient.CreateUserContext(s.userID, nil)
	decision := user.Decide(flagKey, options)

	s.Equal("variation_with_traffic", decision.GetVariationKey())
	s.Equal(true, decision.GetEnabled())
	s.Equal(variablesExpected.ToMap(), decision.GetVariables().ToMap())
	s.Equal("exp_no_audience", decision.GetRuleKey())
	s.Equal(flagKey, decision.GetFlagKey())
	s.Equal(user, decision.GetUserContext())
	s.Len(decision.GetReasons(), 0)
}

func (s *OptimizelyUserContextTestSuite) TestDecideForKeyWithOneFlag() {
	flagKey := "feature_2"
	flagKeys := []string{flagKey}
	variablesExpected, err := s.OptimizelyClient.GetAllFeatureVariables(flagKey, entities.UserContext{ID: s.userID})
	s.Nil(err)

	var options decide.OptimizelyDecideOptions
	user := s.OptimizelyClient.CreateUserContext(s.userID, nil)
	decisions := user.DecideForKeys(flagKeys, options)
	s.Len(decisions, 1)

	decision := decisions[flagKey]
	s.Equal("variation_with_traffic", decision.GetVariationKey())
	s.Equal(true, decision.GetEnabled())
	s.Equal(variablesExpected.ToMap(), decision.GetVariables().ToMap())
	s.Equal("exp_no_audience", decision.GetRuleKey())
	s.Equal(flagKey, decision.GetFlagKey())
	s.Equal(user, decision.GetUserContext())
	s.Len(decision.GetReasons(), 0)
}

func (s *OptimizelyUserContextTestSuite) TestDecideForKeysWithMultipleFlags() {
	flagKey1 := "feature_1"
	flagKey2 := "feature_2"
	flagKeys := []string{flagKey1, flagKey2}
	variablesExpected1, err := s.OptimizelyClient.GetAllFeatureVariables(flagKey1, entities.UserContext{ID: s.userID})
	s.Nil(err)
	variablesExpected2, err := s.OptimizelyClient.GetAllFeatureVariables(flagKey2, entities.UserContext{ID: s.userID})
	s.Nil(err)

	var options decide.OptimizelyDecideOptions
	user := s.OptimizelyClient.CreateUserContext(s.userID, map[string]interface{}{"gender": "f"})
	decisions := user.DecideForKeys(flagKeys, options)
	s.Len(decisions, 2)

	decision1 := decisions[flagKey1]
	s.Equal("a", decision1.GetVariationKey())
	s.Equal(true, decision1.GetEnabled())
	s.Equal(variablesExpected1.ToMap(), decision1.GetVariables().ToMap())
	s.Equal("exp_with_audience", decision1.GetRuleKey())
	s.Equal(flagKey1, decision1.GetFlagKey())
	s.Equal(user, decision1.GetUserContext())
	s.Len(decision1.GetReasons(), 0)

	decision2 := decisions[flagKey2]
	s.Equal("variation_with_traffic", decision2.GetVariationKey())
	s.Equal(true, decision2.GetEnabled())
	s.Equal(variablesExpected2.ToMap(), decision2.GetVariables().ToMap())
	s.Equal("exp_no_audience", decision2.GetRuleKey())
	s.Equal(flagKey2, decision2.GetFlagKey())
	s.Equal(user, decision2.GetUserContext())
	s.Len(decision2.GetReasons(), 0)
}

func (s *OptimizelyUserContextTestSuite) TestDecideAllFlags() {
	flagKey1 := "feature_1"
	flagKey2 := "feature_2"
	flagKey3 := "feature_3"
	variablesExpected1, err := s.OptimizelyClient.GetAllFeatureVariables(flagKey1, entities.UserContext{ID: s.userID})
	s.Nil(err)
	variablesExpected2, err := s.OptimizelyClient.GetAllFeatureVariables(flagKey2, entities.UserContext{ID: s.userID})
	s.Nil(err)
	variablesExpected3 := optimizelyjson.NewOptimizelyJSONfromMap(map[string]interface{}{})

	var options decide.OptimizelyDecideOptions
	user := s.OptimizelyClient.CreateUserContext(s.userID, map[string]interface{}{"gender": "f"})
	decisions := user.DecideAll(options)
	s.Len(decisions, 3)

	decision1 := decisions[flagKey1]
	s.Equal("a", decision1.GetVariationKey())
	s.Equal(true, decision1.GetEnabled())
	s.Equal(variablesExpected1.ToMap(), decision1.GetVariables().ToMap())
	s.Equal("exp_with_audience", decision1.GetRuleKey())
	s.Equal(flagKey1, decision1.GetFlagKey())
	s.Equal(user, decision1.GetUserContext())
	s.Len(decision1.GetReasons(), 0)

	decision2 := decisions[flagKey2]
	s.Equal("variation_with_traffic", decision2.GetVariationKey())
	s.Equal(true, decision2.GetEnabled())
	s.Equal(variablesExpected2.ToMap(), decision2.GetVariables().ToMap())
	s.Equal("exp_no_audience", decision2.GetRuleKey())
	s.Equal(flagKey2, decision2.GetFlagKey())
	s.Equal(user, decision2.GetUserContext())
	s.Len(decision2.GetReasons(), 0)

	decision3 := decisions[flagKey3]
	s.Equal("", decision3.GetVariationKey())
	s.Equal(false, decision3.GetEnabled())
	s.Equal(variablesExpected3.ToMap(), decision3.GetVariables().ToMap())
	s.Equal("", decision3.GetRuleKey())
	s.Equal(flagKey3, decision3.GetFlagKey())
	s.Equal(user, decision3.GetUserContext())
	s.Len(decision3.GetReasons(), 0)
}

func (s *OptimizelyUserContextTestSuite) TestDecideAllEnabledFlagsOnly() {
	flagKey1 := "feature_1"
	variablesExpected1, err := s.OptimizelyClient.GetAllFeatureVariables(flagKey1, entities.UserContext{ID: s.userID})
	s.Nil(err)

	user := s.OptimizelyClient.CreateUserContext(s.userID, map[string]interface{}{"gender": "f"})
	decisions := user.DecideAll(decide.OptimizelyDecideOptions{
		EnabledFlagsOnly: true,
	})
	s.Len(decisions, 2)

	decision1 := decisions[flagKey1]
	s.Equal("a", decision1.GetVariationKey())
	s.Equal(true, decision1.GetEnabled())
	s.Equal(variablesExpected1.ToMap(), decision1.GetVariables().ToMap())
	s.Equal("exp_with_audience", decision1.GetRuleKey())
	s.Equal(flagKey1, decision1.GetFlagKey())
	s.Equal(user, decision1.GetUserContext())
	s.Len(decision1.GetReasons(), 0)
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
	var options decide.OptimizelyDecideOptions

	user := s.OptimizelyClient.CreateUserContext(s.userID, nil)
	decision := user.Decide(flagKey, options)

	s.Equal("variation_with_traffic", decision.GetVariationKey())
	s.True(len(s.eventProcessor.Events) == 1)
	s.Equal(s.userID, s.eventProcessor.Events[0].VisitorID)
	s.Equal(experimentID, s.eventProcessor.Events[0].Impression.ExperimentID)
	s.Equal(variationID, s.eventProcessor.Events[0].Impression.VariationID)
}

func (s *OptimizelyUserContextTestSuite) TestDecideDoNotSendEvent() {
	flagKey := "feature_2"

	user := s.OptimizelyClient.CreateUserContext(s.userID, nil)
	decision := user.Decide(flagKey, decide.OptimizelyDecideOptions{DisableDecisionEvent: true})

	s.Equal("variation_with_traffic", decision.GetVariationKey())
	s.True(len(s.eventProcessor.Events) == 0)
}

func (s *OptimizelyUserContextTestSuite) TestDecisionNotification() {
	flagKey := "feature_2"
	variationKey := "variation_with_traffic"
	enabled := true
	variablesExpected, err := s.OptimizelyClient.GetAllFeatureVariables(flagKey, entities.UserContext{ID: s.userID})
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
	_ = user.Decide(flagKey, decide.OptimizelyDecideOptions{})

	s.Equal(notification.Flag, receivedNotification.Type)
	s.Equal(s.userID, receivedNotification.UserContext.ID)
	s.Equal(attributes, receivedNotification.UserContext.Attributes)
	s.Equal(expectedDecisionInfo, receivedNotification.DecisionInfo)

	receivedNotification = notification.DecisionNotification{}
	expectedDecisionInfo["decisionEventDispatched"] = false
	_ = user.Decide(flagKey, decide.OptimizelyDecideOptions{DisableDecisionEvent: true})
	s.Equal(expectedDecisionInfo, receivedNotification.DecisionInfo)
}

func (s *OptimizelyUserContextTestSuite) TestDecideOptionsBypassUps() {
	flagKey := "feature_2" // embedding experiment: "exp_no_audience"
	experimentID := "10420810910"
	variationID2 := "10418510624"
	variationKey1 := "variation_with_traffic"
	variationKey2 := "variation_no_traffic"

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
	decision := userContext.Decide(flagKey, decide.OptimizelyDecideOptions{})
	// should return variationId2 set by UPS
	s.Equal(variationKey2, decision.GetVariationKey())

	decision = userContext.Decide(flagKey, decide.OptimizelyDecideOptions{IgnoreUserProfileService: true})
	// should ignore variationId2 set by UPS and return variationId1
	s.Equal(variationKey1, decision.GetVariationKey())
	// also should not save either
	userProfileService.AssertNotCalled(s.T(), "Save", mock.Anything)
}

func (s *OptimizelyUserContextTestSuite) TestDecideOptionsExcludeVariables() {
	flagKey := "feature_1"
	var options decide.OptimizelyDecideOptions
	user := s.OptimizelyClient.CreateUserContext(s.userID, nil)

	decision := user.Decide(flagKey, options)
	s.True(len(decision.GetVariables().ToMap()) > 0)

	options.ExcludeVariables = true
	decision = user.Decide(flagKey, options)
	s.Len(decision.GetVariables().ToMap(), 0)
}

func (s *OptimizelyUserContextTestSuite) TestDecideOptionsIncludeReasons() {
	flagKey := "invalid_key"
	var options decide.OptimizelyDecideOptions
	user := s.OptimizelyClient.CreateUserContext(s.userID, nil)

	// invalid flag key
	decision := user.Decide(flagKey, options)
	s.Len(decision.GetReasons(), 1)
	s.Equal(decide.GetDecideMessage(decide.FlagKeyInvalid, flagKey), decision.GetReasons()[0])

	// invalid flag key with includeReasons
	options.IncludeReasons = true
	decision = user.Decide(flagKey, options)
	s.Len(decision.GetReasons(), 1)
	s.Equal(decide.GetDecideMessage(decide.FlagKeyInvalid, flagKey), decision.GetReasons()[0])

	// valid flag key
	options.IncludeReasons = false
	flagKey = "feature_1"
	decision = user.Decide(flagKey, options)
	s.Len(decision.GetReasons(), 0)
}

func (s *OptimizelyUserContextTestSuite) TestDefaultDecideOptions() {
	flagKey := "feature_1"
	options := decide.OptimizelyDecideOptions{ExcludeVariables: true}
	client, _ := s.factory.Client(WithEventProcessor(s.eventProcessor), WithDefaultDecideOptions(options))
	userContext := client.CreateUserContext(s.userID, nil)

	// should be excluded by DefaultDecideOption
	decision := userContext.Decide(flagKey, decide.OptimizelyDecideOptions{})
	s.Len(decision.GetVariables().ToMap(), 0)
}

func (s *OptimizelyUserContextTestSuite) TestDecideSDKNotReady() {
	flagKey := "feature_1"
	factory := OptimizelyFactory{SDKKey: "121"}
	client, _ := factory.Client()
	userContext := client.CreateUserContext(s.userID, nil)
	decision := userContext.Decide(flagKey, decide.OptimizelyDecideOptions{})

	s.Equal("", decision.GetVariationKey())
	s.False(decision.GetEnabled())
	s.Len(decision.GetVariables().ToMap(), 0)
	s.Equal(decision.GetFlagKey(), flagKey)
	s.Equal(decision.GetUserContext(), userContext)
	s.Len(decision.GetReasons(), 1)
	s.Equal(decide.GetDecideMessage(decide.SDKNotReady), decision.GetReasons()[0])
}

func (s *OptimizelyUserContextTestSuite) TestDecideInvalidFeatureKey() {
	flagKey := "invalid_key"
	userContext := s.OptimizelyClient.CreateUserContext(s.userID, nil)
	decision := userContext.Decide(flagKey, decide.OptimizelyDecideOptions{})

	s.Equal("", decision.GetVariationKey())
	s.False(decision.GetEnabled())
	s.Len(decision.GetVariables().ToMap(), 0)
	s.Equal(decision.GetFlagKey(), flagKey)
	s.Equal(decision.GetUserContext(), userContext)
	s.Len(decision.GetReasons(), 1)
	s.Equal(decide.GetDecideMessage(decide.FlagKeyInvalid, flagKey), decision.GetReasons()[0])
}

func (s *OptimizelyUserContextTestSuite) TestDecideForKeySDKNotReady() {
	flagKeys := []string{"feature_1"}
	factory := OptimizelyFactory{SDKKey: "121"}
	client, _ := factory.Client()
	userContext := client.CreateUserContext(s.userID, nil)
	decisions := userContext.DecideForKeys(flagKeys, decide.OptimizelyDecideOptions{})

	s.Len(decisions, 0)
}

func (s *OptimizelyUserContextTestSuite) TestDecideAllSDKNotReady() {
	factory := OptimizelyFactory{SDKKey: "121"}
	client, _ := factory.Client()
	userContext := client.CreateUserContext(s.userID, nil)
	decisions := userContext.DecideAll(decide.OptimizelyDecideOptions{})

	s.Len(decisions, 0)
}

func (s *OptimizelyUserContextTestSuite) TestDecideForKeysErrorDecisionIncluded() {
	flagKey1 := "feature_2"
	flagKey2 := "invalid_key"
	flagKeys := []string{flagKey1, flagKey2}
	variablesExpected1, err := s.OptimizelyClient.GetAllFeatureVariables(flagKey1, entities.UserContext{ID: s.userID})
	s.Nil(err)

	var options decide.OptimizelyDecideOptions
	user := s.OptimizelyClient.CreateUserContext(s.userID, nil)
	decisions := user.DecideForKeys(flagKeys, options)
	s.Len(decisions, 2)

	decision := decisions[flagKey1]
	s.Equal("variation_with_traffic", decision.GetVariationKey())
	s.Equal(true, decision.GetEnabled())
	s.Equal(variablesExpected1.ToMap(), decision.GetVariables().ToMap())
	s.Equal("exp_no_audience", decision.GetRuleKey())
	s.Equal(flagKey1, decision.GetFlagKey())
	s.Equal(user, decision.GetUserContext())
	s.Len(decision.GetReasons(), 0)

	decision = decisions[flagKey2]
	s.Equal(flagKey2, decision.GetFlagKey())
	s.Equal(user, decision.GetUserContext())
	s.Len(decision.GetReasons(), 1)
	s.Equal(decide.GetDecideMessage(decide.FlagKeyInvalid, flagKey2), decision.GetReasons()[0])
}

func TestOptimizelyUserContextTestSuite(t *testing.T) {
	suite.Run(t, new(OptimizelyUserContextTestSuite))
}
