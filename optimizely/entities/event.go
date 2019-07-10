package entities

type Event struct {
	ID string `json:"id"`
	Key string `json:"key"`
	ExperimentIds []string `json:"experimentIds"`
}

