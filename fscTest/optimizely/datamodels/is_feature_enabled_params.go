package datamodels

// IsFeatureEnabledRequestParams represents params required for isFeatureEnabled API
type IsFeatureEnabledRequestParams struct {
	FeatureKey string                 `yaml:"feature_flag_key"`
	UserID     string                 `yaml:"user_id"`
	Attributes map[string]interface{} `yaml:"attributes"`
}
