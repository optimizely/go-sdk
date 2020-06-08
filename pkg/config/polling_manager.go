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

// DatafileURLTemplate is used to construct the endpoint for retrieving regular datafile from the CDN
const DatafileURLTemplate = "https://cdn.optimizely.com/datafiles/%s.json"

// AuthDatafileURLTemplate is used to construct the endpoint for retrieving authenticated datafile from the CDN
const AuthDatafileURLTemplate = "https://config.optimizely.com/datafiles/auth/%s.json"

// Err403Forbidden is 403Forbidden specific error
var Err403Forbidden = errors.New("unable to fetch fresh datafile (consider rechecking SDK key), status code: 403 Forbidden")

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
	logger              logging.OptimizelyLogProducer
	datafileAccessToken string

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

// WithDatafileAccessToken is an optional function, sets a passed datafile access token
func WithDatafileAccessToken(datafileAccessToken string) OptionFunc {
	return func(p *PollingProjectConfigManager) {
		p.datafileAccessToken = datafileAccessToken
	}
}

// SyncConfig downloads datafile and updates projectConfig
func (cm *PollingProjectConfigManager) SyncConfig() {
	var e error
	var code int
	var respHeaders http.Header
	var datafile []byte

	closeMutex := func(e error) {
		cm.err = e
		cm.configLock.Unlock()
	}

	url := fmt.Sprintf(cm.datafileURLTemplate, cm.sdkKey)
	if cm.lastModified != "" {
		lastModifiedHeader := utils.Header{Name: ModifiedSince, Value: cm.lastModified}
		datafile, respHeaders, code, e = cm.requester.Get(url, lastModifiedHeader)
	} else {
		datafile, respHeaders, code, e = cm.requester.Get(url)
	}

	if e != nil {
		msg := "unable to fetch fresh datafile"
		cm.logger.Warning(msg)
		cm.configLock.Lock()

		if code == http.StatusForbidden {
			closeMutex(Err403Forbidden)
			return
		}

		closeMutex(errors.New(fmt.Sprintf("%s, reason (http status code): %s", msg, e.Error())))
		return
	}

	if code == http.StatusNotModified {
		cm.logger.Debug("The datafile was not modified and won't be downloaded again")
		return
	}

	// Save last-modified date from response header
	cm.configLock.Lock()
	lastModified := respHeaders.Get(LastModified)
	if lastModified != "" {
		cm.lastModified = lastModified
	}

	projectConfig, err := datafileprojectconfig.NewDatafileProjectConfig(datafile, logging.GetLogger(cm.sdkKey, "NewDatafileProjectConfig"))
	if err != nil {
		cm.logger.Warning("failed to create project config")
		closeMutex(errors.New("unable to parse datafile"))
		return
	}

	var previousRevision string
	if cm.projectConfig != nil {
		previousRevision = cm.projectConfig.GetRevision()
	}
	if projectConfig.GetRevision() == previousRevision {
		cm.logger.Debug(fmt.Sprintf("No datafile updates. Current revision number: %s", cm.projectConfig.GetRevision()))
		closeMutex(nil)
		return
	}
	err = cm.setConfig(projectConfig)
	closeMutex(err)
	if err == nil {
		cm.logger.Debug(fmt.Sprintf("New datafile set with revision: %s. Old revision: %s", projectConfig.GetRevision(), previousRevision))
		cm.sendConfigUpdateNotification()
	}
}

// Start starts the polling
func (cm *PollingProjectConfigManager) Start(ctx context.Context) {
	if cm.pollingInterval <= 0 {
		cm.logger.Info("Polling Config Manager Disabled")
		return
	}
	cm.logger.Debug("Polling Config Manager Initiated")
	t := time.NewTicker(cm.pollingInterval)
	for {
		select {
		case <-t.C:
			cm.SyncConfig()
		case <-ctx.Done():
			cm.logger.Debug("Polling Config Manager Stopped")
			return
		}
	}
}

func (cm *PollingProjectConfigManager) setAuthHeaderIfDatafileAccessTokenPresent() {
	if cm.datafileAccessToken != "" {
		headers := []utils.Header{{Name: "Content-Type", Value: "application/json"}, {Name: "Accept", Value: "application/json"}}
		headers = append(headers, utils.Header{Name: "Authorization", Value: "Bearer " + cm.datafileAccessToken})
		cm.requester = utils.NewHTTPRequester(logging.GetLogger(cm.sdkKey, "HTTPRequester"), utils.Headers(headers...))
	}
}

func newConfigManager(sdkKey string, logger logging.OptimizelyLogProducer, configOptions ...OptionFunc) *PollingProjectConfigManager {
	pollingProjectConfigManager := PollingProjectConfigManager{
		notificationCenter: registry.GetNotificationCenter(sdkKey),
		pollingInterval:    DefaultPollingInterval,
		requester:          utils.NewHTTPRequester(logging.GetLogger(sdkKey, "HTTPRequester")),
		sdkKey:             sdkKey,
		logger:             logger,
	}

	for _, opt := range configOptions {
		opt(&pollingProjectConfigManager)
	}

	if pollingProjectConfigManager.datafileURLTemplate == "" {
		if pollingProjectConfigManager.datafileAccessToken != "" {
			pollingProjectConfigManager.datafileURLTemplate = AuthDatafileURLTemplate
		} else {
			pollingProjectConfigManager.datafileURLTemplate = DatafileURLTemplate
		}
	}
	pollingProjectConfigManager.setAuthHeaderIfDatafileAccessTokenPresent()
	return &pollingProjectConfigManager
}

// NewPollingProjectConfigManager returns an instance of the polling config manager with the customized configuration
func NewPollingProjectConfigManager(sdkKey string, pollingMangerOptions ...OptionFunc) *PollingProjectConfigManager {

	pollingProjectConfigManager := newConfigManager(sdkKey, logging.GetLogger(sdkKey, "PollingProjectConfigManager"), pollingMangerOptions...)

	if len(pollingProjectConfigManager.initDatafile) > 0 {
		pollingProjectConfigManager.setInitialDatafile(pollingProjectConfigManager.initDatafile)
	} else {
		pollingProjectConfigManager.SyncConfig() // initial poll
	}
	return pollingProjectConfigManager
}

// NewAsyncPollingProjectConfigManager returns an instance of the async polling config manager with the customized configuration
func NewAsyncPollingProjectConfigManager(sdkKey string, pollingMangerOptions ...OptionFunc) *PollingProjectConfigManager {

	pollingProjectConfigManager := newConfigManager(sdkKey, logging.GetLogger(sdkKey, "PollingProjectConfigManager"), pollingMangerOptions...)
	if len(pollingProjectConfigManager.initDatafile) > 0 {
		pollingProjectConfigManager.setInitialDatafile(pollingProjectConfigManager.initDatafile)
	}
	return pollingProjectConfigManager
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
			cm.logger.Warning(fmt.Sprintf("Unable to convert notification payload %v into ProjectConfigUpdateNotification", payload))
		}
	}
	id, err := cm.notificationCenter.AddHandler(notification.ProjectConfigUpdate, handler)
	if err != nil {
		cm.logger.Warning("Problem with adding notification handler")
		return 0, err
	}
	return id, nil
}

// RemoveOnProjectConfigUpdate removes handler for ProjectConfigUpdate notification with given id
func (cm *PollingProjectConfigManager) RemoveOnProjectConfigUpdate(id int) error {
	if err := cm.notificationCenter.RemoveHandler(id, notification.ProjectConfigUpdate); err != nil {
		cm.logger.Warning("Problem with removing notification handler")
		return err
	}
	return nil
}

func (cm *PollingProjectConfigManager) setConfig(projectConfig ProjectConfig) error {
	if projectConfig == nil {
		return errors.New("unable to set nil config")
	}
	cm.projectConfig = projectConfig
	if cm.optimizelyConfig != nil {
		cm.optimizelyConfig = NewOptimizelyConfig(projectConfig)
	}
	return nil
}

func (cm *PollingProjectConfigManager) setInitialDatafile(datafile []byte) {
	if len(datafile) != 0 {
		cm.configLock.Lock()
		defer cm.configLock.Unlock()
		projectConfig, err := datafileprojectconfig.NewDatafileProjectConfig(datafile, logging.GetLogger(cm.sdkKey, "DatafileProjectConfig"))
		if projectConfig != nil {
			err = cm.setConfig(projectConfig)
		}
		cm.err = err
	}
}

func (cm *PollingProjectConfigManager) sendConfigUpdateNotification() {
	if cm.notificationCenter != nil {
		projectConfigUpdateNotification := notification.ProjectConfigUpdateNotification{
			Type:     notification.ProjectConfigUpdate,
			Revision: cm.projectConfig.GetRevision(),
		}
		if err := cm.notificationCenter.Send(notification.ProjectConfigUpdate, projectConfigUpdateNotification); err != nil {
			cm.logger.Warning("Problem with sending notification")
		}
	}
}
