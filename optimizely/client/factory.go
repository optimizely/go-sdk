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

package client

import (
	"github.com/optimizely/go-sdk/optimizely/config"
	"github.com/optimizely/go-sdk/optimizely/config/datafileProjectConfig"
	"github.com/optimizely/go-sdk/optimizely/decision"
)

// OptimizelyFactory is used to construct an instance of the OptimizelyClient
type OptimizelyFactory struct {
	SDKKey   string
	Datafile []byte
}

// Client returns a client initialized with the defaults
func (factory OptimizelyFactory) Client() OptimizelyClient {
	var projectConfig config.ProjectConfig
	var configManager config.ProjectConfigManager
	if factory.Datafile != nil {
		projectConfig = datafileProjectConfig.NewDatafileProjectConfig(factory.Datafile)

		if factory.SDKKey == "" {
			staticConfigManager := config.StaticProjectConfigManager{}
			staticConfigManager.SetConfig(projectConfig)
			configManager = staticConfigManager
		}
	}

	decisionService := decision.NewCompositeService()
	client := OptimizelyClient{
		decisionService: decisionService,
		configManager:   configManager,
	}
	return client
}
