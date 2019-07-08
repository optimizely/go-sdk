package event

import (
	guuid "github.com/google/uuid"
	"github.com/optimizely/go-sdk/optimizely/config"
	"github.com/optimizely/go-sdk/optimizely/entities"
	"strings"
	"time"
)

const impressionKey string = "campaign_activated"
const clientKey string = "go-sdk"
const clientVersion string = "1.0.0"
const attributeType = "custom"
const specialPrefix = "$opt_"
const botFiltering = "$opt_bot_filtering"

func MakeTimestamp() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

func CreateEventContext(config config.ProjectConfig) EventContext {
	context := EventContext{}
	context.ProjectID = config.GetProjectID()
	context.Revision = config.GetRevision()
	context.AccountID = config.GetAccountID()
	context.ClientName = clientKey
	context.ClientVersion = clientVersion
	context.AnonymizeIP = config.GetAnonymizeIP()

	return context
}

func CreateImpressionEvent(config config.ProjectConfig, experiment entities.Experiment,
	variation entities.Variation, attributes map[string]interface{}) ImpressionEvent {

	impression := ImpressionEvent{}
	impression.Key = impressionKey
	impression.EntityID = experiment.LayerID
	impression.Attributes = GetEventAttributes(config, attributes)
	impression.VariationID = variation.ID
	impression.ExperimentID = experiment.ID
	impression.CampaignID = experiment.LayerID

	return impression
}

func CreateImpressionUserEvent(config config.ProjectConfig, experiment entities.Experiment,
	variation entities.Variation,
	userId string,
	attributes map[string]interface{}) UserEvent {

	impression := CreateImpressionEvent(config, experiment, variation, attributes)

	context := CreateEventContext(config)

	userEvent := UserEvent{}
	userEvent.Timestamp = MakeTimestamp()
	userEvent.VisitorID = userId
	userEvent.Uuid = guuid.New().String()
	userEvent.Impression = &impression
	userEvent.EventContext = context

	return userEvent
}

func CreateImpressionBatchEvent(userEvent UserEvent) EventBatch {

	decision := Decision{}
	decision.CampaignID = userEvent.Impression.CampaignID
	decision.ExperimentID = userEvent.Impression.ExperimentID
	decision.VariationID = userEvent.Impression.VariationID

	dispatchEvent := DispatchEvent{}
	dispatchEvent.Timestamp = MakeTimestamp()
	dispatchEvent.Key = userEvent.Impression.Key
	dispatchEvent.EntityID = userEvent.Impression.EntityID
	dispatchEvent.Uuid = guuid.New().String()
	dispatchEvent.Tags = make(map[string]interface{})

	return CreateBatchEvent(userEvent, userEvent.Impression.Attributes, [] Decision{decision}, []DispatchEvent{dispatchEvent})

}

func CreateConversionEvent(config config.ProjectConfig, eventKey string, attributes map[string]interface{}, eventTags map[string]interface{}) ConversionEvent {
	conversion := ConversionEvent{}

	event, err := config.GetEventByKey(eventKey)

	if err != nil {
		return conversion
	}
	conversion.Key = event.Key
	conversion.EntityID = event.ID
	conversion.Tags = eventTags
	conversion.Attributes = GetEventAttributes(config, attributes)

	return conversion
}
func CreateConversionUserEvent(config config.ProjectConfig, eventKey string, userId string, attributes map[string]interface{}, eventTags map[string]interface{}) UserEvent {


	context := CreateEventContext(config)

	userEvent := UserEvent{}
	userEvent.Timestamp = MakeTimestamp()
	userEvent.VisitorID = userId
	userEvent.Uuid = guuid.New().String()

	userEvent.EventContext = context
	conversion := CreateConversionEvent(config, eventKey, attributes, eventTags)
	userEvent.Conversion = &conversion

	return userEvent

}
func CreateConversionBatchEvent(userEvent UserEvent) EventBatch {

	dispatchEvent := DispatchEvent{}
	dispatchEvent.Timestamp = MakeTimestamp()
	dispatchEvent.Key = userEvent.Conversion.Key
	dispatchEvent.EntityID = userEvent.Conversion.EntityID
	dispatchEvent.Uuid = userEvent.Uuid
	dispatchEvent.Tags = userEvent.Conversion.Tags

	return CreateBatchEvent(userEvent, userEvent.Conversion.Attributes, [] Decision{}, []DispatchEvent{dispatchEvent})
}

func CreateBatchEvent(userEvent UserEvent, attributes []EventAttribute,
	decisions []Decision,
	dispatchEvents []DispatchEvent) EventBatch {

	snapShot := Snapshot{}
	snapShot.Decisions = decisions
	snapShot.Events = dispatchEvents

	eventAttributes := attributes

	visitor := Visitor{}

	visitor.Attributes = eventAttributes
	visitor.Snapshots = []Snapshot{snapShot}
	visitor.VisitorID = userEvent.VisitorID

	eventBatch := EventBatch{}
	eventBatch.ProjectID = userEvent.EventContext.ProjectID
	eventBatch.Revision = userEvent.EventContext.Revision
	eventBatch.AccountID = userEvent.EventContext.AccountID
	eventBatch.Visitors = []Visitor{visitor}
	eventBatch.ClientName = clientKey
	eventBatch.ClientVersion = clientVersion
	eventBatch.AnonymizeIP = userEvent.EventContext.AnonymizeIP
	eventBatch.EnrichDecisions = true

	return eventBatch
}

func GetEventAttributes(config config.ProjectConfig, attributes map[string]interface{}) []EventAttribute {
	var eventAttributes = []EventAttribute{}

	for key, value := range attributes {
		if value == nil {
			continue
		}
		attribute := EventAttribute{}
		id := config.GetAttributeID(key)
		if id != "" {
			attribute.EntityID = id
		} else if strings.HasPrefix(key, specialPrefix) {
			attribute.EntityID = key
		} else {
			continue
		}
		attribute.Value = value
		attribute.AttributeType = attributeType

		eventAttributes = append(eventAttributes, attribute)
	}

	attribute := EventAttribute{}
	attribute.Value = config.GetBotFiltering()
	attribute.AttributeType = attributeType
	attribute.Key = botFiltering
	attribute.EntityID = botFiltering
	eventAttributes = append(eventAttributes, attribute)

	return eventAttributes
}