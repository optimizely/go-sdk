package decision

import "github.com/optimizely/go-sdk/optimizely/entities"

// Engine is used to make a decision for a given feature or experiment
type Engine interface {
	GetFeatureDecision(entities.Feature, entities.UserContext) FeatureDecision
}

// FeatureDecision contains the decision information about a feature
type FeatureDecision struct {
	Type           string
	FeatureEnabled bool
}
