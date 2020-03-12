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
	"context"
	"errors"
	"fmt"
	"github.com/optimizely/go-sdk/pkg/logging"
	"sync"

	"github.com/optimizely/go-sdk/pkg/config/datafileprojectconfig"
	"github.com/optimizely/go-sdk/pkg/notification"
	"github.com/optimizely/go-sdk/pkg/utils"
)

// StaticProjectConfigManager maintains a static copy of the project config
type StaticProjectConfigManager struct {
	projectConfig    ProjectConfig
	optimizelyConfig *OptimizelyConfig
	configLock       sync.Mutex
}

// NewStaticProjectConfigManagerFromURL returns new instance of StaticProjectConfigManager for URL
func NewStaticProjectConfigManagerFromURL(ctx context.Context) (*StaticProjectConfigManager, error) {

	requester := utils.NewHTTPRequester(ctx)

	url := fmt.Sprintf(DatafileURLTemplate, logging.GetSdkKey(ctx))
	datafile, _, code, e := requester.Get(url)
	if e != nil {
		logging.GetLogger(ctx, "NewStaticProjectConfigManagerFromURL").Error(fmt.Sprintf("request returned with http code=%d", code), e)
		return nil, e
	}

	return NewStaticProjectConfigManagerFromPayload(ctx, datafile)
}

// NewStaticProjectConfigManagerFromPayload returns new instance of StaticProjectConfigManager for payload
func NewStaticProjectConfigManagerFromPayload(ctx context.Context, payload []byte) (*StaticProjectConfigManager, error) {
	projectConfig, err := datafileprojectconfig.NewDatafileProjectConfig(logging.GetLogger(ctx, "DatafileProjectConfig"), payload)

	if err != nil {
		return nil, err
	}

	return NewStaticProjectConfigManager(projectConfig), nil
}

// NewStaticProjectConfigManager creates a new instance of the manager with the given project config
func NewStaticProjectConfigManager(config ProjectConfig) *StaticProjectConfigManager {
	return &StaticProjectConfigManager{
		projectConfig: config,
	}
}

// GetConfig returns the project config
func (cm *StaticProjectConfigManager) GetConfig() (ProjectConfig, error) {
	cm.configLock.Lock()
	defer cm.configLock.Unlock()
	return cm.projectConfig, nil
}

// GetOptimizelyConfig returns the optimizely project config
func (cm *StaticProjectConfigManager) GetOptimizelyConfig() *OptimizelyConfig {
	cm.configLock.Lock()
	defer cm.configLock.Unlock()
	if cm.optimizelyConfig != nil {
		return cm.optimizelyConfig
	}
	optimizelyConfig := NewOptimizelyConfig(cm.projectConfig)
	cm.optimizelyConfig = optimizelyConfig

	return cm.optimizelyConfig
}

// RemoveOnProjectConfigUpdate here satisfies interface
func (cm *StaticProjectConfigManager) RemoveOnProjectConfigUpdate(id int) error {
	return errors.New("method RemoveOnProjectConfigUpdate does not have any effect on StaticProjectConfigManager")
}

// OnProjectConfigUpdate here satisfies interface
func (cm *StaticProjectConfigManager) OnProjectConfigUpdate(callback func(notification.ProjectConfigUpdateNotification)) (int, error) {
	return 0, errors.New("method OnProjectConfigUpdate does not have any effect on StaticProjectConfigManager")
}
