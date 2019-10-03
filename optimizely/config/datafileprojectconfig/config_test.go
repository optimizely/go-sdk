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
	"testing"

	"github.com/optimizely/go-sdk/optimizely/entities"

	"github.com/stretchr/testify/assert"
)

func TestNewDatafileProjectConfigNil(t *testing.T) {
	projectConfig, err := NewDatafileProjectConfig(nil)
	assert.NotNil(t, err)
	assert.Nil(t, projectConfig)
}

func TestNewDatafileProjectConfigNotNil(t *testing.T) {
	dpc := DatafileProjectConfig{accountID: "123", revision: "1", projectID: "12345"}
	jsonDatafileStr := `{"accountID": "123", "revision": "1", "projectId": "12345"}`
	jsonDatafile := []byte(jsonDatafileStr)
	projectConfig, err := NewDatafileProjectConfig(jsonDatafile)
	assert.Nil(t, err)
	assert.NotNil(t, projectConfig)
	assert.Equal(t, dpc.accountID, projectConfig.accountID)
	assert.Equal(t, dpc.revision, projectConfig.revision)
	assert.Equal(t, dpc.projectID, projectConfig.projectID)
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

	variables := make([]entities.Variable, 1)
	variables[0] = variable

	feature := entities.Feature{
		Variables: variables,
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
	id := "id"
	key := "key"
	attributeKeyToIDMap := make(map[string]string)
	attributeKeyToIDMap[key] = id

	attribute := entities.Attribute{
		Key: key,
	}
	attributeMap := make(map[string]entities.Attribute)
	attributeMap[id] = attribute

	config := &DatafileProjectConfig{
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

func TestGetGroupByIDMissingIDError(t *testing.T) {
	config := &DatafileProjectConfig{}
	_, err := config.GetGroupByID("id")
	if assert.Error(t, err) {
		assert.Equal(t, fmt.Errorf(`group with ID "id" not found`), err)
	}
}
