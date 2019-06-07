package decision

import (
	"github.com/optimizely/go-sdk/optimizely/entities"
)

// ExperimentBucketerService is the default out-of-the-box experiment decision service
type ExperimentBucketerService struct {
	overrides []ExperimentDecisionService
}

// NewExperimentBucketerService returns a new instance of the ExperimentBucketerService
func NewExperimentBucketerService() *ExperimentBucketerService {
	// @TODO(mng): add experiment override service
	return &ExperimentBucketerService{
		overrides: []ExperimentDecisionService{},
	}
}

// GetDecision returns a decision for the given experiment and user context
func (service *ExperimentBucketerService) GetDecision(decisionContext ExperimentDecisionContext, userContext entities.UserContext) (ExperimentDecision, error) {
	// @TODO(mng): use audience evaluator + bucketer to determine the variation to return
	return ExperimentDecision{
		Variation: &entities.Variation{FeatureEnabled: true},
	}, nil
}
