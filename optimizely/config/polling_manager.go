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
	"fmt"
	"sync"
	"time"

	"github.com/optimizely/go-sdk/optimizely"
	"github.com/optimizely/go-sdk/optimizely/config/datafileprojectconfig"
	"github.com/optimizely/go-sdk/optimizely/logging"
	"github.com/optimizely/go-sdk/optimizely/utils"
)

const defaultPollingInterval = 5 * time.Minute // default to 5 minutes for polling

// DatafileURLTemplate is used to construct the endpoint for retrieving the datafile from the CDN
const DatafileURLTemplate = "https://cdn.optimizely.com/datafiles/%s.json"

var cmLogger = logging.GetLogger("PollingConfigManager")

// PollingProjectConfigManagerOptions used to create an instance with custom configuration
type PollingProjectConfigManagerOptions struct {
	Datafile        []byte
	PollingInterval time.Duration
	Requester       utils.Requester
}

// PollingProjectConfigManager maintains a dynamic copy of the project config
type PollingProjectConfigManager struct {
	requester       utils.Requester
	pollingInterval time.Duration
	projectConfig   optimizely.ProjectConfig
	configLock      sync.RWMutex
	err             error

	exeCtx utils.ExecutionCtx // context used for execution control
}

// SyncConfig gets current datafile and updates projectConfig
func (cm *PollingProjectConfigManager) SyncConfig(datafile []byte) {
	var e error
	var code int
	if len(datafile) == 0 {
		datafile, code, e = cm.requester.Get()

		if e != nil {
			cmLogger.Error(fmt.Sprintf("request returned with http code=%d", code), e)
		}
	}

	projectConfig, err := datafileprojectconfig.NewDatafileProjectConfig(datafile)
	if err != nil {
		cmLogger.Error("failed to create project config", err)
	}

	// TODO: Compare revision numbers here and set projectConfig only if the revision number has changed
	cm.configLock.Lock()
	cm.projectConfig = projectConfig
	cm.err = err
	cm.configLock.Unlock()
}

func (cm *PollingProjectConfigManager) start(initialDatafile []byte, init bool) {

	if init {
		cm.SyncConfig(initialDatafile)
		return
	}

	t := time.NewTicker(cm.pollingInterval)
	for {
		select {
		case <-t.C:
			cm.SyncConfig([]byte{})
		case <-cm.exeCtx.GetContext().Done():
			cmLogger.Debug("Polling Config Manager Stopped")
			return
		}
	}
}

// NewPollingProjectConfigManagerWithOptions returns new instance of PollingProjectConfigManager with the given options
func NewPollingProjectConfigManagerWithOptions(exeCtx utils.ExecutionCtx, sdkKey string, options PollingProjectConfigManagerOptions) *PollingProjectConfigManager {

	var requester utils.Requester
	if options.Requester != nil {
		requester = options.Requester
	} else {
		url := fmt.Sprintf(DatafileURLTemplate, sdkKey)
		requester = utils.NewHTTPRequester(url)
	}

	pollingInterval := options.PollingInterval
	if pollingInterval == 0 {
		pollingInterval = defaultPollingInterval
	}

	pollingProjectConfigManager := PollingProjectConfigManager{requester: requester, pollingInterval: pollingInterval, exeCtx: exeCtx}

	pollingProjectConfigManager.SyncConfig(options.Datafile) // initial poll

	cmLogger.Debug("Polling Config Manager Initiated")
	go pollingProjectConfigManager.start([]byte{}, false)
	return &pollingProjectConfigManager
}

// NewPollingProjectConfigManager returns an instance of the polling config manager with the default configuration
func NewPollingProjectConfigManager(exeCtx utils.ExecutionCtx, sdkKey string) *PollingProjectConfigManager {
	options := PollingProjectConfigManagerOptions{}
	configManager := NewPollingProjectConfigManagerWithOptions(exeCtx, sdkKey, options)
	return configManager
}

// GetConfig returns the project config
func (cm *PollingProjectConfigManager) GetConfig() (optimizely.ProjectConfig, error) {
	cm.configLock.RLock()
	defer cm.configLock.RUnlock()
	return cm.projectConfig, cm.err
}
