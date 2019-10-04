package models

// GetEnabledFeaturesParams represents params required for isFeatureEnabled API
type GetEnabledFeaturesParams struct {
	UserID     string                 `yaml:"user_id"`
	Attributes map[string]interface{} `yaml:"attributes"`
}
