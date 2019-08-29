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
	"sync"
	"time"

	"github.com/optimizely/go-sdk/optimizely"
	"github.com/optimizely/go-sdk/optimizely/config/datafileprojectconfig"
	"github.com/optimizely/go-sdk/optimizely/logging"
)

const defaultPollingInterval = 5 * time.Minute // default to 5 minutes for polling

// DatafileURLTemplate is used to construct the endpoint for retrieving the datafile from the CDN
const DatafileURLTemplate = "https://cdn.optimizely.com/datafiles/%s.json"

var cmLogger = logging.GetLogger("PollingConfigManager")

// PollingProjectConfigManagerOptions used to create an instance with custom configuration
type PollingProjectConfigManagerOptions struct {
	Datafile        []byte
	PollingInterval time.Duration
	Requester       Requester
}

// PollingProjectConfigManager maintains a dynamic copy of the project config
type PollingProjectConfigManager struct {
	requester       Requester
	pollingInterval time.Duration
	projectConfig   optimizely.ProjectConfig
	configLock      sync.RWMutex

	ctx context.Context // context used for cancellation
}

func (cm *PollingProjectConfigManager) activate(initialPayload []byte, init bool) {

	update := func() {
		var e error
		var code int
		var payload []byte
		if init && len(initialPayload) > 0 {
			payload = initialPayload
		} else {
			payload, code, e = cm.requester.Get()

			if e != nil {
				cmLogger.Error(fmt.Sprintf("request returned with http code=%d", code), e)
			}
		}

		projectConfig, err := datafileprojectconfig.NewDatafileProjectConfig(payload)
		if err != nil {
			cmLogger.Error("failed to create project config", err)
		}

		cm.configLock.Lock()
		cm.projectConfig = projectConfig
		cm.configLock.Unlock()
	}

	if init {
		update()
		return
	}
	t := time.NewTicker(cm.pollingInterval)
	for {
		select {
		case <-t.C:
			update()
		case <-cm.ctx.Done():
			cmLogger.Debug("Polling Config Manager Stopped")
			return
		}
	}
}

// NewPollingProjectConfigManagerWithOptions returns new instance of PollingProjectConfigManager with the given options
func NewPollingProjectConfigManagerWithOptions(ctx context.Context, sdkKey string, options PollingProjectConfigManagerOptions) *PollingProjectConfigManager {

	var requester Requester
	if options.Requester != nil {
		requester = options.Requester
	} else {
		url := fmt.Sprintf(DatafileURLTemplate, sdkKey)
		requester = NewHTTPRequester(url)
	}

	pollingInterval := options.PollingInterval
	if pollingInterval == 0 {
		pollingInterval = defaultPollingInterval
	}

	pollingProjectConfigManager := PollingProjectConfigManager{requester: requester, pollingInterval: pollingInterval, ctx: ctx}

	pollingProjectConfigManager.activate(options.Datafile, true) // initial poll

	cmLogger.Debug("Polling Config Manager Initiated")
	go pollingProjectConfigManager.activate([]byte{}, false)
	return &pollingProjectConfigManager
}

// NewPollingProjectConfigManager returns an instance of the polling config manager with the default configuration
func NewPollingProjectConfigManager(ctx context.Context, sdkKey string) *PollingProjectConfigManager {
	options := PollingProjectConfigManagerOptions{}
	configManager := NewPollingProjectConfigManagerWithOptions(ctx, sdkKey, options)
	return configManager
}

// GetConfig returns the project config
func (cm *PollingProjectConfigManager) GetConfig() optimizely.ProjectConfig {
	cm.configLock.RLock()
	defer cm.configLock.RUnlock()
	return cm.projectConfig
}
