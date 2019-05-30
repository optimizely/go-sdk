package entities

// Variation represents a variation in the experiment
type Variation struct {
	ID             string
	Key            string
	FeatureEnabled bool
	Variables      map[string]FeatureVariable
}

// Experiment represents an experiment
type Experiment struct {
	ID         string
	Key        string
	Variations []Variation
}
