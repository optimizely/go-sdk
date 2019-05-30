package decision

import (
	"github.com/optimizely/go-sdk/optimizely/entities"
)

// Engine interface is used to make a decision for a given feature or experiment
type Engine interface {
	GetFeatureDecision(entities.Feature, entities.UserContext) FeatureDecision
}

// DefaultDecisionEngine is used out of the box and can be replaced by the user
type DefaultDecisionEngine struct {
}

// GetFeatureDecision returns a decision for the given feature key
func (*DefaultDecisionEngine) GetFeatureDecision(feature entities.Feature, userContext entities.UserContext) FeatureDecision {
	enabled := feature.Key != "" && userContext.ID != ""
	return FeatureDecision{
		FeatureEnabled: enabled,
	}
}
