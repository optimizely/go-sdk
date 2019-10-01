package datamodels

import "github.com/optimizely/go-sdk/optimizely/entities"

// ResponseParams represents result for a scenario
type ResponseParams struct {
	Result         interface{}
	Type           entities.VariableType
	ListenerCalled []DecisionListenerModel
}
