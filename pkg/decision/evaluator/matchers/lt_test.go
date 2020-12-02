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

type LtTestSuite struct {
	suite.Suite
	mockLogger *MockLogger
	reasons    decide.DecisionReasons
	matcher    Matcher
}

func (s *LtTestSuite) SetupTest() {
	s.mockLogger = new(MockLogger)
	s.reasons = decide.NewDecisionReasons(&decide.Options{})
	s.matcher, _ = Get(LtMatchType)
}

func (s *LtTestSuite) TestLtMatcherInt() {
	condition := entities.Condition{
		Match: "lt",
		Value: 42,
		Name:  "int_42",
	}

	// Test match - same type
	user := entities.UserContext{
		Attributes: map[string]interface{}{
			"int_42": 41,
		},
	}

	result, err := s.matcher(condition, user, s.mockLogger, s.reasons)
	s.NoError(err)
	s.True(result)

	// Test match int to float
	user = entities.UserContext{
		Attributes: map[string]interface{}{
			"int_42": 41.9999,
		},
	}

	result, err = s.matcher(condition, user, s.mockLogger, s.reasons)
	s.NoError(err)
	s.True(result)

	// Test no match
	user = entities.UserContext{
		Attributes: map[string]interface{}{
			"int_42": 42.00000,
		},
	}

	result, err = s.matcher(condition, user, s.mockLogger, s.reasons)
	s.NoError(err)
	s.False(result)

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
			"int_42": true,
		},
	}
	s.mockLogger.On("Warning", fmt.Sprintf(logging.InvalidAttributeValueType.String(), "", true, "int_42"))
	result, err = s.matcher(condition, user, s.mockLogger, s.reasons)
	s.Error(err)
	s.False(result)
	s.mockLogger.AssertExpectations(s.T())
}

func (s *LtTestSuite) TestLtMatcherFloat() {
	condition := entities.Condition{
		Match: "lt",
		Value: 4.2,
		Name:  "float_4_2",
	}

	// Test match float to int
	user := entities.UserContext{
		Attributes: map[string]interface{}{
			"float_4_2": 4,
		},
	}

	result, err := s.matcher(condition, user, s.mockLogger, s.reasons)
	s.NoError(err)
	s.True(result)

	// Test match
	user = entities.UserContext{
		Attributes: map[string]interface{}{
			"float_4_2": 4.19999,
		},
	}

	result, err = s.matcher(condition, user, s.mockLogger, s.reasons)
	s.NoError(err)
	s.True(result)

	// Test no match
	user = entities.UserContext{
		Attributes: map[string]interface{}{
			"float_4_2": 4.2,
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
			"float_4_2": true,
		},
	}
	s.mockLogger.On("Warning", fmt.Sprintf(logging.InvalidAttributeValueType.String(), "", true, "float_4_2"))
	result, err = s.matcher(condition, user, s.mockLogger, s.reasons)
	s.Error(err)
	s.False(result)
	s.mockLogger.AssertExpectations(s.T())
}

func (s *LtTestSuite) TestLtMatcherUnsupportedConditionValue() {
	condition := entities.Condition{
		Match: "lt",
		Value: false,
		Name:  "float_4_2",
	}

	// Test match - unsupported condition value
	user := entities.UserContext{
		Attributes: map[string]interface{}{
			"float_4_2": 4.2,
		},
	}
	s.mockLogger.On("Warning", fmt.Sprintf(logging.UnsupportedConditionValue.String(), ""))
	result, err := s.matcher(condition, user, s.mockLogger, s.reasons)
	s.Error(err)
	s.False(result)
	s.mockLogger.AssertExpectations(s.T())
}

func TestLtTestSuite(t *testing.T) {
	suite.Run(t, new(LtTestSuite))
}
