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
	"sync"
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
	userID := "1212121"
	attributes := map[string]interface{}{"key1": 1212, "key2": 1213}
	optimizelyUserContext := newOptimizelyUserContext(s.OptimizelyClient, userID, attributes)

	assert.Equal(s.T(), s.OptimizelyClient, optimizelyUserContext.GetOptimizely())
	assert.Equal(s.T(), userID, optimizelyUserContext.GetUserID())
	assert.Equal(s.T(), attributes, optimizelyUserContext.GetUserAttributes())
}

func (s *OptimizelyUserContextTestSuite) TestOptimizelyUserContextNoAttributes() {
	userID := "1212121"
	var attributes map[string]interface{}
	optimizelyUserContext := newOptimizelyUserContext(s.OptimizelyClient, userID, attributes)

	assert.Equal(s.T(), s.OptimizelyClient, optimizelyUserContext.GetOptimizely())
	assert.Equal(s.T(), userID, optimizelyUserContext.GetUserID())
	assert.Equal(s.T(), attributes, optimizelyUserContext.GetUserAttributes())
}

func (s *OptimizelyUserContextTestSuite) TestUpatingProvidedUserContextHasNoImpactOnOptimizelyUserContext() {
	userID := "1212121"
	attributes := map[string]interface{}{"k1": "v1", "k2": false}

	userContext := entities.UserContext{ID: userID, Attributes: attributes}
	optimizelyUserContext := newOptimizelyUserContext(s.OptimizelyClient, userID, attributes)

	assert.Equal(s.T(), s.OptimizelyClient, optimizelyUserContext.GetOptimizely())
	assert.Equal(s.T(), userID, optimizelyUserContext.GetUserID())
	assert.Equal(s.T(), attributes, optimizelyUserContext.GetUserAttributes())

	userContext.Attributes["k1"] = "v2"
	userContext.Attributes["k2"] = true

	assert.Equal(s.T(), "v1", optimizelyUserContext.GetUserAttributes()["k1"])
	assert.Equal(s.T(), false, optimizelyUserContext.GetUserAttributes()["k2"])

	attributes = optimizelyUserContext.GetUserAttributes()
	attributes["k1"] = "v2"
	attributes["k2"] = true

	assert.Equal(s.T(), "v1", optimizelyUserContext.GetUserAttributes()["k1"])
	assert.Equal(s.T(), false, optimizelyUserContext.GetUserAttributes()["k2"])
}

func (s *OptimizelyUserContextTestSuite) TestSetAttribute() {
	userID := "1212121"
	var attributes map[string]interface{}

	optimizelyUserContext := newOptimizelyUserContext(s.OptimizelyClient, userID, attributes)
	assert.Equal(s.T(), s.OptimizelyClient, optimizelyUserContext.GetOptimizely())

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

	assert.Equal(s.T(), userID, optimizelyUserContext.GetUserID())
	assert.Equal(s.T(), "v1", optimizelyUserContext.GetUserAttributes()["k1"])
	assert.Equal(s.T(), true, optimizelyUserContext.GetUserAttributes()["k2"])
	assert.Equal(s.T(), 100, optimizelyUserContext.GetUserAttributes()["k3"])
	assert.Equal(s.T(), 3.5, optimizelyUserContext.GetUserAttributes()["k4"])
}

func (s *OptimizelyUserContextTestSuite) TestSetAttributeOverride() {
	userID := "1212121"
	attributes := map[string]interface{}{"k1": "v1", "k2": false}
	optimizelyUserContext := newOptimizelyUserContext(s.OptimizelyClient, userID, attributes)

	assert.Equal(s.T(), s.OptimizelyClient, optimizelyUserContext.GetOptimizely())
	assert.Equal(s.T(), userID, optimizelyUserContext.GetUserID())
	assert.Equal(s.T(), attributes, optimizelyUserContext.GetUserAttributes())

	optimizelyUserContext.SetAttribute("k1", "v2")
	optimizelyUserContext.SetAttribute("k2", true)

	assert.Equal(s.T(), "v2", optimizelyUserContext.GetUserAttributes()["k1"])
	assert.Equal(s.T(), true, optimizelyUserContext.GetUserAttributes()["k2"])
}

func TestOptimizelyUserContextTestSuite(t *testing.T) {
	suite.Run(t, new(OptimizelyUserContextTestSuite))
}
