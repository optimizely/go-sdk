package decision

import (
	"github.com/optimizely/go-sdk/optimizely/entities"
)

// DefaultFeatureDecisionService is the default out-of-the-box feature decision service
type DefaultFeatureDecisionService struct {
	experimentDecisionService ExperimentDecisionService
}

// NewDefaultFeatureDecisionService returns a new instance of the DefaultFeatureDecisionService
func NewDefaultFeatureDecisionService() DefaultFeatureDecisionService {
	return DefaultFeatureDecisionService{
		experimentDecisionService: NewDefaultExperimentDecisionService(),
	}
}

// GetDecision returns a decision for the given feature and user context
func (featureService DefaultFeatureDecisionService) GetDecision(decisionContext FeatureDecisionContext, userContext entities.UserContext) (*FeatureDecision, error) {
	featureEnabled := false
	feature := decisionContext.Feature

	// Check if user is bucketed in feature experiment
	experimentDecisionContext := ExperimentDecisionContext{
		Experiment: feature.FeatureExperiments[0],
	}

	experimentDecision, err := featureService.experimentDecisionService.GetDecision(experimentDecisionContext, userContext)
	if err != nil {

	}
	featureEnabled = experimentDecision.Variation.FeatureEnabled

	return &FeatureDecision{
		FeatureEnabled: featureEnabled,
	}, nil
}
