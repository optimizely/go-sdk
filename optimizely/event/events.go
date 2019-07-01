package event

// Context about the event
type LogEvent struct {
	Revision        string  `json:"revision"`
	AccountID       string  `json:"account_id"`
	ClientVersion   string  `json:"client_version"`
	Visitors        []Visitor `json:"visitors"`
	ProjectID       string  `json:"project_id"`
	ClientName      string  `json:"client_name"`
	AnonymizeIP     bool    `json:"anonymize_ip"`
	EnrichDecisions bool    `json:"enrich_decisions"`
}

type Visitor struct {
	Attributes []EventAttribute `json:"attributes"`
	Snapshots  []Snapshot       `json:"snapshots"`
	VisitorID  string           `json:"visitor_id"`
}

type EventAttribute struct {
	Value         interface{} `json:"value"`
	Key           string	`json:"key"`
	AttributeType string	`json:"type"`
	EntityID      string	`json:"entity_id"`
}

type Snapshot struct {
	Decisions []Decision `json:"decisions"`
 	Events []DispatchEvent `json:"events"`
}

type Decision struct {
	VariationID  string `json:"variation_id"`
	CampaignID   string `json:"campaign_id"`
	ExperimentID string `json:"experiment_id"`
}

type DispatchEvent struct {
	EventID  string `json:"event_id"`
	Key       string `json:"key"`
	Timestamp int64 `json:"timestamp"`
	Uuid      string `json:"uuid"`
	Tags      map[string]interface{} `json:"tags"`
//	Revenue   int `json:"revenue"`
//	Value     float32 `json:"value"`
}