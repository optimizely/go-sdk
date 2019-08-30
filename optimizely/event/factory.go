package event

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/optimizely/go-sdk/optimizely"

	guuid "github.com/google/uuid"
	"github.com/optimizely/go-sdk/optimizely/entities"
	"github.com/optimizely/go-sdk/optimizely/logging"
	"github.com/optimizely/go-sdk/optimizely/utils"
)

var efLogger = logging.GetLogger("EventFactory")

const impressionKey string = "campaign_activated"
const clientKey string = optimizely.ClientName
const clientVersion string = optimizely.Version
const attributeType = "custom"
const specialPrefix = "$opt_"
const botFilteringKey = "$opt_bot_filtering"
const eventEndPoint = "https://logx.optimizely.com/v1/events"
const revenueKey = "revenue"
const valueKey = "value"

func createLogEvent(event Batch) LogEvent {
	return LogEvent{EndPoint: eventEndPoint, Event: event}
}

func makeTimestamp() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

// CreateEventContext creates and returns EventContext
func CreateEventContext(projectConfig optimizely.ProjectConfig) Context {
	context := Context{}
	context.ProjectID = projectConfig.GetProjectID()
	context.Revision = projectConfig.GetRevision()
	context.AccountID = projectConfig.GetAccountID()
	context.ClientName = clientKey
	context.ClientVersion = clientVersion
	context.AnonymizeIP = projectConfig.GetAnonymizeIP()
	context.BotFiltering = projectConfig.GetBotFiltering()

	return context
}

func createImpressionEvent(projectConfig optimizely.ProjectConfig, experiment entities.Experiment,
	variation entities.Variation, attributes map[string]interface{}) ImpressionEvent {

	impression := ImpressionEvent{}
	impression.Key = impressionKey
	impression.EntityID = experiment.LayerID
	impression.Attributes = getEventAttributes(projectConfig, attributes)
	impression.VariationID = variation.ID
	impression.ExperimentID = experiment.ID
	impression.CampaignID = experiment.LayerID

	return impression
}

// CreateImpressionUserEvent creates and returns ImpressionEvent for user
func CreateImpressionUserEvent(projectConfig optimizely.ProjectConfig, experiment entities.Experiment,
	variation entities.Variation,
	userContext entities.UserContext) UserEvent {

	impression := createImpressionEvent(projectConfig, experiment, variation, userContext.Attributes)

	userEvent := UserEvent{}
	userEvent.Timestamp = makeTimestamp()
	userEvent.VisitorID = userContext.ID
	userEvent.UUID = guuid.New().String()
	userEvent.Impression = &impression
	userEvent.EventContext = CreateEventContext(projectConfig)

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
func createConversionEvent(projectConfig optimizely.ProjectConfig, event entities.Event, attributes map[string]interface{}, eventTags map[string]interface{}) ConversionEvent {
	conversion := ConversionEvent{}

	conversion.Key = event.Key
	conversion.EntityID = event.ID
	conversion.Tags = eventTags
	conversion.Attributes = getEventAttributes(projectConfig, attributes)

	return conversion
}

// CreateConversionUserEvent creates and returns ConversionEvent for user
func CreateConversionUserEvent(projectConfig optimizely.ProjectConfig, event entities.Event, userContext entities.UserContext, eventTags map[string]interface{}) UserEvent {

	userEvent := UserEvent{}
	userEvent.Timestamp = makeTimestamp()
	userEvent.VisitorID = userContext.ID
	userEvent.UUID = guuid.New().String()

	userEvent.EventContext = CreateEventContext(projectConfig)
	conversion := createConversionEvent(projectConfig, event, userContext.Attributes, eventTags)
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
	eventBatch.ClientName = userEvent.EventContext.ClientName
	eventBatch.ClientVersion = userEvent.EventContext.ClientVersion
	eventBatch.AnonymizeIP = userEvent.EventContext.AnonymizeIP
	eventBatch.EnrichDecisions = true

	return eventBatch
}

// get visitor attributes from user attributes
func getEventAttributes(projectConfig optimizely.ProjectConfig, attributes map[string]interface{}) []VisitorAttribute {
	var eventAttributes = []VisitorAttribute{}

	for key, value := range attributes {
		if value == nil {
			continue
		}
		visitorAttribute := VisitorAttribute{}
		attribute, _ := projectConfig.GetAttributeByKey(key)
		if attribute.ID != "" {
			visitorAttribute.EntityID = attribute.ID
		} else if strings.HasPrefix(key, specialPrefix) {
			visitorAttribute.EntityID = key
		} else {
			efLogger.Debug(fmt.Sprintf("Unrecognized attribute %s provided. Pruning before sending event to Optimizely.", key))
			continue
		}
		visitorAttribute.Key = attribute.Key
		visitorAttribute.Value = value
		visitorAttribute.AttributeType = attributeType

		eventAttributes = append(eventAttributes, visitorAttribute)
	}

	visitorAttribute := VisitorAttribute{}
	visitorAttribute.Value = projectConfig.GetBotFiltering()
	visitorAttribute.AttributeType = attributeType
	visitorAttribute.Key = botFilteringKey
	visitorAttribute.EntityID = botFilteringKey
	eventAttributes = append(eventAttributes, visitorAttribute)

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
