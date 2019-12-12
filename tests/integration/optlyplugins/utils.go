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
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"time"

	"github.com/optimizely/go-sdk/pkg/config"
	"github.com/optimizely/go-sdk/tests/integration/models"
)

const localDatafileURLTemplate = "http://localhost:3001/datafiles/%s.json?request_id="

// SyncConfig doesn't request for new datafile if we provide a valid datafile
// this requires us to keep defaultPollingInterval low so that the request
// initiated from Start method is executed quickly
const defaultPollingInterval = time.Duration(1000) * time.Millisecond

// Since notificationManager is mapped against sdkKey, we need a unique sdkKey for every scenario
var sdkKey int

// CreatePollingConfigManager creates a pollingConfigManager with given configuration
func CreatePollingConfigManager(sdkKey, scenarioID string, notificationManager *NotificationManager) config.ProjectConfigManager {

	// Add revision as delta to waitgroup for projectConfigNotification
	if notificationManager.APIOptions.DFMConfiguration.Mode == models.KeyWaitForConfigUpdate {
		// default as 1 For cases where we are just waiting for a single project notification
		revision := 1
		if notificationManager.APIOptions.DFMConfiguration.Revision != nil {
			revision = *(notificationManager.APIOptions.DFMConfiguration.Revision)
		}
		notificationManager.WaitGroup.Add(revision)
	}

	var pollingConfigManagerOptions []config.OptionFunc
	// Setting up optional initial datafile
	if notificationManager.APIOptions.DatafileName != "" {
		datafile, err := GetDatafile(notificationManager.APIOptions.DatafileName)
		if err != nil {
			log.Fatal(err)
		}
		pollingConfigManagerOptions = append(pollingConfigManagerOptions, config.WithInitialDatafile(datafile))
	}
	// Setting up polling interval
	pollingTimeInterval := defaultPollingInterval
	if notificationManager.APIOptions.DFMConfiguration.UpdateInterval != nil {
		pollingTimeInterval = time.Duration((*notificationManager.APIOptions.DFMConfiguration.UpdateInterval)) * time.Millisecond
	}
	pollingConfigManagerOptions = append(pollingConfigManagerOptions, config.WithPollingInterval(pollingTimeInterval))
	// Setting DatafileURLTemplate
	urlString := localDatafileURLTemplate + scenarioID
	pollingConfigManagerOptions = append(pollingConfigManagerOptions, config.WithDatafileURLTemplate(urlString))
	// Adding callbacks and creating config manager with options
	notificationManager.SubscribeProjectConfigUpdateNotifications(sdkKey)
	configManager := config.NewPollingProjectConfigManager(
		sdkKey,
		pollingConfigManagerOptions...,
	)
	return configManager
}

// GetDatafile returns datafile,error for the provided datafileName
func GetDatafile(datafileName string) ([]byte, error) {
	datafileDir := os.Getenv("DATAFILES_DIR")
	return ioutil.ReadFile(filepath.Clean(path.Join(datafileDir, datafileName)))
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
