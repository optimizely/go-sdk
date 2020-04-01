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
	jsonRepr map[string]interface{}
}

// NewOptimizelyJSON constructs the object
func NewOptimizelyJSON(jsonRepr map[string]interface{}) *OptimizelyJSON {
	return &OptimizelyJSON{jsonRepr: jsonRepr}
}

// ToString returns the string representation of json
func (optlyJson OptimizelyJSON) ToString() (string, error) {
	jsonBytes, err := json.Marshal(optlyJson.jsonRepr)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

// ToDict returns the native representation of json (map of interface)
func (optlyJson OptimizelyJSON) ToDict() map[string]interface{} {
	return optlyJson.jsonRepr
}

// GetValue populates the schema passed by the user - it takes primitive types and complex struct type
func (optlyJson OptimizelyJSON) GetValue(jsonKey string, schema interface{}) error {

	populateSchema := func(v interface{}) error {
		jsonBytes, err := json.Marshal(v)
		if err != nil {
			return err
		}
		err = json.Unmarshal(jsonBytes, schema)
		return err
	}

	if jsonKey == "" { // populate the whole schema
		return populateSchema(optlyJson.jsonRepr)
	}

	splitJSONKey := strings.Split(jsonKey, ".")
	internalMap := optlyJson.jsonRepr
	lastIndex := len(splitJSONKey) - 1

	for i := 0; i < len(splitJSONKey); i++ {

		if splitJSONKey[i] == "" {
			return errors.New("json key cannot be empty")
		}

		if item, ok := internalMap[splitJSONKey[i]]; ok {
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
			return fmt.Errorf(`json key "%s" not found`, splitJSONKey[i])
		}
	}

	return nil
}
