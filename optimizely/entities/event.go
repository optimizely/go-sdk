package entities

// Event represents a conversion event
type Event struct {
	ID            string   `json:"id"`
	Key           string   `json:"key"`
	ExperimentIds []string `json:"experimentIds"`
}
