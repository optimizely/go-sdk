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

package mappers

import (
	"fmt"
	"testing"

	datafileConfig "github.com/optimizely/go-sdk/pkg/config/datafileprojectconfig/entities"
	"github.com/optimizely/go-sdk/pkg/entities"
	"github.com/stretchr/testify/assert"
)

func _TestTemp(t *testing.T) {
	jsons := `["1", "2"]`
	jsons = `["1"]`
	var cond interface{}
	json.Unmarshal([]byte(jsons), &cond)
	cTree, _ := buildAudienceConditionTree(cond)
	fmt.Println(cTree)
	t.FailNow()
}
func TestConfigV2(t *testing.T) {
	aMap := map[string]string{
		"1":  "us",
		"2":  "female",
		"3":  "adult",
		"11": "fr",
		"12": "male",
		"13": "kid",
	}
	_ = aMap

	audience_input_json := []string{
		`            []`,
		`            ["or", "1", "2"]`,
		`            ["and", "1", "2", "3"]`,
		`            ["not", "1"]`,
		`            ["or", "1"]`,
		`            ["and", "1"]`,
		`            ["1"]`,
		`            ["1", "2"]`,
		`            ["and", ["or", "1", "2"], "3"]`,
		`            ["and", ["or", "1", ["and", "2", "3"]], ["and", "11", ["or", "12", "13"]]]`,
		`            ["not", ["and", "1", "2"]]`,
		`            ["or", "1", "100000"]`,
		`            ["and", "and"]`,
		`            ["and"]`,
		`            ["and", ["or", "1", ["and", "2", "3"]], ["and", "11", ["or", "12", "3"]]]`}

	audiences_output := []string{
		``,
		`"us" OR "female"`,
		`"us" AND "female" AND "adult"`,
		`NOT "us"`,
		`"us"`,
		`"us"`,
		`"us"`,
		`"us" OR "female"`,
		`("us" OR "female") AND "adult"`,
		`("us" OR ("female" AND "adult")) AND ("fr" AND ("male" OR "kid"))`,
		`NOT ("us" AND "female")`,
		`"us" OR "100000"`,
		``,
		``,
		`("us" OR ("female" AND "adult")) AND ("fr" AND ("male" OR "adult"))`}
	_ = audiences_output
	for k, v := range audience_input_json {
		var conditions interface{}
		json.Unmarshal([]byte(v), &conditions)
		fmt.Println("aud", v)
		conditionTree, err := buildAudienceConditionTree(conditions)
		if err != nil {
			t.Fatal(err)
		}
		aud := conditionTree.GetAudienceString(aMap, true)
		fmt.Println(k, v, aud)
		if audiences_output[k] != aud {
			t.FailNow()
		}
	}

}

func TestBuildAudienceConditionTreeEmpty(t *testing.T) {
	conditionString := ""
	var conditions interface{}
	json.Unmarshal([]byte(conditionString), &conditions)
	conditionTree, err := buildAudienceConditionTree(conditions)

	expectedTree := &entities.TreeNode{Operator: "or"}
	assert.NoError(t, err)
	assert.Equal(t, expectedTree, conditionTree)
}

func TestBuildAudienceConditionTreeSimpleAudienceCondition(t *testing.T) {
	conditionString := "[ \"and\", [ \"or\", [ \"or\",  \"12\", \"123\", \"1234\"] ] ]"
	audienceString := `"12" OR "123" OR "1234"`
	var conditions interface{}
	json.Unmarshal([]byte(conditionString), &conditions)
	conditionTree, err := buildAudienceConditionTree(conditions)
	if err != nil {
		assert.Fail(t, err.Error())
	}

	expectedConditionTree := &entities.TreeNode{
		Operator: "and",
		Nodes: []*entities.TreeNode{
			{
				Operator: "or",
				Nodes: []*entities.TreeNode{
					{
						Operator: "or",
						Nodes: []*entities.TreeNode{
							{
								Item: "12",
							},
							{
								Item: "123",
							},
							{
								Item: "1234",
							},
						},
					},
				},
			},
		},
	}
	expectedAudienceString := expectedConditionTree.GetAudienceString(make(map[string]string), false)
	assert.Equal(t, expectedConditionTree, conditionTree)
	return
	if audienceString != expectedAudienceString {
		// t.Fatal(audienceString, "\n", expectedAudienceString)
	}
}

func TestBuildAudienceConditionTreeNoOperators(t *testing.T) {
	conditions := []string{"123"}
	expectedConditionTree := &entities.TreeNode{
		Operator: "or",
		Nodes: []*entities.TreeNode{
			{
				Item: "123",
			},
		},
	}

	conditionTree, err := buildAudienceConditionTree(conditions)
	assert.Equal(t, expectedConditionTree, conditionTree)
	assert.NoError(t, err)
}

func TestBuildConditionTreeUsingDatafileAudienceConditions(t *testing.T) {

	audience := datafileConfig.Audience{
		ID:         "12567320080",
		Name:       "message",
		Conditions: "[\"and\", [\"or\", [\"or\", {\"name\": \"s_foo\", \"type\": \"custom_attribute\", \"value\": \"foo\"}]]]",
	}

	conditionTree, err := buildConditionTree(audience.Conditions)
	if err != nil {
		assert.Fail(t, err.Error())
	}

	expectedConditionTree := &entities.TreeNode{
		Operator: "and",
		Nodes: []*entities.TreeNode{
			{
				Operator: "or",
				Nodes: []*entities.TreeNode{
					{
						Operator: "or",
						Nodes: []*entities.TreeNode{
							{
								Item: entities.Condition{
									Name:                 "s_foo",
									Type:                 "custom_attribute",
									Value:                "foo",
									StringRepresentation: `{"name":"s_foo","type":"custom_attribute","value":"foo"}`,
								},
							},
						},
					},
				},
			},
		},
	}
	assert.Equal(t, expectedConditionTree, conditionTree)
}

func TestBuildConditionTreeSimpleAudienceCondition(t *testing.T) {
	conditionString := "[ \"and\", [ \"or\", [ \"or\", { \"type\": \"custom_attribute\", \"name\": \"s_foo\", \"match\": \"exact\", \"value\": \"foo\" } ] ] ]"
	var conditions interface{}
	json.Unmarshal([]byte(conditionString), &conditions)
	conditionTree, err := buildConditionTree(conditions)
	if err != nil {
		assert.Fail(t, err.Error())
	}

	expectedConditionTree := &entities.TreeNode{
		Operator: "and",
		Nodes: []*entities.TreeNode{
			{
				Operator: "or",
				Nodes: []*entities.TreeNode{
					{
						Operator: "or",
						Nodes: []*entities.TreeNode{
							{
								Item: entities.Condition{
									Name:                 "s_foo",
									Match:                "exact",
									Type:                 "custom_attribute",
									Value:                "foo",
									StringRepresentation: `{"match":"exact","name":"s_foo","type":"custom_attribute","value":"foo"}`,
								},
							},
						},
					},
				},
			},
		},
	}
	assert.Equal(t, expectedConditionTree, conditionTree)
}

func TestBuildConditionTreeWithLeafNode(t *testing.T) {
	conditionString := "{ \"type\": \"custom_attribute\", \"name\": \"s_foo\", \"match\": \"exact\", \"value\": \"foo\" }"
	var conditions interface{}
	json.Unmarshal([]byte(conditionString), &conditions)
	conditionTree, err := buildConditionTree(conditions)
	assert.NoError(t, err)

	expectedConditionTree := &entities.TreeNode{
		Operator: "or",
		Nodes: []*entities.TreeNode{
			{
				Item: entities.Condition{
					Name:                 "s_foo",
					Match:                "exact",
					Type:                 "custom_attribute",
					Value:                "foo",
					StringRepresentation: `{"match":"exact","name":"s_foo","type":"custom_attribute","value":"foo"}`,
				},
			},
		},
	}
	assert.Equal(t, expectedConditionTree, conditionTree)
}
