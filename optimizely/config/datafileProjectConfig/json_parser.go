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

	"github.com/optimizely/go-sdk/optimizely/entities"
)

// FeatureFlag represents a FeatureFlag object from the Optimizely datafile
type FeatureFlag struct {
	ID            string   `json:"id"`
	RolloutID     string   `json:"rolloutId"`
	Key           string   `json:"key"`
	ExperimentIDs []string `json:"experimentIds"`
	Variables     []string `json:"variables"`
}

// Datafile represents the datafile we get from Optimizely
type Datafile struct {
	Version      string        `json:"version"`
	AnonymizeIP  bool          `json:"anonymizeIP"`
	ProjectID    string        `json:"projectId"`
	Variables    []string      `json:"variables"`
	FeatureFlags []FeatureFlag `json:"featureFlags"`
	BotFiltering bool          `json:"botFiltering"`
	AccountID    string        `json:"accountId"`
	Revision     string        `json:"revision"`
}

// DatafileJSONParser implements the DatafileParser interface and parses a JSON-based datafile into a DatafileProjectConfig
type DatafileJSONParser struct {
}

// Parse parses the json datafile
func (parser DatafileJSONParser) Parse(jsonDatafile []byte) (*DatafileProjectConfig, error) {
	datafile := Datafile{}
	projectConfig := &DatafileProjectConfig{}

	err := json.Unmarshal(jsonDatafile, &datafile)
	if err != nil {
		// @TODO(mng): return error
	}

	// convert the Datafile into a ProjectConfig
	projectConfig.features = mapFeatureFlags(datafile.FeatureFlags)
	return projectConfig, nil
}

func mapFeatureFlags(featureFlags []FeatureFlag) map[string]entities.Feature {
	featureMap := make(map[string]entities.Feature)
	for _, featureFlag := range featureFlags {
		featureMap[featureFlag.Key] = entities.Feature{
			Key: featureFlag.Key,
			ID:  featureFlag.ID,
		}
	}
	return featureMap
}
