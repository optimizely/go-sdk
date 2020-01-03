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
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/optimizely/go-sdk/pkg/config/datafileprojectconfig"
	"github.com/optimizely/go-sdk/pkg/logging"
	"github.com/optimizely/go-sdk/pkg/notification"
	"github.com/optimizely/go-sdk/pkg/registry"
	"github.com/optimizely/go-sdk/pkg/utils"

	"github.com/pkg/errors"
)

// DefaultPollingInterval sets default interval for polling manager
const DefaultPollingInterval = 5 * time.Minute // default to 5 minutes for polling

// ModifiedSince header key for request
const ModifiedSince = "If-Modified-Since"

// LastModified header key for response
const LastModified = "Last-Modified"

// DatafileURLTemplate is used to construct the endpoint for retrieving the datafile from the CDN
const DatafileURLTemplate = "https://cdn.optimizely.com/datafiles/%s.json"

var cmLogger = logging.GetLogger("PollingConfigManager")

// PollingProjectConfigManager maintains a dynamic copy of the project config by continuously polling for the datafile
// from the Optimizely CDN at a given (configurable) interval.
type PollingProjectConfigManager struct {
	datafileURLTemplate string
	initDatafile        []byte
	lastModified        string
	notificationCenter  notification.Center
	pollingInterval     time.Duration
	requester           utils.Requester
	sdkKey              string

	configLock       sync.RWMutex
	err              error
	projectConfig    ProjectConfig
	optimizelyConfig *OptimizelyConfig
}

// OptionFunc is used to provide custom configuration to the PollingProjectConfigManager.
type OptionFunc func(*PollingProjectConfigManager)

// WithRequester is an optional function, sets a passed requester
func WithRequester(requester utils.Requester) OptionFunc {
	return func(p *PollingProjectConfigManager) {
		p.requester = requester
	}
}

// WithDatafileURLTemplate is an optional function, sets a passed datafile URL template
func WithDatafileURLTemplate(datafileTemplate string) OptionFunc {
	return func(p *PollingProjectConfigManager) {
		p.datafileURLTemplate = datafileTemplate
	}
}

// WithPollingInterval is an optional function, sets a passed polling interval
func WithPollingInterval(interval time.Duration) OptionFunc {
	return func(p *PollingProjectConfigManager) {
		p.pollingInterval = interval
	}
}

// WithInitialDatafile is an optional function, sets a passed datafile
func WithInitialDatafile(datafile []byte) OptionFunc {
	return func(p *PollingProjectConfigManager) {
		p.initDatafile = datafile
	}
}

// SyncConfig gets current datafile and updates projectConfig
func (cm *PollingProjectConfigManager) SyncConfig(datafile []byte) {
	var e error
	var code int
	var respHeaders http.Header

	closeMutex := func(e error) {
		cm.err = e
		cm.configLock.Unlock()
	}

	url := fmt.Sprintf(cm.datafileURLTemplate, cm.sdkKey)
	if len(datafile) == 0 {
		if cm.lastModified != "" {
			lastModifiedHeader := utils.Header{Name: ModifiedSince, Value: cm.lastModified}
			datafile, respHeaders, code, e = cm.requester.Get(url, lastModifiedHeader)
		} else {
			datafile, respHeaders, code, e = cm.requester.Get(url)
		}

		if e != nil {
			msg := "unable to fetch fresh datafile"
			cmLogger.Warning(msg)
			cm.configLock.Lock()
			closeMutex(errors.New(fmt.Sprintf("%s, reason (http status code): %s", msg, e.Error())))
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
		cmLogger.Warning("failed to create project config")
		closeMutex(errors.New("unable to parse datafile"))
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
	closeMutex(nil)
	_ = cm.setConfig(projectConfig)
}

// Start starts the polling
func (cm *PollingProjectConfigManager) Start(ctx context.Context) {
	cmLogger.Debug("Polling Config Manager Initiated")
	t := time.NewTicker(cm.pollingInterval)
	for {
		select {
		case <-t.C:
			cm.SyncConfig([]byte{})
		case <-ctx.Done():
			cmLogger.Debug("Polling Config Manager Stopped")
			return
		}
	}
}

// NewPollingProjectConfigManager returns an instance of the polling config manager with the customized configuration
func NewPollingProjectConfigManager(sdkKey string, pollingMangerOptions ...OptionFunc) *PollingProjectConfigManager {

	pollingProjectConfigManager := PollingProjectConfigManager{
		pollingInterval:     DefaultPollingInterval,
		requester:           utils.NewHTTPRequester(),
		datafileURLTemplate: DatafileURLTemplate,
		sdkKey:              sdkKey,
	}

	for _, opt := range pollingMangerOptions {
		opt(&pollingProjectConfigManager)
	}

	initDatafile := pollingProjectConfigManager.initDatafile
	pollingProjectConfigManager.SyncConfig(initDatafile) // initial poll
	pollingProjectConfigManager.notificationCenter = registry.GetNotificationCenter(sdkKey)
	return &pollingProjectConfigManager
}

// NewAsyncPollingProjectConfigManager returns an instance of the async polling config manager with the customized configuration
func NewAsyncPollingProjectConfigManager(sdkKey string, pollingMangerOptions ...OptionFunc) *PollingProjectConfigManager {

	pollingProjectConfigManager := PollingProjectConfigManager{
		pollingInterval:     DefaultPollingInterval,
		requester:           utils.NewHTTPRequester(),
		datafileURLTemplate: DatafileURLTemplate,
		sdkKey:              sdkKey,
	}

	for _, opt := range pollingMangerOptions {
		opt(&pollingProjectConfigManager)
	}

	initDatafile := pollingProjectConfigManager.initDatafile
	if len(initDatafile) != 0 {
		if projectConfig, _ := datafileprojectconfig.NewDatafileProjectConfig(initDatafile); projectConfig != nil {
			_ = pollingProjectConfigManager.setConfig(projectConfig)
		}
	}
	pollingProjectConfigManager.notificationCenter = registry.GetNotificationCenter(sdkKey)
	return &pollingProjectConfigManager
}

// GetConfig returns the project config
func (cm *PollingProjectConfigManager) GetConfig() (ProjectConfig, error) {
	cm.configLock.RLock()
	defer cm.configLock.RUnlock()
	if cm.projectConfig == nil {
		return cm.projectConfig, cm.err
	}
	return cm.projectConfig, nil
}

func (cm *PollingProjectConfigManager) setConfig(projectConfig ProjectConfig) error {
	if projectConfig == nil {
		return cm.err
	}

	cm.configLock.Lock()
	cm.projectConfig = projectConfig
	if cm.optimizelyConfig != nil {
		cm.optimizelyConfig = NewOptimizelyConfig(projectConfig)
	}
	cm.err = nil
	cm.configLock.Unlock()

	if cm.notificationCenter != nil {
		projectConfigUpdateNotification := notification.ProjectConfigUpdateNotification{
			Type:     notification.ProjectConfigUpdate,
			Revision: cm.projectConfig.GetRevision(),
		}
		if err := cm.notificationCenter.Send(notification.ProjectConfigUpdate, projectConfigUpdateNotification); err != nil {
			cmLogger.Warning("Problem with sending notification")
		}
	}

	return nil
}

// GetOptimizelyConfig returns the optimizely project config
func (cm *PollingProjectConfigManager) GetOptimizelyConfig() *OptimizelyConfig {
	cm.configLock.RLock()
	defer cm.configLock.RUnlock()
	if cm.optimizelyConfig != nil {
		return cm.optimizelyConfig
	}
	optimizelyConfig := NewOptimizelyConfig(cm.projectConfig)
	cm.optimizelyConfig = optimizelyConfig
	return cm.optimizelyConfig
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
