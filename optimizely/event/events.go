package event

import (
	"time"

	"github.com/optimizely/go-sdk/optimizely/entities"
)

// Context about the event
type Context struct {
	AccountID     string `json:"account_id"`
	ProjectID     string `json:"project_id"`
	ClientName    string `json:"client_name"`
	ClientVersion string `json:"client_version"`
	Revision      string `json:"revision"`
	AnonymizeIP   bool   `json:"anonymize_ip"`
	BotFiltering  bool   `json:"bot_filtering"`
}

// Event is the base event type
type Event struct {
	UUID      string 	`json:"uuid"`
	Timestamp time.Time `json:"time_stamp"`
	Context   Context
}

// Impression event
type Impression struct {
	Event
	User         entities.UserContext
	LayerID      string
	ExperimentID string
	VariationID  string
}
