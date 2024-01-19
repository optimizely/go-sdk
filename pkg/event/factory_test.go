/****************************************************************************
 * Copyright 2019,2022 Optimizely, Inc. and contributors                    *
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

// Package event //
package event

import (
	"context"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/optimizely/go-sdk/v2/pkg/config"
	"github.com/optimizely/go-sdk/v2/pkg/decision"
	"github.com/optimizely/go-sdk/v2/pkg/entities"
)

var testExperiment = entities.Experiment{
	Key:     "background_experiment",
	LayerID: "15399420423",
	ID:      "15402980349",
}

var testVariation = entities.Variation{
	Key: "variation_a",
	ID:  "15410990633",
}

type TestConfig struct {
	config.ProjectConfig
	sendFlagDecisions bool
}

func (TestConfig) GetAttributeByKey(string) (entities.Attribute, error) {
	return entities.Attribute{ID: "100000", Key: "sample_attribute"}, nil
}

func (TestConfig) GetEventByKey(string) (entities.Event, error) {
	return entities.Event{ExperimentIds: []string{"15402980349"}, ID: "15368860886", Key: "sample_conversion"}, nil
}

func (TestConfig) GetFeatureByKey(string) (entities.Feature, error) {
	return entities.Feature{}, nil
}

func (TestConfig) GetProjectID() string {
	return "15389410617"
}
func (TestConfig) GetRevision() string {
	return "7"
}
func (TestConfig) GetAccountID() string {
	return "8362480420"
}
func (TestConfig) GetAnonymizeIP() bool {
	return true
}
func (TestConfig) GetAttributeID(key string) string { // returns "" if there is no id
	return ""
}
func (TestConfig) GetBotFiltering() bool {
	return false
}
func (TestConfig) GetClientName() string {
	return "go-sdk"
}
func (TestConfig) GetClientVersion() string {
	return "1.0.0"
}
func (t TestConfig) SendFlagDecisions() bool {
	return t.sendFlagDecisions
}

func RandomString(len int) string {
	bytes := make([]byte, len)
	rand.Seed(time.Now().UnixNano())
	for i := 0; i < len; i++ {
		bytes[i] = byte(65 + rand.Intn(25)) //A=65 and Z = 65+25
	}
	return string(bytes)
}

var userID = RandomString(10)
var userContext = entities.UserContext{
	ID:         userID,
	Attributes: map[string]interface{}{"test": "val"},
}

func BuildTestImpressionEvent() UserEvent {
	tc := TestConfig{}
	impressionUserEvent, _ := CreateImpressionUserEvent(tc, testExperiment, &testVariation, userContext, "", testExperiment.Key, "experiment", true)
	return impressionUserEvent
}

func BuildTestConversionEvent() UserEvent {
	tc := TestConfig{}
	conversionUserEvent := CreateConversionUserEvent(tc, entities.Event{ExperimentIds: []string{"15402980349"}, ID: "15368860886", Key: "sample_conversion"}, userContext, make(map[string]interface{}))

	return conversionUserEvent
}

func TestVisitorTimestampMatchesUserEventTimestamp(t *testing.T) {
	impressionUserEvent := BuildTestImpressionEvent()
	visitor := createVisitorFromUserEvent(impressionUserEvent)
	assert.Equal(t, impressionUserEvent.Timestamp, visitor.Snapshots[0].Events[0].Timestamp)
}

func TestCreateEmptyEvent(t *testing.T) {
	impressionUserEvent := BuildTestImpressionEvent()

	impressionUserEvent.Impression = nil
	impressionUserEvent.Conversion = nil

	visitor := createVisitorFromUserEvent(impressionUserEvent)

	assert.Nil(t, visitor.Snapshots)

}

func TestCreateAndSendImpressionEvent(t *testing.T) {
	impressionUserEvent := BuildTestImpressionEvent()
	assert.Equal(t, userContext.ID, impressionUserEvent.VisitorID)
	assert.Equal(t, "experiment", impressionUserEvent.Impression.Metadata.RuleType)
	assert.Equal(t, testVariation.Key, impressionUserEvent.Impression.Metadata.VariationKey)
	assert.Equal(t, testExperiment.Key, impressionUserEvent.Impression.Metadata.RuleKey)

	processor := NewBatchEventProcessor(WithBatchSize(10), WithQueueSize(100),
		WithFlushInterval(10),
		WithEventDispatcher(&MockDispatcher{Events: NewInMemoryQueue(100)}))

	go processor.Start(context.Background())

	processor.ProcessEvent(impressionUserEvent)

	assert.Equal(t, 1, processor.eventsCount())

	time.Sleep(100 * time.Millisecond)

	assert.Equal(t, 0, processor.eventsCount())
}

func TestCreateAndSendConversionEvent(t *testing.T) {

	conversionUserEvent := BuildTestConversionEvent()

	processor := NewBatchEventProcessor(WithFlushInterval(10),
		WithEventDispatcher(&MockDispatcher{Events: NewInMemoryQueue(100)}))

	go processor.Start(context.Background())

	processor.ProcessEvent(conversionUserEvent)

	assert.Equal(t, 1, processor.eventsCount())

	time.Sleep(100 * time.Millisecond)

	assert.Equal(t, 0, processor.eventsCount())
}

func TestCreateConversionEventRevenue(t *testing.T) {
	eventTags := map[string]interface{}{"revenue": 55.0, "value": 25.1}
	tc := TestConfig{}

	conversionUserEvent := CreateConversionUserEvent(tc, entities.Event{ExperimentIds: []string{"15402980349"}, ID: "15368860886", Key: "sample_conversion"}, userContext, eventTags)

	assert.Equal(t, int64(55), *conversionUserEvent.Conversion.Revenue)
	assert.Equal(t, 25.1, *conversionUserEvent.Conversion.Value)

	batch := createBatchEvent(conversionUserEvent, createVisitorFromUserEvent(conversionUserEvent))
	assert.Equal(t, conversionUserEvent.Timestamp, batch.Visitors[0].Snapshots[0].Events[0].Timestamp)
	assert.Equal(t, int64(55), *batch.Visitors[0].Snapshots[0].Events[0].Revenue)
	assert.Equal(t, 25.1, *batch.Visitors[0].Snapshots[0].Events[0].Value)

}

func TestCreateImpressionUserEvent(t *testing.T) {
	tc := TestConfig{}

	scenarios := []struct {
		flagType string
		expected bool
	}{
		{decision.FeatureTest, true},
		{"experiment", true},
		{"anything-else", true},
		{decision.Rollout, false},
	}

	for _, scenario := range scenarios {
		userEvent, ok := CreateImpressionUserEvent(tc, testExperiment, &testVariation, userContext, "", testExperiment.Key, scenario.flagType, true)
		assert.Equal(t, ok, scenario.expected)

		if ok {
			metaData := userEvent.Impression.Metadata
			assert.Equal(t, "", metaData.FlagKey)
			assert.Equal(t, testExperiment.Key, metaData.RuleKey)
			assert.Equal(t, scenario.flagType, metaData.RuleType)
			assert.Equal(t, true, metaData.Enabled)
		}
	}

	// nil variation should _always_ return false
	for _, scenario := range scenarios {
		userEvent, ok := CreateImpressionUserEvent(tc, testExperiment, nil, userContext, "", testExperiment.Key, scenario.flagType, false)
		assert.False(t, ok)
		if ok {
			metaData := userEvent.Impression.Metadata
			assert.Equal(t, "", metaData.FlagKey)
			assert.Equal(t, testExperiment.Key, metaData.RuleKey)
			assert.Equal(t, scenario.flagType, metaData.RuleType)
			assert.Equal(t, false, metaData.Enabled)
		}
	}

	// should _always_ return true if sendFlagDecisions is set
	tc.sendFlagDecisions = true
	for _, scenario := range scenarios {
		userEvent, ok := CreateImpressionUserEvent(tc, testExperiment, nil, userContext, "", testExperiment.Key, scenario.flagType, true)
		assert.True(t, ok)
		if ok {
			metaData := userEvent.Impression.Metadata
			assert.Equal(t, "", metaData.FlagKey)
			assert.Equal(t, testExperiment.Key, metaData.RuleKey)
			assert.Equal(t, scenario.flagType, metaData.RuleType)
			assert.Equal(t, true, metaData.Enabled)
		}
	}
}
