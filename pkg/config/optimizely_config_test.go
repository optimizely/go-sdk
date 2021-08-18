/****************************************************************************
 * Copyright 2019-2021, Optimizely, Inc. and contributors                   *
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
	"encoding/json"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/suite"
)

type OptimizelyConfigTestSuite struct {
	suite.Suite
	projectConfig            ProjectConfig
	expectedOptimizelyConfig OptimizelyConfig
}

func (s *OptimizelyConfigTestSuite) SetupTest() {

	// reading expected json file validates public access for OptimizelyConfig and its members
	outputFileName := "testdata/optimizely_config_expected.json"
	expectedOutput, err := ioutil.ReadFile(outputFileName)
	if err != nil {
		s.Fail("error opening file " + outputFileName)
	}

	err = json.Unmarshal(expectedOutput, &s.expectedOptimizelyConfig)
	if err != nil {
		s.Fail("unable to parse expected file")
	}

	dataFileName := "testdata/optimizely_config_datafile.json"
	dataFile, err := ioutil.ReadFile(dataFileName)
	if err != nil {
		s.Fail("error opening file " + dataFileName)
	}
	s.expectedOptimizelyConfig.datafile = string(dataFile)

	projectMgr := NewStaticProjectConfigManagerWithOptions("", WithInitialDatafile(dataFile))

	s.projectConfig = projectMgr.projectConfig

}

func (s *OptimizelyConfigTestSuite) TestOptlyConfig() {
	optimizelyConfig := NewOptimizelyConfig(s.projectConfig)

	s.Equal(s.expectedOptimizelyConfig.FeaturesMap, optimizelyConfig.FeaturesMap)
	s.Equal(s.expectedOptimizelyConfig.ExperimentsMap, optimizelyConfig.ExperimentsMap)
	s.Equal(s.expectedOptimizelyConfig.Revision, optimizelyConfig.Revision)
	s.Equal(s.expectedOptimizelyConfig.datafile, optimizelyConfig.datafile)
	s.Equal(s.expectedOptimizelyConfig.SdkKey, optimizelyConfig.SdkKey)
	s.Equal(s.expectedOptimizelyConfig.EnvironmentKey, optimizelyConfig.EnvironmentKey)

	s.Equal(s.expectedOptimizelyConfig, *optimizelyConfig)
}

func (s *OptimizelyConfigTestSuite) TestOptlyConfigUnMarshalEmptySDKKeyAndEnvironmentKey() {
	datafile := []byte(`{"version":"4"}`)
	projectMgr := NewStaticProjectConfigManagerWithOptions("", WithInitialDatafile(datafile))
	optimizelyConfig := NewOptimizelyConfig(projectMgr.projectConfig)
	s.Equal("", optimizelyConfig.SdkKey)
	s.Equal("", optimizelyConfig.EnvironmentKey)

	var jsonMap map[string]interface{}
	bytesData, _ := json.Marshal(optimizelyConfig)
	json.Unmarshal(bytesData, &jsonMap)

	_, keyExists := jsonMap["sdkKey"]
	s.True(keyExists)

	_, keyExists = jsonMap["environmentKey"]
	s.True(keyExists)
}

func _TestTemp(t *testing.T) {
	dataFileName := "../../dataV2.json"
	dataFile, err := ioutil.ReadFile(dataFileName)
	if err != nil {
		panic(err)
	}
	projectMgr := NewStaticProjectConfigManagerWithOptions("", WithInitialDatafile(dataFile))
	oc := NewOptimizelyConfig(projectMgr.projectConfig)
	ba, err := json.Marshal(oc)
	if err != nil {
		panic(err)
	}
	ioutil.WriteFile("../../oc.json", ba, 0644)
}

func (s *OptimizelyConfigTestSuite) TestOptlyConfigUnMarshalNonEmptySDKKeyAndEnvironmentKey() {
	datafile := []byte(`{"version":"4", "sdkKey":"a", "environmentKey": "production"}`)
	projectMgr := NewStaticProjectConfigManagerWithOptions("", WithInitialDatafile(datafile))
	optimizelyConfig := NewOptimizelyConfig(projectMgr.projectConfig)
	s.Equal("a", optimizelyConfig.SdkKey)
	s.Equal("production", optimizelyConfig.EnvironmentKey)

	var jsonMap map[string]interface{}
	bytesData, _ := json.Marshal(optimizelyConfig)
	json.Unmarshal(bytesData, &jsonMap)

	_, keyExists := jsonMap["sdkKey"]
	s.True(keyExists)

	_, keyExists = jsonMap["environmentKey"]
	s.True(keyExists)
}

func (s *OptimizelyConfigTestSuite) TestOptlyConfigNullProjectConfig() {
	optimizelyConfig := NewOptimizelyConfig(nil)

	s.Nil(optimizelyConfig)
}

func (s *OptimizelyConfigTestSuite) TestOptlyConfigGetDatafile() {
	datafile := []byte(`{"version":"4"}`)
	projectMgr := NewStaticProjectConfigManagerWithOptions("", WithInitialDatafile(datafile))
	optimizelyConfig := NewOptimizelyConfig(projectMgr.projectConfig)
	s.NotNil(optimizelyConfig.datafile)
	s.Equal(string(datafile), optimizelyConfig.GetDatafile())
}

func TestOptimizelyConfigTestSuite(t *testing.T) {
	suite.Run(t, new(OptimizelyConfigTestSuite))
}
