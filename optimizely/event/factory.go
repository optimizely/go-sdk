package event

import (
	"errors"
	"strings"
	"time"

	guuid "github.com/google/uuid"
	"github.com/optimizely/go-sdk/optimizely/entities"
	"github.com/optimizely/go-sdk/optimizely/utils"
)

const impressionKey string = "campaign_activated"
const clientKey string = "go-sdk"
const clientVersion string = "1.0.0"
const attributeType = "custom"
const specialPrefix = "$opt_"
const botFilteringKey = "$opt_bot_filtering"
const eventEndPoint = "https://logx.optimizely.com/v1/events"
const revenueKey = "revenue"
const valueKey = "value"

func createLogEvent(event Batch) LogEvent {
	return LogEvent{endPoint: eventEndPoint, event: event}
}

func makeTimestamp() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

// CreateEventContext creates and returns EventContext
func CreateEventContext(projectID string, revision string, accountID string, anonymizeIP bool, botFiltering bool, attributeKeyToIDMap map[string]string) Context {
	context := Context{}
	context.ProjectID = projectID
	context.Revision = revision
	context.AccountID = accountID
	context.ClientName = clientKey
	context.ClientVersion = clientVersion
	context.AnonymizeIP = anonymizeIP
	context.BotFiltering = botFiltering
	context.attributeKeyToIDMap = attributeKeyToIDMap

	return context
}

func createImpressionEvent(context Context, experiment entities.Experiment,
	variation entities.Variation, attributes map[string]interface{}) ImpressionEvent {

	impression := ImpressionEvent{}
	impression.Key = impressionKey
	impression.EntityID = experiment.LayerID
	impression.Attributes = getEventAttributes(context.attributeKeyToIDMap, attributes, context.BotFiltering)
	impression.VariationID = variation.ID
	impression.ExperimentID = experiment.ID
	impression.CampaignID = experiment.LayerID

	return impression
}

// CreateImpressionUserEvent creates and returns ImpressionEvent for user
func CreateImpressionUserEvent(context Context, experiment entities.Experiment,
	variation entities.Variation,
	userContext entities.UserContext) UserEvent {

	impression := createImpressionEvent(context, experiment, variation, userContext.Attributes)

	userEvent := UserEvent{}
	userEvent.Timestamp = makeTimestamp()
	userEvent.VisitorID = userContext.ID
	userEvent.UUID = guuid.New().String()
	userEvent.Impression = &impression
	userEvent.EventContext = context

	return userEvent
}

// create an impression visitor
func createImpressionVisitor(userEvent UserEvent) Visitor {
	decision := Decision{}
	decision.CampaignID = userEvent.Impression.CampaignID
	decision.ExperimentID = userEvent.Impression.ExperimentID
	decision.VariationID = userEvent.Impression.VariationID

	dispatchEvent := SnapshotEvent{}
	dispatchEvent.Timestamp = makeTimestamp()
	dispatchEvent.Key = userEvent.Impression.Key
	dispatchEvent.EntityID = userEvent.Impression.EntityID
	dispatchEvent.UUID = guuid.New().String()
	dispatchEvent.Tags = make(map[string]interface{})

	visitor := createVisitor(userEvent, userEvent.Impression.Attributes, []Decision{decision}, []SnapshotEvent{dispatchEvent})

	return visitor
}

// create a conversion event
func createConversionEvent(attributeKeyToIDMap map[string]string, event entities.Event, attributes map[string]interface{}, eventTags map[string]interface{}, botFiltering bool) ConversionEvent {
	conversion := ConversionEvent{}

	conversion.Key = event.Key
	conversion.EntityID = event.ID
	conversion.Tags = eventTags
	conversion.Attributes = getEventAttributes(attributeKeyToIDMap, attributes, botFiltering)

	return conversion
}

// CreateConversionUserEvent creates and returns ConversionEvent for user
func CreateConversionUserEvent(context Context, event entities.Event, userContext entities.UserContext, attributeKeyToIDMap map[string]string, eventTags map[string]interface{}) UserEvent {

	userEvent := UserEvent{}
	userEvent.Timestamp = makeTimestamp()
	userEvent.VisitorID = userContext.ID
	userEvent.UUID = guuid.New().String()

	userEvent.EventContext = context
	conversion := createConversionEvent(attributeKeyToIDMap, event, userContext.Attributes, eventTags, context.BotFiltering)
	revenue, err := getRevenueValue(eventTags)
	if err == nil {
		conversion.Revenue = &revenue
	}
	// get value if available
	value, err := getTagValue(eventTags)
	if err == nil {
		conversion.Value = &value
	}
	userEvent.Conversion = &conversion

	return userEvent

}

// create visitor from user event
func createVisitorFromUserEvent(event UserEvent) Visitor {
	if event.Impression != nil {
		return createImpressionVisitor(event)
	}
	if event.Conversion != nil {
		return createConversionVisitor(event)
	}

	return Visitor{}
}

// create a conversion visitor
func createConversionVisitor(userEvent UserEvent) Visitor {

	dispatchEvent := SnapshotEvent{}
	dispatchEvent.Timestamp = makeTimestamp()
	dispatchEvent.Key = userEvent.Conversion.Key
	dispatchEvent.EntityID = userEvent.Conversion.EntityID
	dispatchEvent.UUID = userEvent.UUID
	dispatchEvent.Tags = userEvent.Conversion.Tags
	if userEvent.Conversion.Revenue != nil {
		dispatchEvent.Revenue = userEvent.Conversion.Revenue
	}
	if userEvent.Conversion.Value != nil {
		dispatchEvent.Value = userEvent.Conversion.Value
	}

	visitor := createVisitor(userEvent, userEvent.Conversion.Attributes, []Decision{}, []SnapshotEvent{dispatchEvent})

	return visitor
}

// create a visitor
func createVisitor(userEvent UserEvent, attributes []VisitorAttribute,
	decisions []Decision,
	dispatchEvents []SnapshotEvent) Visitor {
	snapShot := Snapshot{}
	snapShot.Decisions = decisions
	snapShot.Events = dispatchEvents

	eventAttributes := attributes

	visitor := Visitor{}

	visitor.Attributes = eventAttributes
	visitor.Snapshots = []Snapshot{snapShot}
	visitor.VisitorID = userEvent.VisitorID

	return visitor
}

// create a batch event with visitor
func createBatchEvent(userEvent UserEvent, visitor Visitor) Batch {


	eventBatch := Batch{}
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

// get visitor attributes from user attributes
func getEventAttributes(attributeKeyToIDMap map[string]string, attributes map[string]interface{}, botFiltering bool) []VisitorAttribute {
	var eventAttributes = []VisitorAttribute{}

	for key, value := range attributes {
		if value == nil {
			continue
		}
		attribute := VisitorAttribute{}
		id := attributeKeyToIDMap[key]
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

// get revenue attribute
func getRevenueValue(eventTags map[string]interface{}) (int64, error) {
	if value, ok := eventTags[revenueKey]; ok {
		return utils.GetIntValue(value)
	}

	return 0, errors.New("No event tag found for revenue")
}

// get a value attribute
func getTagValue(eventTags map[string]interface{}) (float64, error) {
	if value, ok := eventTags[valueKey]; ok {
		return utils.GetFloatValue(value)
	}

	return 0, errors.New("No event tag found for value")
}
