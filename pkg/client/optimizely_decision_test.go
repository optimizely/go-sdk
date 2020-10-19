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

	"github.com/optimizely/go-sdk/pkg/entities"
	"github.com/optimizely/go-sdk/pkg/optimizelyjson"

	"github.com/stretchr/testify/assert"
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

	userContext := entities.UserContext{ID: "testUser1", Attributes: map[string]interface{}{"key": 1212}}
	optimizelyUserContext := s.OptimizelyClient.CreateUserContext(userContext)
	decision := NewOptimizelyDecision(variationKey, ruleKey, flagKey, enabled, *variables, optimizelyUserContext, reasons)

	assert.Equal(s.T(), variationKey, decision.GetVariationKey())
	assert.Equal(s.T(), enabled, decision.GetEnabled())
	assert.Equal(s.T(), variables, decision.GetVariables())
	assert.Equal(s.T(), ruleKey, decision.GetRuleKey())
	assert.Equal(s.T(), flagKey, decision.GetFlagKey())
	assert.Equal(s.T(), reasons, decision.GetReasons())
	assert.Equal(s.T(), optimizelyUserContext, decision.GetUserContext())
}

func (s *OptimizelyDecisionTestSuite) TestCreateErrorDecision() {
	flagKey := "flag1"
	errorString := "SDK has an error"
	userContext := entities.UserContext{ID: "testUser1", Attributes: map[string]interface{}{"key": 1212}}
	optimizelyUserContext := s.OptimizelyClient.CreateUserContext(userContext)
	decision := CreateErrorDecision(flagKey, optimizelyUserContext, errors.New(errorString))

	assert.Equal(s.T(), "", decision.GetVariationKey())
	assert.Equal(s.T(), false, decision.GetEnabled())
	assert.Nil(s.T(), decision.GetVariables())
	assert.Equal(s.T(), "", decision.GetRuleKey())
	assert.Equal(s.T(), flagKey, decision.GetFlagKey())
	assert.Equal(s.T(), 1, len(decision.GetReasons()))
	assert.Equal(s.T(), optimizelyUserContext, decision.GetUserContext())
	assert.Equal(s.T(), errorString, decision.GetReasons()[0])
}

func TestOptimizelyDecisionTestSuite(t *testing.T) {
	suite.Run(t, new(OptimizelyUserContextTestSuite))
}
