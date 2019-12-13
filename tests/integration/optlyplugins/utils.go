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
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"github.com/optimizely/go-sdk/pkg/config"
	"github.com/optimizely/go-sdk/pkg/notification"
	"github.com/optimizely/go-sdk/pkg/registry"
	"github.com/optimizely/go-sdk/tests/integration/models"
)

const localDatafileURLTemplate = "http://localhost:3001/datafiles/%s.json?request_id=%s"

// SyncConfig doesn't request for new datafile if we provide a valid datafile
// this requires us to keep defaultPollingInterval low so that the request
// initiated from Start method is executed quickly
const defaultPollingInterval = time.Duration(1000) * time.Millisecond

// Since notificationManager is mapped against sdkKey, we need a unique sdkKey for every scenario
var sdkKey int

func RegisterNotification(sdkKey string, notificationType notification.Type, callback func(interface{})) {
	registry.GetNotificationCenter(sdkKey).AddHandler(notificationType, callback)
}

func RegisterConfigUpdateBlockModeHandler(wg *sync.WaitGroup, sdkKey string, datafileOptions models.DataFileManagerConfiguration) {
	defer func() {
		if r := recover(); r != nil {
			// TODO: Need to add proper switch
			// switch t := r.(type) {
			// case error:
			// 	err = t
			// case string:
			// 	err = errors.New(t)
			// default:
			// 	err = errors.New("unexpected error")
			// }
			// errorMessage := fmt.Sprintf("optimizely SDK is panicking with the error:")
			// logger.Error(errorMessage, err)
			// logger.Debug(string(debug.Stack()))
		}
	}()

	revision := 0

	switch datafileOptions.Mode {
	case models.KeyWaitForOnReady:
		revision = 1
		break
	case models.KeyWaitForConfigUpdate:
		if datafileOptions.Revision != nil {
			revision = *datafileOptions.Revision
		}
	}

	wg.Add(revision)

	handler := func(payload interface{}) {
		// need to checkout
		if revision > 0 {
			wg.Done()
		}
		revision = revision - 1
	}

	RegisterNotification(sdkKey, notification.ProjectConfigUpdate, handler)
}

func WaitOrTimeoutWG(wg *sync.WaitGroup, timeout time.Duration) bool {
	c := make(chan struct{})
	go func() {
		defer close(c)
		wg.Wait()
	}()

	select {
	case <-c:
		return true // completed normally
	case <-time.After(timeout):
		// timed out, call done and exit
		return false
	}
}

// CreatePollingConfigManager creates a pollingConfigManager with given configuration
func CreatePollingConfigManager(sdkKey, scenarioID string, apiOptions models.APIOptions) config.ProjectConfigManager {
	var datafile []byte

	pollingInterval := defaultPollingInterval

	if apiOptions.DatafileName != "" {
		datafile, _ = GetDatafile(apiOptions.DatafileName)
	}

	// Setting up polling interval
	if apiOptions.DFMConfiguration.UpdateInterval != nil {
		pollingInterval = time.Duration(*apiOptions.DFMConfiguration.UpdateInterval) * time.Millisecond
	}

	// Setting DatafileURLTemplate
	urlString := fmt.Sprintf(localDatafileURLTemplate, sdkKey, scenarioID)

	configManager := config.NewPollingProjectConfigManager(
		sdkKey,
		config.WithInitialDatafile(datafile),
		config.WithPollingInterval(pollingInterval),
		config.WithDatafileURLTemplate(urlString),
	)

	return configManager
}

// GetDatafile returns datafile,error for the provided datafileName
func GetDatafile(datafileName string) ([]byte, error) {
	datafileDir := os.Getenv("DATAFILES_DIR")

	datafile, err := ioutil.ReadFile(filepath.Clean(path.Join(datafileDir, datafileName)))

	if err != nil {
		log.Fatal(err)
	}

	// TODO: we need to bubble up err
	return datafile, err
}

// GetSDKKey returns SDKKey for configuration
func GetSDKKey(configuration *models.DataFileManagerConfiguration) string {
	if configuration == nil {
		sdkKey++
		return strconv.Itoa(sdkKey)
	}
	key := configuration.SDKKey
	if configuration.DatafileCondition != "" {
		key += "_" + configuration.DatafileCondition
	}
	return key
}
