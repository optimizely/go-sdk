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
	"github.com/optimizely/go-sdk/optimizely/notification"
	"github.com/optimizely/go-sdk/optimizely/utils"
)

const defaultPollingInterval = 5 * time.Minute // default to 5 minutes for polling

// DatafileURLTemplate is used to construct the endpoint for retrieving the datafile from the CDN
const DatafileURLTemplate = "https://cdn.optimizely.com/datafiles/%s.json"

var cmLogger = logging.GetLogger("PollingConfigManager")

// PollingProjectConfigManagerOptions used to create an instance with custom configuration
type PollingProjectConfigManagerOptions struct {
	Datafile           []byte
	PollingInterval    time.Duration
	Requester          utils.Requester
	NotificationCenter notification.Center
}

// PollingProjectConfigManager maintains a dynamic copy of the project config
type PollingProjectConfigManager struct {
	requester          utils.Requester
	pollingInterval    time.Duration
	projectConfig      optimizely.ProjectConfig
	configLock         sync.RWMutex
	err                error
	notificationCenter notification.Center

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

	cm.configLock.Lock()
	defer func() {
		cm.err = err
		cm.configLock.Unlock()
	}()
	if err != nil {
		cmLogger.Error("failed to create project config", err)
		return
	}

	var previousRevision string
	if cm.projectConfig != nil {
		previousRevision = cm.projectConfig.GetRevision()
	}
	if projectConfig.GetRevision() == previousRevision {
		cmLogger.Debug(fmt.Sprintf("No datafile updates. Current revision number: %s", cm.projectConfig.GetRevision()))
		return
	}
	cmLogger.Debug(fmt.Sprintf("New datafile set with revision: %s. Old revision: %s", projectConfig.GetRevision(), previousRevision))
	cm.projectConfig = projectConfig

	if cm.notificationCenter != nil {
		projectConfigUpdateNotification := notification.ProjectConfigUpdateNotification{
			Type:     notification.ProjectConfigUpdate,
			Revision: cm.projectConfig.GetRevision(),
		}
		if err = cm.notificationCenter.Send(notification.ProjectConfigUpdate, projectConfigUpdateNotification); err != nil {
			cmLogger.Warning("Problem with sending notification")
		}
	}
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

	pollingProjectConfigManager := PollingProjectConfigManager{requester: requester, pollingInterval: pollingInterval, notificationCenter: options.NotificationCenter, exeCtx: exeCtx}

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

// OnProjectConfigUpdate registers a handler for ProjectConfigUpdate notifications
func (cm *PollingProjectConfigManager) OnProjectConfigUpdate(callback func(notification.ProjectConfigUpdateNotification)) (int, error) {
	handler := func(payload interface{}) {
		if projectConfigUpdateNotification, ok := payload.(notification.ProjectConfigUpdateNotification); ok {
			callback(projectConfigUpdateNotification)
		} else {
			cmLogger.Warning(fmt.Sprintf("Unable to convert notification payload %v into ProjectConfigUpdateNotification", payload))
		}
	}
	id, err := cm.notificationCenter.AddHandler(notification.ProjectConfigUpdate, handler)
	if err != nil {
		cmLogger.Warning("Problem with adding notification handler")
		return 0, err
	}
	return id, nil
}

// RemoveOnProjectConfigUpdate removes handler for ProjectConfigUpdate notification with given id
func (cm *PollingProjectConfigManager) RemoveOnProjectConfigUpdate(id int) error {
	if err := cm.notificationCenter.RemoveHandler(id, notification.ProjectConfigUpdate); err != nil {
		cmLogger.Warning("Problem with removing notification handler")
		return err
	}
	return nil
}
