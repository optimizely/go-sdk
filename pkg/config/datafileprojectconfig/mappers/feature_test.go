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

func TestMapFeatures(t *testing.T) {
	const testFeatureFlagString = `{
		"id": "21111",
		"key": "test_feature_21111",
		"rolloutId": "41111",
		"experimentIds": ["31111", "31112"],
		"variables": [{"defaultValue":"1","id":"1","key":"test","type":"integer"}]
	}`

	var rawFeatureFlag datafileEntities.FeatureFlag
	var json = jsoniter.ConfigCompatibleWithStandardLibrary
	json.Unmarshal([]byte(testFeatureFlagString), &rawFeatureFlag)

	rawFeatureFlags := []datafileEntities.FeatureFlag{rawFeatureFlag}
	rollout := entities.Rollout{ID: "41111"}
	rolloutMap := map[string]entities.Rollout{
		"41111": rollout,
	}
	experiment31111 := entities.Experiment{ID: "31111"}
	experiment31112 := entities.Experiment{ID: "31112"}
	experimentMap := map[string]entities.Experiment{
		"31111": experiment31111,
		"31112": experiment31112,
	}
	featureMap, experimentFeatureMap := MapFeatures(rawFeatureFlags, rolloutMap, experimentMap)
	variable := entities.Variable{
		ID:           "1",
		DefaultValue: "1",
		Key:          "test",
		Type:         "integer",
	}
	expectedFeatureMap := map[string]entities.Feature{
		"test_feature_21111": entities.Feature{
			ID:                 "21111",
			Key:                "test_feature_21111",
			Rollout:            rollout,
			FeatureExperiments: []entities.Experiment{experiment31111, experiment31112},
			VariableMap:        map[string]entities.Variable{variable.Key: variable},
		},
	}

	expectedExperimentFeatureMap := map[string][]string{}
	expectedExperimentFeatureMap["31111"] = []string{"21111"}
	expectedExperimentFeatureMap["31112"] = []string{"21111"}

	assert.Equal(t, expectedFeatureMap, featureMap)
	assert.Equal(t, expectedExperimentFeatureMap, experimentFeatureMap)
}
