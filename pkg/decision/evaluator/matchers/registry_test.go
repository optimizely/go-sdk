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

// Package matchers //
package matchers

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/optimizely/go-sdk/pkg/decide"
	"github.com/optimizely/go-sdk/pkg/entities"
	"github.com/optimizely/go-sdk/pkg/logging"
)

func TestRegister(t *testing.T) {
	expected := func(entities.Condition, entities.UserContext, logging.OptimizelyLogProducer, decide.DecisionReasons) (bool, error) {
		return false, nil
	}
	Register("test", expected)
	actual := assertMatcher(t, "test")
	matches, err := actual(entities.Condition{}, entities.UserContext{}, nil, decide.NewDecisionReasons(decide.OptimizelyDecideOptions{}))
	assert.False(t, matches)
	assert.NoError(t, err)
}

func TestInit(t *testing.T) {
	assertMatcher(t, ExactMatchType)
	assertMatcher(t, ExistsMatchType)
	assertMatcher(t, LtMatchType)
	assertMatcher(t, GtMatchType)
	assertMatcher(t, SubstringMatchType)
}

func assertMatcher(t *testing.T, name string) Matcher {
	actual, ok := Get(name)
	assert.True(t, ok)
	assert.NotNil(t, actual)
	return actual
}
