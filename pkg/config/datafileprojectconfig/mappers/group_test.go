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

	datafileEntities "github.com/optimizely/go-sdk/v2/pkg/config/datafileprojectconfig/entities"
	"github.com/optimizely/go-sdk/v2/pkg/entities"
	"github.com/stretchr/testify/assert"
)

func TestMapGroups(t *testing.T) {

	rawGroup := datafileEntities.Group{
		Policy: "random",
		ID:     "14",
		TrafficAllocation: []datafileEntities.TrafficAllocation{
			datafileEntities.TrafficAllocation{
				EntityID:   "13",
				EndOfRange: 4000,
			},
		},
	}

	rawGroups := []datafileEntities.Group{rawGroup}
	groupMap, _ := MapGroups(rawGroups)

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
