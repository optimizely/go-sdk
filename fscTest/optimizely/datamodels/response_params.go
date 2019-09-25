package datamodels

// ResponseParams represents result for a scenario
type ResponseParams struct {
	Result         interface{}
	ListenerCalled []map[string]interface{}
}
