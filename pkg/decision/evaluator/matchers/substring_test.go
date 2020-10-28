/****************************************************************************
 * Copyright 2019-2020, Optimizely, Inc. and contributors                   *
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

package matchers

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/optimizely/go-sdk/pkg/decide"
	"github.com/optimizely/go-sdk/pkg/entities"
	"github.com/optimizely/go-sdk/pkg/logging"
)

type SubstringTestSuite struct {
	suite.Suite
	mockLogger *MockLogger
	reasons    *decide.DecisionReasons
	matcher    Matcher
}

func (s *SubstringTestSuite) SetupTest() {
	s.mockLogger = new(MockLogger)
	s.reasons = decide.NewDecisionReasons(decide.OptimizelyDecideOptions{})
	s.matcher, _ = Get(SubstringMatchType)
}

func (s *SubstringTestSuite) TestSubstringMatcher() {
	condition := entities.Condition{
		Match: "substring",
		Value: "foo",
		Name:  "string_foo",
	}

	// Test match
	user := entities.UserContext{
		Attributes: map[string]interface{}{
			"string_foo": "foobar",
		},
	}

	result, err := s.matcher(condition, user, s.mockLogger, s.reasons)
	s.NoError(err)
	s.True(result)

	// Test no match
	user = entities.UserContext{
		Attributes: map[string]interface{}{
			"string_foo": "bar",
		},
	}

	result, err = s.matcher(condition, user, s.mockLogger, s.reasons)
	s.NoError(err)
	s.False(result)

	// Test attribute not found
	user = entities.UserContext{
		Attributes: map[string]interface{}{
			"not_string_foo": "foo",
		},
	}

	s.mockLogger.On("Debug", fmt.Sprintf(logging.NullUserAttribute.String(), "", "string_foo"))
	_, err = s.matcher(condition, user, s.mockLogger, s.reasons)
	s.Error(err)

	// Test attribute of different type
	user = entities.UserContext{
		Attributes: map[string]interface{}{
			"string_foo": true,
		},
	}
	s.mockLogger.On("Warning", fmt.Sprintf(logging.InvalidAttributeValueType.String(), "", true, "string_foo"))
	result, err = s.matcher(condition, user, s.mockLogger, s.reasons)
	s.Error(err)
	s.False(result)
	s.mockLogger.AssertExpectations(s.T())
}

func (s *SubstringTestSuite) TestSubstringMatcherUnsupportedConditionValue() {
	condition := entities.Condition{
		Match: "substring",
		Value: false,
		Name:  "string_foo",
	}

	// Test match - unsupported condition value
	user := entities.UserContext{
		Attributes: map[string]interface{}{
			"string_foo": "foo",
		},
	}
	s.mockLogger.On("Warning", fmt.Sprintf(logging.UnsupportedConditionValue.String(), ""))
	result, err := s.matcher(condition, user, s.mockLogger, s.reasons)
	s.Error(err)
	s.False(result)
	s.mockLogger.AssertExpectations(s.T())
}

func TestSubstringTestSuite(t *testing.T) {
	suite.Run(t, new(SubstringTestSuite))
}
