package models

import (
	"github.com/optimizely/go-sdk/optimizely/notification"
)

// DecisionListener represents a decision notification
type DecisionListener struct {
	Type         notification.DecisionNotificationType `yaml:"type"`
	UserID       string                                `yaml:"user_id"`
	Attributes   map[string]interface{}                `yaml:"attributes"`
	DecisionInfo map[string]interface{}                `yaml:"decision_info"`
}
