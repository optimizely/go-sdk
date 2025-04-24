/****************************************************************************
 * Copyright 2019,2021-2022, Optimizely, Inc. and contributors              *
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
	"os"
	"path/filepath"
	"testing"

	"github.com/optimizely/go-sdk/v2/pkg/entities"
	"github.com/optimizely/go-sdk/v2/pkg/logging"

	"github.com/stretchr/testify/assert"
)

func TestNewDatafileProjectConfigNil(t *testing.T) {
	projectConfig, err := NewDatafileProjectConfig(nil, logging.GetLogger("", "DatafileProjectConfig"))
	assert.Error(t, err)
	assert.Nil(t, projectConfig)

	jsonDatafileStr := `{"accountID": "123", "revision": "1", "projectId": "12345", "version": "2"}`
	projectConfig, err = NewDatafileProjectConfig([]byte(jsonDatafileStr), logging.GetLogger("", "DatafileProjectConfig"))
	assert.Error(t, err)
	assert.Nil(t, projectConfig)
}

func TestNewDatafileProjectConfigNotNil(t *testing.T) {
	dpc := DatafileProjectConfig{accountID: "123", revision: "1", projectID: "12345", sdkKey: "a", environmentKey: "production", eventMap: map[string]entities.Event{"event_single_targeted_exp": {Key: "event_single_targeted_exp"}}, attributeMap: map[string]entities.Attribute{"10401066170": {ID: "10401066170"}}, integrations: []entities.Integration{{PublicKey: "123", Host: "www.123.com", Key: "odp"}}}
	jsonDatafileStr := `{"accountID":"123","revision":"1","projectId":"12345","version":"4","sdkKey":"a","environmentKey":"production","events":[{"key":"event_single_targeted_exp"}],"attributes":[{"id":"10401066170"}],"integrations": [{"publicKey": "123", "host": "www.123.com", "key": "odp"}]}`
	jsonDatafile := []byte(jsonDatafileStr)
	projectConfig, err := NewDatafileProjectConfig(jsonDatafile, logging.GetLogger("", "DatafileProjectConfig"))
	assert.Nil(t, err)
	assert.NotNil(t, projectConfig)
	assert.Equal(t, dpc.accountID, projectConfig.accountID)
	assert.Equal(t, dpc.revision, projectConfig.revision)
	assert.Equal(t, dpc.projectID, projectConfig.projectID)
	assert.Equal(t, dpc.environmentKey, projectConfig.environmentKey)
	assert.Equal(t, dpc.sdkKey, projectConfig.sdkKey)
	assert.Equal(t, dpc.integrations, projectConfig.integrations)
}

func TestNewDatafileProjectConfigWithODP(t *testing.T) {
	expectedIntegrations := [][]entities.Integration{
		{{PublicKey: "123", Host: "www.123.com", Key: "odp"}},
		{{PublicKey: "123", Host: "www.123.com", Key: "odp"}},
		{{Key: "odp"}},
		{{PublicKey: "123", Host: "www.123.com"}},
		{{PublicKey: "123", Host: "www.123.com"}},
	}

	expectedErrors := []error{
		nil,
		nil,
		nil,
		errors.New(""),
		nil,
	}

	jsonDatafileStrs := []string{
		// odp without extra keys
		`{"version":"4","integrations": [{"publicKey": "123", "host": "www.123.com", "key": "odp"}]}`,
		// odp with extra keys
		`{"version":"4","integrations": [{"publicKey": "123", "host": "www.123.com", "key": "odp", "key1": "odp", "key2": "odp"}]}`,
		// odp with missing host and public key keys
		`{"version":"4","integrations": [{"key": "odp"}]}`,
		// odp with missing key
		`{"version":"4","integrations": [{"publicKey": "123", "host": "www.123.com"}]}`,
		// odp with empty key
		`{"version":"4","integrations": [{"key": "", "publicKey": "123", "host": "www.123.com"}]}`,
	}

	for i := 0; i < len(expectedIntegrations); i++ {
		jsonDatafile := []byte(jsonDatafileStrs[i])
		projectConfig, err := NewDatafileProjectConfig(jsonDatafile, logging.GetLogger("", "DatafileProjectConfig"))
		assert.Equal(t, expectedErrors[i] == nil, err == nil)
		if expectedErrors[i] == nil {
			assert.NotNil(t, projectConfig)
			assert.Equal(t, expectedIntegrations[i], projectConfig.integrations)
		}
	}
}

func TestGetDatafile(t *testing.T) {
	jsonDatafileStr := `{"accountID": "123", "revision": "1", "projectId": "12345", "version": "4", "sdkKey": "a", "environmentKey": "production","integrations": [{"publicKey": "123", "host": "www.123.com", "key": "odp"}]}`
	jsonDatafile := []byte(jsonDatafileStr)
	config := &DatafileProjectConfig{
		datafile: string(jsonDatafile),
	}

	assert.Equal(t, string(jsonDatafile), config.GetDatafile())
}

func TestGetHostAndPublicKeyForValidODP(t *testing.T) {
	jsonDatafileStr := `{"version": "4","integrations": [{"publicKey": "1234", "host": "www.1234.com", "key": "non-odp"},{"randomKey":"123", "publicKey": "123", "host": "www.123.com", "key": "123"},{"publicKey": "123", "host": "www.123.com", "key": "odp"}]}`
	jsonDatafile := []byte(jsonDatafileStr)
	config, err := NewDatafileProjectConfig(jsonDatafile, logging.GetLogger("", ""))
	assert.NoError(t, err)
	assert.Equal(t, string(jsonDatafile), config.GetDatafile())
	assert.Equal(t, "www.123.com", config.GetHostForODP())
	assert.Equal(t, "123", config.GetPublicKeyForODP())
}

func TestGetHostAndPublicKeyForInvalidODP(t *testing.T) {
	jsonDatafileStr := `{"version": "4","integrations": [{"publicKey": "1234", "host": "www.1234.com", "key": "non-odp"},{"randomKey":"123", "publicKey": "123", "host": "www.123.com", "key": "123"}]}`
	jsonDatafile := []byte(jsonDatafileStr)
	config, err := NewDatafileProjectConfig(jsonDatafile, logging.GetLogger("", ""))
	assert.NoError(t, err)
	assert.Equal(t, string(jsonDatafile), config.GetDatafile())
	assert.Equal(t, "", config.GetHostForODP())
	assert.Equal(t, "", config.GetPublicKeyForODP())
}

func TestGetProjectID(t *testing.T) {
	projectID := "projectID"
	config := &DatafileProjectConfig{
		projectID: projectID,
	}

	assert.Equal(t, projectID, config.GetProjectID())
}

func TestGetRevision(t *testing.T) {
	revision := "revision"
	config := &DatafileProjectConfig{
		revision: revision,
	}

	assert.Equal(t, revision, config.GetRevision())
}

func TestGetAccountID(t *testing.T) {
	accountID := "accountID"
	config := &DatafileProjectConfig{
		accountID: accountID,
	}

	assert.Equal(t, accountID, config.GetAccountID())
}

func TestGetAnonymizeIP(t *testing.T) {
	anonymizeIP := true
	config := &DatafileProjectConfig{
		anonymizeIP: anonymizeIP,
	}

	assert.Equal(t, anonymizeIP, config.GetAnonymizeIP())
}

func TestGetAttributes(t *testing.T) {
	config := &DatafileProjectConfig{
		attributeMap: map[string]entities.Attribute{"id1": {ID: "id1", Key: "key"}, "id2": {ID: "id1", Key: "key"}},
	}

	assert.Equal(t, []entities.Attribute{config.attributeMap["id1"], config.attributeMap["id2"]}, config.GetAttributes())
}

func TestGetAttributeID(t *testing.T) {
	id := "id"
	key := "key"
	attributeKeyToIDMap := make(map[string]string)
	attributeKeyToIDMap[key] = id
	config := &DatafileProjectConfig{
		attributeKeyToIDMap: attributeKeyToIDMap,
	}

	assert.Equal(t, id, config.GetAttributeID(key))
}

func TestGetSdkKey(t *testing.T) {
	config := &DatafileProjectConfig{
		sdkKey: "1",
	}

	assert.Equal(t, "1", config.GetSdkKey())
}

func TestGetEnvironmentKey(t *testing.T) {
	config := &DatafileProjectConfig{
		environmentKey: "production",
	}
	assert.Equal(t, "production", config.GetEnvironmentKey())
}

func TestGetEvents(t *testing.T) {
	config := &DatafileProjectConfig{
		eventMap: map[string]entities.Event{"key": {ID: "5", Key: "key"}},
	}
	assert.Equal(t, []entities.Event{config.eventMap["key"]}, config.GetEvents())
}

func TestGetBotFiltering(t *testing.T) {
	botFiltering := true
	config := &DatafileProjectConfig{
		botFiltering: botFiltering,
	}

	assert.Equal(t, botFiltering, config.GetBotFiltering())
}

func TestGetEventByKey(t *testing.T) {
	key := "key"
	event := entities.Event{
		Key: key,
	}
	eventMap := make(map[string]entities.Event)
	eventMap[key] = event
	config := &DatafileProjectConfig{
		eventMap: eventMap,
	}

	actual, err := config.GetEventByKey(key)
	assert.Nil(t, err)
	assert.Equal(t, event, actual)
}

func TestGetEventByKeyWithError(t *testing.T) {
	config := &DatafileProjectConfig{}
	_, err := config.GetEventByKey("key")
	if assert.Error(t, err) {
		assert.Equal(t, fmt.Errorf(`event with key "key" not found`), err)
	}
}

func TestGetFeatureByKey(t *testing.T) {
	key := "key"
	feature := entities.Feature{
		Key: key,
	}
	featureMap := make(map[string]entities.Feature)
	featureMap[key] = feature
	config := &DatafileProjectConfig{
		featureMap: featureMap,
	}

	actual, err := config.GetFeatureByKey(key)
	assert.Nil(t, err)
	assert.Equal(t, feature, actual)
}

func TestGetFeatureByKeyWithError(t *testing.T) {
	config := &DatafileProjectConfig{}
	_, err := config.GetFeatureByKey("key")
	if assert.Error(t, err) {
		assert.Equal(t, fmt.Errorf(`feature with key "key" not found`), err)
	}
}

func TestGetVariableByKey(t *testing.T) {
	featureKey := "featureKey"
	variableKey := "variable"

	variable := entities.Variable{
		Key: variableKey,
	}

	variableMap := map[string]entities.Variable{variable.Key: variable}

	feature := entities.Feature{
		VariableMap: variableMap,
	}

	featureMap := make(map[string]entities.Feature)
	featureMap[featureKey] = feature

	config := &DatafileProjectConfig{
		featureMap: featureMap,
	}

	actual, err := config.GetVariableByKey(featureKey, variableKey)
	assert.Nil(t, err)
	assert.Equal(t, variable, actual)
}

func TestGetVariableByKeyWithMissingFeatureError(t *testing.T) {
	config := &DatafileProjectConfig{}
	_, err := config.GetVariableByKey("featureKey", "variableKey")
	if assert.Error(t, err) {
		assert.Equal(t, fmt.Errorf(`variable with key "featureKey" not found`), err)
	}
}

func TestGetVariableByKeyWithMissingVariableError(t *testing.T) {
	featureKey := "featureKey"
	variableKey := "variableKey"

	feature := entities.Feature{}

	featureMap := make(map[string]entities.Feature)
	featureMap[featureKey] = feature

	config := &DatafileProjectConfig{
		featureMap: featureMap,
	}

	_, err := config.GetVariableByKey(featureKey, variableKey)
	if assert.Error(t, err) {
		assert.Equal(t, fmt.Errorf(`variable with key "featureKey" not found`), err)
	}
}

func TestGetAttributeByKey(t *testing.T) {
	key := "key"
	attribute := entities.Attribute{
		Key: key,
	}

	// The old and new mappings to ensure backward compatibility
	attributeKeyMap := make(map[string]entities.Attribute)
	attributeKeyMap[key] = attribute

	id := "id"
	attributeKeyToIDMap := make(map[string]string)
	attributeKeyToIDMap[key] = id

	attributeMap := make(map[string]entities.Attribute)
	attributeMap[id] = attribute

	config := &DatafileProjectConfig{
		attributeKeyMap:     attributeKeyMap,
		attributeKeyToIDMap: attributeKeyToIDMap,
		attributeMap:        attributeMap,
	}

	actual, err := config.GetAttributeByKey(key)
	assert.Nil(t, err)
	assert.Equal(t, attribute, actual)
}

func TestGetAttributeByKeyWithMissingKeyError(t *testing.T) {
	config := &DatafileProjectConfig{}

	_, err := config.GetAttributeByKey("key")
	if assert.Error(t, err) {
		assert.Equal(t, fmt.Errorf(`attribute with key "key" not found`), err)
	}
}

func TestGetAttributeByKeyWithMissingAttributeId(t *testing.T) {
	id := "id"
	key := "key"
	attributeKeyToIDMap := make(map[string]string)
	attributeKeyToIDMap[key] = id

	config := &DatafileProjectConfig{
		attributeKeyToIDMap: attributeKeyToIDMap,
	}

	_, err := config.GetAttributeByKey(key)
	if assert.Error(t, err) {
		assert.Equal(t, fmt.Errorf(`attribute with key "key" not found`), err)
	}
}

func TestGetFeatureList(t *testing.T) {
	key := "key"
	feature := entities.Feature{
		Key: key,
	}
	featureMap := make(map[string]entities.Feature)
	featureMap[key] = feature
	config := &DatafileProjectConfig{
		featureMap: featureMap,
	}

	features := config.GetFeatureList()

	assert.Equal(t, 1, len(features))
	assert.Equal(t, feature, features[0])
}

func TestGetExperimentList(t *testing.T) {
	id := "id"
	key := "key"
	experimentKeyToIDMap := make(map[string]string)
	experimentKeyToIDMap[key] = id

	experiment := entities.Experiment{
		Key: key,
	}

	experimentMap := make(map[string]entities.Experiment)
	experimentMap[id] = experiment

	config := &DatafileProjectConfig{
		experimentKeyToIDMap: experimentKeyToIDMap,
		experimentMap:        experimentMap,
	}
	experiments := config.GetExperimentList()

	assert.Equal(t, 1, len(experiments))
	assert.Equal(t, experiment, experiments[0])
}

func TestGetRolloutList(t *testing.T) {
	config := &DatafileProjectConfig{
		rollouts: []entities.Rollout{{ID: "5"}},
	}
	assert.Equal(t, config.rollouts, config.GetRolloutList())
}

func TestGetIntegrationListODP(t *testing.T) {
	jsonDatafileStr := `{"version": "4","integrations": [{"publicKey": "1234", "host": "www.1234.com", "key": "non-odp"},{"publicKey": "123", "host": "www.123.com", "key": "odp"},{"randomKey":"123", "publicKey": "123", "host": "www.123.com", "key": "123"}]}`
	jsonDatafile := []byte(jsonDatafileStr)
	config, err := NewDatafileProjectConfig(jsonDatafile, logging.GetLogger("", ""))
	assert.NoError(t, err)
	assert.Equal(t, string(jsonDatafile), config.GetDatafile())
	expectedIntegrations := []entities.Integration{{Host: "www.1234.com", PublicKey: "1234", Key: "non-odp"}, {Host: "www.123.com", PublicKey: "123", Key: "odp"}, {Host: "www.123.com", PublicKey: "123", Key: "123"}}
	assert.Equal(t, expectedIntegrations, config.GetIntegrationList())
}

func TestGetSegmentList(t *testing.T) {
	jsonDatafileStr := `{"version": "4","typedAudiences": [{"id": "22290411365", "conditions": ["and", ["or", ["or", {"value": "order_last_7_days", "type": "third_party_dimension", "name": "odp.audiences", "match": "qualified"}, {"value": "favoritecolorred", "type": "third_party_dimension", "name": "odp.audiences", "match": "qualified"}]]], "name": "sohail"}, {"id": "22282671333", "conditions": ["and", ["or", ["or", {"value": "has_email_or_phone_opted_in", "type": "third_party_dimension", "name": "odp.audiences", "match": "qualified"}]]], "name": "audience_age"}]}`
	jsonDatafile := []byte(jsonDatafileStr)
	expectedSegmentList := []string{"order_last_7_days", "favoritecolorred", "has_email_or_phone_opted_in"}
	config, err := NewDatafileProjectConfig(jsonDatafile, logging.GetLogger("", ""))
	assert.NoError(t, err)
	assert.Equal(t, expectedSegmentList, config.GetSegmentList())
}

func TestGetAudienceList(t *testing.T) {
	config := &DatafileProjectConfig{
		audienceMap: map[string]entities.Audience{"5": {ID: "5", Name: "one"}, "6": {ID: "6", Name: "two"}},
	}
	assert.ElementsMatch(t, []entities.Audience{config.audienceMap["5"], config.audienceMap["6"]}, config.GetAudienceList())
}

func TestGetAudienceByID(t *testing.T) {
	id := "id"
	audience := entities.Audience{
		ID: id,
	}
	audienceMap := make(map[string]entities.Audience)
	audienceMap[id] = audience
	config := &DatafileProjectConfig{
		audienceMap: audienceMap,
	}

	actual, err := config.GetAudienceByID(id)
	assert.Nil(t, err)
	assert.Equal(t, audience, actual)
}

func TestGetAudienceByIDMissingIDError(t *testing.T) {
	config := &DatafileProjectConfig{}
	_, err := config.GetAudienceByID("id")
	if assert.Error(t, err) {
		assert.Equal(t, fmt.Errorf(`audience with ID "id" not found`), err)
	}
}

func TestGetAudienceMap(t *testing.T) {
	audienceMap := make(map[string]entities.Audience)
	audienceMap["key"] = entities.Audience{}

	config := &DatafileProjectConfig{
		audienceMap: audienceMap,
	}

	assert.Equal(t, audienceMap, config.GetAudienceMap())
}

func TestGetExperimentByKey(t *testing.T) {
	id := "id"
	key := "key"
	experimentKeyToIDMap := make(map[string]string)
	experimentKeyToIDMap[key] = id

	experiment := entities.Experiment{
		Key: key,
	}
	experimentMap := make(map[string]entities.Experiment)
	experimentMap[id] = experiment

	config := &DatafileProjectConfig{
		experimentKeyToIDMap: experimentKeyToIDMap,
		experimentMap:        experimentMap,
	}

	actual, err := config.GetExperimentByKey(key)
	assert.Nil(t, err)
	assert.Equal(t, experiment, actual)
}

func TestGetExperimentByKeyMissingKeyError(t *testing.T) {
	config := &DatafileProjectConfig{}
	_, err := config.GetExperimentByKey("key")
	if assert.Error(t, err) {
		assert.Equal(t, fmt.Errorf(`experiment with key "key" not found`), err)
	}
}

func TestGetExperimentByKeyMissingIDError(t *testing.T) {
	id := "id"
	key := "key"
	experimentKeyToIDMap := make(map[string]string)
	experimentKeyToIDMap[key] = id

	config := &DatafileProjectConfig{
		experimentKeyToIDMap: experimentKeyToIDMap,
	}

	_, err := config.GetExperimentByKey(key)
	if assert.Error(t, err) {
		assert.Equal(t, fmt.Errorf(`experiment with key "key" not found`), err)
	}
}

func TestGetGroupByID(t *testing.T) {
	id := "id"
	group := entities.Group{
		ID: id,
	}
	groupMap := make(map[string]entities.Group)
	groupMap[id] = group
	config := &DatafileProjectConfig{
		groupMap: groupMap,
	}

	actual, err := config.GetGroupByID(id)
	assert.Nil(t, err)
	assert.Equal(t, group, actual)
}

func TestSendFlagDecisions(t *testing.T) {
	config := &DatafileProjectConfig{
		sendFlagDecisions: true,
	}
	assert.Equal(t, config.sendFlagDecisions, config.SendFlagDecisions())
}

func TestGetGroupByIDMissingIDError(t *testing.T) {
	config := &DatafileProjectConfig{}
	_, err := config.GetGroupByID("id")
	if assert.Error(t, err) {
		assert.Equal(t, fmt.Errorf(`group with ID "id" not found`), err)
	}
}

func TestGetFlagVariationsMap(t *testing.T) {
	absPath, _ := filepath.Abs("../../../test-data/decide-test-datafile.json")
	datafile, err := os.ReadFile(absPath)
	assert.NoError(t, err)
	config, err := NewDatafileProjectConfig(datafile, logging.GetLogger("", ""))
	assert.NoError(t, err)
	flagVariationsMap := config.GetFlagVariationsMap()

	variationsMap := map[string]bool{"a": true, "b": true, "3324490633": true, "3324490562": true, "18257766532": true}
	for _, variation := range flagVariationsMap["feature_1"] {
		assert.True(t, variationsMap[variation.Key])
	}

	variationsMap = map[string]bool{"variation_with_traffic": true, "variation_no_traffic": true}
	for _, variation := range flagVariationsMap["feature_2"] {
		assert.True(t, variationsMap[variation.Key])
	}

	assert.NotNil(t, flagVariationsMap["feature_3"])
	assert.Len(t, flagVariationsMap["feature_3"], 0)
}

func TestCmabExperiments(t *testing.T) {
	// Load the decide-test-datafile.json
	absPath, _ := filepath.Abs("../../../test-data/decide-test-datafile.json")
	datafile, err := os.ReadFile(absPath)
	assert.NoError(t, err)

	// Parse the datafile to modify it
	var datafileJSON map[string]interface{}
	err = json.Unmarshal(datafile, &datafileJSON)
	assert.NoError(t, err)

	// Add CMAB to the first experiment with traffic allocation as an integer
	experiments := datafileJSON["experiments"].([]interface{})
	exp0 := experiments[0].(map[string]interface{})
	exp0["cmab"] = map[string]interface{}{
		"attributeIds":      []string{"808797688", "808797689"},
		"trafficAllocation": 5000, // Changed from array to integer
	}

	// Convert back to JSON
	modifiedDatafile, err := json.Marshal(datafileJSON)
	assert.NoError(t, err)

	// Create project config from modified datafile
	config, err := NewDatafileProjectConfig(modifiedDatafile, logging.GetLogger("", "DatafileProjectConfig"))
	assert.NoError(t, err)

	// Get the experiment key from the datafile
	exp0Key := exp0["key"].(string)

	// Test that Cmab fields are correctly mapped for experiment 0
	experiment0, err := config.GetExperimentByKey(exp0Key)
	assert.NoError(t, err)
	assert.NotNil(t, experiment0.Cmab)
	if experiment0.Cmab != nil {
		// Test attribute IDs
		assert.Equal(t, 2, len(experiment0.Cmab.AttributeIds))
		assert.Contains(t, experiment0.Cmab.AttributeIds, "808797688")
		assert.Contains(t, experiment0.Cmab.AttributeIds, "808797689")

		// Test traffic allocation as integer
		assert.Equal(t, 5000, experiment0.Cmab.TrafficAllocation)
	}
}

func TestCmabExperimentsNil(t *testing.T) {
	// Load the decide-test-datafile.json (which doesn't have CMAB by default)
	absPath, _ := filepath.Abs("../../../test-data/decide-test-datafile.json")
	datafile, err := os.ReadFile(absPath)
	assert.NoError(t, err)

	// Create project config from the original datafile
	config, err := NewDatafileProjectConfig(datafile, logging.GetLogger("", "DatafileProjectConfig"))
	assert.NoError(t, err)

	// Parse the datafile to get experiment keys
	var datafileJSON map[string]interface{}
	err = json.Unmarshal(datafile, &datafileJSON)
	assert.NoError(t, err)

	experiments := datafileJSON["experiments"].([]interface{})
	exp0 := experiments[0].(map[string]interface{})
	exp0Key := exp0["key"].(string)

	// Test that Cmab field is nil for experiment 0
	experiment0, err := config.GetExperimentByKey(exp0Key)
	assert.NoError(t, err)
	assert.Nil(t, experiment0.Cmab, "CMAB field should be nil when not present in datafile")

	// Test another experiment if available
	if len(experiments) > 1 {
		exp1 := experiments[1].(map[string]interface{})
		exp1Key := exp1["key"].(string)

		experiment1, err := config.GetExperimentByKey(exp1Key)
		assert.NoError(t, err)
		assert.Nil(t, experiment1.Cmab, "CMAB field should be nil when not present in datafile")
	}
}

func TestGetAttributeKeyByID(t *testing.T) {
	// Setup
	id := "id"
	key := "key"
	attributeIDToKeyMap := make(map[string]string)
	attributeIDToKeyMap[id] = key

	config := &DatafileProjectConfig{
		attributeIDToKeyMap: attributeIDToKeyMap,
	}

	// Test successful case
	actual, err := config.GetAttributeKeyByID(id)
	assert.Nil(t, err)
	assert.Equal(t, key, actual)
}

func TestGetAttributeKeyByIDWithMissingIDError(t *testing.T) {
	// Setup
	config := &DatafileProjectConfig{}

	// Test error case
	_, err := config.GetAttributeKeyByID("id")
	if assert.Error(t, err) {
		assert.Equal(t, fmt.Errorf(`attribute with ID "id" not found`), err)
	}
}

func TestGetAttributeByKeyWithDirectMapping(t *testing.T) {
	key := "key"
	attribute := entities.Attribute{
		Key: key,
	}

	attributeKeyMap := make(map[string]entities.Attribute)
	attributeKeyMap[key] = attribute

	config := &DatafileProjectConfig{
		attributeKeyMap: attributeKeyMap,
	}

	actual, err := config.GetAttributeByKey(key)
	assert.Nil(t, err)
	assert.Equal(t, attribute, actual)
}
