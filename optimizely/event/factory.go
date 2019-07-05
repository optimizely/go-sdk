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

func CreateImpressionEvent(config config.ProjectConfig, experiment entities.Experiment,
	variation entities.Variation,
	userId string,
	attributes map[string]interface{}) EventBatch {

	decision := Decision{}
	decision.CampaignID = experiment.LayerID
	decision.ExperimentID = experiment.ID
	decision.VariationID = variation.ID

	dispatchEvent := DispatchEvent{}
	dispatchEvent.Timestamp = MakeTimestamp()
	dispatchEvent.Key = impressionKey
	dispatchEvent.EntityID = experiment.LayerID
	dispatchEvent.Uuid = guuid.New().String()
	dispatchEvent.Tags = make(map[string]interface{})

	return CreateBatchEvent(config, userId, attributes, [] Decision{decision}, []DispatchEvent{dispatchEvent})
}

func CreateConversionEvent(config config.ProjectConfig, eventKey string, userId string, attributes map[string]interface{}, eventTags map[string]interface{}) (EventBatch, error) {


	event, err := config.GetEventByKey(eventKey)

	if err != nil {
		return EventBatch{}, err
	}
	dispatchEvent := DispatchEvent{}
	dispatchEvent.Timestamp = MakeTimestamp()
	dispatchEvent.Key = event.Key
	dispatchEvent.EntityID = event.ID
	dispatchEvent.Uuid = guuid.New().String()
	dispatchEvent.Tags = eventTags

	return CreateBatchEvent(config, userId, attributes, [] Decision{}, []DispatchEvent{dispatchEvent}), nil

}

func CreateBatchEvent(config config.ProjectConfig,
	userId string,
	attributes map[string]interface{},
	decisions []Decision,
	dispatchEvents []DispatchEvent) EventBatch {

	snapShot := Snapshot{}
	snapShot.Decisions = decisions
	snapShot.Events = dispatchEvents

	eventAttributes := GetEventAttributes(config, attributes)

	visitor := Visitor{}

	visitor.Attributes = eventAttributes
	visitor.Snapshots = []Snapshot{snapShot}
	visitor.VisitorID = userId


	eventBatch := EventBatch{}
	eventBatch.ProjectID = config.GetProjectID()
	eventBatch.Revision = config.GetRevision()
	eventBatch.AccountID = config.GetAccountID()
	eventBatch.Visitors = []Visitor{visitor}
	eventBatch.ClientName = clientKey
	eventBatch.ClientVersion = clientVersion
	eventBatch.AnonymizeIP = config.GetAnonymizeIP()
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