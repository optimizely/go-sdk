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

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/optimizely/go-sdk/optimizely"
	"github.com/optimizely/go-sdk/optimizely/config/datafileprojectconfig"
	"github.com/optimizely/go-sdk/optimizely/logging"
)

const defaultPollingWait = time.Duration(5 * time.Minute) // default 5 minutes for polling wait

var cmLogger = logging.GetLogger("PollingConfigManager")

// PollingProjectConfigManager maintains a dynamic copy of the project config
type PollingProjectConfigManager struct {
	requester     Requester
	pollingWait   time.Duration
	projectConfig optimizely.ProjectConfig
	configLock    sync.RWMutex

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
	t := time.NewTicker(cm.pollingWait)
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

// NewPollingProjectConfigManager returns new instance of PollingProjectConfigManager
func NewPollingProjectConfigManager(ctx context.Context, requester Requester, initialPayload []byte, pollingWait time.Duration) *PollingProjectConfigManager {

	if pollingWait == 0 {
		pollingWait = defaultPollingWait
	}

	pollingProjectConfigManager := PollingProjectConfigManager{requester: requester, pollingWait: pollingWait, ctx: ctx}

	pollingProjectConfigManager.activate(initialPayload, true) // initial poll

	cmLogger.Debug("Polling Config Manager Initiated")
	go pollingProjectConfigManager.activate([]byte{}, false)
	return &pollingProjectConfigManager
}

// GetConfig returns the project config
func (cm *PollingProjectConfigManager) GetConfig() optimizely.ProjectConfig {
	cm.configLock.RLock()
	defer cm.configLock.RUnlock()
	return cm.projectConfig
}
