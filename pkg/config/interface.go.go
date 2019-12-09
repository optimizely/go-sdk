package config

import (
	"github.com/optimizely/go-sdk/pkg/entities"
	"github.com/optimizely/go-sdk/pkg/notification"
)

// ProjectConfig represents the project's experiments and feature flags and contains methods for accessing the them.
type ProjectConfig interface {
	GetAccountID() string
	GetAnonymizeIP() bool
	GetAttributeID(id string) string // returns "" if there is no id
	GetAttributeByKey(key string) (entities.Attribute, error)
	GetAudienceByID(string) (entities.Audience, error)
	GetAudienceMap() map[string]entities.Audience
	GetBotFiltering() bool
	GetEventByKey(string) (entities.Event, error)
	GetExperimentByKey(string) (entities.Experiment, error)
	GetFeatureByKey(string) (entities.Feature, error)
	GetVariableByKey(featureKey string, variableKey string) (entities.Variable, error)
	GetFeatureList() []entities.Feature
	GetGroupByID(string) (entities.Group, error)
	GetProjectID() string
	GetRevision() string
}

// ProjectConfigManager maintains an instance of the ProjectConfig
type ProjectConfigManager interface {
	GetConfig() (ProjectConfig, error)
	RemoveOnProjectConfigUpdate(id int) error
	OnProjectConfigUpdate(callback func(notification.ProjectConfigUpdateNotification)) (int, error)
}
