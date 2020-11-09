/****************************************************************************
 * Copyright 2019-2020, Optimizely, Inc. and contributors                   *
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
	"errors"
	"strings"
	"time"

	guuid "github.com/google/uuid"

	"github.com/optimizely/go-sdk/pkg/config"
	decisionPkg "github.com/optimizely/go-sdk/pkg/decision"
	"github.com/optimizely/go-sdk/pkg/entities"
	"github.com/optimizely/go-sdk/pkg/utils"
)

const impressionKey string = "campaign_activated"
const attributeType = "custom"
const specialPrefix = "$opt_"
const botFilteringKey = "$opt_bot_filtering"
const revenueKey = "revenue"
const valueKey = "value"

func createLogEvent(event Batch, eventEndPoint string) LogEvent {
	return LogEvent{EndPoint: eventEndPoint, Event: event}
}

func makeTimestamp() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

// CreateEventContext creates and returns EventContext
func CreateEventContext(projectConfig config.ProjectConfig) Context {
	context := Context{}
	context.ProjectID = projectConfig.GetProjectID()
	context.Revision = projectConfig.GetRevision()
	context.AccountID = projectConfig.GetAccountID()
	context.ClientName = ClientName
	context.ClientVersion = Version
	context.AnonymizeIP = projectConfig.GetAnonymizeIP()
	context.BotFiltering = projectConfig.GetBotFiltering()

	return context
}

func createImpressionEvent(
	projectConfig config.ProjectConfig,
	experiment entities.Experiment,
	variation *entities.Variation,
	attributes map[string]interface{},
	flagKey, ruleKey, ruleType string, enabled bool,
) ImpressionEvent {

	metadata := DecisionMetadata{
		FlagKey:  flagKey,
		RuleKey:  ruleKey,
		RuleType: ruleType,
		Enabled:  enabled,
	}

	var variationID string
	if variation != nil {
		metadata.VariationKey = variation.Key
		variationID = variation.ID
	}

	event := ImpressionEvent{
		Attributes:   getEventAttributes(projectConfig, attributes),
		CampaignID:   experiment.LayerID,
		EntityID:     experiment.LayerID,
		ExperimentID: experiment.ID,
		Key:          impressionKey,
		Metadata:     metadata,
		VariationID:  variationID,
	}

	return event
}

// CreateImpressionUserEvent creates and returns ImpressionEvent for user
func CreateImpressionUserEvent(projectConfig config.ProjectConfig, experiment entities.Experiment,
	variation *entities.Variation, userContext entities.UserContext, flagKey, ruleKey, ruleType string, enabled bool) (UserEvent, bool) {

	if (ruleType == decisionPkg.Rollout || variation == nil) && !projectConfig.SendFlagDecisions() {
		return UserEvent{}, false
	}

	impression := createImpressionEvent(projectConfig, experiment, variation, userContext.Attributes, flagKey, ruleKey, ruleType, enabled)

	userEvent := UserEvent{}
	userEvent.Timestamp = makeTimestamp()
	userEvent.VisitorID = userContext.ID
	userEvent.UUID = guuid.New().String()
	userEvent.Impression = &impression
	userEvent.EventContext = CreateEventContext(projectConfig)

	return userEvent, true
}

// create an impression visitor
func createImpressionVisitor(userEvent UserEvent) Visitor {
	decision := Decision{}
	decision.CampaignID = userEvent.Impression.CampaignID
	decision.ExperimentID = userEvent.Impression.ExperimentID
	decision.VariationID = userEvent.Impression.VariationID
	decision.Metadata = userEvent.Impression.Metadata

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
func createConversionEvent(projectConfig config.ProjectConfig, event entities.Event, attributes, eventTags map[string]interface{}) ConversionEvent {
	conversion := ConversionEvent{}

	conversion.Key = event.Key
	conversion.EntityID = event.ID
	conversion.Tags = eventTags
	conversion.Attributes = getEventAttributes(projectConfig, attributes)

	return conversion
}

// CreateConversionUserEvent creates and returns ConversionEvent for user
func CreateConversionUserEvent(projectConfig config.ProjectConfig, event entities.Event, userContext entities.UserContext, eventTags map[string]interface{}) UserEvent {

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
func getEventAttributes(projectConfig config.ProjectConfig, attributes map[string]interface{}) []VisitorAttribute {
	var eventAttributes = []VisitorAttribute{}

	for key, value := range attributes {
		if value == nil {
			continue
		}
		visitorAttribute := VisitorAttribute{}
		attribute, _ := projectConfig.GetAttributeByKey(key)

		switch {
		case attribute.ID != "":
			visitorAttribute.EntityID = attribute.ID
		case strings.HasPrefix(key, specialPrefix):
			visitorAttribute.EntityID = key
		default:
			continue
		}
		visitorAttribute.Key = key
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

	return 0, errors.New("no event tag found for revenue")
}

// get a value attribute
func getTagValue(eventTags map[string]interface{}) (float64, error) {
	if value, ok := eventTags[valueKey]; ok {
		return utils.GetFloatValue(value)
	}

	return 0, errors.New("no event tag found for value")
}
