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

package evaluator

import (
	"fmt"
	"testing"

	"github.com/optimizely/go-sdk/pkg/entities"
	"github.com/optimizely/go-sdk/pkg/logging"
	"github.com/stretchr/testify/suite"
)

type ConditionTestSuite struct {
	suite.Suite
	mockLogger         *MockLogger
	conditionEvaluator *CustomAttributeConditionEvaluator
}

func (s *ConditionTestSuite) SetupTest() {
	s.mockLogger = new(MockLogger)
	s.conditionEvaluator = NewCustomAttributeConditionEvaluator(s.mockLogger)
}

func (s *ConditionTestSuite) TestCustomAttributeConditionEvaluator() {
	condition := entities.Condition{
		Match: "exact",
		Value: "foo",
		Name:  "string_foo",
		Type:  "custom_attribute",
	}

	// Test condition passes
	user := entities.UserContext{
		Attributes: map[string]interface{}{
			"string_foo": "foo",
		},
	}

	condTreeParams := entities.NewTreeParameters(&user, map[string]entities.Audience{})
	result, _ := s.conditionEvaluator.Evaluate(condition, condTreeParams)
	s.Equal(result, true)

	// Test condition fails
	user = entities.UserContext{
		Attributes: map[string]interface{}{
			"string_foo": "not_foo",
		},
	}
	result, _ = s.conditionEvaluator.Evaluate(condition, condTreeParams)
	s.Equal(result, false)
}

func (s *ConditionTestSuite) TestCustomAttributeConditionEvaluatorWithoutMatchType() {
	condition := entities.Condition{
		Value: "foo",
		Name:  "string_foo",
		Type:  "custom_attribute",
	}

	// Test condition passes
	user := entities.UserContext{
		Attributes: map[string]interface{}{
			"string_foo": "foo",
		},
	}

	condTreeParams := entities.NewTreeParameters(&user, map[string]entities.Audience{})
	result, _ := s.conditionEvaluator.Evaluate(condition, condTreeParams)
	s.Equal(result, true)

	// Test condition fails
	user = entities.UserContext{
		Attributes: map[string]interface{}{
			"string_foo": "not_foo",
		},
	}
	result, _ = s.conditionEvaluator.Evaluate(condition, condTreeParams)
	s.Equal(result, false)
}

func (s *ConditionTestSuite) TestCustomAttributeConditionEvaluatorWithInvalidMatchType() {
	condition := entities.Condition{
		Value: "foo",
		Name:  "string_foo",
		Type:  "custom_attribute",
		Match: "invalid",
	}

	// Test condition fails
	user := entities.UserContext{
		Attributes: map[string]interface{}{
			"string_foo": "foo",
		},
	}

	condTreeParams := entities.NewTreeParameters(&user, map[string]entities.Audience{})
	s.mockLogger.On("Warning", fmt.Sprintf(logging.UnknownMatchType.String(), ""))
	result, _ := s.conditionEvaluator.Evaluate(condition, condTreeParams)
	s.Equal(result, false)
	s.mockLogger.AssertExpectations(s.T())
}

func (s *ConditionTestSuite) TestCustomAttributeConditionEvaluatorWithUnknownType() {
	condition := entities.Condition{
		Value: "foo",
		Name:  "string_foo",
		Type:  "",
	}

	// Test condition fails
	user := entities.UserContext{
		Attributes: map[string]interface{}{
			"string_foo": "foo",
		},
	}

	condTreeParams := entities.NewTreeParameters(&user, map[string]entities.Audience{})
	s.mockLogger.On("Warning", fmt.Sprintf(logging.UnknownConditionType.String(), ""))
	result, _ := s.conditionEvaluator.Evaluate(condition, condTreeParams)
	s.Equal(result, false)
	s.mockLogger.AssertExpectations(s.T())
}

func TestConditionTestSuite(t *testing.T) {
	suite.Run(t, new(ConditionTestSuite))
}
