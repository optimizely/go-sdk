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
	attributes map[string]interface{}) LogEvent {

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

	return CreateLogEvent(config, userId, attributes, [] Decision{decision}, []DispatchEvent{dispatchEvent})
}


func CreateLogEvent(config config.ProjectConfig,
	userId string,
	attributes map[string]interface{},
	decisions []Decision,
	dispatchEvents []DispatchEvent) LogEvent {

	snapShot := Snapshot{}
	snapShot.Decisions = decisions
	snapShot.Events = dispatchEvents

	eventAttributes := GetEventAttributes(config, attributes)

	visitor := Visitor{}

	visitor.Attributes = eventAttributes
	visitor.Snapshots = []Snapshot{snapShot}
	visitor.VisitorID = userId


	logEvent := LogEvent{}
	logEvent.ProjectID = config.GetProjectID()
	logEvent.Revision = config.GetRevision()
	logEvent.AccountID = config.GetAccountID()
	logEvent.Visitors = []Visitor{visitor}
	logEvent.ClientName = clientKey
	logEvent.ClientVersion = clientVersion
	logEvent.AnonymizeIP = config.GetAnonymizeIP()
	logEvent.EnrichDecisions = true

	return logEvent
}

func GetEventAttributes(config config.ProjectConfig, attributes map[string]interface{}) []EventAttribute {
	var eventAttributes = []EventAttribute{}

	for key, value := range attributes {
		if value == nil {
			continue
		}
		attribute := EventAttribute{}
		st := config.GetAttributeID(key)
		if st != "" {
			attribute.EntityID = st
		} else if strings.HasPrefix(key, specialPrefix) {
			attribute.EntityID = key
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