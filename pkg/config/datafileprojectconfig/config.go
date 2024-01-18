/****************************************************************************
 * Copyright 2019-2022, Optimizely, Inc. and contributors                   *
 *                                                                          *
 * Licensed under the Apache License, Version 2.0 (the "License");          *
 * you may not use this file except in compliance with the License.         *
 * You may obtain a copy of the License at                                  *
 *                                                                          *
 *    https://www.apache.org/licenses/LICENSE-2.0                           *
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
	"errors"
	"fmt"

	"github.com/optimizely/go-sdk/v2/pkg/config/datafileprojectconfig/mappers"
	"github.com/optimizely/go-sdk/v2/pkg/entities"
	"github.com/optimizely/go-sdk/v2/pkg/logging"
)

var datafileVersions = map[string]struct{}{
	"4": {},
}

// DatafileProjectConfig is a project config backed by a datafile
type DatafileProjectConfig struct {
	datafile             string
	hostForODP           string
	publicKeyForODP      string
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
	rollouts             []entities.Rollout
	integrations         []entities.Integration
	segments             []string
	rolloutMap           map[string]entities.Rollout
	anonymizeIP          bool
	botFiltering         bool
	sendFlagDecisions    bool
	sdkKey               string
	environmentKey       string

	flagVariationsMap map[string][]entities.Variation
}

// GetDatafile returns a string representation of the environment's datafile
func (c DatafileProjectConfig) GetDatafile() string {
	return c.datafile
}

// GetHostForODP returns hostForODP
func (c DatafileProjectConfig) GetHostForODP() string {
	return c.hostForODP
}

// GetPublicKeyForODP returns publicKeyForODP
func (c DatafileProjectConfig) GetPublicKeyForODP() string {
	return c.publicKeyForODP
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

// GetAttributes returns attributes
func (c DatafileProjectConfig) GetAttributes() (attributeList []entities.Attribute) {
	for _, attribute := range c.attributeMap {
		attributeList = append(attributeList, attribute)
	}
	return attributeList
}

// GetAttributeID returns attributeID
func (c DatafileProjectConfig) GetAttributeID(key string) string {
	return c.attributeKeyToIDMap[key]
}

// GetBotFiltering returns botFiltering
func (c DatafileProjectConfig) GetBotFiltering() bool {
	return c.botFiltering
}

// GetSdkKey returns sdkKey for specific environment.
func (c DatafileProjectConfig) GetSdkKey() string {
	return c.sdkKey
}

// GetEnvironmentKey returns current environment of the datafile.
func (c DatafileProjectConfig) GetEnvironmentKey() string {
	return c.environmentKey
}

// GetEvents returns all events
func (c DatafileProjectConfig) GetEvents() (eventList []entities.Event) {
	for _, event := range c.eventMap {
		eventList = append(eventList, event)
	}
	return eventList
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

// GetExperimentList returns an array of all the experiments
func (c DatafileProjectConfig) GetExperimentList() (experimentList []entities.Experiment) {
	for _, experiment := range c.experimentMap {
		experimentList = append(experimentList, experiment)
	}
	return experimentList
}

// GetIntegrationList returns an array of all the integrations
func (c DatafileProjectConfig) GetIntegrationList() (integrationList []entities.Integration) {
	return c.integrations
}

// GetSegmentList returns an array of all the segments
func (c DatafileProjectConfig) GetSegmentList() (segmentList []string) {
	return c.segments
}

// GetRolloutList returns an array of all the rollouts
func (c DatafileProjectConfig) GetRolloutList() (rolloutList []entities.Rollout) {
	return c.rollouts
}

// GetAudienceList returns an array of all the audiences
func (c DatafileProjectConfig) GetAudienceList() (audienceList []entities.Audience) {
	for _, audience := range c.audienceMap {
		audienceList = append(audienceList, audience)
	}
	return audienceList
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

// SendFlagDecisions determines whether impressions events are sent for ALL decision types
func (c DatafileProjectConfig) SendFlagDecisions() bool {
	return c.sendFlagDecisions
}

// GetFlagVariationsMap returns map containing all variations for each flag
func (c DatafileProjectConfig) GetFlagVariationsMap() map[string][]entities.Variation {
	return c.flagVariationsMap
}

// NewDatafileProjectConfig initializes a new datafile from a json byte array using the default JSON datafile parser
func NewDatafileProjectConfig(jsonDatafile []byte, logger logging.OptimizelyLogProducer) (*DatafileProjectConfig, error) {
	datafile, err := Parse(jsonDatafile)
	if err != nil {
		logger.Error("Error parsing datafile", err)
		return nil, err
	}

	if _, ok := datafileVersions[datafile.Version]; !ok {
		err = errors.New("unsupported datafile version")
		logger.Error(fmt.Sprintf("Version %s of datafile not supported", datafile.Version), err)
		return nil, err
	}

	var hostForODP, publicKeyForODP string
	for _, integration := range datafile.Integrations {
		if integration.Key == nil {
			err = errors.New("unsupported key in integrations")
			logger.Error("Error parsing datafile", err)
			return nil, err
		}
		if *integration.Key == "odp" {
			hostForODP = integration.Host
			publicKeyForODP = integration.PublicKey
			break
		}
	}

	attributeMap, attributeKeyToIDMap := mappers.MapAttributes(datafile.Attributes)
	allExperiments := mappers.MergeExperiments(datafile.Experiments, datafile.Groups)
	groupMap, experimentGroupMap := mappers.MapGroups(datafile.Groups)
	experimentIDMap, experimentKeyMap := mappers.MapExperiments(allExperiments, experimentGroupMap)

	rollouts, rolloutMap := mappers.MapRollouts(datafile.Rollouts)
	integrations := []entities.Integration{}
	for _, integration := range datafile.Integrations {
		integrations = append(integrations, entities.Integration{Key: *integration.Key, Host: integration.Host, PublicKey: integration.PublicKey})
	}
	eventMap := mappers.MapEvents(datafile.Events)
	featureMap := mappers.MapFeatures(datafile.FeatureFlags, rolloutMap, experimentIDMap)
	audienceMap, audienceSegmentList := mappers.MapAudiences(append(datafile.TypedAudiences, datafile.Audiences...))
	flagVariationsMap := mappers.MapFlagVariations(featureMap)

	config := &DatafileProjectConfig{
		hostForODP:           hostForODP,
		publicKeyForODP:      publicKeyForODP,
		datafile:             string(jsonDatafile),
		accountID:            datafile.AccountID,
		anonymizeIP:          datafile.AnonymizeIP,
		attributeKeyToIDMap:  attributeKeyToIDMap,
		audienceMap:          audienceMap,
		attributeMap:         attributeMap,
		botFiltering:         datafile.BotFiltering,
		sdkKey:               datafile.SDKKey,
		environmentKey:       datafile.EnvironmentKey,
		experimentKeyToIDMap: experimentKeyMap,
		experimentMap:        experimentIDMap,
		groupMap:             groupMap,
		eventMap:             eventMap,
		featureMap:           featureMap,
		projectID:            datafile.ProjectID,
		revision:             datafile.Revision,
		rollouts:             rollouts,
		integrations:         integrations,
		segments:             audienceSegmentList,
		rolloutMap:           rolloutMap,
		sendFlagDecisions:    datafile.SendFlagDecisions,
		flagVariationsMap:    flagVariationsMap,
	}

	logger.Info("Datafile is valid.")
	return config, nil
}
