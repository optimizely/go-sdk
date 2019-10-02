package models

// RequestParams represents parameters for a scenario
type RequestParams struct {
	APIName      string
	Arguments    string
	DatafileName string
	Listeners    map[string]int
}
