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
