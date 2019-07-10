package event

import (
	"github.com/optimizely/go-sdk/optimizely/entities"
	"github.com/stretchr/testify/assert"
	"math/rand"
	"testing"
	"time"
)

type TestConfig struct {
}

func (TestConfig) GetEventByKey(string) (entities.Event, error) {
	return entities.Event{ExperimentIds: []string{"15402980349"}, ID: "15368860886", Key: "sample_conversion"}, nil
}

func (TestConfig)GetFeatureByKey(string) (entities.Feature, error) {
	return entities.Feature{}, nil
}

func (TestConfig)GetProjectID() string {
	return "15389410617"
}
func (TestConfig)GetRevision()  string {
	return "7"
}
func (TestConfig)GetAccountID() string {
	return "8362480420"
}
func (TestConfig)GetAnonymizeIP() bool {
	return true
}
func (TestConfig)GetAttributeID(key string) string { // returns "" if there is no id
	return ""
}
func (TestConfig)GetBotFiltering() bool {
	return false
}

func RandomString(len int) string {
	bytes := make([]byte, len)
	rand.Seed(time.Now().UnixNano())
	for i := 0; i < len; i++ {
		bytes[i] = byte(65 + rand.Intn(25))  //A=65 and Z = 65+25
	}
	return string(bytes)
}

var userId = RandomString(10)
var userContext = entities.UserContext{userId, entities.UserAttributes{make(map[string]interface{})}}

func BuildTestImpressionEvent() UserEvent {
	config := TestConfig{}
	context := CreateEventContext(config.GetProjectID(), config.GetRevision(), config.GetAccountID(), config.GetAnonymizeIP(), config.GetBotFiltering(), make(map[string]string))

	experiment := entities.Experiment{}
	experiment.Key = "background_experiment"
	experiment.LayerID = "15399420423"
	experiment.ID = "15402980349"

	variation := entities.Variation{}
	variation.Key = "variation_a"
	variation.ID = "15410990633"

	impressionUserEvent := CreateImpressionUserEvent(context, experiment, variation, userContext)

	return impressionUserEvent
}

func BuildTestConversionEvent() UserEvent {
	config := TestConfig{}
	context := CreateEventContext(config.GetProjectID(), config.GetRevision(), config.GetAccountID(), config.GetAnonymizeIP(), config.GetBotFiltering(), make(map[string]string))

	conversionUserEvent := CreateConversionUserEvent(context, entities.Event{ExperimentIds: []string{"15402980349"}, ID: "15368860886", Key: "sample_conversion"}, userContext,make(map[string]string), make(map[string]interface{}))

	return conversionUserEvent
}

func TestCreateAndSendImpressionEvent(t *testing.T) {

	impressionUserEvent := BuildTestImpressionEvent()

	processor := NewEventProcessor(100, 100)

	processor.ProcessEvent(impressionUserEvent)

	result, ok := processor.(*QueueingEventProcessor)

	if ok {
		assert.Equal(t, 1, result.EventsCount())

		time.Sleep(2000 * time.Millisecond)

		assert.Equal(t, 0, result.EventsCount())
	}
}

func TestCreateAndSendConversionEvent(t *testing.T) {

	conversionUserEvent := BuildTestConversionEvent()

	processor := NewEventProcessor(100, 100)

	processor.ProcessEvent(conversionUserEvent)

	result, ok := processor.(*QueueingEventProcessor)

	if ok {
		assert.Equal(t, 1, result.EventsCount())

		time.Sleep(2000 * time.Millisecond)

		assert.Equal(t, 0, result.EventsCount())
	}
}

func TestCreateConversionEventRevenue(t *testing.T) {
	eventTags := map[string]interface{}{"revenue":55.0, "value":25.1}
	config := TestConfig{}
	context := CreateEventContext(config.GetProjectID(), config.GetRevision(), config.GetAccountID(), config.GetAnonymizeIP(), config.GetBotFiltering(), make(map[string]string))

	conversionUserEvent := CreateConversionUserEvent(context, entities.Event{ExperimentIds: []string{"15402980349"}, ID: "15368860886", Key: "sample_conversion"}, userContext,make(map[string]string), eventTags)

	assert.Equal(t, int64(55), *conversionUserEvent.Conversion.Revenue)
	assert.Equal(t, 25.1, *conversionUserEvent.Conversion.Value)

	batch := createConversionBatchEvent(conversionUserEvent)
	assert.Equal(t, int64(55), *batch.Visitors[0].Snapshots[0].Events[0].Revenue)
	assert.Equal(t, 25.1, *batch.Visitors[0].Snapshots[0].Events[0].Value)


}
