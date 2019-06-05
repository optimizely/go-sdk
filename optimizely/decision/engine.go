package decision

import (
	"github.com/optimizely/go-sdk/optimizely/entities"
)

// Engine interface is used to make a decision for a given feature or experiment
type Engine interface {
	GetFeatureDecision(FeatureDecisionContext, entities.UserContext) (FeatureDecision, error)
}

// DefaultDecisionEngine is used out of the box and can be replaced by the user
type DefaultDecisionEngine struct {
	featureDecisionService FeatureDecisionService
	customFeatureOverrides []FeatureDecisionService
}

// NewDefaultDecisionEngine returns a new instance of the DefeaultDecisionEngine
func NewDefaultDecisionEngine() *DefaultDecisionEngine {
	return &DefaultDecisionEngine{
		featureDecisionService: NewDefaultFeatureDecisionService(),
	}
}

// GetFeatureDecision returns a decision for the given feature key
func (engine DefaultDecisionEngine) GetFeatureDecision(featureDecisionContext FeatureDecisionContext, userContext entities.UserContext) (FeatureDecision, error) {
	// Run through overrides, if any, and return those decisions first
	featureDecision, err := engine.featureDecisionService.GetDecision(featureDecisionContext, userContext)
	return *featureDecision, err
}
