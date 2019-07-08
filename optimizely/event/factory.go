package event

import (
	guuid "github.com/google/uuid"
	"github.com/optimizely/go-sdk/optimizely/entities"
	"strings"
	"time"
)

const impressionKey string = "campaign_activated"
const clientKey string = "go-sdk"
const clientVersion string = "1.0.0"
const attributeType = "custom"
const specialPrefix = "$opt_"
const botFilteringKey = "$opt_bot_filtering"
const eventEndPoint = "https://logx.optimizely.com/v1/events"

func createLogEvent(event EventBatch) LogEvent {
	return LogEvent{endPoint:eventEndPoint,event:event}
}

func makeTimestamp() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

func CreateEventContext(projectID string, revision string, accountID string, anonymizeIP bool, botFiltering bool, attributeKeyToIdMap map[string]string) EventContext {
	context := EventContext{}
	context.ProjectID = projectID
	context.Revision = revision
	context.AccountID = accountID
	context.ClientName = clientKey
	context.ClientVersion = clientVersion
	context.AnonymizeIP = anonymizeIP
	context.BotFiltering = botFiltering
	context.AttributeKeyToIdMap = attributeKeyToIdMap

	return context
}

func createImpressionEvent(context EventContext, experiment entities.Experiment,
	variation entities.Variation, attributes map[string]interface{}) ImpressionEvent {

	impression := ImpressionEvent{}
	impression.Key = impressionKey
	impression.EntityID = experiment.LayerID
	impression.Attributes = getEventAttributes(context.AttributeKeyToIdMap, attributes, context.BotFiltering)
	impression.VariationID = variation.ID
	impression.ExperimentID = experiment.ID
	impression.CampaignID = experiment.LayerID

	return impression
}

func CreateImpressionUserEvent(context EventContext, experiment entities.Experiment,
	variation entities.Variation,
	userId string,
	attributes map[string]interface{}) UserEvent {

	impression := createImpressionEvent(context, experiment, variation, attributes)

	userEvent := UserEvent{}
	userEvent.Timestamp = makeTimestamp()
	userEvent.VisitorID = userId
	userEvent.Uuid = guuid.New().String()
	userEvent.Impression = &impression
	userEvent.EventContext = context

	return userEvent
}

func createImpressionBatchEvent(userEvent UserEvent) EventBatch {

	decision := Decision{}
	decision.CampaignID = userEvent.Impression.CampaignID
	decision.ExperimentID = userEvent.Impression.ExperimentID
	decision.VariationID = userEvent.Impression.VariationID

	dispatchEvent := DispatchEvent{}
	dispatchEvent.Timestamp = makeTimestamp()
	dispatchEvent.Key = userEvent.Impression.Key
	dispatchEvent.EntityID = userEvent.Impression.EntityID
	dispatchEvent.Uuid = guuid.New().String()
	dispatchEvent.Tags = make(map[string]interface{})

	return createBatchEvent(userEvent, userEvent.Impression.Attributes, [] Decision{decision}, []DispatchEvent{dispatchEvent})

}

func CreateConversionEvent(attributeKeyToIdMap map[string]string, event entities.Event, attributes map[string]interface{}, eventTags map[string]interface{}, botFiltering bool) ConversionEvent {
	conversion := ConversionEvent{}

	conversion.Key = event.Key
	conversion.EntityID = event.ID
	conversion.Tags = eventTags
	conversion.Attributes = getEventAttributes(attributeKeyToIdMap, attributes, botFiltering)

	return conversion
}
func CreateConversionUserEvent(context EventContext, event entities.Event, userId string, attributes map[string]interface{}, attributeKeyToIdMap map[string]string, eventTags map[string]interface{}) UserEvent {


	userEvent := UserEvent{}
	userEvent.Timestamp = makeTimestamp()
	userEvent.VisitorID = userId
	userEvent.Uuid = guuid.New().String()

	userEvent.EventContext = context
	conversion := CreateConversionEvent(attributeKeyToIdMap, event, attributes, eventTags, context.BotFiltering)
	userEvent.Conversion = &conversion

	return userEvent

}
func createConversionBatchEvent(userEvent UserEvent) EventBatch {

	dispatchEvent := DispatchEvent{}
	dispatchEvent.Timestamp = makeTimestamp()
	dispatchEvent.Key = userEvent.Conversion.Key
	dispatchEvent.EntityID = userEvent.Conversion.EntityID
	dispatchEvent.Uuid = userEvent.Uuid
	dispatchEvent.Tags = userEvent.Conversion.Tags

	return createBatchEvent(userEvent, userEvent.Conversion.Attributes, [] Decision{}, []DispatchEvent{dispatchEvent})
}

func createBatchEvent(userEvent UserEvent, attributes []VisitorAttribute,
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

func getEventAttributes(attributeKeyToIdMap map[string]string, attributes map[string]interface{}, botFiltering bool) []VisitorAttribute {
	var eventAttributes = []VisitorAttribute{}

	for key, value := range attributes {
		if value == nil {
			continue
		}
		attribute := VisitorAttribute{}
		id := attributeKeyToIdMap[key]
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

	attribute := VisitorAttribute{}
	attribute.Value = botFiltering
	attribute.AttributeType = attributeType
	attribute.Key = botFilteringKey
	attribute.EntityID = botFilteringKey
	eventAttributes = append(eventAttributes, attribute)

	return eventAttributes
}