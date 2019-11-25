/****************************************************************************
 * Copyright 2019, Optimizely, Inc. and contributors                        *
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

// Package pkg //
package pkg

import (
	"github.com/optimizely/go-sdk/pkg/entities"
	"github.com/optimizely/go-sdk/pkg/notification"
)

// ProjectConfig contains the parsed project entities
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

// ProjectConfigManager manages the config
type ProjectConfigManager interface {
	GetConfig() (ProjectConfig, error)
	RemoveOnProjectConfigUpdate(id int) error
	OnProjectConfigUpdate(callback func(notification.ProjectConfigUpdateNotification)) (int, error)
}
