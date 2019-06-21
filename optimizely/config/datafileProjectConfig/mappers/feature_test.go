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
	"encoding/json"
	"testing"

	datafileEntities "github.com/optimizely/go-sdk/optimizely/config/datafileProjectConfig/entities"
	"github.com/optimizely/go-sdk/optimizely/entities"
	"github.com/stretchr/testify/assert"
)

func TestMapFeatures(t *testing.T) {
	const testFeatureFlagString = `{
		"id": "21111",
		"key": "test_feature_21111"
	}`

	var rawFeatureFlag datafileEntities.FeatureFlag
	json.Unmarshal([]byte(testFeatureFlagString), &rawFeatureFlag)

	rawFeatureFlags := []datafileEntities.FeatureFlag{rawFeatureFlag}
	featureMap := MapFeatureFlags(rawFeatureFlags)
	expectedFeatureMap := map[string]entities.Feature{
		"test_feature_21111": entities.Feature{
			ID:  "21111",
			Key: "test_feature_21111",
		},
	}

	assert.Equal(t, expectedFeatureMap, featureMap)
}
