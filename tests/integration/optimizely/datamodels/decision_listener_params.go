package datamodels

import (
	"github.com/optimizely/go-sdk/optimizely/notification"
)

// DecisionListenerModel represents a decision notification
type DecisionListenerModel struct {
	Type         notification.DecisionNotificationType `yaml:"type"`
	UserID       string                                `yaml:"user_id"`
	Attributes   map[string]interface{}                `yaml:"attributes"`
	DecisionInfo map[string]interface{}                `yaml:"decision_info"`
}
