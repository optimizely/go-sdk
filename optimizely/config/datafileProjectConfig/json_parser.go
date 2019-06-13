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
	"encoding/json"

	"github.com/optimizely/go-sdk/optimizely/config/datafileProjectConfig/entities"
)

// JSONParser implements the DatafileParser interface and parses a JSON-based datafile into a DatafileProjectConfig
type JSONParser struct {
}

// Parse parses the json datafile
func (p JSONParser) Parse(jsonDatafile []byte) (*entities.Datafile, error) {

	datafile := &entities.Datafile{}

	err := json.Unmarshal(jsonDatafile, &datafile)
	if err != nil {
		return nil, err
	}

	return datafile, nil
}
