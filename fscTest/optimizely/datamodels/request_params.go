package datamodels

// RequestParams represents parameters for a scenario
type RequestParams struct {
	ApiName      string
	Arguments    string
	DatafileName string
	Listeners    map[string]int
}
