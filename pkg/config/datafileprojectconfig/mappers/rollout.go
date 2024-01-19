/****************************************************************************
 * Copyright 2019,2021 Optimizely, Inc. and contributors                    *
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

// Package mappers ...
package mappers

import (
	datafileEntities "github.com/optimizely/go-sdk/v2/pkg/config/datafileprojectconfig/entities"
	"github.com/optimizely/go-sdk/v2/pkg/entities"
)

// MapRollouts maps the raw datafile rollout entities to SDK Rollout entities
func MapRollouts(rollouts []datafileEntities.Rollout) (rolloutList []entities.Rollout, rolloutMap map[string]entities.Rollout) {
	rolloutList = []entities.Rollout{}
	rolloutMap = make(map[string]entities.Rollout)
	for _, rollout := range rollouts {
		mappedRollout := mapRollout(rollout)
		rolloutList = append(rolloutList, mappedRollout)
		rolloutMap[rollout.ID] = mappedRollout
	}

	return rolloutList, rolloutMap
}

func mapRollout(datafileRollout datafileEntities.Rollout) entities.Rollout {
	rolloutExperiments := make([]entities.Experiment, len(datafileRollout.Experiments))
	for i, datafileExperiment := range datafileRollout.Experiments {
		experiment := mapExperiment(datafileExperiment)
		rolloutExperiments[i] = experiment
	}

	return entities.Rollout{
		ID:          datafileRollout.ID,
		Experiments: rolloutExperiments,
	}
}
