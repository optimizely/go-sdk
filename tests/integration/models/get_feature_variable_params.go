package models

// GetFeatureVariableParams represents params required for GetFeatureVariable API's
type GetFeatureVariableParams struct {
	FeatureKey  string                 `yaml:"feature_flag_key"`
	VariableKey string                 `yaml:"variable_key"`
	UserID      string                 `yaml:"user_id"`
	Attributes  map[string]interface{} `yaml:"attributes"`
}
