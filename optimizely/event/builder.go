package event

import (
	guuid "github.com/google/uuid"
	"github.com/optimizely/go-sdk/optimizely/config"
	"github.com/optimizely/go-sdk/optimizely/config/datafileProjectConfig/entities"
	"strings"
	"time"
)

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
	dispatchEvent.Key = "campaign_activated"
	dispatchEvent.EntityID = experiment.LayerID
	dispatchEvent.Uuid = guuid.New().String()

	return CreateBatchEvent(config, userId, attributes, [] Decision{decision}, []DispatchEvent{dispatchEvent})
}


func CreateBatchEvent(config config.ProjectConfig,
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
	logEvent.ClientName = "go-sdk"
	logEvent.ClientVersion = "1.0.0"
	logEvent.AnonymizeIP = config.GetAnonymizeIP()
	logEvent.EnrichDecisions = true

	return logEvent
}

func GetEventAttributes(config config.ProjectConfig, attributes map[string]interface{}) []EventAttribute {
	var eventAttributes = []EventAttribute{}

	if attributes != nil {
		for key, value := range attributes {
			if value != nil {
				attribute := EventAttribute{}
				st := config.GetAttributeID(key)
				if st != "" {
					attribute.EntityID = st
				} else if strings.HasPrefix(key, "$opt_") {
					attribute.EntityID = key
				}
				attribute.Value = value
				attribute.AttributeType = "custom"

				eventAttributes = append(eventAttributes, attribute)
			}
		}
	}
	if config.GetBotFiltering() {
		attribute := EventAttribute{}
		attribute.Value = true
		attribute.AttributeType = "custom"
		attribute.Key = "$opt_bot_filtering"
		attribute.EntityID = "$opt_bot_filtering"
		eventAttributes = append(eventAttributes, attribute)
	}

	return eventAttributes
}