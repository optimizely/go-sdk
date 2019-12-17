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
	"encoding/json"
	"github.com/optimizely/go-sdk/pkg/entities"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/suite"
)

type OptimizelyConfigTestSuite struct {
	suite.Suite
	featureList              []entities.Feature
	experimentList           []entities.Experiment
	expectedOptimizelyConfig OptimizelyConfig
}

var featurelist = `[{"ID":"4482920077","Key":"mutex_group_feature","FeatureExperiments":[{"AudienceIds":[],"ID":"12198292373","LayerID":"12235440722","Key":"experiment_4000",
"Variations":{"12098126626":{"ID":"12098126626","Variables":{"2687470094":{"ID":"2687470094","Value":"50"},"2687470095":{"ID":"2687470095","Value":"50.5"},"2687470096":{"ID":"2687470096","Value":"false"},"2687470097":{"ID":"2687470097","Value":"s1"}},
"Key":"all_traffic_variation_exp_1","FeatureEnabled":true},"12107729995":{"ID":"12107729995","Variables":{},"Key":"no_traffic_variation_exp_1","FeatureEnabled":true}},"VariationKeyToIDMap":{"all_traffic_variation_exp_1":"12098126626","no_traffic_variation_exp_1":"12107729995"},
"TrafficAllocation":[{"EntityID":"12098126626","EndOfRange":10000}],"GroupID":"12115595438","AudienceConditionTree":null,"Whitelist":{},"IsFeatureExperiment":true},{"AudienceIds":[],"ID":"12198292374","LayerID":"12187694825","Key":"experiment_8000","Variations":
{"12252360417":{"ID":"12252360417","Variables":{"2687470094":{"ID":"2687470094","Value":"50"},"2687470095":{"ID":"2687470095","Value":"50.5"}},"Key":"no_traffic_variation_exp_2","FeatureEnabled":false}},"VariationKeyToIDMap":{"no_traffic_variation_exp_2":"12252360417"},
"TrafficAllocation":[{"EntityID":"12232050369","EndOfRange":10000}],"GroupID":"12115595438","AudienceConditionTree":null,"Whitelist":{},"IsFeatureExperiment":true}],"Rollout":{"ID":"","Experiments":null},"VariableMap":
{"b_true":{"DefaultValue":"true","ID":"2687470096","Key":"b_true","Type":"boolean"},"d_4_2":{"DefaultValue":"42.2","ID":"2687470095","Key":"d_4_2","Type":"double"},"i_42":{"DefaultValue":"42","ID":"2687470094","Key":"i_42","Type":"integer"},
"s_foo":{"DefaultValue":"foo","ID":"2687470097","Key":"s_foo","Type":"string"}}},{"ID":"4482920079","Key":"feature_exp_no_traffic","FeatureExperiments":[{"AudienceIds":[],"ID":"12198292376","LayerID":"12187694826","Key":"no_traffic_experiment",
"Variations":{"12098126629":{"ID":"12098126629","Variables":{},"Key":"variation_5000","FeatureEnabled":true},"12098126630":{"ID":"12098126630","Variables":{},"Key":"variation_10000","FeatureEnabled":true}},"VariationKeyToIDMap":
{"variation_10000":"12098126630","variation_5000":"12098126629"},"TrafficAllocation":[{"EntityID":"12098126629","EndOfRange":5000},{"EntityID":"12098126630","EndOfRange":10000}],"GroupID":"12115595439",
"AudienceConditionTree":null,"Whitelist":{},"IsFeatureExperiment":true}],"Rollout":{"ID":"","Experiments":null},"VariableMap":{}}]`

var experimentList = `[{"AudienceIds":[],"ID":"12198292374","LayerID":"12187694825","Key":"experiment_8000","Variations":{"12252360417":{"ID":"12252360417","Variables":{"2687470094":{"ID":"2687470094","Value":"50"},
"2687470095":{"ID":"2687470095","Value":"50.5"}},"Key":"no_traffic_variation_exp_2","FeatureEnabled":false}},"VariationKeyToIDMap":{"no_traffic_variation_exp_2":"12252360417"},"TrafficAllocation":[{"EntityID":"12232050369","EndOfRange":10000}],
"GroupID":"12115595438","AudienceConditionTree":null,"Whitelist":{},"IsFeatureExperiment":true},{"AudienceIds":[],"ID":"12198292375","LayerID":"12235440723","Key":"all_traffic_experiment","Variations":
{"12098126627":{"ID":"12098126627","Variables":{"2687470094":{"ID":"2687470094","Value":"4"},"2687470095":{"ID":"2687470095","Value":"10.6"},"2687470096":{"ID":"2687470096","Value":"true"},"2687470097":{"ID":"2687470097","Value":"s1 foo"}},
"Key":"all_traffic_variation","FeatureEnabled":true},"12098126628":{"ID":"12098126628","Variables":{},"Key":"no_traffic_variation","FeatureEnabled":true}},"VariationKeyToIDMap":{"all_traffic_variation":"12098126627","no_traffic_variation":"12098126628"},
"TrafficAllocation":[{"EntityID":"12098126627","EndOfRange":10000}],"GroupID":"12115595439","AudienceConditionTree":null,"Whitelist":{},"IsFeatureExperiment":false},{"AudienceIds":[],"ID":"12198292376","LayerID":"12187694826","Key":"no_traffic_experiment",
"Variations":{"12098126629":{"ID":"12098126629","Variables":{},"Key":"variation_5000","FeatureEnabled":true},"12098126630":{"ID":"12098126630","Variables":{},"Key":"variation_10000","FeatureEnabled":true}},"VariationKeyToIDMap":
{"variation_10000":"12098126630","variation_5000":"12098126629"},"TrafficAllocation":[{"EntityID":"12098126629","EndOfRange":5000},{"EntityID":"12098126630","EndOfRange":10000}],"GroupID":"12115595439","AudienceConditionTree":null,"Whitelist":{},"IsFeatureExperiment":true},
{"AudienceIds":[],"ID":"10390977673","LayerID":"10420273888","Key":"exp_with_audience","Variations":{"10389729780":{"ID":"10389729780","Variables":{},"Key":"a","FeatureEnabled":true},"10416523121":{"ID":"10416523121","Variables":{},"Key":"b","FeatureEnabled":false}},
"VariationKeyToIDMap":{"a":"10389729780","b":"10416523121"},"TrafficAllocation":[{"EntityID":"10389729780","EndOfRange":10000}],"GroupID":"","AudienceConditionTree":null,"Whitelist":{},"IsFeatureExperiment":false},
{"AudienceIds":[],"ID":"12198292373","LayerID":"12235440722","Key":"experiment_4000","Variations":{"12098126626":{"ID":"12098126626","Variables":{"2687470094":{"ID":"2687470094","Value":"50"},"2687470095":{"ID":"2687470095","Value":"50.5"},
"2687470096":{"ID":"2687470096","Value":"false"},"2687470097":{"ID":"2687470097","Value":"s1"}},"Key":"all_traffic_variation_exp_1","FeatureEnabled":true},"12107729995":{"ID":"12107729995","Variables":{},"Key":"no_traffic_variation_exp_1","FeatureEnabled":true}},
"VariationKeyToIDMap":{"all_traffic_variation_exp_1":"12098126626","no_traffic_variation_exp_1":"12107729995"},"TrafficAllocation":[{"EntityID":"12098126626","EndOfRange":10000}],"GroupID":"12115595438","AudienceConditionTree":null,"Whitelist":{},"IsFeatureExperiment":true}]
`

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

	err = json.Unmarshal([]byte(featurelist), &s.featureList)
	if err != nil {
		s.Fail("unable to parse features")
	}

	err = json.Unmarshal([]byte(experimentList), &s.experimentList)
	if err != nil {
		s.Fail("unable to parse experiments")
	}

}

func (s *OptimizelyConfigTestSuite) TestOptlyConfig() {
	optimizelyConfig := NewOptimizelyConfig(s.featureList, s.experimentList, "9")

	s.Equal(s.expectedOptimizelyConfig.FeaturesMap, optimizelyConfig.FeaturesMap)
	s.Equal(s.expectedOptimizelyConfig.ExperimentsMap, optimizelyConfig.ExperimentsMap)
	s.Equal(s.expectedOptimizelyConfig.Revision, optimizelyConfig.Revision)

	s.Equal(s.expectedOptimizelyConfig, *optimizelyConfig)

}

func TestOptimizelyConfigTestSuite(t *testing.T) {
	suite.Run(t, new(OptimizelyConfigTestSuite))
}

func TestGetOptimizelyConfig(t *testing.T) {
	dataFileName := "test/optimizely_config_datafile.json"
	datafile, err := ioutil.ReadFile(dataFileName)
	if err != nil {
		assert.Fail(t, "error opening file:"+dataFileName)
	}
	projectConfig, e := NewDatafileProjectConfig(datafile)
	if e != nil {
		assert.Fail(t, "error parsing datafile")
	}

	// reading expected json file validates public access for OptimizelyConfig and its members
	outputFileName := "test/optimizely_config_expected.json"
	expectedOutput, er := ioutil.ReadFile(outputFileName)
	if er != nil {
		assert.Fail(t, "error opening file "+outputFileName)
	}

	var expectedOptimizelyConfig = entities.OptimizelyConfig{}
	err = json.Unmarshal(expectedOutput, &expectedOptimizelyConfig)
	if err != nil {
		assert.Fail(t, "unable to parse expected file")
	}
	optimizelyConfig := projectConfig.GetOptimizelyConfig()

	assert.Equal(t, expectedOptimizelyConfig.FeaturesMap, optimizelyConfig.FeaturesMap)
	assert.Equal(t, expectedOptimizelyConfig.ExperimentsMap, optimizelyConfig.ExperimentsMap)
	assert.Equal(t, expectedOptimizelyConfig.Revision, optimizelyConfig.Revision)

	assert.Equal(t, expectedOptimizelyConfig, *optimizelyConfig)

}
