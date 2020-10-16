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
	"testing"

	"github.com/optimizely/go-sdk/pkg/entities"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type OptimizelyUserContextTestSuite struct {
	suite.Suite
	*OptimizelyClient
}

func (s *OptimizelyUserContextTestSuite) SetupTest() {
	factory := OptimizelyFactory{SDKKey: "1212"}
	s.OptimizelyClient, _ = factory.Client()
}

func (s *OptimizelyUserContextTestSuite) TestOptimizelyUserContextWithAttributes() {
	userContext := entities.UserContext{ID: "1212121", Attributes: map[string]interface{}{"key1": 1212, "key2": 1213}}
	optimizelyUserContext := NewOptimizelyUserContext(s.OptimizelyClient, userContext)

	assert.Equal(s.T(), s.OptimizelyClient, optimizelyUserContext.GetOptimizely())
	assert.Equal(s.T(), userContext, optimizelyUserContext.GetUserContext())
}

func (s *OptimizelyUserContextTestSuite) TestOptimizelyUserContextNoAttributes() {
	userContext := entities.UserContext{ID: "1212121"}
	optimizelyUserContext := NewOptimizelyUserContext(s.OptimizelyClient, userContext)

	assert.Equal(s.T(), s.OptimizelyClient, optimizelyUserContext.GetOptimizely())
	assert.Equal(s.T(), map[string]interface{}{}, optimizelyUserContext.GetUserContext().Attributes)
}

func (s *OptimizelyUserContextTestSuite) TestSetAttribute() {
	userContext := entities.UserContext{ID: "1212121"}
	optimizelyUserContext := NewOptimizelyUserContext(s.OptimizelyClient, userContext)
	assert.Equal(s.T(), s.OptimizelyClient, optimizelyUserContext.GetOptimizely())

	optimizelyUserContext.SetAttribute("k1", "v1")
	optimizelyUserContext.SetAttribute("k2", true)
	optimizelyUserContext.SetAttribute("k3", 100)
	optimizelyUserContext.SetAttribute("k4", 3.5)

	assert.Equal(s.T(), userContext.ID, optimizelyUserContext.GetUserContext().ID)
	assert.Equal(s.T(), "v1", optimizelyUserContext.GetUserContext().Attributes["k1"])
	assert.Equal(s.T(), true, optimizelyUserContext.GetUserContext().Attributes["k2"])
	assert.Equal(s.T(), 100, optimizelyUserContext.GetUserContext().Attributes["k3"])
	assert.Equal(s.T(), 3.5, optimizelyUserContext.GetUserContext().Attributes["k4"])
}

func (s *OptimizelyUserContextTestSuite) TestSetAttributeOverride() {
	userContext := entities.UserContext{ID: "1212121", Attributes: map[string]interface{}{"k1": "v1", "k2": false}}
	optimizelyUserContext := NewOptimizelyUserContext(s.OptimizelyClient, userContext)

	assert.Equal(s.T(), s.OptimizelyClient, optimizelyUserContext.GetOptimizely())
	assert.Equal(s.T(), userContext, optimizelyUserContext.GetUserContext())

	optimizelyUserContext.SetAttribute("k1", "v2")
	optimizelyUserContext.SetAttribute("k2", true)

	assert.Equal(s.T(), "v2", optimizelyUserContext.GetUserContext().Attributes["k1"])
	assert.Equal(s.T(), true, optimizelyUserContext.GetUserContext().Attributes["k2"])
}

func TestOptimizelyUserContextTestSuite(t *testing.T) {
	suite.Run(t, new(OptimizelyUserContextTestSuite))
}
