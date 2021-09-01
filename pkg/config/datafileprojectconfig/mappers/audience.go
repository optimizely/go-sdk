/****************************************************************************
 * Copyright 2019,2021, Optimizely, Inc. and contributors                   *
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
func MapAudiences(audiences []datafileEntities.Audience) (audienceList []entities.Audience, audienceMap map[string]entities.Audience) {

	audienceList = []entities.Audience{}
	audienceMap = make(map[string]entities.Audience)
	// Since typed audiences were added prior to audiences,
	// they will be given priority in the audienceMap and list
	for _, audience := range audiences {
		_, ok := audienceMap[audience.ID]
		if !ok {
			audience := entities.Audience{
				ID:         audience.ID,
				Name:       audience.Name,
				Conditions: audience.Conditions,
			}
			conditionTree, err := buildConditionTree(audience.Conditions)
			if err == nil {
				audience.ConditionTree = conditionTree
			}

			audienceMap[audience.ID] = audience
			audienceList = append(audienceList, audience)
		}
	}
	return audienceList, audienceMap
}
