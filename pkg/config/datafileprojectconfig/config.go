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

// Package datafileprojectconfig //
package datafileprojectconfig

import (
	"fmt"

	"github.com/optimizely/go-sdk/pkg/config/datafileprojectconfig/mappers"
	"github.com/optimizely/go-sdk/pkg/entities"
	"github.com/optimizely/go-sdk/pkg/logging"
)

var logger = logging.GetLogger("DatafileProjectConfig")

// DatafileProjectConfig is a project config backed by a datafile
type DatafileProjectConfig struct {
	accountID            string
	projectID            string
	revision             string
	experimentKeyToIDMap map[string]string
	audienceMap          map[string]entities.Audience
	attributeMap         map[string]entities.Attribute
	eventMap             map[string]entities.Event
	attributeKeyToIDMap  map[string]string
	experimentMap        map[string]entities.Experiment
	featureMap           map[string]entities.Feature
	groupMap             map[string]entities.Group
	rolloutMap           map[string]entities.Rollout
	anonymizeIP          bool
	botFiltering         bool
}

// GetProjectID returns projectID
func (c DatafileProjectConfig) GetProjectID() string {
	return c.projectID
}

// GetRevision returns revision
func (c DatafileProjectConfig) GetRevision() string {
	return c.revision
}

// GetAccountID returns accountID
func (c DatafileProjectConfig) GetAccountID() string {
	return c.accountID
}

// GetAnonymizeIP returns anonymizeIP
func (c DatafileProjectConfig) GetAnonymizeIP() bool {
	return c.anonymizeIP
}

// GetAttributeID returns attributeID
func (c DatafileProjectConfig) GetAttributeID(key string) string {
	return c.attributeKeyToIDMap[key]
}

// GetBotFiltering returns GetBotFiltering
func (c DatafileProjectConfig) GetBotFiltering() bool {
	return c.botFiltering
}

// GetEventByKey returns the event with the given key
func (c DatafileProjectConfig) GetEventByKey(eventKey string) (entities.Event, error) {
	if event, ok := c.eventMap[eventKey]; ok {
		return event, nil
	}

	return entities.Event{}, fmt.Errorf(`event with key "%s" not found`, eventKey)
}

// GetFeatureByKey returns the feature with the given key
func (c DatafileProjectConfig) GetFeatureByKey(featureKey string) (entities.Feature, error) {
	if feature, ok := c.featureMap[featureKey]; ok {
		return feature, nil
	}

	return entities.Feature{}, fmt.Errorf(`feature with key "%s" not found`, featureKey)
}

// GetVariableByKey returns the featureVariable with the given key
func (c DatafileProjectConfig) GetVariableByKey(featureKey, variableKey string) (entities.Variable, error) {

	var variable entities.Variable
	var err = fmt.Errorf(`variable with key "%s" not found`, featureKey)
	if feature, ok := c.featureMap[featureKey]; ok {

		if v, ok := feature.VariableMap[variableKey]; ok {
			variable = v
			err = nil
		}
	}
	return variable, err
}

// GetAttributeByKey returns the attribute with the given key
func (c DatafileProjectConfig) GetAttributeByKey(key string) (entities.Attribute, error) {
	if attributeID, ok := c.attributeKeyToIDMap[key]; ok {
		if attribute, ok := c.attributeMap[attributeID]; ok {
			return attribute, nil
		}
	}

	return entities.Attribute{}, fmt.Errorf(`attribute with key "%s" not found`, key)
}

// GetFeatureList returns an array of all the features
func (c DatafileProjectConfig) GetFeatureList() (featureList []entities.Feature) {
	for _, feature := range c.featureMap {
		featureList = append(featureList, feature)
	}
	return featureList
}

// GetAudienceByID returns the audience with the given ID
func (c DatafileProjectConfig) GetAudienceByID(audienceID string) (entities.Audience, error) {
	if audience, ok := c.audienceMap[audienceID]; ok {
		return audience, nil
	}

	return entities.Audience{}, fmt.Errorf(`audience with ID "%s" not found`, audienceID)
}

// GetAudienceMap returns the audience map
func (c DatafileProjectConfig) GetAudienceMap() map[string]entities.Audience {
	return c.audienceMap
}

// GetExperimentByKey returns the experiment with the given key
func (c DatafileProjectConfig) GetExperimentByKey(experimentKey string) (entities.Experiment, error) {
	if experimentID, ok := c.experimentKeyToIDMap[experimentKey]; ok {
		if experiment, ok := c.experimentMap[experimentID]; ok {
			return experiment, nil
		}
	}

	return entities.Experiment{}, fmt.Errorf(`experiment with key "%s" not found`, experimentKey)
}

// GetGroupByID returns the group with the given ID
func (c DatafileProjectConfig) GetGroupByID(groupID string) (entities.Group, error) {
	if group, ok := c.groupMap[groupID]; ok {
		return group, nil
	}

	return entities.Group{}, fmt.Errorf(`group with ID "%s" not found`, groupID)
}

// NewDatafileProjectConfig initializes a new datafile from a json byte array using the default JSON datafile parser
func NewDatafileProjectConfig(jsonDatafile []byte) (*DatafileProjectConfig, error) {
	datafile, err := Parse(jsonDatafile)
	if err != nil {
		logger.Error("Error parsing datafile.", err)
		return nil, err
	}

	attributeMap, attributeKeyToIDMap := mappers.MapAttributes(datafile.Attributes)
	experimentMap, experimentKeyMap := mappers.MapExperiments(datafile.Experiments)
	rolloutMap := mappers.MapRollouts(datafile.Rollouts)
	eventMap := mappers.MapEvents(datafile.Events)
	mergedAudiences := append(datafile.TypedAudiences, datafile.Audiences...)
	config := &DatafileProjectConfig{
		accountID:            datafile.AccountID,
		anonymizeIP:          datafile.AnonymizeIP,
		attributeKeyToIDMap:  attributeKeyToIDMap,
		audienceMap:          mappers.MapAudiences(mergedAudiences),
		attributeMap:         attributeMap,
		botFiltering:         datafile.BotFiltering,
		experimentKeyToIDMap: experimentKeyMap,
		experimentMap:        experimentMap,
		eventMap:             eventMap,
		featureMap:           mappers.MapFeatures(datafile.FeatureFlags, rolloutMap, experimentMap),
		projectID:            datafile.ProjectID,
		revision:             datafile.Revision,
		rolloutMap:           rolloutMap,
	}

	logger.Info("Datafile is valid.")
	return config, nil
}
