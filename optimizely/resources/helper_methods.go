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

package resources

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/optimizely/go-sdk/optimizely/config/datafileProjectConfig/entities"
)

var goPath = os.Getenv("GOPATH")

// GetTestDataFileJSON will attempt to return datafile entity for datafileName
func GetTestDataFileJSON(datafileName string) (*entities.Datafile, error) {
	directoryPath := goPath + "/src/github.com/optimizely/go-sdk/optimizely/resources/testdatafiles"
	filePath, _ := filepath.Abs(filepath.Join(directoryPath, datafileName))
	jsonDatafile, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	datafile := &entities.Datafile{}
	err1 := json.Unmarshal(jsonDatafile, &datafile)
	if err1 != nil {
		return nil, err1
	}
	return datafile, nil
}
