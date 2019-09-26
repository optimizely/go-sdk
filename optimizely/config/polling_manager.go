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
	"github.com/optimizely/go-sdk/optimizely/registry"
	"github.com/optimizely/go-sdk/optimizely/utils"
)

// DefaultPollingInterval sets default interval for polling manager
const DefaultPollingInterval = 5 * time.Minute // default to 5 minutes for polling

// DatafileURLTemplate is used to construct the endpoint for retrieving the datafile from the CDN
const DatafileURLTemplate = "https://cdn.optimizely.com/datafiles/%s.json"

var cmLogger = logging.GetLogger("PollingConfigManager")

// PollingProjectConfigManager maintains a dynamic copy of the project config
type PollingProjectConfigManager struct {
	requester          utils.Requester
	pollingInterval    time.Duration
	notificationCenter notification.Center
	initDatafile       []byte

	configLock    sync.RWMutex
	err           error
	projectConfig optimizely.ProjectConfig
	exeCtx        utils.ExecutionCtx // context used for execution control
}

// OptionFunc is a type to a proper func
type OptionFunc func(*PollingProjectConfigManager)

// DefaultRequester is an optional function, sets default requester based on a key.
func DefaultRequester(sdkKey string) OptionFunc {
	return func(p *PollingProjectConfigManager) {

		url := fmt.Sprintf(DatafileURLTemplate, sdkKey)
		requester := utils.NewHTTPRequester(url)

		p.requester = requester
	}
}

// Requester is an optional function, sets a passed requester
func Requester(requester utils.Requester) OptionFunc {
	return func(p *PollingProjectConfigManager) {
		p.requester = requester
	}
}

// PollingInterval is an optional function, sets a passed polling interval
func PollingInterval(interval time.Duration) OptionFunc {
	return func(p *PollingProjectConfigManager) {
		p.pollingInterval = interval
	}
}

// InitialDatafile is an optional function, sets a passed datafile
func InitialDatafile(datafile []byte) OptionFunc {
	return func(p *PollingProjectConfigManager) {
		p.initDatafile = datafile
	}
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

	cm.configLock.Lock()
	if cm.projectConfig != nil {
		if cm.projectConfig.GetRevision() == projectConfig.GetRevision() {
			cmLogger.Debug(fmt.Sprintf("No datafile updates. Current revision number: %s", cm.projectConfig.GetRevision()))
		} else {
			cmLogger.Debug(fmt.Sprintf("Received new datafile and updated config. Old revision number: %s. New revision number: %s", cm.projectConfig.GetRevision(), projectConfig.GetRevision()))
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
	} else {
		cm.projectConfig = projectConfig
	}
	cm.err = err
	cm.configLock.Unlock()
}

func (cm *PollingProjectConfigManager) start() {

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

// NewPollingProjectConfigManager returns an instance of the polling config manager with the customized configuration
func NewPollingProjectConfigManager(exeCtx utils.ExecutionCtx, sdkKey string, pollingMangerOptions ...OptionFunc) *PollingProjectConfigManager {
	url := fmt.Sprintf(DatafileURLTemplate, sdkKey)

	pollingProjectConfigManager := PollingProjectConfigManager{
		exeCtx:             exeCtx,
		notificationCenter: registry.GetNotificationCenter(sdkKey),
		pollingInterval:    DefaultPollingInterval,
		requester:          utils.NewHTTPRequester(url),
	}

	for _, opt := range pollingMangerOptions {
		opt(&pollingProjectConfigManager)
	}

	initDatafile := pollingProjectConfigManager.initDatafile
	pollingProjectConfigManager.SyncConfig(initDatafile) // initial poll

	cmLogger.Debug("Polling Config Manager Initiated")
	go pollingProjectConfigManager.start()
	return &pollingProjectConfigManager
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
