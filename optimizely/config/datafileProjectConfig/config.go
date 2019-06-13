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
	"errors"
	"fmt"

	"github.com/optimizely/go-sdk/optimizely/config/datafileProjectConfig/mappers"
	"github.com/optimizely/go-sdk/optimizely/entities"
)

// DatafileProjectConfig is a project config backed by a datafile
type DatafileProjectConfig struct {
	audienceMap          map[string]entities.Audience
	experimentMap        map[string]entities.Experiment
	experimentKeyToIDMap map[string]string
	featureMap           map[string]entities.Feature
}

// NewDatafileProjectConfig initializes a new datafile from a json byte array using the default JSON datafile parser
func NewDatafileProjectConfig(jsonDatafile []byte) *DatafileProjectConfig {
	parser := JSONParser{}
	datafile, err := parser.Parse(jsonDatafile)
	if err != nil {
		// @TODO(mng): handle error
	}

	experiments, experimentKeyMap := mappers.MapExperiments(datafile.Experiments)
	config := &DatafileProjectConfig{
		audienceMap:          mappers.MapAudiences(datafile.Audiences),
		experimentMap:        experiments,
		experimentKeyToIDMap: experimentKeyMap,
	}

	return config
}

// GetFeatureByKey returns the feature with the given key
func (config DatafileProjectConfig) GetFeatureByKey(featureKey string) (entities.Feature, error) {
	if feature, ok := config.featureMap[featureKey]; ok {
		return feature, nil
	}

	errMessage := fmt.Sprintf("Feature with key %s not found", featureKey)
	return entities.Feature{}, errors.New(errMessage)
}
