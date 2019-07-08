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
	"github.com/optimizely/go-sdk/optimizely/logging"
)

var logger = logging.GetLogger("DatafileProjectConfig")

// DatafileProjectConfig is a project config backed by a datafile
type DatafileProjectConfig struct {
	audienceMap          map[string]entities.Audience
	experimentMap        map[string]entities.Experiment
	experimentKeyToIDMap map[string]string
	featureMap           map[string]entities.Feature
	attributeKeyToIDMap  map[string]string
	eventMap             map[string]entities.Event
	projectID			 string
	revision			 string
	accountID			 string
	anonymizeIP			 bool
	botFiltering		 bool

}

func (config DatafileProjectConfig) GetProjectID() string {
	return config.projectID
}

func (config DatafileProjectConfig) GetRevision() string {
	return config.revision
}

func (config DatafileProjectConfig) GetAccountID() string {
	return config.accountID
}

func (config DatafileProjectConfig) GetAnonymizeIP() bool {
	return config.anonymizeIP
}

func (config DatafileProjectConfig) GetAttributeID(key string) string {
	return config.attributeKeyToIDMap[key]
}

func (config DatafileProjectConfig) GetBotFiltering() bool {
	return config.botFiltering
}

// NewDatafileProjectConfig initializes a new datafile from a json byte array using the default JSON datafile parser
func NewDatafileProjectConfig(jsonDatafile []byte) *DatafileProjectConfig {
	datafile, err := Parse(jsonDatafile)
	if err != nil {
		logger.Error("Error parsing datafile.", err)
	}

	experiments, experimentKeyMap := mappers.MapExperiments(datafile.Experiments)
	config := &DatafileProjectConfig{
		audienceMap:          mappers.MapAudiences(datafile.Audiences),
		experimentMap:        experiments,
		experimentKeyToIDMap: experimentKeyMap,
	}

	logger.Info("Datafile is valid.")
	return config
}

// GetEventByKey returns the event with the given key
func (config DatafileProjectConfig) GetEventByKey(eventKey string) (entities.Event, error) {
	if event, ok := config.eventMap[eventKey]; ok {
		return event, nil
	}

	errMessage := fmt.Sprintf("Event with key %s not found", eventKey)
	return entities.Event{}, errors.New(errMessage)
}

// GetFeatureByKey returns the feature with the given key
func (config DatafileProjectConfig) GetFeatureByKey(featureKey string) (entities.Feature, error) {
	if feature, ok := config.featureMap[featureKey]; ok {
		return feature, nil
	}

	errMessage := fmt.Sprintf("Feature with key %s not found", featureKey)
	return entities.Feature{}, errors.New(errMessage)
}
