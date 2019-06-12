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

import "sync"

// StaticProjectConfigManager maintains a static copy of the project config
type StaticProjectConfigManager struct {
	projectConfig ProjectConfig
	configLock    *sync.Mutex
}

// GetConfig returns the project config
func (cm StaticProjectConfigManager) GetConfig() ProjectConfig {
	cm.configLock.Lock()
	defer cm.configLock.Unlock()
	return cm.projectConfig
}

// SetConfig sets the project config
func (cm *StaticProjectConfigManager) SetConfig(config ProjectConfig) {
	cm.configLock.Lock()
	defer cm.configLock.Unlock()
	cm.projectConfig = config
}
