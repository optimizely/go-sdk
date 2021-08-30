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

package mappers

import (
	"testing"

	datafileEntities "github.com/optimizely/go-sdk/pkg/config/datafileprojectconfig/entities"
	"github.com/optimizely/go-sdk/pkg/entities"
	"github.com/stretchr/testify/assert"
)

func TestMapRollouts(t *testing.T) {
	const testRolloutString = `{
		 "id": "21111",
		 "experiments": [
			 { "id": "11111", "key": "exp_11111" },
			 { "id": "11112", "key": "exp_11112" }
		 ]
	 }`

	var rawRollout datafileEntities.Rollout
	json.Unmarshal([]byte(testRolloutString), &rawRollout)

	rawRollouts := []datafileEntities.Rollout{rawRollout}
	rolloutList, rolloutMap := MapRollouts(rawRollouts)
	expectedRolloutMap := map[string]entities.Rollout{
		"21111": {
			ID: "21111",
			Experiments: []entities.Experiment{
				{ID: "11111", Key: "exp_11111", Variations: map[string]entities.Variation{}, VariationKeyToIDMap: map[string]string{}, TrafficAllocation: []entities.Range{}},
				{ID: "11112", Key: "exp_11112", Variations: map[string]entities.Variation{}, VariationKeyToIDMap: map[string]string{}, TrafficAllocation: []entities.Range{}},
			},
		},
	}

	assert.Equal(t, []entities.Rollout{expectedRolloutMap["21111"]}, rolloutList)
	assert.Equal(t, expectedRolloutMap, rolloutMap)
}
