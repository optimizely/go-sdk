package config

import (
	"github.com/optimizely/go-sdk/optimizely/entities"
)

// ProjectConfig contains the parsed project entities
type ProjectConfig interface {
	GetFeatureByKey(string) (entities.Feature, error)
}

// ProjectConfigManager manages the config
type ProjectConfigManager interface {
	GetConfig() ProjectConfig
}
