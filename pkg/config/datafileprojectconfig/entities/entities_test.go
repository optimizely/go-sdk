/****************************************************************************
 * Copyright 2021, Optimizely, Inc. and contributors                        *
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

package entities

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUnMarshalSdkKeyAndEnvironmentFromDatafile(t *testing.T) {
	datafile := Datafile{}
	var jsonMap map[string]interface{}
	bytesData, _ := json.Marshal(datafile)
	json.Unmarshal(bytesData, &jsonMap)

	_, keyExists := jsonMap["sdkKey"]
	assert.False(t, keyExists)

	_, keyExists = jsonMap["environment"]
	assert.False(t, keyExists)

	datafile.SDKKey = "a"
	datafile.Environment = "production"
	bytesData, _ = json.Marshal(datafile)
	json.Unmarshal(bytesData, &jsonMap)

	_, keyExists = jsonMap["sdkKey"]
	assert.True(t, keyExists)

	_, keyExists = jsonMap["environment"]
	assert.True(t, keyExists)
}
