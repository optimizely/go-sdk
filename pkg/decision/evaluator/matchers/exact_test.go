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

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/optimizely/go-sdk/pkg/decide"
	"github.com/optimizely/go-sdk/pkg/entities"
	"github.com/optimizely/go-sdk/pkg/logging"
)

type MockLogger struct {
	mock.Mock
}

func (m *MockLogger) Debug(message string) {
	m.Called(message)
}

func (m *MockLogger) Info(message string) {
	m.Called(message)
}

func (m *MockLogger) Warning(message string) {
	m.Called(message)
}

func (m *MockLogger) Error(message string, err interface{}) {
	m.Called(message, err)
}

type ExactTestSuite struct {
	suite.Suite
	mockLogger *MockLogger
	reasons    decide.DecisionReasons
	matcher    Matcher
}

func (s *ExactTestSuite) SetupTest() {
	s.mockLogger = new(MockLogger)
	s.reasons = decide.NewDecisionReasons(&decide.Options{})
	s.matcher, _ = Get(ExactMatchType)
}

func (s *ExactTestSuite) TestExactMatcherString() {
	condition := entities.Condition{
		Match: "exact",
		Value: "foo",
		Name:  "string_foo",
	}

	// Test match
	user := entities.UserContext{
		Attributes: map[string]interface{}{
			"string_foo": "foo",
		},
	}
	result, err := s.matcher(condition, user, s.mockLogger, s.reasons)
	s.NoError(err)
	s.True(result)

	// Test no match
	user = entities.UserContext{
		Attributes: map[string]interface{}{
			"string_foo": "not_foo",
		},
	}

	result, err = s.matcher(condition, user, s.mockLogger, s.reasons)
	s.NoError(err)
	s.False(result)

	// Test attribute not found
	user = entities.UserContext{
		Attributes: map[string]interface{}{
			"string_not_foo": "foo",
		},
	}
	s.mockLogger.On("Debug", fmt.Sprintf(logging.NullUserAttribute.String(), "", "string_foo"))
	_, err = s.matcher(condition, user, s.mockLogger, s.reasons)
	s.Error(err)

	// Test attribute of different type
	user = entities.UserContext{
		Attributes: map[string]interface{}{
			"string_foo": 121,
		},
	}
	s.mockLogger.On("Warning", fmt.Sprintf(logging.InvalidAttributeValueType.String(), "", 121, "string_foo"))
	result, err = s.matcher(condition, user, s.mockLogger, s.reasons)
	s.Error(err)
	s.False(false)
	s.mockLogger.AssertExpectations(s.T())
}

func (s *ExactTestSuite) TestExactMatcherBool() {
	condition := entities.Condition{
		Match: "exact",
		Value: true,
		Name:  "bool_true",
	}

	// Test match
	user := entities.UserContext{
		Attributes: map[string]interface{}{
			"bool_true": true,
		},
	}
	result, err := s.matcher(condition, user, s.mockLogger, s.reasons)
	s.NoError(err)
	s.True(result)

	// Test no match
	user = entities.UserContext{
		Attributes: map[string]interface{}{
			"bool_true": false,
		},
	}

	result, err = s.matcher(condition, user, s.mockLogger, s.reasons)
	s.NoError(err)
	s.False(result)

	// Test attribute not found
	user = entities.UserContext{
		Attributes: map[string]interface{}{
			"not_bool_true": true,
		},
	}

	s.mockLogger.On("Debug", fmt.Sprintf(logging.NullUserAttribute.String(), "", "bool_true"))
	_, err = s.matcher(condition, user, s.mockLogger, s.reasons)
	s.Error(err)

	// Test attribute of different type
	user = entities.UserContext{
		Attributes: map[string]interface{}{
			"bool_true": 121,
		},
	}
	s.mockLogger.On("Warning", fmt.Sprintf(logging.InvalidAttributeValueType.String(), "", 121, "bool_true"))
	result, err = s.matcher(condition, user, s.mockLogger, s.reasons)
	s.Error(err)
	s.False(result)
	s.mockLogger.AssertExpectations(s.T())
}

func (s *ExactTestSuite) TestExactMatcherInt() {
	condition := entities.Condition{
		Match: "exact",
		Value: 42,
		Name:  "int_42",
	}

	// Test match - same type
	user := entities.UserContext{
		Attributes: map[string]interface{}{
			"int_42": 42,
		},
	}
	result, err := s.matcher(condition, user, s.mockLogger, s.reasons)
	s.NoError(err)
	s.True(result)

	// Test match int to float
	user = entities.UserContext{
		Attributes: map[string]interface{}{
			"int_42": 42.0,
		},
	}

	result, err = s.matcher(condition, user, s.mockLogger, s.reasons)
	s.NoError(err)
	s.True(result)

	// Test no match
	user = entities.UserContext{
		Attributes: map[string]interface{}{
			"int_42": 43,
		},
	}

	result, err = s.matcher(condition, user, s.mockLogger, s.reasons)
	s.NoError(err)
	s.False(result)

	// Test attribute not found
	user = entities.UserContext{
		Attributes: map[string]interface{}{
			"int_43": 42,
		},
	}
	s.mockLogger.On("Debug", fmt.Sprintf(logging.NullUserAttribute.String(), "", "int_42"))
	_, err = s.matcher(condition, user, s.mockLogger, s.reasons)
	s.Error(err)

	// Test attribute of different type
	user = entities.UserContext{
		Attributes: map[string]interface{}{
			"int_42": "test",
		},
	}
	s.mockLogger.On("Warning", fmt.Sprintf(logging.InvalidAttributeValueType.String(), "", "test", "int_42"))
	result, err = s.matcher(condition, user, s.mockLogger, s.reasons)
	s.Error(err)
	s.False(result)
	s.mockLogger.AssertExpectations(s.T())
}

func (s *ExactTestSuite) TestExactMatcherFloat() {
	condition := entities.Condition{
		Match: "exact",
		Value: 4.2,
		Name:  "float_4_2",
	}

	// Test match
	user := entities.UserContext{
		Attributes: map[string]interface{}{
			"float_4_2": 4.2,
		},
	}
	result, err := s.matcher(condition, user, s.mockLogger, s.reasons)
	s.NoError(err)
	s.True(result)

	// Test no match
	user = entities.UserContext{
		Attributes: map[string]interface{}{
			"float_4_2": 4.3,
		},
	}

	result, err = s.matcher(condition, user, s.mockLogger, s.reasons)
	s.NoError(err)
	s.False(result)

	// Test attribute not found
	user = entities.UserContext{
		Attributes: map[string]interface{}{
			"float_4_3": 4.2,
		},
	}
	s.mockLogger.On("Debug", fmt.Sprintf(logging.NullUserAttribute.String(), "", "float_4_2"))
	_, err = s.matcher(condition, user, s.mockLogger, s.reasons)
	s.Error(err)

	// Test attribute of different type
	user = entities.UserContext{
		Attributes: map[string]interface{}{
			"float_4_2": "test",
		},
	}
	s.mockLogger.On("Warning", fmt.Sprintf(logging.InvalidAttributeValueType.String(), "", "test", "float_4_2"))
	result, err = s.matcher(condition, user, s.mockLogger, s.reasons)
	s.Error(err)
	s.False(result)
	s.mockLogger.AssertExpectations(s.T())
}

func (s *ExactTestSuite) TestExactMatcherUnsupportedConditionValue() {
	condition := entities.Condition{
		Match: "exact",
		Value: map[string]interface{}{},
		Name:  "int_42",
	}

	// Test match - unsupported condition value
	user := entities.UserContext{
		Attributes: map[string]interface{}{
			"int_42": 42,
		},
	}
	s.mockLogger.On("Warning", fmt.Sprintf(logging.UnsupportedConditionValue.String(), ""))
	result, err := s.matcher(condition, user, s.mockLogger, s.reasons)
	s.Error(err)
	s.False(result)
	s.mockLogger.AssertExpectations(s.T())
}

func TestExactTestSuite(t *testing.T) {
	suite.Run(t, new(ExactTestSuite))
}
