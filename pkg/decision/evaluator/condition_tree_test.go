package evaluator

import (
	"testing"

	"github.com/stretchr/testify/assert"

	e "github.com/optimizely/go-sdk/pkg/entities"
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
	conditionTreeEvaluator := NewMixedTreeEvaluator()
	conditionTree := &e.TreeNode{
		Operator: "or",
		Nodes: []*e.TreeNode{
			&e.TreeNode{
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
	result, _ := conditionTreeEvaluator.Evaluate(conditionTree, condTreeParams)
	assert.True(t, result)

	// Test no match
	user = e.UserContext{
		Attributes: map[string]interface{}{
			"string_foo": "not foo",
		},
	}
	result, _ = conditionTreeEvaluator.Evaluate(conditionTree, condTreeParams)
	assert.False(t, result)
}

func TestConditionTreeEvaluateMultipleOrConditions(t *testing.T) {
	conditionTreeEvaluator := NewMixedTreeEvaluator()
	conditionTree := &e.TreeNode{
		Operator: "or",
		Nodes: []*e.TreeNode{
			&e.TreeNode{
				Item: stringFooCondition,
			},
			&e.TreeNode{
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
	result, _ := conditionTreeEvaluator.Evaluate(conditionTree, condTreeParams)
	assert.True(t, result)

	// Test match bool
	user = e.UserContext{
		Attributes: map[string]interface{}{
			"bool_true": true,
		},
	}
	result, _ = conditionTreeEvaluator.Evaluate(conditionTree, condTreeParams)
	assert.True(t, result)

	// Test match both
	user = e.UserContext{
		Attributes: map[string]interface{}{
			"string_foo": "foo",
			"bool_true":  true,
		},
	}
	result, _ = conditionTreeEvaluator.Evaluate(conditionTree, condTreeParams)
	assert.True(t, result)

	// Test no match
	user = e.UserContext{
		Attributes: map[string]interface{}{
			"string_foo": "not foo",
			"bool_true":  false,
		},
	}
	result, _ = conditionTreeEvaluator.Evaluate(conditionTree, condTreeParams)
	assert.False(t, result)
}

func TestConditionTreeEvaluateMultipleAndConditions(t *testing.T) {
	conditionTreeEvaluator := NewMixedTreeEvaluator()
	conditionTree := &e.TreeNode{
		Operator: "and",
		Nodes: []*e.TreeNode{
			&e.TreeNode{
				Item: stringFooCondition,
			},
			&e.TreeNode{
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
	result, _ := conditionTreeEvaluator.Evaluate(conditionTree, condTreeParams)
	assert.False(t, result)

	// Test only bool match with NULL bubbling
	user = e.UserContext{
		Attributes: map[string]interface{}{
			"bool_true": true,
		},
	}
	result, _ = conditionTreeEvaluator.Evaluate(conditionTree, condTreeParams)
	assert.False(t, result)

	// Test match both
	user = e.UserContext{
		Attributes: map[string]interface{}{
			"string_foo": "foo",
			"bool_true":  true,
		},
	}
	result, _ = conditionTreeEvaluator.Evaluate(conditionTree, condTreeParams)
	assert.True(t, result)

	// Test no match
	user = e.UserContext{
		Attributes: map[string]interface{}{
			"string_foo": "not foo",
			"bool_true":  false,
		},
	}
	result, _ = conditionTreeEvaluator.Evaluate(conditionTree, condTreeParams)
	assert.False(t, result)
}

func TestConditionTreeEvaluateNotCondition(t *testing.T) {
	conditionTreeEvaluator := NewMixedTreeEvaluator()
	// [or, [not, stringFooCondition], [not, boolTrueCondition]]
	conditionTree := &e.TreeNode{
		Operator: "or",
		Nodes: []*e.TreeNode{
			&e.TreeNode{
				Operator: "not",
				Nodes: []*e.TreeNode{
					&e.TreeNode{
						Item: stringFooCondition,
					},
				},
			},
			&e.TreeNode{
				Operator: "not",
				Nodes: []*e.TreeNode{
					&e.TreeNode{
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
	result, _ := conditionTreeEvaluator.Evaluate(conditionTree, condTreeParams)
	assert.True(t, result)

	// Test match bool
	user = e.UserContext{
		Attributes: map[string]interface{}{
			"bool_true": false,
		},
	}
	result, _ = conditionTreeEvaluator.Evaluate(conditionTree, condTreeParams)
	assert.True(t, result)

	// Test match both
	user = e.UserContext{
		Attributes: map[string]interface{}{
			"string_foo": "not foo",
			"bool_true":  false,
		},
	}
	result, _ = conditionTreeEvaluator.Evaluate(conditionTree, condTreeParams)
	assert.True(t, result)

	// Test no match
	user = e.UserContext{
		Attributes: map[string]interface{}{
			"string_foo": "foo",
			"bool_true":  true,
		},
	}
	result, _ = conditionTreeEvaluator.Evaluate(conditionTree, condTreeParams)
	assert.False(t, result)
}

func TestConditionTreeEvaluateMultipleMixedConditions(t *testing.T) {
	conditionTreeEvaluator := NewMixedTreeEvaluator()
	// [or, [and, stringFooCondition, boolTrueCondition], [or, [not, stringFooCondition], int42Condition]]
	conditionTree := &e.TreeNode{
		Operator: "or",
		Nodes: []*e.TreeNode{
			&e.TreeNode{
				Operator: "and",
				Nodes: []*e.TreeNode{
					&e.TreeNode{
						Item: stringFooCondition,
					},
					&e.TreeNode{
						Item: boolTrueCondition,
					},
				},
			},
			&e.TreeNode{
				Operator: "or",
				Nodes: []*e.TreeNode{
					&e.TreeNode{
						Operator: "not",
						Nodes: []*e.TreeNode{
							&e.TreeNode{
								Item: stringFooCondition,
							},
						},
					},
					&e.TreeNode{
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
	result, _ := conditionTreeEvaluator.Evaluate(conditionTree, condTreeParams)
	assert.True(t, result)

	// Test only match the NOT condition
	user = e.UserContext{
		Attributes: map[string]interface{}{
			"string_foo": "not foo",
			"bool_true":  true,
			"int_42":     43,
		},
	}
	result, _ = conditionTreeEvaluator.Evaluate(conditionTree, condTreeParams)
	assert.True(t, result)

	// Test only match the int condition
	user = e.UserContext{
		Attributes: map[string]interface{}{
			"string_foo": "foo",
			"bool_true":  false,
			"int_42":     42,
		},
	}
	result, _ = conditionTreeEvaluator.Evaluate(conditionTree, condTreeParams)
	assert.True(t, result)

	// Test no match
	user = e.UserContext{
		Attributes: map[string]interface{}{
			"string_foo": "foo",
			"bool_true":  false,
			"int_42":     43,
		},
	}
	result, _ = conditionTreeEvaluator.Evaluate(conditionTree, condTreeParams)
	assert.False(t, result)
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
			&e.TreeNode{
				Operator: "or",
				Nodes: []*e.TreeNode{
					&e.TreeNode{
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
			&e.TreeNode{
				Operator: "and",
				Nodes: []*e.TreeNode{
					&e.TreeNode{
						Item: boolTrueCondition,
					},
					&e.TreeNode{
						Item: int42Condition,
					},
				},
			},
		},
	},
}

func TestConditionTreeEvaluateAnAudienceTreeSingleAudience(t *testing.T) {
	audienceTree := &e.TreeNode{
		Operator: "or",
		Nodes: []*e.TreeNode{
			&e.TreeNode{
				Item: audience11111.ID,
			},
		},
	}

	conditionTreeEvaluator := NewMixedTreeEvaluator()

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
	result, _ := conditionTreeEvaluator.Evaluate(audienceTree, treeParams)
	assert.True(t, result)
}

func TestConditionTreeEvaluateAnAudienceTreeMultipleAudiences(t *testing.T) {
	audienceTree := &e.TreeNode{
		Operator: "or",
		Nodes: []*e.TreeNode{
			&e.TreeNode{
				Item: audience11111.ID,
			},
			&e.TreeNode{
				Item: audience11112.ID,
			},
		},
	}

	conditionTreeEvaluator := NewMixedTreeEvaluator()

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
	result, _ := conditionTreeEvaluator.Evaluate(audienceTree, treeParams)
	assert.True(t, result)

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
	result, _ = conditionTreeEvaluator.Evaluate(audienceTree, treeParams)
	assert.True(t, result)
}
