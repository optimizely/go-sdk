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
	"errors"
	"sync"

	"github.com/optimizely/go-sdk/pkg/logging"
	"github.com/optimizely/go-sdk/pkg/notification"
)

// StaticProjectConfigManager maintains a static copy of the project config
type StaticProjectConfigManager struct {
	projectConfig    ProjectConfig
	optimizelyConfig *OptimizelyConfig
	configLock       sync.Mutex
}

// NewStaticProjectConfigManager creates a new instance of the manager with the given sdk key and some options
func NewStaticProjectConfigManager(sdkKey string, configMangerOptions ...OptionFunc) *StaticProjectConfigManager {

	logger := logging.GetLogger(sdkKey, "StaticProjectConfigManager")
	staticProjectConfigManager := newConfigManager(sdkKey, logger, configMangerOptions...)
	if sdkKey != "" {
		staticProjectConfigManager.SyncConfig()
	} else if len(staticProjectConfigManager.initDatafile) > 0 {
		staticProjectConfigManager.setInitialDatafile(staticProjectConfigManager.initDatafile)
	}
	projectConfig, err := staticProjectConfigManager.GetConfig()
	if err != nil {
		logger.Error("unable to get project config, error returned:", err)
		return nil
	}

	return &StaticProjectConfigManager{
		projectConfig: projectConfig,
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
