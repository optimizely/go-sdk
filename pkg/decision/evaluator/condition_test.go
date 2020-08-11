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
	"github.com/optimizely/go-sdk/pkg/entities"
	"github.com/stretchr/testify/suite"
)

type ConditionTestSuite struct {
	suite.Suite
	user               entities.UserContext
	conditionEvaluator CustomAttributeConditionEvaluator
}

func (s *ConditionTestSuite) SetupTest() {
	s.conditionEvaluator = CustomAttributeConditionEvaluator{}
	s.user = entities.UserContext{
		Attributes: map[string]interface{}{},
	}
}

func (s *ConditionTestSuite) TestCustomAttributeConditionEvaluator() {
	condition := entities.Condition{
		Match: "exact",
		Value: "foo",
		Name:  "string_foo",
		Type:  "custom_attribute",
	}

	// Test condition passes
	s.user.Attributes["string_foo"] = "foo"

	condTreeParams := entities.NewTreeParameters(&s.user, map[string]entities.Audience{})
	result, _ := s.conditionEvaluator.Evaluate(condition, condTreeParams)
	s.Equal(result, true)

	// Test condition fails
	s.user.Attributes["string_foo"] = "not_foo"
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
	s.user.Attributes["string_foo"] = "foo"

	condTreeParams := entities.NewTreeParameters(&s.user, map[string]entities.Audience{})
	result, _ := s.conditionEvaluator.Evaluate(condition, condTreeParams)
	s.Equal(result, true)

	// Test condition fails
	s.user.Attributes["string_foo"] = "not_foo"
	result, _ = s.conditionEvaluator.Evaluate(condition, condTreeParams)
	s.Equal(result, false)
}

func (s *ConditionTestSuite) TestCustomAttributeConditionEvaluatorSemanticSame() {
	// Test if same when all target is only major.minor
	condition := entities.Condition{
		Value: "2.0.0",
		Match: "semver_eq",
		Name:  "version",
		Type:  "custom_attribute",
	}

	// Test condition passes
	s.user.Attributes["version"] = "2.0"

	condTreeParams := entities.NewTreeParameters(&s.user, map[string]entities.Audience{})
	result, _ := s.conditionEvaluator.Evaluate(condition, condTreeParams)
	s.Equal(result, true)
}

func (s *ConditionTestSuite) TestCustomAttributeConditionEvaluatorSemanticSameFull() {
	// Test when target is full semantic version major.minor.patch
	condition := entities.Condition{
		Value: "3.0.0",
		Match: "semver_eq",
		Name:  "version",
		Type:  "custom_attribute",
	}

	// Test condition passes
	s.user.Attributes["version"] = "3.0.0"

	condTreeParams := entities.NewTreeParameters(&s.user, map[string]entities.Audience{})
	result, _ := s.conditionEvaluator.Evaluate(condition, condTreeParams)
	s.Equal(result, true)
}

func (s *ConditionTestSuite) TestCustomAttributeConditionEvaluatorSemanticLess() {
	// Test compare less when target is only major.minor
	condition := entities.Condition{
		Value: "2.1.6",
		Match: "semver_lt",
		Name:  "version",
		Type:  "custom_attribute",
	}

	// Test condition passes
	s.user.Attributes["version"] = "2.2"

	condTreeParams := entities.NewTreeParameters(&s.user, map[string]entities.Audience{})
	result, _ := s.conditionEvaluator.Evaluate(condition, condTreeParams)
	s.Equal(result, true)
}

func (s *ConditionTestSuite) TestCustomAttributeConditionEvaluatorSemanticFullLess() {
	// Test compare less when target is full major.minor.patch
	condition := entities.Condition{
		Value: "2.1.6",
		Match: "semver_lt",
		Name:  "version",
		Type:  "custom_attribute",
	}

	// Test condition passes
	s.user.Attributes["version"] = "2.1.9"

	condTreeParams := entities.NewTreeParameters(&s.user, map[string]entities.Audience{})
	result, _ := s.conditionEvaluator.Evaluate(condition, condTreeParams)
	s.Equal(result, true)
}

func (s *ConditionTestSuite) TestCustomAttributeConditionEvaluatorSemanticMore() {
	// Test compare greater when target is only major.minor
	condition := entities.Condition{
		Value: "2.3.6",
		Match: "semver_gt",
		Name:  "version",
		Type:  "custom_attribute",
	}

	// Test condition passes
	s.user.Attributes["version"] = "2.2"

	condTreeParams := entities.NewTreeParameters(&s.user, map[string]entities.Audience{})
	result, _ := s.conditionEvaluator.Evaluate(condition, condTreeParams)
	s.Equal(result, true)
}

func (s *ConditionTestSuite) TestCustomAttributeConditionEvaluatorSemanticFullMore() {
	// Test compare greater when target is major.minor.patch
	condition := entities.Condition{
		Value: "2.1.9",
		Match: "semver_gt",
		Name:  "version",
		Type:  "custom_attribute",
	}

	// Test condition passes
	s.user.Attributes["version"] = "2.1.6"

	condTreeParams := entities.NewTreeParameters(&s.user, map[string]entities.Audience{})
	result, _ := s.conditionEvaluator.Evaluate(condition, condTreeParams)
	s.Equal(result, true)
}

func (s *ConditionTestSuite) TestCustomAttributeConditionEvaluatorSemanticFullEqual() {
	// Test compare equal when target is major.minor.patch-beta
	condition := entities.Condition{
		Value: "2.1.9-beta",
		Match: "semver_eq",
		Name:  "version",
		Type:  "custom_attribute",
	}

	// Test condition passes
	s.user.Attributes["version"] = "2.1.9-beta"

	condTreeParams := entities.NewTreeParameters(&s.user, map[string]entities.Audience{})
	result, _ := s.conditionEvaluator.Evaluate(condition, condTreeParams)
	s.Equal(result, true)
}
