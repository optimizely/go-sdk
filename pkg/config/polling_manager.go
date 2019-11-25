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
	"net/http"
	"sync"
	"time"

	"github.com/optimizely/go-sdk/pkg"
	"github.com/optimizely/go-sdk/pkg/config/datafileprojectconfig"
	"github.com/optimizely/go-sdk/pkg/logging"
	"github.com/optimizely/go-sdk/pkg/notification"
	"github.com/optimizely/go-sdk/pkg/registry"
	"github.com/optimizely/go-sdk/pkg/utils"
)

// DefaultPollingInterval sets default interval for polling manager
const DefaultPollingInterval = 5 * time.Minute // default to 5 minutes for polling

// ModifiedSince header key for request
const ModifiedSince = "If-Modified-Since"

// LastModified header key for response
const LastModified = "Last-Modified"

var cmLogger = logging.GetLogger("PollingConfigManager")

// PollingProjectConfigManager maintains a dynamic copy of the project config
type PollingProjectConfigManager struct {
	requester          utils.Requester
	pollingInterval    time.Duration
	notificationCenter notification.Center
	initDatafile       []byte
	lastModified       string

	configLock    sync.RWMutex
	err           error
	projectConfig pkg.ProjectConfig
}

// OptionFunc is a type to a proper func
type OptionFunc func(*PollingProjectConfigManager)

// DefaultRequester is an optional function, sets default requester
func DefaultRequester() OptionFunc {
	return func(p *PollingProjectConfigManager) {

		requester := utils.NewHTTPRequester()
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
func (cm *PollingProjectConfigManager) SyncConfig(sdkKey string, datafile []byte) {
	var e error
	var code int
	var respHeaders http.Header

	closeMutex := func(e error) {
		cm.err = e
		cm.configLock.Unlock()
	}
	uri := "/" + sdkKey + ".json"
	if len(datafile) == 0 {
		if cm.lastModified != "" {
			lastModifiedHeader := utils.Header{Name: ModifiedSince, Value: cm.lastModified}
			datafile, respHeaders, code, e = cm.requester.Get(uri, lastModifiedHeader)
		} else {
			datafile, respHeaders, code, e = cm.requester.Get(uri)
		}

		if e != nil {
			cmLogger.Error(fmt.Sprintf("request returned with http code=%d", code), e)
			cm.configLock.Lock()
			closeMutex(e)
			return
		}

		if code == http.StatusNotModified {
			cmLogger.Debug("The datafile was not modified and won't be downloaded again")
			return
		}

		// Save last-modified date from response header
		lastModified := respHeaders.Get(LastModified)
		if lastModified != "" {
			cm.configLock.Lock()
			cm.lastModified = lastModified
			cm.configLock.Unlock()
		}
	}

	projectConfig, err := datafileprojectconfig.NewDatafileProjectConfig(datafile)

	cm.configLock.Lock()

	if err != nil {
		cmLogger.Error("failed to create project config", err)
		closeMutex(err)
		return
	}

	var previousRevision string
	if cm.projectConfig != nil {
		previousRevision = cm.projectConfig.GetRevision()
	}
	if projectConfig.GetRevision() == previousRevision {
		cmLogger.Debug(fmt.Sprintf("No datafile updates. Current revision number: %s", cm.projectConfig.GetRevision()))
		closeMutex(nil)
		return
	}
	cmLogger.Debug(fmt.Sprintf("New datafile set with revision: %s. Old revision: %s", projectConfig.GetRevision(), previousRevision))
	cm.projectConfig = projectConfig
	closeMutex(nil)

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

// Start starts the polling
func (cm *PollingProjectConfigManager) Start(sdkKey string, exeCtx utils.ExecutionCtx) {
	go func() {
		cmLogger.Debug("Polling Config Manager Initiated")
		t := time.NewTicker(cm.pollingInterval)
		for {
			select {
			case <-t.C:
				cm.SyncConfig(sdkKey, []byte{})
			case <-exeCtx.GetContext().Done():
				cmLogger.Debug("Polling Config Manager Stopped")
				return
			}
		}
	}()
}

// NewPollingProjectConfigManager returns an instance of the polling config manager with the customized configuration
func NewPollingProjectConfigManager(sdkKey string, pollingMangerOptions ...OptionFunc) *PollingProjectConfigManager {

	pollingProjectConfigManager := PollingProjectConfigManager{
		notificationCenter: registry.GetNotificationCenter(sdkKey),
		pollingInterval:    DefaultPollingInterval,
		requester:          utils.NewHTTPRequester(),
	}

	for _, opt := range pollingMangerOptions {
		opt(&pollingProjectConfigManager)
	}

	initDatafile := pollingProjectConfigManager.initDatafile
	pollingProjectConfigManager.SyncConfig(sdkKey, initDatafile) // initial poll
	return &pollingProjectConfigManager
}

// GetConfig returns the project config
func (cm *PollingProjectConfigManager) GetConfig() (pkg.ProjectConfig, error) {
	cm.configLock.RLock()
	defer cm.configLock.RUnlock()
	if cm.projectConfig == nil {
		return cm.projectConfig, cm.err
	}
	return cm.projectConfig, nil
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
