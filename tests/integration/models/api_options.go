package models

// APIOptions represents parameters for a scenario
type APIOptions struct {
	APIName   string
	Arguments string
	Listeners map[string]int
}
