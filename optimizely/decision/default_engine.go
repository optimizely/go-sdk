package decision

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
