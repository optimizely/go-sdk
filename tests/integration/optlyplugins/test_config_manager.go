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

package optlyplugins

import (
	"time"

	"github.com/optimizely/go-sdk/pkg"
	"github.com/optimizely/go-sdk/pkg/notification"
	"github.com/optimizely/go-sdk/tests/integration/models"
)

// DefaultInitializationTimeout defines default timeout for datafile sync
const DefaultInitializationTimeout = time.Duration(3000) * time.Millisecond

// TestConfigManager represents a ProjectConfigManager with custom implementations
type TestConfigManager struct {
	pkg.ProjectConfigManager
	listenersCalled []notification.ProjectConfigUpdateNotification
}

// CreateListenerCallbacks - Creates Notification Listeners
func (c *TestConfigManager) CreateListenerCallbacks(apiOptions models.APIOptions) (listeners []func(notification notification.ProjectConfigUpdateNotification)) {

	projectConfigUpdateCallback := func(notification notification.ProjectConfigUpdateNotification) {
		c.listenersCalled = append(c.listenersCalled, notification)
	}

	for listenerType, count := range apiOptions.Listeners {
		for i := 1; i <= count; i++ {
			switch listenerType {
			case "Config-update":
				listeners = append(listeners, projectConfigUpdateCallback)
				break
			default:
				break
			}
		}
	}
	return listeners
}

// Verify - Verifies configuration tests
func (c *TestConfigManager) Verify(apiOptions models.APIOptions) {
	timeout := DefaultInitializationTimeout
	if apiOptions.DFMConfiguration.Timeout != nil {
		timeout = time.Duration(*(apiOptions.DFMConfiguration.Timeout)) * time.Millisecond
	}

	start := time.Now()
	switch apiOptions.DFMConfiguration.Mode {
	case "wait_for_on_ready":
		for {
			t := time.Now()
			elapsed := t.Sub(start)
			if elapsed >= timeout {
				break
			}
			// Check if projectconfig is ready
			_, err := c.GetConfig()
			if err == nil {
				break
			}
		}
		break
	case "wait_for_config_update":
		revision := 0
		if apiOptions.DFMConfiguration.Revision != nil {
			revision = *(apiOptions.DFMConfiguration.Revision)
		}
		for {
			t := time.Now()
			elapsed := t.Sub(start)
			if elapsed >= timeout {
				break
			}
			if revision > 0 {
				// This means we want the manager to poll until we get to a specific revision
				if revision == len(c.listenersCalled) {
					break
				}
			} else if len(c.listenersCalled) == 1 {
				// For cases where we are just waiting for config listener
				break
			}
		}
		break
	default:
		break
	}
}

// GetListenersCalled - Returns listeners called
func (c *TestConfigManager) GetListenersCalled() []notification.ProjectConfigUpdateNotification {
	listenerCalled := c.listenersCalled
	// Since for every scenario, a new sdk instance is created, emptying listenersCalled is required for scenario's
	// where multiple requests are executed but no session is to be maintained among them.
	// @TODO: Make it optional once event-batching(sessioned) tests are implemented.
	c.listenersCalled = nil
	return listenerCalled
}
