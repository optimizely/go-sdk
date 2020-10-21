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

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/optimizely/go-sdk/pkg/decide"
	e "github.com/optimizely/go-sdk/pkg/entities"
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

var stringFooCondition = e.Condition{
	Type:  "custom_attribute",
	Match: "exact",
	Name:  "string_foo",
	Value: "foo",
}

var boolTrueCondition = e.Condition{
	Type:  "custom_attribute",
	Match: "exact",
	Name:  "bool_true",
	Value: true,
}

var int42Condition = e.Condition{
	Type:  "custom_attribute",
	Match: "exact",
	Name:  "int_42",
	Value: 42,
}

type ConditionTreeTestSuite struct {
	suite.Suite
	mockLogger             *MockLogger
	options                []decide.Options
	reasons                decide.DecisionReasons
	conditionTreeEvaluator *MixedTreeEvaluator
}

func (s *ConditionTreeTestSuite) SetupTest() {
	s.mockLogger = new(MockLogger)
	s.options = []decide.Options{}
	s.reasons = decide.NewDecisionReasons(s.options)
	s.conditionTreeEvaluator = NewMixedTreeEvaluator(s.mockLogger)
}

func (s *ConditionTreeTestSuite) TestConditionTreeEvaluateSimpleCondition() {
	conditionTree := &e.TreeNode{
		Operator: "or",
		Nodes: []*e.TreeNode{
			{
				Item: stringFooCondition,
			},
		},
	}

	// Test match
	user := e.UserContext{
		Attributes: map[string]interface{}{
			"string_foo": "foo",
		},
	}
	condTreeParams := e.NewTreeParameters(&user, map[string]e.Audience{})
	result, _ := s.conditionTreeEvaluator.Evaluate(conditionTree, condTreeParams, s.options, s.reasons)
	s.True(result)

	// Test no match
	user = e.UserContext{
		Attributes: map[string]interface{}{
			"string_foo": "not foo",
		},
	}
	result, _ = s.conditionTreeEvaluator.Evaluate(conditionTree, condTreeParams, s.options, s.reasons)
	s.False(result)
}

func (s *ConditionTreeTestSuite) TestConditionTreeEvaluateMultipleOrConditions() {
	conditionTree := &e.TreeNode{
		Operator: "or",
		Nodes: []*e.TreeNode{
			{
				Item: stringFooCondition,
			},
			{
				Item: boolTrueCondition,
			},
		},
	}

	// Test match string
	user := e.UserContext{
		Attributes: map[string]interface{}{
			"string_foo": "foo",
		},
	}

	condTreeParams := e.NewTreeParameters(&user, map[string]e.Audience{})
	result, _ := s.conditionTreeEvaluator.Evaluate(conditionTree, condTreeParams, s.options, s.reasons)
	s.True(result)

	// Test match bool
	user = e.UserContext{
		Attributes: map[string]interface{}{
			"bool_true": true,
		},
	}
	s.mockLogger.On("Debug", fmt.Sprintf(logging.NullUserAttribute.String(), "", "string_foo"))
	result, _ = s.conditionTreeEvaluator.Evaluate(conditionTree, condTreeParams, s.options, s.reasons)

	s.True(result)

	// Test match both
	user = e.UserContext{
		Attributes: map[string]interface{}{
			"string_foo": "foo",
			"bool_true":  true,
		},
	}
	result, _ = s.conditionTreeEvaluator.Evaluate(conditionTree, condTreeParams, s.options, s.reasons)
	s.True(result)

	// Test no match
	user = e.UserContext{
		Attributes: map[string]interface{}{
			"string_foo": "not foo",
			"bool_true":  false,
		},
	}
	result, _ = s.conditionTreeEvaluator.Evaluate(conditionTree, condTreeParams, s.options, s.reasons)
	s.False(result)
	s.mockLogger.AssertExpectations(s.T())
}

func (s *ConditionTreeTestSuite) TestConditionTreeEvaluateMultipleAndConditions() {
	conditionTree := &e.TreeNode{
		Operator: "and",
		Nodes: []*e.TreeNode{
			{
				Item: stringFooCondition,
			},
			{
				Item: boolTrueCondition,
			},
		},
	}

	// Test only string match with NULL bubbling
	user := e.UserContext{
		Attributes: map[string]interface{}{
			"string_foo": "foo",
		},
	}

	condTreeParams := e.NewTreeParameters(&user, map[string]e.Audience{})
	s.mockLogger.On("Debug", fmt.Sprintf(logging.NullUserAttribute.String(), "", "bool_true"))
	result, _ := s.conditionTreeEvaluator.Evaluate(conditionTree, condTreeParams, s.options, s.reasons)
	s.False(result)

	// Test only bool match with NULL bubbling
	user = e.UserContext{
		Attributes: map[string]interface{}{
			"bool_true": true,
		},
	}

	s.mockLogger.On("Debug", fmt.Sprintf(logging.NullUserAttribute.String(), "", "string_foo"))
	result, _ = s.conditionTreeEvaluator.Evaluate(conditionTree, condTreeParams, s.options, s.reasons)
	s.False(result)

	// Test match both
	user = e.UserContext{
		Attributes: map[string]interface{}{
			"string_foo": "foo",
			"bool_true":  true,
		},
	}
	result, _ = s.conditionTreeEvaluator.Evaluate(conditionTree, condTreeParams, s.options, s.reasons)
	s.True(result)

	// Test no match
	user = e.UserContext{
		Attributes: map[string]interface{}{
			"string_foo": "not foo",
			"bool_true":  false,
		},
	}
	result, _ = s.conditionTreeEvaluator.Evaluate(conditionTree, condTreeParams, s.options, s.reasons)
	s.False(result)
	s.mockLogger.AssertExpectations(s.T())
}

func (s *ConditionTreeTestSuite) TestConditionTreeEvaluateNotCondition() {
	// [or, [not, stringFooCondition], [not, boolTrueCondition]]
	conditionTree := &e.TreeNode{
		Operator: "or",
		Nodes: []*e.TreeNode{
			{
				Operator: "not",
				Nodes: []*e.TreeNode{
					{
						Item: stringFooCondition,
					},
				},
			},
			{
				Operator: "not",
				Nodes: []*e.TreeNode{
					{
						Item: boolTrueCondition,
					},
				},
			},
		},
	}

	// Test match string
	user := e.UserContext{
		Attributes: map[string]interface{}{
			"string_foo": "not foo",
		},
	}

	condTreeParams := e.NewTreeParameters(&user, map[string]e.Audience{})
	result, _ := s.conditionTreeEvaluator.Evaluate(conditionTree, condTreeParams, s.options, s.reasons)
	s.True(result)

	// Test match bool
	user = e.UserContext{
		Attributes: map[string]interface{}{
			"bool_true": false,
		},
	}
	s.mockLogger.On("Debug", fmt.Sprintf(logging.NullUserAttribute.String(), "", "string_foo"))
	result, _ = s.conditionTreeEvaluator.Evaluate(conditionTree, condTreeParams, s.options, s.reasons)
	s.True(result)

	// Test match both
	user = e.UserContext{
		Attributes: map[string]interface{}{
			"string_foo": "not foo",
			"bool_true":  false,
		},
	}
	result, _ = s.conditionTreeEvaluator.Evaluate(conditionTree, condTreeParams, s.options, s.reasons)
	s.True(result)

	// Test no match
	user = e.UserContext{
		Attributes: map[string]interface{}{
			"string_foo": "foo",
			"bool_true":  true,
		},
	}
	result, _ = s.conditionTreeEvaluator.Evaluate(conditionTree, condTreeParams, s.options, s.reasons)
	s.False(result)
	s.mockLogger.AssertExpectations(s.T())
}

func (s *ConditionTreeTestSuite) TestConditionTreeEvaluateMultipleMixedConditions() {
	// [or, [and, stringFooCondition, boolTrueCondition], [or, [not, stringFooCondition], int42Condition]]
	conditionTree := &e.TreeNode{
		Operator: "or",
		Nodes: []*e.TreeNode{
			{
				Operator: "and",
				Nodes: []*e.TreeNode{
					{
						Item: stringFooCondition,
					},
					{
						Item: boolTrueCondition,
					},
				},
			},
			{
				Operator: "or",
				Nodes: []*e.TreeNode{
					{
						Operator: "not",
						Nodes: []*e.TreeNode{
							{
								Item: stringFooCondition,
							},
						},
					},
					{
						Item: int42Condition,
					},
				},
			},
		},
	}

	// Test only match AND condition
	user := e.UserContext{
		Attributes: map[string]interface{}{
			"string_foo": "foo",
			"bool_true":  true,
			"int_42":     43,
		},
	}

	condTreeParams := e.NewTreeParameters(&user, map[string]e.Audience{})
	result, _ := s.conditionTreeEvaluator.Evaluate(conditionTree, condTreeParams, s.options, s.reasons)
	s.True(result)

	// Test only match the NOT condition
	user = e.UserContext{
		Attributes: map[string]interface{}{
			"string_foo": "not foo",
			"bool_true":  true,
			"int_42":     43,
		},
	}
	result, _ = s.conditionTreeEvaluator.Evaluate(conditionTree, condTreeParams, s.options, s.reasons)
	s.True(result)

	// Test only match the int condition
	user = e.UserContext{
		Attributes: map[string]interface{}{
			"string_foo": "foo",
			"bool_true":  false,
			"int_42":     42,
		},
	}
	result, _ = s.conditionTreeEvaluator.Evaluate(conditionTree, condTreeParams, s.options, s.reasons)
	s.True(result)

	// Test no match
	user = e.UserContext{
		Attributes: map[string]interface{}{
			"string_foo": "foo",
			"bool_true":  false,
			"int_42":     43,
		},
	}
	result, _ = s.conditionTreeEvaluator.Evaluate(conditionTree, condTreeParams, s.options, s.reasons)
	s.False(result)
}

var audienceMap = map[string]e.Audience{
	"11111": audience11111,
	"11112": audience11112,
}

var audience11111 = e.Audience{
	ID: "11111",
	ConditionTree: &e.TreeNode{
		Operator: "or",
		Nodes: []*e.TreeNode{
			{
				Operator: "or",
				Nodes: []*e.TreeNode{
					{
						Item: stringFooCondition,
					},
				},
			},
		},
	},
}

var audience11112 = e.Audience{
	ID: "11112",
	ConditionTree: &e.TreeNode{
		Operator: "or",
		Nodes: []*e.TreeNode{
			{
				Operator: "and",
				Nodes: []*e.TreeNode{
					{
						Item: boolTrueCondition,
					},
					{
						Item: int42Condition,
					},
				},
			},
		},
	},
}

func (s *ConditionTreeTestSuite) TestConditionTreeEvaluateAnAudienceTreeSingleAudience() {
	audienceTree := &e.TreeNode{
		Operator: "or",
		Nodes: []*e.TreeNode{
			{
				Item: audience11111.ID,
			},
		},
	}

	// Test matches audience 11111
	treeParams := &e.TreeParameters{
		User: &e.UserContext{
			ID: "test_user_1",
			Attributes: map[string]interface{}{
				"string_foo": "foo",
			},
		},
		AudienceMap: audienceMap,
	}

	s.mockLogger.On("Debug", fmt.Sprintf(logging.AudienceEvaluationStarted.String(), "11111"))
	s.mockLogger.On("Debug", fmt.Sprintf(logging.AudienceEvaluatedTo.String(), "11111", true))
	result, _ := s.conditionTreeEvaluator.Evaluate(audienceTree, treeParams, s.options, s.reasons)
	s.True(result)
	s.mockLogger.AssertExpectations(s.T())
}

func (s *ConditionTreeTestSuite) TestConditionTreeEvaluateAnAudienceTreeMultipleAudiences() {
	audienceTree := &e.TreeNode{
		Operator: "or",
		Nodes: []*e.TreeNode{
			{
				Item: audience11111.ID,
			},
			{
				Item: audience11112.ID,
			},
		},
	}

	// Test only matches audience 11111
	treeParams := &e.TreeParameters{
		User: &e.UserContext{
			ID: "test_user_1",
			Attributes: map[string]interface{}{
				"string_foo": "foo",
			},
		},
		AudienceMap: audienceMap,
	}
	s.mockLogger.On("Debug", fmt.Sprintf(logging.AudienceEvaluationStarted.String(), "11111"))
	s.mockLogger.On("Debug", fmt.Sprintf(logging.AudienceEvaluatedTo.String(), "11111", true))
	result, _ := s.conditionTreeEvaluator.Evaluate(audienceTree, treeParams, s.options, s.reasons)
	s.True(result)

	// Test only matches audience 11112
	treeParams = &e.TreeParameters{
		User: &e.UserContext{
			ID: "test_user_1",
			Attributes: map[string]interface{}{
				"bool_true": true,
				"int_42":    42,
			},
		},
		AudienceMap: audienceMap,
	}
	s.mockLogger.On("Debug", fmt.Sprintf(logging.AudienceEvaluationStarted.String(), "11111"))
	s.mockLogger.On("Debug", fmt.Sprintf(logging.NullUserAttribute.String(), "", "string_foo"))
	s.mockLogger.On("Debug", fmt.Sprintf(logging.AudienceEvaluationStarted.String(), "11112"))
	s.mockLogger.On("Debug", fmt.Sprintf(logging.AudienceEvaluatedTo.String(), "11112", true))

	result, _ = s.conditionTreeEvaluator.Evaluate(audienceTree, treeParams, s.options, s.reasons)
	s.True(result)
	s.mockLogger.AssertExpectations(s.T())
}

func TestConditionTreeTestSuite(t *testing.T) {
	suite.Run(t, new(ConditionTreeTestSuite))
}
