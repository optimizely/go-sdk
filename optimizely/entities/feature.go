package entities

// Feature represents a feature flag
type Feature struct {
	ID      string
	Key     string
	Rollout Rollout
}
