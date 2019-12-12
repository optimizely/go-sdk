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
	"os"
	"path"
	"path/filepath"
	"strconv"

	"github.com/optimizely/go-sdk/tests/integration/models"
)

// Since notificationManager is mapped against sdkKey, we need a unique sdkKey for every scenario
var sdkKey int

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
