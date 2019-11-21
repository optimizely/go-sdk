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

// MapGroups maps the raw group entity from the datafile to an SDK Group entity
func MapGroups(rawGroups []datafileEntities.Group) (groupMap map[string]entities.Group, experimentGroupMap map[string]string) {
	groupMap = make(map[string]entities.Group)
	experimentGroupMap = make(map[string]string)
	for _, group := range rawGroups {
		groupEntity := entities.Group{
			ID:     group.ID,
			Policy: group.Policy,
		}
		for _, allocation := range group.TrafficAllocation {
			groupEntity.TrafficAllocation = append(groupEntity.TrafficAllocation, entities.Range(allocation))
		}
		groupMap[group.ID] = groupEntity
		for _, experiment := range group.Experiments {
			experimentGroupMap[experiment.ID] = group.ID
		}
	}
	return groupMap, experimentGroupMap
}
