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

package datafileprojectconfig

import (
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/suite"
)

type OptimizelyConfigTestSuite struct {
	suite.Suite
	projectConfig            *DatafileProjectConfig
	expectedOptimizelyConfig OptimizelyConfig
}

func (s *OptimizelyConfigTestSuite) SetupTest() {
	dataFileName := "test/optimizely_config_datafile.json"
	datafile, err := ioutil.ReadFile(dataFileName)
	if err != nil {
		s.Fail("error opening file:" + dataFileName)
	}
	projectConfig, e := NewDatafileProjectConfig(datafile)
	if e != nil {
		s.Fail("error parsing datafile")
	}
	s.projectConfig = projectConfig

	// reading expected json file validates public access for OptimizelyConfig and its members
	outputFileName := "test/optimizely_config_expected.json"
	expectedOutput, er := ioutil.ReadFile(outputFileName)
	if er != nil {
		s.Fail("error opening file " + outputFileName)
	}

	err = json.Unmarshal(expectedOutput, &s.expectedOptimizelyConfig)
	if err != nil {
		s.Fail("unable to parse expected file")
	}

}

func (s *OptimizelyConfigTestSuite) TestOptlyConfig() {
	optimizelyConfig, err := s.projectConfig.GetOptimizelyConfig()

	s.NoError(err)

	s.Equal(s.expectedOptimizelyConfig.FeaturesMap, optimizelyConfig.FeaturesMap)
	s.Equal(s.expectedOptimizelyConfig.ExperimentsMap, optimizelyConfig.ExperimentsMap)
	s.Equal(s.expectedOptimizelyConfig.Revision, optimizelyConfig.Revision)

	s.Equal(s.expectedOptimizelyConfig, *optimizelyConfig)

}

func TestOptimizelyConfigTestSuite(t *testing.T) {
	suite.Run(t, new(OptimizelyConfigTestSuite))
}
