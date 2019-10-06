package models

import (
	"github.com/optimizely/go-sdk/pkg/entities"
)

// APIResponse represents result for a scenario
type APIResponse struct {
	Result         interface{}
	Type           entities.VariableType
	ListenerCalled []DecisionListener
}
