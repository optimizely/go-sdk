/****************************************************************************
 * Copyright 2019-2025, Optimizely, Inc. and contributors                   *
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

	datafileEntities "github.com/optimizely/go-sdk/v2/pkg/config/datafileprojectconfig/entities"
	"github.com/optimizely/go-sdk/v2/pkg/config/datafileprojectconfig/mappers"
	"github.com/optimizely/go-sdk/v2/pkg/entities"
	"github.com/optimizely/go-sdk/v2/pkg/featuretoggle"
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
	attributeKeyMap      map[string]entities.Attribute
	eventMap             map[string]entities.Event
	attributeKeyToIDMap  map[string]string
	attributeIDToKeyMap  map[string]string
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
	region               string
	flagVariationsMap    map[string][]entities.Variation
	holdoutIDMap         map[string]entities.Holdout
	globalHoldouts       []entities.Holdout
	includedHoldouts     map[string][]entities.Holdout
	excludedHoldouts     map[string][]entities.Holdout
	flagHoldoutsMap      map[string][]entities.Holdout
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

// GetAttributeByKey returns the attribute with the given key
func (c DatafileProjectConfig) GetAttributeByKey(key string) (entities.Attribute, error) {
	if attribute, ok := c.attributeKeyMap[key]; ok {
		return attribute, nil
	}

	return entities.Attribute{}, fmt.Errorf(`attribute with key "%s" not found`, key)
}

// GetAttributeKeyByID returns the attribute key for the given ID
func (c DatafileProjectConfig) GetAttributeKeyByID(id string) (string, error) {
	if key, ok := c.attributeIDToKeyMap[id]; ok {
		return key, nil
	}

	return "", fmt.Errorf(`attribute with ID "%s" not found`, id)
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

// GetExperimentByID returns the experiment with the given ID
func (c DatafileProjectConfig) GetExperimentByID(experimentID string) (entities.Experiment, error) {
	if experiment, ok := c.experimentMap[experimentID]; ok {
		return experiment, nil
	}

	return entities.Experiment{}, fmt.Errorf(`experiment with ID "%s" not found`, experimentID)
}

// GetGroupByID returns the group with the given ID
func (c DatafileProjectConfig) GetGroupByID(groupID string) (entities.Group, error) {
	if group, ok := c.groupMap[groupID]; ok {
		return group, nil
	}

	return entities.Group{}, fmt.Errorf(`group with ID "%s" not found`, groupID)
}

// GetHoldoutsForFlag returns the holdouts that apply to a specific flag
func (c *DatafileProjectConfig) GetHoldoutsForFlag(flagKey string) []entities.Holdout {
	// Return empty if holdouts are disabled
	if !featuretoggle.HoldoutEnabled() {
		return []entities.Holdout{}
	}
	// Get flag ID from key
	feature, exists := c.featureMap[flagKey]
	if !exists {
		return []entities.Holdout{}
	}

	flagID := feature.ID

	// Check cache first
	if cachedHoldouts, exists := c.flagHoldoutsMap[flagID]; exists {
		return cachedHoldouts
	}

	holdouts := []entities.Holdout{}

	// Add global holdouts that don't exclude this flag
	for _, holdout := range c.globalHoldouts {
		isExcluded := false
		for _, excludedFlagID := range holdout.ExcludedFlags {
			if excludedFlagID == flagID {
				isExcluded = true
				break
			}
		}
		if !isExcluded {
			holdouts = append(holdouts, holdout)
		}
	}

	// Add holdouts that specifically include this flag
	if includedHoldouts, exists := c.includedHoldouts[flagID]; exists {
		holdouts = append(holdouts, includedHoldouts...)
	}

	// Cache the result
	c.flagHoldoutsMap[flagID] = holdouts

	return holdouts
}

// GetHoldout returns a holdout by its ID
func (c DatafileProjectConfig) GetHoldout(holdoutID string) (entities.Holdout, error) {
	// Return error if holdouts are disabled
	if !featuretoggle.HoldoutEnabled() {
		return entities.Holdout{}, fmt.Errorf(`holdout with ID "%s" not found`, holdoutID)
	}
	if holdout, ok := c.holdoutIDMap[holdoutID]; ok {
		return holdout, nil
	}
	return entities.Holdout{}, fmt.Errorf(`holdout with ID "%s" not found`, holdoutID)
}

// SendFlagDecisions determines whether impressions events are sent for ALL decision types
func (c DatafileProjectConfig) SendFlagDecisions() bool {
	return c.sendFlagDecisions
}

// GetFlagVariationsMap returns map containing all variations for each flag
func (c DatafileProjectConfig) GetFlagVariationsMap() map[string][]entities.Variation {
	return c.flagVariationsMap
}

// GetRegion returns the region for the datafile
func (c DatafileProjectConfig) GetRegion() string {
	return c.region
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

	// Process holdouts
	holdoutIDMap := make(map[string]entities.Holdout)
	globalHoldouts := []entities.Holdout{}
	includedHoldouts := make(map[string][]entities.Holdout)
	excludedHoldouts := make(map[string][]entities.Holdout)
	flagHoldoutsMap := make(map[string][]entities.Holdout)

	// Only process holdouts if the feature is enabled
	if featuretoggle.HoldoutEnabled() {
		for _, datafileHoldout := range datafile.Holdouts {
			// Only process running holdouts
			if datafileHoldout.Status != datafileEntities.HoldoutStatusRunning {
				continue
			}

			// Create runtime holdout entity
			holdout := entities.Holdout{
				ID:            datafileHoldout.ID,
				Key:           datafileHoldout.Key,
				Status:        entities.HoldoutStatus(datafileHoldout.Status),
				IncludedFlags: datafileHoldout.IncludedFlags,
				ExcludedFlags: datafileHoldout.ExcludedFlags,
			}

			// Add to ID map
			holdoutIDMap[holdout.ID] = holdout

			// Categorize holdouts based on flag targeting
			if len(datafileHoldout.IncludedFlags) == 0 {
				// This is a global holdout (applies to all flags unless excluded)
				globalHoldouts = append(globalHoldouts, holdout)

				// Add to excluded flags map
				for _, flagID := range datafileHoldout.ExcludedFlags {
					excludedHoldouts[flagID] = append(excludedHoldouts[flagID], holdout)
				}
			} else {
				// This holdout specifically includes certain flags
				for _, flagID := range datafileHoldout.IncludedFlags {
					includedHoldouts[flagID] = append(includedHoldouts[flagID], holdout)
				}
			}
		}
	}

	attributeKeyMap := make(map[string]entities.Attribute)
	attributeIDToKeyMap := make(map[string]string)

	for id, attribute := range attributeMap {
		attributeIDToKeyMap[id] = attribute.Key
		attributeKeyMap[attribute.Key] = attribute
	}

	region := datafile.Region
	if region == "" {
		region = "US"
	}

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
		attributeKeyMap:      attributeKeyMap,
		attributeIDToKeyMap:  attributeIDToKeyMap,
		region:               region,
		holdoutIDMap:         holdoutIDMap,
		globalHoldouts:       globalHoldouts,
		includedHoldouts:     includedHoldouts,
		excludedHoldouts:     excludedHoldouts,
		flagHoldoutsMap:      flagHoldoutsMap,
	}

	logger.Info("Datafile is valid.")
	return config, nil
}
