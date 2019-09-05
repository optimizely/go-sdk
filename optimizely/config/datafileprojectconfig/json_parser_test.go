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

package datafileprojectconfig

import (
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/optimizely/go-sdk/optimizely/config/datafileprojectconfig/entities"
	"github.com/stretchr/testify/assert"
)

func TestParseDatafilePasses(t *testing.T) {
	testFeatureKey := "feature_test_1"
	testFeatureID := "feature_id_123"
	datafileString := fmt.Sprintf(`{
		"projectId": "1337",
		"accountId": "1338",
		"version": "4",
		"featureFlags": [
			{
				"key": "%s",
				"id" : "%s"
			}
		]
	}`, testFeatureKey, testFeatureID)

	rawDatafile := []byte(datafileString)
	parsedDatafile, err := Parse(rawDatafile)
	if err != nil {
		assert.Fail(t, err.Error())
	}

	expectedDatafile := &entities.Datafile{
		AccountID: "1338",
		ProjectID: "1337",
		Version:   "4",
		FeatureFlags: []entities.FeatureFlag{
			entities.FeatureFlag{
				Key: testFeatureKey,
				ID:  testFeatureID,
			},
		},
	}

	assert.Equal(t, expectedDatafile, parsedDatafile)
}

func BenchmarkParseDatafilePasses(b *testing.B) {
	for n := 0; n < b.N; n++ {
		datafile, _ := ioutil.ReadFile("test/100_entities.json")

		Parse(datafile)

	}

}
