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
	"errors"
	"testing"

	"github.com/optimizely/go-sdk/pkg/optimizelyjson"

	"github.com/stretchr/testify/suite"
)

type OptimizelyDecisionTestSuite struct {
	suite.Suite
	*OptimizelyClient
}

func (s *OptimizelyDecisionTestSuite) SetupTest() {
	factory := OptimizelyFactory{SDKKey: "1212"}
	s.OptimizelyClient, _ = factory.Client()
}

func (s *OptimizelyDecisionTestSuite) TestOptimizelyDecision() {
	variationKey := "var1"
	enabled := true
	variables, _ := optimizelyjson.NewOptimizelyJSONfromString(`{"k1":"v1"}`)
	var ruleKey string
	flagKey := "flag1"
	reasons := []string{}
	userID := "testUser1"
	attributes := map[string]interface{}{"key": 1212}

	optimizelyUserContext := s.OptimizelyClient.CreateUserContext(userID, attributes)
	decision := NewOptimizelyDecision(variationKey, ruleKey, flagKey, enabled, variables, optimizelyUserContext, reasons)

	s.Equal(variationKey, decision.GetVariationKey())
	s.Equal(enabled, decision.GetEnabled())
	s.Equal(variables, decision.GetVariables())
	s.Equal(ruleKey, decision.GetRuleKey())
	s.Equal(flagKey, decision.GetFlagKey())
	s.Equal(reasons, decision.GetReasons())
	s.Equal(optimizelyUserContext, decision.GetUserContext())
}

func (s *OptimizelyDecisionTestSuite) TestNewErrorDecision() {
	flagKey := "flag1"
	errorString := "SDK has an error"
	userID := "testUser1"
	attributes := map[string]interface{}{"key": 1212}
	optimizelyUserContext := s.OptimizelyClient.CreateUserContext(userID, attributes)
	decision := NewErrorDecision(flagKey, optimizelyUserContext, errors.New(errorString))

	s.Equal("", decision.GetVariationKey())
	s.Equal(false, decision.GetEnabled())
	s.Equal(&optimizelyjson.OptimizelyJSON{}, decision.GetVariables())
	s.Equal("", decision.GetRuleKey())
	s.Equal(flagKey, decision.GetFlagKey())
	s.Equal(1, len(decision.GetReasons()))
	s.Equal(optimizelyUserContext, decision.GetUserContext())
	s.Equal(errorString, decision.GetReasons()[0])
}

func TestOptimizelyDecisionTestSuite(t *testing.T) {
	suite.Run(t, new(OptimizelyDecisionTestSuite))
}
