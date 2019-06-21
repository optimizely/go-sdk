package evaluator

import (
	"testing"

	"github.com/stretchr/testify/assert"

	e "github.com/optimizely/go-sdk/optimizely/entities"
)

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

func TestConditionTreeEvaluateSimpleCondition(t *testing.T) {
	conditionTreeEvaluator := NewConditionTreeEvaluator()
	conditionTree := &e.ConditionTreeNode{
		Operator: "or",
		Nodes: []*e.ConditionTreeNode{
			&e.ConditionTreeNode{
				Condition: stringFooCondition,
			},
		},
	}

	// Test match
	user := e.UserContext{
		Attributes: e.UserAttributes{
			Attributes: map[string]interface{}{
				"string_foo": "foo",
			},
		},
	}
	result := conditionTreeEvaluator.Evaluate(conditionTree, user)
	assert.True(t, result)

	// Test no match
	user = e.UserContext{
		Attributes: e.UserAttributes{
			Attributes: map[string]interface{}{
				"string_foo": "not foo",
			},
		},
	}
	result = conditionTreeEvaluator.Evaluate(conditionTree, user)
	assert.False(t, result)
}

func TestConditionTreeEvaluateMultipleOrConditions(t *testing.T) {
	conditionTreeEvaluator := NewConditionTreeEvaluator()
	conditionTree := &e.ConditionTreeNode{
		Operator: "or",
		Nodes: []*e.ConditionTreeNode{
			&e.ConditionTreeNode{
				Condition: stringFooCondition,
			},
			&e.ConditionTreeNode{
				Condition: boolTrueCondition,
			},
		},
	}

	// Test match string
	user := e.UserContext{
		Attributes: e.UserAttributes{
			Attributes: map[string]interface{}{
				"string_foo": "foo",
			},
		},
	}
	result := conditionTreeEvaluator.Evaluate(conditionTree, user)
	assert.True(t, result)

	// Test match bool
	user = e.UserContext{
		Attributes: e.UserAttributes{
			Attributes: map[string]interface{}{
				"bool_true": true,
			},
		},
	}
	result = conditionTreeEvaluator.Evaluate(conditionTree, user)
	assert.True(t, result)

	// Test match both
	user = e.UserContext{
		Attributes: e.UserAttributes{
			Attributes: map[string]interface{}{
				"string_foo": "foo",
				"bool_true":  true,
			},
		},
	}
	result = conditionTreeEvaluator.Evaluate(conditionTree, user)
	assert.True(t, result)

	// Test no match
	user = e.UserContext{
		Attributes: e.UserAttributes{
			Attributes: map[string]interface{}{
				"string_foo": "not foo",
				"bool_true":  false,
			},
		},
	}
	result = conditionTreeEvaluator.Evaluate(conditionTree, user)
	assert.False(t, result)
}

func TestConditionTreeEvaluateMultipleAndConditions(t *testing.T) {
	conditionTreeEvaluator := NewConditionTreeEvaluator()
	conditionTree := &e.ConditionTreeNode{
		Operator: "and",
		Nodes: []*e.ConditionTreeNode{
			&e.ConditionTreeNode{
				Condition: stringFooCondition,
			},
			&e.ConditionTreeNode{
				Condition: boolTrueCondition,
			},
		},
	}

	// Test only string match with NULL bubbling
	user := e.UserContext{
		Attributes: e.UserAttributes{
			Attributes: map[string]interface{}{
				"string_foo": "foo",
			},
		},
	}
	result := conditionTreeEvaluator.Evaluate(conditionTree, user)
	assert.False(t, result)

	// Test only bool match with NULL bubbling
	user = e.UserContext{
		Attributes: e.UserAttributes{
			Attributes: map[string]interface{}{
				"bool_true": true,
			},
		},
	}
	result = conditionTreeEvaluator.Evaluate(conditionTree, user)
	assert.False(t, result)

	// Test match both
	user = e.UserContext{
		Attributes: e.UserAttributes{
			Attributes: map[string]interface{}{
				"string_foo": "foo",
				"bool_true":  true,
			},
		},
	}
	result = conditionTreeEvaluator.Evaluate(conditionTree, user)
	assert.True(t, result)

	// Test no match
	user = e.UserContext{
		Attributes: e.UserAttributes{
			Attributes: map[string]interface{}{
				"string_foo": "not foo",
				"bool_true":  false,
			},
		},
	}
	result = conditionTreeEvaluator.Evaluate(conditionTree, user)
	assert.False(t, result)
}

func TestConditionTreeEvaluateNotCondition(t *testing.T) {
	conditionTreeEvaluator := NewConditionTreeEvaluator()
	// [or, [not, stringFooCondition], [not, boolTrueCondition]]
	conditionTree := &e.ConditionTreeNode{
		Operator: "or",
		Nodes: []*e.ConditionTreeNode{
			&e.ConditionTreeNode{
				Operator: "not",
				Nodes: []*e.ConditionTreeNode{
					&e.ConditionTreeNode{
						Condition: stringFooCondition,
					},
				},
			},
			&e.ConditionTreeNode{
				Operator: "not",
				Nodes: []*e.ConditionTreeNode{
					&e.ConditionTreeNode{
						Condition: boolTrueCondition,
					},
				},
			},
		},
	}

	// Test match string
	user := e.UserContext{
		Attributes: e.UserAttributes{
			Attributes: map[string]interface{}{
				"string_foo": "not foo",
			},
		},
	}
	result := conditionTreeEvaluator.Evaluate(conditionTree, user)
	assert.True(t, result)

	// Test match bool
	user = e.UserContext{
		Attributes: e.UserAttributes{
			Attributes: map[string]interface{}{
				"bool_true": false,
			},
		},
	}
	result = conditionTreeEvaluator.Evaluate(conditionTree, user)
	assert.True(t, result)

	// Test match both
	user = e.UserContext{
		Attributes: e.UserAttributes{
			Attributes: map[string]interface{}{
				"string_foo": "not foo",
				"bool_true":  false,
			},
		},
	}
	result = conditionTreeEvaluator.Evaluate(conditionTree, user)
	assert.True(t, result)

	// Test no match
	user = e.UserContext{
		Attributes: e.UserAttributes{
			Attributes: map[string]interface{}{
				"string_foo": "foo",
				"bool_true":  true,
			},
		},
	}
	result = conditionTreeEvaluator.Evaluate(conditionTree, user)
	assert.False(t, result)
}

func TestConditionTreeEvaluateMultipleMixedConditions(t *testing.T) {
	conditionTreeEvaluator := NewConditionTreeEvaluator()
	// [or, [and, stringFooCondition, boolTrueCondition], [or, [not, stringFooCondition], int42Condition]]
	conditionTree := &e.ConditionTreeNode{
		Operator: "or",
		Nodes: []*e.ConditionTreeNode{
			&e.ConditionTreeNode{
				Operator: "and",
				Nodes: []*e.ConditionTreeNode{
					&e.ConditionTreeNode{
						Condition: stringFooCondition,
					},
					&e.ConditionTreeNode{
						Condition: boolTrueCondition,
					},
				},
			},
			&e.ConditionTreeNode{
				Operator: "or",
				Nodes: []*e.ConditionTreeNode{
					&e.ConditionTreeNode{
						Operator: "not",
						Nodes: []*e.ConditionTreeNode{
							&e.ConditionTreeNode{
								Condition: stringFooCondition,
							},
						},
					},
					&e.ConditionTreeNode{
						Condition: int42Condition,
					},
				},
			},
		},
	}

	// Test only match AND condition
	user := e.UserContext{
		Attributes: e.UserAttributes{
			Attributes: map[string]interface{}{
				"string_foo": "foo",
				"bool_true":  true,
				"int_42":     43,
			},
		},
	}
	result := conditionTreeEvaluator.Evaluate(conditionTree, user)
	assert.True(t, result)

	// Test only match the NOT condition
	user = e.UserContext{
		Attributes: e.UserAttributes{
			Attributes: map[string]interface{}{
				"string_foo": "not foo",
				"bool_true":  true,
				"int_42":     43,
			},
		},
	}
	result = conditionTreeEvaluator.Evaluate(conditionTree, user)
	assert.True(t, result)

	// Test only match the int condition
	user = e.UserContext{
		Attributes: e.UserAttributes{
			Attributes: map[string]interface{}{
				"string_foo": "foo",
				"bool_true":  false,
				"int_42":     42,
			},
		},
	}
	result = conditionTreeEvaluator.Evaluate(conditionTree, user)
	assert.True(t, result)

	// Test no match
	user = e.UserContext{
		Attributes: e.UserAttributes{
			Attributes: map[string]interface{}{
				"string_foo": "foo",
				"bool_true":  false,
				"int_42":     43,
			},
		},
	}
	result = conditionTreeEvaluator.Evaluate(conditionTree, user)
	assert.False(t, result)
}
