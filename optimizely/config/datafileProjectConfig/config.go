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
	accountID            string
	anonymizeIP          bool
	attributeKeyToIDMap  map[string]string
	audienceMap          map[string]entities.Audience
	botFiltering         bool
	eventMap             map[string]entities.Event
	experimentKeyToIDMap map[string]string
	experimentMap        map[string]entities.Experiment
	featureMap           map[string]entities.Feature
	groupMap             map[string]entities.Group
	projectID            string
	revision             string
	rolloutMap           map[string]entities.Rollout
}

func (c DatafileProjectConfig) GetProjectID() string {
	return c.projectID
}

func (c DatafileProjectConfig) GetRevision() string {
	return c.revision
}

func (c DatafileProjectConfig) GetAccountID() string {
	return c.accountID
}

func (c DatafileProjectConfig) GetAnonymizeIP() bool {
	return c.anonymizeIP
}

func (c DatafileProjectConfig) GetAttributeID(key string) string {
	return c.attributeKeyToIDMap[key]
}

func (c DatafileProjectConfig) GetBotFiltering() bool {
	return c.botFiltering
}

// NewDatafileProjectConfig initializes a new datafile from a json byte array using the default JSON datafile parser
func NewDatafileProjectConfig(jsonDatafile []byte) (*DatafileProjectConfig, error) {
	datafile, err := Parse(jsonDatafile)
	if err != nil {
		logger.Error("Error parsing datafile.", err)
		return nil, err
	}

	experiments, experimentKeyMap := mappers.MapExperiments(datafile.Experiments)
	rolloutMap := mappers.MapRollouts(datafile.Rollouts)
	config := &DatafileProjectConfig{
		audienceMap:          mappers.MapAudiences(datafile.Audiences),
		experimentMap:        experiments,
		experimentKeyToIDMap: experimentKeyMap,
		rolloutMap:           rolloutMap,
		featureMap:           mappers.MapFeatureFlags(datafile.FeatureFlags, rolloutMap),
	}

	logger.Info("Datafile is valid.")
	return config, nil
}

// GetEventByKey returns the event with the given key
func (c DatafileProjectConfig) GetEventByKey(eventKey string) (entities.Event, error) {
	if event, ok := c.eventMap[eventKey]; ok {
		return event, nil
	}

	errMessage := fmt.Sprintf("Event with key %s not found", eventKey)
	return entities.Event{}, errors.New(errMessage)
}

// GetFeatureByKey returns the feature with the given key
func (c DatafileProjectConfig) GetFeatureByKey(featureKey string) (entities.Feature, error) {
	if feature, ok := c.featureMap[featureKey]; ok {
		return feature, nil
	}

	errMessage := fmt.Sprintf("Feature with key %s not found", featureKey)
	return entities.Feature{}, errors.New(errMessage)
}

// GetAudienceByID returns the audience with the given ID
func (c DatafileProjectConfig) GetAudienceByID(audienceID string) (entities.Audience, error) {
	if audience, ok := c.audienceMap[audienceID]; ok {
		return audience, nil
	}

	errMessage := fmt.Sprintf(`Audience with ID "%s" not found`, audienceID)
	return entities.Audience{}, errors.New(errMessage)
}

// GetAudienceMap returns the audience map
func (c DatafileProjectConfig) GetAudienceMap() map[string]entities.Audience {
	return c.audienceMap
}

// GetExperimentByKey returns the experiment with the given key
func (c DatafileProjectConfig) GetExperimentByKey(experimentKey string) (entities.Experiment, error) {
	if experimentID, ok := c.experimentKeyToIDMap[experimentKey]; ok {
		experiment := c.experimentMap[experimentID]
		return experiment, nil
	}

	errMessage := fmt.Sprintf(`Experiment with key "%s" not found`, experimentKey)
	return entities.Experiment{}, errors.New(errMessage)
}

// GetGroupByID returns the group with the given ID
func (c DatafileProjectConfig) GetGroupByID(groupID string) (entities.Group, error) {
	if group, ok := c.groupMap[groupID]; ok {
		return group, nil
	}

	errMessage := fmt.Sprintf(`Group with ID "%s" not found`, groupID)
	return entities.Group{}, errors.New(errMessage)
}
