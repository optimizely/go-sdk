package decision

import (
	"github.com/optimizely/go-sdk/optimizely/entities"
)

// DecisionService interface is used to make a decision for a given feature or experiment
type DecisionService interface {
	GetFeatureDecision(FeatureDecisionContext, entities.UserContext) (FeatureDecision, error)
}

// ExperimentDecisionService can make a decision about an experiment
type ExperimentDecisionService interface {
	GetDecision(decisionContext ExperimentDecisionContext, userContext entities.UserContext) (ExperimentDecision, error)
}

// FeatureDecisionService can make a decision about a Feature Flag (can be feature test or rollout)
type FeatureDecisionService interface {
	GetDecision(decisionContext FeatureDecisionContext, userContext entities.UserContext) (FeatureDecision, error)
}
