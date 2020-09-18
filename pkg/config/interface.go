/****************************************************************************
 * Copyright 2019-2020, Optimizely, Inc. and contributors                   *
 *                                                                          *
 * Licensed under the Apache License, Version 2.0 (the "License");          *
 * you may not use this file except in compliance with the License.         *
 * You may obtain a copy of the License at                                  *
 *                                                                          *
 *    http://www.apache.org/licenses/LICENSE-2.0                            *
 *                                                                          *
 * Unless required by applicable law or agreed to in writing, software      *
 * distributed under the License is distributed on an "AS IS" BASIS,        *
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. *
 * See the License for the specific language governing permissions and      *
 * limitations under the License.                                           *
 ***************************************************************************/

// Package config //
package config

import (
	"github.com/optimizely/go-sdk/pkg/entities"
	"github.com/optimizely/go-sdk/pkg/notification"
)

// ProjectConfig represents the project's experiments and feature flags and contains methods for accessing the them.
type ProjectConfig interface {
	GetDatafile() string
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
	GetExperimentList() []entities.Experiment
	GetFeatureList() []entities.Feature
	GetGroupByID(string) (entities.Group, error)
	GetProjectID() string
	GetRevision() string
	SendFlagDecisions() bool
}

// ProjectConfigManager maintains an instance of the ProjectConfig
type ProjectConfigManager interface {
	GetConfig() (ProjectConfig, error)
	GetOptimizelyConfig() *OptimizelyConfig
	RemoveOnProjectConfigUpdate(id int) error
	OnProjectConfigUpdate(callback func(notification.ProjectConfigUpdateNotification)) (int, error)
}
