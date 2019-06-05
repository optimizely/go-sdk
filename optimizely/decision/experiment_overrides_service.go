package decision

import (
	"github.com/optimizely/go-sdk/optimizely/entities"
)

// ExperimentOverridesDecisionService allows for intercepting and overriding experiment decisions
// based on experiment whitelisting and/or forced bucketing
type ExperimentOverridesDecisionService struct {
	forcedVariations map[string]map[string]string
}

// NewExperimentOverridesDecisionService returns a new instance of the ExperimentOverridesDecisionService
func NewExperimentOverridesDecisionService() *ExperimentOverridesDecisionService {
	return &ExperimentOverridesDecisionService{
		forcedVariations: make(map[string]map[string]string),
	}
}

// GetDecision returns a decision for the given experiment and user context
func (service ExperimentOverridesDecisionService) GetDecision(decisionContext ExperimentDecisionContext, userContext entities.UserContext) (*ExperimentDecision, error) {
	experiment := decisionContext.Experiment
	userID := userContext.ID
	experimentKey := experiment.Key

	// Check for forced bucketing
	if variationKey, ok := service.forcedVariations[userID][experimentKey]; ok {
		if variation, ok := experiment.Variations[variationKey]; ok {
			return &ExperimentDecision{
				Variation: variation,
			}, nil
		}
	}

	return nil, nil
}

// ForceUserIntoVariation will map the given user's ID to the given experiment's variation key
func (service *ExperimentOverridesDecisionService) ForceUserIntoVariation(userContext entities.UserContext, experimentKey string, variationKey string) bool {
	userID := userContext.ID
	var userBucketingMap map[string]string
	if userBucketingMap, ok := service.forcedVariations[userID]; !ok {
		userBucketingMap = map[string]string{}
		userBucketingMap[experimentKey] = variationKey
		service.forcedVariations[userID] = userBucketingMap
		return true
	}

	if variationKey, ok := userBucketingMap[experimentKey]; !ok {
		service.forcedVariations[userID][experimentKey] = variationKey
		return true
	}

	return false
}
