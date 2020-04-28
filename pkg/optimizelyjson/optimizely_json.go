/****************************************************************************
 * Copyright 2020, Optimizely, Inc. and contributors                        *
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

// Package optimizelyjson //
package optimizelyjson

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

// OptimizelyJSON holds the underlying structure of the object
type OptimizelyJSON struct {
	payload string

	data map[string]interface{}
}

// NewOptimizelyJSONfromString constructs the object out of string payload
func NewOptimizelyJSONfromString(payload string) (*OptimizelyJSON, error) {

	var data map[string]interface{}
	err := json.Unmarshal([]byte(payload), &data)
	if err != nil {
		return nil, err
	}

	return &OptimizelyJSON{payload: payload, data: data}, err
}

// NewOptimizelyJSONfromMap constructs the object
func NewOptimizelyJSONfromMap(data map[string]interface{}) *OptimizelyJSON {
	return &OptimizelyJSON{data: data}
}

// ToString returns the string representation of json
func (optlyJson *OptimizelyJSON) ToString() (string, error) {
	if optlyJson.payload == "" {
		jsonBytes, err := json.Marshal(optlyJson.data)
		if err != nil {
			return "", err
		}
		optlyJson.payload = string(jsonBytes)

	}
	return optlyJson.payload, nil
}

// ToMap returns the native representation of json (map of interface)
func (optlyJson *OptimizelyJSON) ToMap() map[string]interface{} {
	return optlyJson.data
}

// GetValue populates the schema passed by the user - it takes primitive types and complex struct type
func (optlyJson *OptimizelyJSON) GetValue(jsonPath string, schema interface{}) error {

	populateSchema := func(v interface{}) error {
		jsonBytes, err := json.Marshal(v)
		if err != nil {
			return err
		}
		err = json.Unmarshal(jsonBytes, schema)
		return err
	}

	if jsonPath == "" { // populate the whole schema
		if optlyJson.payload != "" {
			return json.Unmarshal([]byte(optlyJson.payload), schema)
		}
		return populateSchema(optlyJson.data)
	}

	splitJSONPath := strings.Split(jsonPath, ".")
	lastIndex := len(splitJSONPath) - 1

	internalMap := optlyJson.data

	for i := 0; i < len(splitJSONPath); i++ {

		if splitJSONPath[i] == "" {
			return errors.New("json key cannot be empty")
		}

		if item, ok := internalMap[splitJSONPath[i]]; ok {
			switch v := item.(type) {

			case map[string]interface{}:
				internalMap = v
				if i == lastIndex {
					return populateSchema(v)
				}

			default:
				if i == lastIndex {
					return populateSchema(v)
				}
			}
		} else {
			return fmt.Errorf(`json key "%s" not found`, splitJSONPath[i])
		}
	}

	return nil
}
