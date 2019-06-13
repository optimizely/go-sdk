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

package datafileProjectConfig

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseDatafilePasses(t *testing.T) {
	testFeatureKey := "feature_test_1"
	testFeatureID := "feature_id_123"
	datafileString := fmt.Sprintf(`{
		"projectId": "1337",
		"featureFlags": [
			{
				"key": "%s",
				"id" : "%s"
			}
		]
	}`, testFeatureKey, testFeatureID)

	rawDatafile := []byte(datafileString)
	parser := DatafileJSONParser{}
	projectConfig, err := parser.Parse(rawDatafile)
	if err != nil {
		assert.Fail(t, err.Error())
	}
	feature, err := projectConfig.GetFeatureByKey("feature_test_1")
	if err != nil {
		assert.Fail(t, err.Error())
	}

	assert.Equal(t, "feature_id_123", feature.ID)
}
