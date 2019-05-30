package event

import (
	"time"

	"github.com/optimizely/go-sdk/optimizely/entities"
)

// Context about the event
type Context struct {
	accountID     string
	projectID     string
	clientName    string
	clientVersion string
	revision      string
	anonimizeIP   bool
	botFiltering  bool
}

// Event is the base event type
type Event struct {
	UUID      string
	Timestamp time.Time
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
