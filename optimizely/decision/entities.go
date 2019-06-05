package decision

import "github.com/optimizely/go-sdk/optimizely/entities"

// ExperimentDecisionContext contains the information needed to be able to make a decision for a given experiment
type ExperimentDecisionContext struct {
	Experiment entities.Experiment
	Group      entities.Group
}

// FeatureDecisionContext contains the information needed to be able to make a decision for a given feature
type FeatureDecisionContext struct {
	Feature entities.Feature
	Group   entities.Group
}

// FeatureDecision contains the decision information about a feature
type FeatureDecision struct {
	Type           string
	FeatureEnabled bool
}

// ExperimentDecision contains the decision information about an experiment
type ExperimentDecision struct {
	Variation entities.Variation
}
