package models

// GetEnabledFeaturesParams represents params required for GetEnabledFeatures API
type GetEnabledFeaturesParams struct {
	UserID     string                 `yaml:"user_id"`
	Attributes map[string]interface{} `yaml:"attributes"`
}
