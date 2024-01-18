/****************************************************************************
 * Copyright 2019,2021-2022, Optimizely, Inc. and contributors              *
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

// Package mappers  ...
package mappers

import (
	datafileEntities "github.com/optimizely/go-sdk/v2/pkg/config/datafileprojectconfig/entities"
	"github.com/optimizely/go-sdk/v2/pkg/entities"
)

// MapAudiences maps the raw datafile audience entities to SDK Audience entities
func MapAudiences(audiences []datafileEntities.Audience) (audienceMap map[string]entities.Audience, audienceSegmentList []string) {

	audienceMap = make(map[string]entities.Audience)
	// To keep unique segments only
	odpSegmentsMap := map[string]bool{}
	audienceSegmentList = []string{}
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
			conditionTree, fSegments, err := buildConditionTree(audience.Conditions)
			if err == nil {
				audience.ConditionTree = conditionTree
			}
			// Only add unique segments to the list
			for _, s := range fSegments {
				if !odpSegmentsMap[s] {
					odpSegmentsMap[s] = true
					audienceSegmentList = append(audienceSegmentList, s)
				}
			}
			audienceMap[audience.ID] = audience
		}
	}
	return audienceMap, audienceSegmentList
}
