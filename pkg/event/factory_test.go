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

// Package event //
package event

import (
	"math/rand"
	"testing"
	"time"

	"github.com/optimizely/go-sdk/pkg"
	"github.com/optimizely/go-sdk/pkg/entities"
	"github.com/optimizely/go-sdk/pkg/utils"
	"github.com/stretchr/testify/assert"
)

type TestConfig struct {
	pkg.ProjectConfig
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
	Attributes: make(map[string]interface{}),
}

func BuildTestImpressionEvent() UserEvent {
	config := TestConfig{}

	experiment := entities.Experiment{}
	experiment.Key = "background_experiment"
	experiment.LayerID = "15399420423"
	experiment.ID = "15402980349"

	variation := entities.Variation{}
	variation.Key = "variation_a"
	variation.ID = "15410990633"

	impressionUserEvent := CreateImpressionUserEvent(config, experiment, variation, userContext)

	return impressionUserEvent
}

func BuildTestConversionEvent() UserEvent {
	config := TestConfig{}
	conversionUserEvent := CreateConversionUserEvent(config, entities.Event{ExperimentIds: []string{"15402980349"}, ID: "15368860886", Key: "sample_conversion"}, userContext, make(map[string]interface{}))

	return conversionUserEvent
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

	processor := NewEventProcessor(BatchSize(10), QueueSize(100), FlushInterval(100))

	processor.Start(utils.NewCancelableExecutionCtx())

	processor.ProcessEvent(impressionUserEvent)

	assert.Equal(t, 1, processor.EventsCount())

	time.Sleep(2000 * time.Millisecond)

	assert.Equal(t, 0, processor.EventsCount())
}

func TestCreateAndSendConversionEvent(t *testing.T) {

	conversionUserEvent := BuildTestConversionEvent()

	processor := NewEventProcessor(FlushInterval(100))

	processor.Start(utils.NewCancelableExecutionCtx())

	processor.ProcessEvent(conversionUserEvent)

	assert.Equal(t, 1, processor.EventsCount())

	time.Sleep(2000 * time.Millisecond)

	assert.Equal(t, 0, processor.EventsCount())
}

func TestCreateConversionEventRevenue(t *testing.T) {
	eventTags := map[string]interface{}{"revenue": 55.0, "value": 25.1}
	config := TestConfig{}

	conversionUserEvent := CreateConversionUserEvent(config, entities.Event{ExperimentIds: []string{"15402980349"}, ID: "15368860886", Key: "sample_conversion"}, userContext, eventTags)

	assert.Equal(t, int64(55), *conversionUserEvent.Conversion.Revenue)
	assert.Equal(t, 25.1, *conversionUserEvent.Conversion.Value)

	batch := createBatchEvent(conversionUserEvent, createVisitorFromUserEvent(conversionUserEvent))
	assert.Equal(t, int64(55), *batch.Visitors[0].Snapshots[0].Events[0].Revenue)
	assert.Equal(t, 25.1, *batch.Visitors[0].Snapshots[0].Events[0].Value)

}
