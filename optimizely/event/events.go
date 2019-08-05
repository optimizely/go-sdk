package event

// EventContext respresents event context of a UserEvent
type EventContext struct {
	Revision            string            `json:"revision"`
	AccountID           string            `json:"account_id"`
	ClientVersion       string            `json:"client_version"`
	ProjectID           string            `json:"project_id"`
	ClientName          string            `json:"client_name"`
	AnonymizeIP         bool              `json:"anonymize_ip"`
	BotFiltering        bool              `json:"bot_filtering"`
	attributeKeyToIDMap map[string]string `json:"attributeKeyToIdMap"`
}

// UserEvent respresents a user event
type UserEvent struct {
	Timestamp    int64  `json:"timestamp"`
	UUID         string `json:"uuid"`
	EventContext EventContext
	VisitorID    string
	Impression   *ImpressionEvent
	Conversion   *ConversionEvent
}

// ImpressionEvent respresents a impression event
type ImpressionEvent struct {
	EntityID     string `json:"entity_id"`
	Key          string `json:"key"`
	Attributes   []VisitorAttribute
	VariationID  string `json:"variation_id"`
	CampaignID   string `json:"campaign_id"`
	ExperimentID string `json:"experiment_id"`
}

// ConversionEvent respresents a conversion event
type ConversionEvent struct {
	EntityID   string `json:"entity_id"`
	Key        string `json:"key"`
	Attributes []VisitorAttribute
	Tags       map[string]interface{} `json:"tags,omitempty"`
	// these need to be pointers because 0 is a valid Revenue or Value.
	// 0 is equivalent to omitempty for json marshalling.
	Revenue *int64   `json:"revenue,omitempty"`
	Value   *float64 `json:"value,omitempty"`
}

// LogEvent respresents a log event
type LogEvent struct {
	endPoint string
	event    EventBatch
}

// EventBatch respresents Context about the event
type EventBatch struct {
	Revision        string    `json:"revision"`
	AccountID       string    `json:"account_id"`
	ClientVersion   string    `json:"client_version"`
	Visitors        []Visitor `json:"visitors"`
	ProjectID       string    `json:"project_id"`
	ClientName      string    `json:"client_name"`
	AnonymizeIP     bool      `json:"anonymize_ip"`
	EnrichDecisions bool      `json:"enrich_decisions"`
}

// Visitor respresents a visitor of an eventbatch
type Visitor struct {
	Attributes []VisitorAttribute `json:"attributes"`
	Snapshots  []Snapshot         `json:"snapshots"`
	VisitorID  string             `json:"visitor_id"`
}

// VisitorAttribute respresents an attribute of a visitor
type VisitorAttribute struct {
	Value         interface{} `json:"value"`
	Key           string      `json:"key"`
	AttributeType string      `json:"type"`
	EntityID      string      `json:"entity_id"`
}

// Snapshot respresents a snapshot of a visitor
type Snapshot struct {
	Decisions []Decision      `json:"decisions"`
	Events    []SnapshotEvent `json:"events"`
}

// Decision respresents a decision of a snapshot
type Decision struct {
	VariationID  string `json:"variation_id"`
	CampaignID   string `json:"campaign_id"`
	ExperimentID string `json:"experiment_id"`
}

// SnapshotEvent respresents an event of a snapshot
type SnapshotEvent struct {
	EntityID  string                 `json:"entity_id"`
	Key       string                 `json:"key"`
	Timestamp int64                  `json:"timestamp"`
	UUID      string                 `json:"uuid"`
	Tags      map[string]interface{} `json:"tags,omitempty"`
	Revenue   *int64                 `json:"revenue,omitempty"`
	Value     *float64               `json:"value,omitempty"`
}
