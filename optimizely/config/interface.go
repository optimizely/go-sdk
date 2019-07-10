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

package config

import (
	"github.com/optimizely/go-sdk/optimizely/entities"
)

// ProjectConfig contains the parsed project entities
type ProjectConfig interface {
	GetFeatureByKey(string) (entities.Feature, error)
	GetProjectID() string
	GetRevision()  string
	GetAccountID() string
	GetAnonymizeIP() bool
	GetAttributeID(key string) string // returns "" if there is no id
	GetBotFiltering() bool
	GetEventByKey(string) (entities.Event, error)
}

// ProjectConfigManager manages the config
type ProjectConfigManager interface {
	GetConfig() ProjectConfig
}
