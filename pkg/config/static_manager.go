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

// Package config //
package config

import (
	"errors"
	"fmt"
	"sync"

	"github.com/optimizely/go-sdk/pkg"
	"github.com/optimizely/go-sdk/pkg/config/datafileprojectconfig"
	"github.com/optimizely/go-sdk/pkg/notification"
	"github.com/optimizely/go-sdk/pkg/utils"
)

// StaticProjectConfigManager maintains a static copy of the project config
type StaticProjectConfigManager struct {
	projectConfig pkg.ProjectConfig
	configLock    sync.Mutex
}

// NewStaticProjectConfigManagerFromURL returns new instance of StaticProjectConfigManager for URL
func NewStaticProjectConfigManagerFromURL(sdkKey string) (*StaticProjectConfigManager, error) {

	requester := utils.NewHTTPRequester(fmt.Sprintf(DatafileURLTemplate, sdkKey))
	datafile, code, e := requester.Get()
	if e != nil {
		cmLogger.Error(fmt.Sprintf("request returned with http code=%d", code), e)
	}

	return NewStaticProjectConfigManagerFromPayload(datafile)
}

// NewStaticProjectConfigManagerFromPayload returns new instance of StaticProjectConfigManager for payload
func NewStaticProjectConfigManagerFromPayload(payload []byte) (*StaticProjectConfigManager, error) {
	projectConfig, err := datafileprojectconfig.NewDatafileProjectConfig(payload)

	if err != nil {
		return nil, err
	}

	return NewStaticProjectConfigManager(projectConfig), nil
}

// NewStaticProjectConfigManager creates a new instance of the manager with the given project config
func NewStaticProjectConfigManager(config pkg.ProjectConfig) *StaticProjectConfigManager {
	return &StaticProjectConfigManager{
		projectConfig: config,
	}
}

// GetConfig returns the project config
func (cm *StaticProjectConfigManager) GetConfig() (pkg.ProjectConfig, error) {
	cm.configLock.Lock()
	defer cm.configLock.Unlock()
	return cm.projectConfig, nil
}

// RemoveOnProjectConfigUpdate here satisfies interface
func (cm *StaticProjectConfigManager) RemoveOnProjectConfigUpdate(id int) error {
	return errors.New("method RemoveOnProjectConfigUpdate does not have any effect on StaticProjectConfigManager")
}

// OnProjectConfigUpdate here satisfies interface
func (cm *StaticProjectConfigManager) OnProjectConfigUpdate(callback func(notification.ProjectConfigUpdateNotification)) (int, error) {
	return 0, errors.New("method OnProjectConfigUpdate does not have any effect on StaticProjectConfigManager")
}
