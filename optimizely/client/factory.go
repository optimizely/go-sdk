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

	"github.com/optimizely/go-sdk/optimizely"
	"fmt"
	"github.com/optimizely/go-sdk/optimizely/config"
	"github.com/optimizely/go-sdk/optimizely/decision"
)

// OptimizelyFactory is used to construct an instance of the OptimizelyClient
type OptimizelyFactory struct {
	SDKKey   string
	Datafile []byte
}

// Client returns a client initialized with the defaults
func (f OptimizelyFactory) Client() (*OptimizelyClient, error) {
	var configManager optimizely.ProjectConfigManager

	if f.SDKKey != "" {
		url := fmt.Sprintf("https://cdn.optimizely.com/datafiles/%s.json", f.SDKKey)
		staticConfigManager, err := config.NewStaticProjectConfigManagerFromUrl(url)

		if err != nil {
			return nil, err
		}

		configManager = staticConfigManager

	} else if f.Datafile != nil {
		staticConfigManager, err := config.NewStaticProjectConfigManagerFromPayload(f.Datafile)

		if err != nil {
			return nil, err
		}

		configManager = staticConfigManager
	}

	decisionService := decision.NewCompositeService()
	client := OptimizelyClient{
		decisionService: decisionService,
		configManager:   configManager,
		isValid:         true,
	}
	return &client, nil
}
