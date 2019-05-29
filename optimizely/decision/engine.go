package decision

// Engine is used to make a decision for a given feature or experiment
type Engine interface {
	GetFeatureDecision(string, string) FeatureDecision
}

// FeatureDecision contains the decision information about a feature
type FeatureDecision struct {
	Type           string
	FeatureEnabled bool
}

// DefaultDecisionEngine is used out of the box and can be replaced by the user
type DefaultDecisionEngine struct {
}

// GetFeatureDecision returns a decision for the given feature key
func (*DefaultDecisionEngine) GetFeatureDecision(featureKey string, userID string) FeatureDecision {
	enabled := featureKey != "" && userID != ""
	return FeatureDecision{
		FeatureEnabled: enabled,
	}
}
