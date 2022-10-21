/****************************************************************************
 * Copyright 2019,2021-2022 Optimizely, Inc. and contributors               *
 *                                                                          *
 * Licensed under the Apache License, Version 2.0 (the "License");          *
 * you may not use this file except in compliance with the License.         *
 * You may obtain a copy of the License at                                  *
 *                                                                          *
 *    https://www.apache.org/licenses/LICENSE-2.0                           *
 *                                                                          *
 * Unless required by applicable law or agreed to in writing, software      *
 * distributed under the License is distributed on an "AS IS" BASIS,        *
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. *
 * See the License for the specific language governing permissions and      *
 * limitations under the License.                                           *
 ***************************************************************************/

// Package mappers //
package mappers

import (
	"testing"

	datafileEntities "github.com/optimizely/go-sdk/pkg/config/datafileprojectconfig/entities"
	"github.com/optimizely/go-sdk/pkg/entities"

	"github.com/stretchr/testify/assert"
)

func TestMapAudiencesEmptyList(t *testing.T) {

	audienceMap, audienceSegmentList := MapAudiences(nil)
	expectedAudienceMap := map[string]entities.Audience{}

	assert.Equal(t, expectedAudienceMap, audienceMap)
	assert.Empty(t, audienceSegmentList)
}

func TestMapAudiences(t *testing.T) {

	expectedConditions := "[\"and\", [\"or\", [\"or\", {\"name\": \"s_foo\", \"type\": \"custom_attribute\", \"value\": \"foo\"}]]]"
	audienceList := []datafileEntities.Audience{{ID: "1", Name: "one", Conditions: expectedConditions}, {ID: "2", Name: "two"},
		{ID: "3", Name: "three"}, {ID: "2", Name: "four"}, {ID: "1", Name: "one"}}
	audienceMap, audienceSegmentList := MapAudiences(audienceList)

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
	expectedAudienceMap := map[string]entities.Audience{"1": {ID: "1", Name: "one", ConditionTree: expectedConditionTree, Conditions: expectedConditions}, "2": {ID: "2", Name: "two"},
		"3": {ID: "3", Name: "three"}}

	assert.Equal(t, expectedAudienceMap, audienceMap)
	assert.Empty(t, audienceSegmentList)
}

func TestMapAudiencesWithDuplicateSegments(t *testing.T) {

	expectedConditions1 := "[\"and\", [\"or\", [\"or\", {\"name\": \"odp.audiences1\", \"type\": \"third_party_dimension\", \"value\": \"favoritecolorred\", \"match\": \"qualified\"}]]]"
	expectedConditions2 := "[\"and\", [\"or\", [\"or\", {\"name\": \"odp.audiences2\", \"type\": \"third_party_dimension\", \"value\": \"favoritecolorred\", \"match\": \"qualified\"}]]]"
	audienceList := []datafileEntities.Audience{{ID: "1", Name: "one", Conditions: expectedConditions1}, {ID: "2", Name: "two", Conditions: expectedConditions2},
		{ID: "3", Name: "three"}, {ID: "2", Name: "four"}, {ID: "1", Name: "one"}}
	audienceMap, audienceSegmentList := MapAudiences(audienceList)

	expectedConditionTree1 := &entities.TreeNode{
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
									Name:                 "odp.audiences1",
									Type:                 "third_party_dimension",
									Value:                "favoritecolorred",
									Match:                "qualified",
									StringRepresentation: `{"match":"qualified","name":"odp.audiences1","type":"third_party_dimension","value":"favoritecolorred"}`,
								},
							},
						},
					},
				},
			},
		},
	}
	expectedConditionTree2 := &entities.TreeNode{
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
									Name:                 "odp.audiences2",
									Type:                 "third_party_dimension",
									Value:                "favoritecolorred",
									Match:                "qualified",
									StringRepresentation: `{"match":"qualified","name":"odp.audiences2","type":"third_party_dimension","value":"favoritecolorred"}`,
								},
							},
						},
					},
				},
			},
		},
	}

	expectedAudienceMap := map[string]entities.Audience{"1": {ID: "1", Name: "one", ConditionTree: expectedConditionTree1, Conditions: expectedConditions1}, "2": {ID: "2", Name: "two", ConditionTree: expectedConditionTree2, Conditions: expectedConditions2},
		"3": {ID: "3", Name: "three"}}
	expectedaudienceSegmentList := []string{"favoritecolorred"}

	assert.Equal(t, expectedAudienceMap, audienceMap)
	assert.Equal(t, expectedaudienceSegmentList, audienceSegmentList)
}
