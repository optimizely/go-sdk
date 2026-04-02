/****************************************************************************
 * Copyright 2026, Optimizely, Inc. and contributors                       *
 *                                                                          *
 * Licensed under the Apache License, Version 2.0 (the "License");          *
 * you may not use this file except in compliance with the License.         *
 * You may obtain a copy of the License at                                  *
 *                                                                          *
 *    https://www.apache.org/licenses/LICENSE-2.0                           *
 *                                                                          *
 * Unless required by applicable law or agreed to in writing, software      *
 * distributed under the License is distributed on an "AS IS" BASIS,        *
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. *
 * See the License for the specific language governing permissions and      *
 * limitations under the License.                                           *
 ***************************************************************************/

package datafileprojectconfig

import (
	"testing"

	"github.com/optimizely/go-sdk/v2/pkg/entities"
	"github.com/optimizely/go-sdk/v2/pkg/logging"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const featureRolloutDatafile = `{
  "accountId": "12345",
  "anonymizeIP": false,
  "sendFlagDecisions": true,
  "botFiltering": false,
  "projectId": "67890",
  "revision": "1",
  "sdkKey": "FeatureRolloutTest",
  "environmentKey": "production",
  "version": "4",
  "audiences": [],
  "typedAudiences": [],
  "attributes": [],
  "events": [],
  "groups": [],
  "integrations": [],
  "experiments": [
    {
      "id": "exp_rollout_1",
      "key": "feature_rollout_experiment",
      "status": "Running",
      "layerId": "layer_1",
      "audienceIds": [],
      "forcedVariations": {},
      "type": "fr",
      "variations": [
        {
          "id": "var_rollout_1",
          "key": "rollout_variation",
          "featureEnabled": true
        }
      ],
      "trafficAllocation": [
        {
          "entityId": "var_rollout_1",
          "endOfRange": 5000
        }
      ]
    },
    {
      "id": "exp_ab_1",
      "key": "ab_test_experiment",
      "status": "Running",
      "layerId": "layer_2",
      "audienceIds": [],
      "forcedVariations": {},
      "type": "ab",
      "variations": [
        {
          "id": "var_ab_1",
          "key": "control",
          "featureEnabled": false
        },
        {
          "id": "var_ab_2",
          "key": "treatment",
          "featureEnabled": true
        }
      ],
      "trafficAllocation": [
        {
          "entityId": "var_ab_1",
          "endOfRange": 5000
        },
        {
          "entityId": "var_ab_2",
          "endOfRange": 10000
        }
      ]
    },
    {
      "id": "exp_no_type",
      "key": "no_type_experiment",
      "status": "Running",
      "layerId": "layer_3",
      "audienceIds": [],
      "forcedVariations": {},
      "variations": [
        {
          "id": "var_notype_1",
          "key": "variation_1",
          "featureEnabled": true
        }
      ],
      "trafficAllocation": [
        {
          "entityId": "var_notype_1",
          "endOfRange": 10000
        }
      ]
    },
    {
      "id": "exp_rollout_no_rollout_id",
      "key": "rollout_no_rollout_id_experiment",
      "status": "Running",
      "layerId": "layer_4",
      "audienceIds": [],
      "forcedVariations": {},
      "type": "fr",
      "variations": [
        {
          "id": "var_no_rollout_1",
          "key": "rollout_no_rollout_variation",
          "featureEnabled": true
        }
      ],
      "trafficAllocation": [
        {
          "entityId": "var_no_rollout_1",
          "endOfRange": 5000
        }
      ]
    }
  ],
  "featureFlags": [
    {
      "id": "flag_1",
      "key": "feature_with_rollout",
      "rolloutId": "rollout_1",
      "experimentIds": ["exp_rollout_1"],
      "variables": []
    },
    {
      "id": "flag_2",
      "key": "feature_with_ab",
      "rolloutId": "rollout_2",
      "experimentIds": ["exp_ab_1"],
      "variables": []
    },
    {
      "id": "flag_3",
      "key": "feature_no_rollout_id",
      "rolloutId": "",
      "experimentIds": ["exp_rollout_no_rollout_id"],
      "variables": []
    }
  ],
  "rollouts": [
    {
      "id": "rollout_1",
      "experiments": [
        {
          "id": "rollout_exp_1",
          "key": "rollout_rule_1",
          "status": "Running",
          "layerId": "rollout_layer_1",
          "audienceIds": [],
          "forcedVariations": {},
          "variations": [
            {
              "id": "rollout_var_1",
              "key": "rollout_enabled",
              "featureEnabled": true
            }
          ],
          "trafficAllocation": [
            {
              "entityId": "rollout_var_1",
              "endOfRange": 10000
            }
          ]
        },
        {
          "id": "rollout_exp_everyone",
          "key": "everyone_else_rule",
          "status": "Running",
          "layerId": "rollout_layer_everyone",
          "audienceIds": [],
          "forcedVariations": {},
          "variations": [
            {
              "id": "everyone_else_var",
              "key": "everyone_else_variation",
              "featureEnabled": false
            }
          ],
          "trafficAllocation": [
            {
              "entityId": "everyone_else_var",
              "endOfRange": 10000
            }
          ]
        }
      ]
    },
    {
      "id": "rollout_2",
      "experiments": [
        {
          "id": "rollout_exp_2",
          "key": "rollout_rule_2",
          "status": "Running",
          "layerId": "rollout_layer_2",
          "audienceIds": [],
          "forcedVariations": {},
          "variations": [
            {
              "id": "rollout_var_2",
              "key": "rollout_variation_2",
              "featureEnabled": true
            }
          ],
          "trafficAllocation": [
            {
              "entityId": "rollout_var_2",
              "endOfRange": 10000
            }
          ]
        }
      ]
    }
  ]
}`

func loadFeatureRolloutConfig(t *testing.T) *DatafileProjectConfig {
	config, err := NewDatafileProjectConfig([]byte(featureRolloutDatafile), logging.GetLogger("", "FeatureRolloutTest"))
	require.NoError(t, err)
	require.NotNil(t, config)
	return config
}

// Test 1: Backward compatibility - experiments without type field have type="" (zero value)
func TestExperimentWithoutTypeFieldHasEmptyType(t *testing.T) {
	config := loadFeatureRolloutConfig(t)
	experiment, err := config.GetExperimentByKey("no_type_experiment")
	assert.NoError(t, err)
	assert.Empty(t, experiment.Type, "Type should be empty for experiments without type field")
}

// Test 2: Core injection - feature_rollout experiments get everyone else variation + trafficAllocation injected
func TestFeatureRolloutExperimentGetsEveryoneElseVariationInjected(t *testing.T) {
	config := loadFeatureRolloutConfig(t)
	experiment, err := config.GetExperimentByKey("feature_rollout_experiment")
	assert.NoError(t, err)
	assert.Equal(t, entities.ExperimentTypeFR, experiment.Type)

	// Should have 2 variations: original + everyone else
	assert.Equal(t, 2, len(experiment.Variations), "Should have 2 variations after injection")

	// Check the injected variation exists
	injectedVariation, ok := experiment.Variations["everyone_else_var"]
	assert.True(t, ok, "Should contain injected variation by ID")
	assert.Equal(t, "everyone_else_variation", injectedVariation.Key)

	// Check the injected traffic allocation
	assert.Equal(t, 2, len(experiment.TrafficAllocation), "Should have 2 traffic allocations after injection")
	lastAllocation := experiment.TrafficAllocation[len(experiment.TrafficAllocation)-1]
	assert.Equal(t, "everyone_else_var", lastAllocation.EntityID)
	assert.Equal(t, 10000, lastAllocation.EndOfRange)
}

// Test 3: Variation maps updated - VariationKeyToIDMap contains the injected variation
func TestVariationMapsContainInjectedVariation(t *testing.T) {
	config := loadFeatureRolloutConfig(t)
	experiment, err := config.GetExperimentByKey("feature_rollout_experiment")
	assert.NoError(t, err)

	// Check VariationKeyToIDMap contains the injected variation
	variationID, ok := experiment.VariationKeyToIDMap["everyone_else_variation"]
	assert.True(t, ok, "VariationKeyToIDMap should contain injected variation key")
	assert.Equal(t, "everyone_else_var", variationID)
}

// Test 4: Non-rollout unchanged - A/B experiments are not modified
func TestABTestExperimentNotModified(t *testing.T) {
	config := loadFeatureRolloutConfig(t)
	experiment, err := config.GetExperimentByKey("ab_test_experiment")
	assert.NoError(t, err)
	assert.Equal(t, entities.ExperimentTypeAB, experiment.Type)

	// Should still have exactly 2 original variations
	assert.Equal(t, 2, len(experiment.Variations), "A/B test should keep original 2 variations")
	assert.Equal(t, 2, len(experiment.TrafficAllocation), "A/B test should keep original 2 traffic allocations")
}

// Test 5: No rollout edge case - feature_rollout with empty rolloutId does not crash
func TestFeatureRolloutWithEmptyRolloutIdDoesNotCrash(t *testing.T) {
	config := loadFeatureRolloutConfig(t)
	experiment, err := config.GetExperimentByKey("rollout_no_rollout_id_experiment")
	assert.NoError(t, err)
	assert.Equal(t, entities.ExperimentTypeFR, experiment.Type)

	// Should keep only original variation since rollout cannot be resolved
	assert.Equal(t, 1, len(experiment.Variations), "Should keep only original variation")
}

// Test 6: Type field parsed - experiments with type field have the value correctly preserved
func TestTypeFieldCorrectlyParsed(t *testing.T) {
	config := loadFeatureRolloutConfig(t)

	rolloutExp, err := config.GetExperimentByKey("feature_rollout_experiment")
	assert.NoError(t, err)
	assert.Equal(t, entities.ExperimentTypeFR, rolloutExp.Type)

	abExp, err := config.GetExperimentByKey("ab_test_experiment")
	assert.NoError(t, err)
	assert.Equal(t, entities.ExperimentTypeAB, abExp.Type)

	noTypeExp, err := config.GetExperimentByKey("no_type_experiment")
	assert.NoError(t, err)
	assert.Empty(t, noTypeExp.Type)
}

// Test 7: Unknown experiment type accepted - config parsing succeeds with unknown type value
func TestUnknownExperimentTypeAccepted(t *testing.T) {
	datafile := `{
  "accountId": "12345",
  "anonymizeIP": false,
  "sendFlagDecisions": true,
  "botFiltering": false,
  "projectId": "67890",
  "revision": "1",
  "sdkKey": "UnknownTypeTest",
  "environmentKey": "production",
  "version": "4",
  "audiences": [],
  "typedAudiences": [],
  "attributes": [],
  "events": [],
  "groups": [],
  "integrations": [],
  "experiments": [
    {
      "id": "exp_unknown",
      "key": "unknown_type_experiment",
      "status": "Running",
      "layerId": "layer_1",
      "audienceIds": [],
      "forcedVariations": {},
      "type": "new_unknown_type",
      "variations": [
        {
          "id": "var_1",
          "key": "variation_1",
          "featureEnabled": true
        }
      ],
      "trafficAllocation": [
        {
          "entityId": "var_1",
          "endOfRange": 10000
        }
      ]
    }
  ],
  "featureFlags": [
    {
      "id": "flag_1",
      "key": "test_flag",
      "rolloutId": "",
      "experimentIds": ["exp_unknown"],
      "variables": []
    }
  ],
  "rollouts": []
}`

	logger := logging.GetLogger("test", "TestUnknownExperimentTypeAccepted")
	config, err := NewDatafileProjectConfig([]byte(datafile), logger)
	require.NoError(t, err, "Config parsing should succeed with unknown experiment type")
	require.NotNil(t, config)

	experiment, err := config.GetExperimentByKey("unknown_type_experiment")
	assert.NoError(t, err)
	assert.Equal(t, entities.ExperimentType("new_unknown_type"), experiment.Type)
}
