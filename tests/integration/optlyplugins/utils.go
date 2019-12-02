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
	"time"

	"github.com/optimizely/go-sdk/pkg/config"
	"github.com/optimizely/go-sdk/pkg/utils"
	"github.com/optimizely/go-sdk/tests/integration/models"
)

const localDatafileURLTemplate = "http://localhost:3001/datafiles/%s.json?request_id="

// SyncConfig doesn't request for new datafile if we provide a valid datafile
// this requires us to keep defaultPollingInterval low so that the request
// initiated from Start method is executed quickly
const defaultPollingInterval = time.Duration(1000) * time.Millisecond

// CreatePollingConfigManager creates a pollingConfigManager with given configuration
func CreatePollingConfigManager(options models.APIOptions) *TestProjectConfigManager {
	var pollingConfigManagerOptions []config.OptionFunc

	// Setting up initial datafile
	if options.DatafileName != "" {
		datafile, err := GetDatafile(options.DatafileName)
		if err != nil {
			log.Fatal(err)
		}
		pollingConfigManagerOptions = append(pollingConfigManagerOptions, config.WithInitialDatafile(datafile))
	}
	// Setting up polling interval
	pollingTimeInterval := defaultPollingInterval
	if options.DFMConfiguration.UpdateInterval != nil {
		pollingTimeInterval = time.Duration((*options.DFMConfiguration.UpdateInterval)) * time.Millisecond
	}
	pollingConfigManagerOptions = append(pollingConfigManagerOptions, config.WithPollingInterval(pollingTimeInterval))
	sdkKey := GetSDKKey(options.DFMConfiguration)
	urlString := localDatafileURLTemplate + options.ScenarioID
	pollingConfigManagerOptions = append(pollingConfigManagerOptions, config.WithDatafileURLTemplate(urlString))

	testProjectConfigManagerInstance := &TestProjectConfigManager{}
	pollingConfigManagerOptions = append(pollingConfigManagerOptions, config.WithNotificationHandlers(testProjectConfigManagerInstance.GetListenerCallbacks(options)...))

	configManager := config.NewPollingProjectConfigManager(
		sdkKey,
		pollingConfigManagerOptions...,
	)
	testProjectConfigManagerInstance.ProjectConfigManager = configManager
	// Since we are using TestProjectConfigManager over ProjectConfigManager, factory will
	// not call the start method for ProjectConfigManager, so we have to do it manually
	exeCtx := utils.NewCancelableExecutionCtx()
	configManager.Start(sdkKey, exeCtx)
	// Verify datafile configuration tests
	testProjectConfigManagerInstance.Verify(options.DFMConfiguration)

	return testProjectConfigManagerInstance
}

// GetDatafile returns datafile,error for the provided datafileName
func GetDatafile(datafileName string) ([]byte, error) {
	datafileDir := os.Getenv("DATAFILES_DIR")
	return ioutil.ReadFile(filepath.Clean(path.Join(datafileDir, datafileName)))
}

// GetSDKKey returns SDKKey for configuration
func GetSDKKey(configuration models.DataFileManagerConfiguration) (sdkKey string) {
	sdkKey = configuration.SDKKey
	if configuration.DatafileCondition != "" {
		sdkKey += "_" + configuration.DatafileCondition
	}
	return sdkKey
}
