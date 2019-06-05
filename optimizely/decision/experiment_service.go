package decision

import (
	"github.com/optimizely/go-sdk/optimizely/entities"
)

// DefaultExperimentDecisionService is the default out-of-the-box experiment decision service
type DefaultExperimentDecisionService struct {
	overrides []ExperimentDecisionService
}

// NewDefaultExperimentDecisionService returns a new instance of the DefaultExperimentDecisionService
func NewDefaultExperimentDecisionService() *DefaultExperimentDecisionService {
	experimentOverridesDecisionService := NewExperimentOverridesDecisionService()
	return &DefaultExperimentDecisionService{
		overrides: []ExperimentDecisionService{experimentOverridesDecisionService},
	}
}

// GetDecision returns a decision for the given experiment and user context
func (DefaultExperimentDecisionService) GetDecision(decisionContext ExperimentDecisionContext, userContext entities.UserContext) (*ExperimentDecision, error) {
	// TODO: use audience evaluator + bucketer to determine the variation to return
	return &ExperimentDecision{
		Variation: entities.Variation{FeatureEnabled: true},
	}, nil
}
