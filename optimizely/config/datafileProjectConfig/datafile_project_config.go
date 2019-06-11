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

	"github.com/optimizely/go-sdk/optimizely/entities"
)

// DatafileParser parses a datafile into a DatafileProjectConfig
type DatafileParser interface {
	Parse([]byte) (*DatafileProjectConfig, error)
}

// DatafileProjectConfig is a project config backed by a datafile
type DatafileProjectConfig struct {
	features map[string]entities.Feature
	parser   DatafileParser
}

// NewDatafileProjectConfig initializes a new datafile from a json byte array using the default JSON datafile parser
func NewDatafileProjectConfig(jsonDatafile []byte) *DatafileProjectConfig {
	parser := DatafileJSONParser{}
	return NewDatafileProjectConfigWithParser(parser, jsonDatafile)
}

// NewDatafileProjectConfigWithParser initializes a new datafile from a json byte array using the given parser
func NewDatafileProjectConfigWithParser(parser DatafileParser, jsonDatafile []byte) *DatafileProjectConfig {
	projectConfig, err := parser.Parse(jsonDatafile)
	if err != nil {
		// @TODO(mng): handle the error
	}
	return projectConfig
}

// GetFeatureByKey returns the feature with the given key
func (config DatafileProjectConfig) GetFeatureByKey(featureKey string) (entities.Feature, error) {
	if feature, ok := config.features[featureKey]; ok {
		return feature, nil
	}

	errMessage := fmt.Sprintf("Feature with key %s not found", featureKey)
	return entities.Feature{}, errors.New(errMessage)
}
