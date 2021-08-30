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
}

func (s *OptimizelyConfigTestSuite) TestOptlyConfig() {

	test := func(datafileOne string, datafileTwo string) {
		// Cleaned here due to linter warning
		folder := "testdata/"
		dataFile, err := ioutil.ReadFile(folder + datafileOne)
		if err != nil {
			s.Fail("error opening file " + datafileOne)
		}

		projectMgr := NewStaticProjectConfigManagerWithOptions("", WithInitialDatafile(dataFile))
		projConfig, err := projectMgr.GetConfig()
		if err != nil {
			s.Fail(err.Error())
		}
		optimizelyConfig := NewOptimizelyConfig(projConfig)

		dataFile2, err := ioutil.ReadFile(folder + datafileTwo)
		if err != nil {
			s.Fail("error opening file " + datafileTwo)
		}
		var expectedConfig OptimizelyConfig
		err = json.Unmarshal(dataFile2, &expectedConfig)
		if err != nil {
			s.Fail("unable to parse expected file")
		}
		expectedConfig.datafile = string(dataFile)

		s.Equal(expectedConfig.Attributes, optimizelyConfig.Attributes)
		s.Equal(expectedConfig.Audiences, optimizelyConfig.Audiences)
		s.Equal(expectedConfig.FeaturesMap, optimizelyConfig.FeaturesMap)
		s.Equal(expectedConfig.ExperimentsMap, optimizelyConfig.ExperimentsMap)
		s.Equal(expectedConfig.Revision, optimizelyConfig.Revision)
		s.Equal(expectedConfig.datafile, optimizelyConfig.datafile)
		s.Equal(expectedConfig.SdkKey, optimizelyConfig.SdkKey)
		s.Equal(expectedConfig.EnvironmentKey, optimizelyConfig.EnvironmentKey)
		s.Equal(expectedConfig.Events, optimizelyConfig.Events)

		s.Equal(expectedConfig, *optimizelyConfig)
	}

	// Simple datafile
	test("optimizely_config_datafile.json", "optimizely_config_expected.json")
	// Similar keys datafile
	test("similar_exp_keys_datafile.json", "similar_exp_keys_expected.json")
	// Similar rule keys bucketing datafile
	test("similar_rule_keys_bucketing_datafile.json", "similar_rule_keys_bucketing_expected.json")
}

func (s *OptimizelyConfigTestSuite) TestSerializeAudiences() {

	dataFileName := "testdata/typed_audience_datafile.json"
	dataFile, err := ioutil.ReadFile(dataFileName)
	if err != nil {
		s.Fail("error opening file " + dataFileName)
	}

	projectMgr := NewStaticProjectConfigManagerWithOptions("", WithInitialDatafile(dataFile))
	projConfig, err := projectMgr.GetConfig()
	if err != nil {
		s.Fail(err.Error())
	}

	conditions := []interface{}{
		[]interface{}{"or", "3468206642", "3988293898"},
		[]interface{}{"or", "3468206642", "3988293898", "3468206646"},
		[]interface{}{"not", "3468206642"},
		[]interface{}{"or", "3468206642"},
		[]interface{}{"and", "3468206642"},
		[]interface{}{"3468206642"},
		[]interface{}{"3468206642", "3988293898"},
		[]interface{}{"and", []interface{}{"or", "3468206642", "3988293898"}, "3468206646"},
		[]interface{}{"and", []interface{}{"or", "3468206642", []interface{}{"and", "3988293898", "3468206646"}}, []interface{}{"and", "3988293899", []interface{}{"or", "3468206647", "3468206643"}}},
		[]interface{}{"and", "and"},
		[]interface{}{"not", []interface{}{"and", "3468206642", "3988293898"}},
		[]interface{}{},
		[]interface{}{"or", "3468206642", "999999999"},
	}

	expectedOutputs := []string{
		"\"exactString\" OR \"substringString\"",
		"\"exactString\" OR \"substringString\" OR \"exactNumber\"",
		"NOT \"exactString\"",
		"\"exactString\"",
		"\"exactString\"",
		"\"exactString\"",
		"\"exactString\" OR \"substringString\"",
		"(\"exactString\" OR \"substringString\") AND \"exactNumber\"",
		"(\"exactString\" OR (\"substringString\" AND \"exactNumber\")) AND (\"exists\" AND (\"gtNumber\" OR \"exactBoolean\"))",
		"",
		"NOT (\"exactString\" AND \"substringString\")",
		"",
		"\"exactString\" OR \"999999999\"",
	}

	for i, condition := range conditions {
		result := getSerializedAudiences(condition, projConfig.GetAudienceMap())
		s.Equal(expectedOutputs[i], result)
	}

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

	s.Equal("", jsonMap["sdkKey"])
	s.Equal("", jsonMap["environmentKey"])
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
