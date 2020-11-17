/****************************************************************************
 * Copyright 2020, Optimizely, Inc. and contributors                        *
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

// Context holds project-related contextual information about a UserEvent
type Context struct {
	Revision      string `json:"revision"`
	AccountID     string `json:"account_id"`
	ClientVersion string `json:"client_version"`
	ProjectID     string `json:"project_id"`
	ClientName    string `json:"client_name"`
	AnonymizeIP   bool   `json:"anonymize_ip"`
	BotFiltering  bool   `json:"bot_filtering"`
}

// UserEvent represents a user event
type UserEvent struct {
	Timestamp    int64  `json:"timestamp"`
	UUID         string `json:"uuid"`
	EventContext Context
	VisitorID    string
	Impression   *ImpressionEvent
	Conversion   *ConversionEvent
}

// ImpressionEvent represents an impression event
type ImpressionEvent struct {
	Attributes   []VisitorAttribute
	CampaignID   string           `json:"campaign_id"`
	EntityID     string           `json:"entity_id"`
	ExperimentID string           `json:"experiment_id"`
	Key          string           `json:"key"`
	Metadata     DecisionMetadata `json:"metadata"`
	VariationID  string           `json:"variation_id"`
}

// DecisionMetadata captures additional information regarding the decision
type DecisionMetadata struct {
	FlagKey      string `json:"flag_key"`
	RuleKey      string `json:"rule_key"`
	RuleType     string `json:"rule_type"`
	VariationKey string `json:"variation_key"`
	Enabled      bool   `json:"enabled"`
}

// ConversionEvent represents a conversion event
type ConversionEvent struct {
	EntityID   string `json:"entity_id"`
	Key        string `json:"key"`
	Attributes []VisitorAttribute
	Tags       map[string]interface{} `json:"tags,omitempty"`
	// these need to be pointers because 0 is a valid Revenue or Value.
	// 0 is equivalent to omitempty for json marshaling.
	Revenue *int64   `json:"revenue,omitempty"`
	Value   *float64 `json:"value,omitempty"`
}

// LogEvent represents a log event
type LogEvent struct {
	EndPoint string
	Event    Batch
}

// Batch - Context about the event to send in batch
type Batch struct {
	Revision        string    `json:"revision"`
	AccountID       string    `json:"account_id"`
	ClientVersion   string    `json:"client_version"`
	Visitors        []Visitor `json:"visitors"`
	ProjectID       string    `json:"project_id"`
	ClientName      string    `json:"client_name"`
	AnonymizeIP     bool      `json:"anonymize_ip"`
	EnrichDecisions bool      `json:"enrich_decisions"`
}

// Visitor represents a visitor of an eventbatch
type Visitor struct {
	Attributes []VisitorAttribute `json:"attributes"`
	Snapshots  []Snapshot         `json:"snapshots"`
	VisitorID  string             `json:"visitor_id"`
}

// VisitorAttribute represents an attribute of a visitor
type VisitorAttribute struct {
	Value         interface{} `json:"value"`
	Key           string      `json:"key"`
	AttributeType string      `json:"type"`
	EntityID      string      `json:"entity_id"`
}

// Snapshot represents a snapshot of a visitor
type Snapshot struct {
	Decisions []Decision      `json:"decisions,omitempty"`
	Events    []SnapshotEvent `json:"events"`
}

// Decision represents a decision of a snapshot
type Decision struct {
	VariationID  string           `json:"variation_id"`
	CampaignID   string           `json:"campaign_id"`
	ExperimentID string           `json:"experiment_id"`
	Metadata     DecisionMetadata `json:"metadata"`
}

// SnapshotEvent represents an event of a snapshot
type SnapshotEvent struct {
	EntityID  string                 `json:"entity_id"`
	Key       string                 `json:"key"`
	Timestamp int64                  `json:"timestamp"`
	UUID      string                 `json:"uuid"`
	Tags      map[string]interface{} `json:"tags,omitempty"`
	Revenue   *int64                 `json:"revenue,omitempty"`
	Value     *float64               `json:"value,omitempty"`
}
