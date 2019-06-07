package decision

import (
	"github.com/optimizely/go-sdk/optimizely/entities"
)

// CompositeFeatureService is the default out-of-the-box feature decision service
type CompositeFeatureService struct {
	experimentDecisionService        ExperimentDecisionService
	rolloutExperimentDecisionService ExperimentDecisionService
}

// NewCompositeFeatureService returns a new instance of the CompositeFeatureService
func NewCompositeFeatureService(experimentDecisionService ExperimentDecisionService) *CompositeFeatureService {
	if experimentDecisionService == nil {
		experimentDecisionService = NewExperimentBucketerService()
	}
	return &CompositeFeatureService{
		experimentDecisionService: experimentDecisionService,
	}
}

// GetDecision returns a decision for the given feature and user context
func (featureService *CompositeFeatureService) GetDecision(decisionContext FeatureDecisionContext, userContext entities.UserContext) (FeatureDecision, error) {
	featureEnabled := false
	feature := decisionContext.Feature

	// Check if user is bucketed in feature experiment
	experimentDecisionContext := ExperimentDecisionContext{
		Experiment: feature.FeatureExperiments[0],
	}

	experimentDecision, err := featureService.experimentDecisionService.GetDecision(experimentDecisionContext, userContext)
	if err != nil {
		// @TODO(mng): handle error here
	}
	featureEnabled = experimentDecision.Variation.FeatureEnabled

	return FeatureDecision{
		FeatureEnabled: featureEnabled,
	}, nil
}
