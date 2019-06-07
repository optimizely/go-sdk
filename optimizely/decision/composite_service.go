package decision

import (
	"github.com/optimizely/go-sdk/optimizely/entities"
)

// CompositeService is the entrypoint into the decision service. It provides out of the box decision making for Features and Experiments.
type CompositeService struct {
	experimentDecisionServices []ExperimentDecisionService
	featureDecisionServices    []FeatureDecisionService
}

// NewCompositeService returns a new instance of the DefeaultDecisionEngine
func NewCompositeService() *CompositeService {
	experimentDecisionService := NewExperimentBucketerService()
	featureDecisionService := NewCompositeFeatureService(experimentDecisionService)
	return &CompositeService{
		experimentDecisionServices: []ExperimentDecisionService{experimentDecisionService},
		featureDecisionServices:    []FeatureDecisionService{featureDecisionService},
	}
}

// GetFeatureDecision returns a decision for the given feature key
func (service CompositeService) GetFeatureDecision(featureDecisionContext FeatureDecisionContext, userContext entities.UserContext) (FeatureDecision, error) {
	var featureDecision FeatureDecision

	// loop through the different features decision services until we get a decision
	for _, decisionService := range service.featureDecisionServices {
		featureDecision, err := decisionService.GetDecision(featureDecisionContext, userContext)
		if err != nil {
			// @TODO: log error
		}

		if featureDecision.DecisionMade {
			return featureDecision, err
		}
	}

	// @TODO: add errors
	return featureDecision, nil
}
