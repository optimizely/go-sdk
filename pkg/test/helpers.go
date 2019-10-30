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

// Package test //
package test

import (
	"fmt"

	"github.com/optimizely/go-sdk/pkg/entities"
)

// MakeTestExperiment creates a new experiment with the given key
func MakeTestExperiment(experimentKey string) entities.Experiment {
	return entities.Experiment{
		ID:  fmt.Sprintf("%s_id", experimentKey),
		Key: experimentKey,
		Variations: map[string]entities.Variation{
			"v1": entities.Variation{ID: "v1", Key: "v1"},
			"v2": entities.Variation{ID: "v2", Key: "v2"},
		},
	}
}

// MakeTestVariation creates a new variation with the given key and params
func MakeTestVariation(variationKey string, featureEnabled bool) entities.Variation {
	return entities.Variation{
		ID:             fmt.Sprintf("test_variation_%s", variationKey),
		Key:            variationKey,
		FeatureEnabled: featureEnabled,
	}
}

// MakeTestExperimentWithVariations creates an experiment with the given key and variations
func MakeTestExperimentWithVariations(experimentKey string, variations []entities.Variation) entities.Experiment {
	variationsMap := make(map[string]entities.Variation)
	for _, variation := range variations {
		variationsMap[variation.ID] = variation
	}
	return entities.Experiment{
		Key:        experimentKey,
		ID:         fmt.Sprintf("test_experiment_%s", experimentKey),
		Variations: variationsMap,
	}
}

// MakeTestFeatureWithExperiment creates a new feature test with the given key and experiment
func MakeTestFeatureWithExperiment(featureKey string, experiment entities.Experiment) entities.Feature {
	testFeature := entities.Feature{
		ID:                 fmt.Sprintf("test_feature_%s", featureKey),
		Key:                featureKey,
		FeatureExperiments: []entities.Experiment{experiment},
	}

	return testFeature
}
