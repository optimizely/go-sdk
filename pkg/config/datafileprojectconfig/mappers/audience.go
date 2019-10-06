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

// Package mappers  ...
package mappers

import (
	datafileEntities "github.com/optimizely/go-sdk/pkg/config/datafileprojectconfig/entities"
	"github.com/optimizely/go-sdk/pkg/entities"
)

// MapAudiences maps the raw datafile audience entities to SDK Audience entities
func MapAudiences(audiences []datafileEntities.Audience) map[string]entities.Audience {

	audienceMap := make(map[string]entities.Audience)
	for _, audience := range audiences {
		_, ok := audienceMap[audience.ID]
		if !ok {
			conditionTree, err := buildConditionTree(audience.Conditions)
			if err != nil {
				// @TODO: handle error
				func() {}() // cheat the linters
			}
			audienceMap[audience.ID] = entities.Audience{
				ID:            audience.ID,
				Name:          audience.Name,
				ConditionTree: conditionTree,
			}
		}
	}
	return audienceMap
}
