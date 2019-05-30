package entities

// Feature represents a feature flag
type Feature struct {
	ID                 string
	Key                string
	FeatureExperiments map[string]Experiment
	RolloutExperiments []Experiment
	Variables          map[string]FeatureVariable
}

// FeatureVariable represents a variable
type FeatureVariable struct {
	Key   string
	Type  string
	Value string
}
