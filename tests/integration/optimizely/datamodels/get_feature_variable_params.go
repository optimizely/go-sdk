package datamodels

// GetFeatureVariableRequestParams represents params required for isFeatureEnabled API
type GetFeatureVariableRequestParams struct {
	FeatureKey  string                 `yaml:"feature_flag_key"`
	VariableKey string                 `yaml:"variable_key"`
	UserID      string                 `yaml:"user_id"`
	Attributes  map[string]interface{} `yaml:"attributes"`
}
