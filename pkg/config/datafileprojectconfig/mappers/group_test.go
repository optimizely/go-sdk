/****************************************************************************
 * Copyright 2019, Optimizely, Inc. and contributors                        *
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
	"testing"

	jsoniter "github.com/json-iterator/go"
	datafileEntities "github.com/optimizely/go-sdk/pkg/config/datafileprojectconfig/entities"
	"github.com/optimizely/go-sdk/pkg/entities"
	"github.com/stretchr/testify/assert"
)

func TestMapGroups(t *testing.T) {
	const testGroupString = `{
		"policy": "random",
		"trafficAllocation": [
		  {
			"entityId": "13",
			"endOfRange": 4000
		  }
		],
		"id": "14"
	  }`

	var rawGroup datafileEntities.Group
	var json = jsoniter.ConfigCompatibleWithStandardLibrary
	json.Unmarshal([]byte(testGroupString), &rawGroup)

	rawGroups := []datafileEntities.Group{rawGroup}
	groupMap := MapGroups(rawGroups)

	expectedGroupsMap := map[string]entities.Group{
		"14": entities.Group{
			ID:     "14",
			Policy: "random",
			TrafficAllocation: []entities.Range{
				entities.Range{
					EntityID:   "13",
					EndOfRange: 4000,
				},
			},
		},
	}

	assert.Equal(t, expectedGroupsMap, groupMap)
}
