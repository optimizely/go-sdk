/****************************************************************************
 * Copyright 2020, Optimizely, Inc. and contributors                        *
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

// Package decision //
package decision

import (
	"testing"

	"github.com/optimizely/go-sdk/pkg/entities"
	"github.com/optimizely/go-sdk/pkg/notification"

	"github.com/stretchr/testify/assert"
)

func TestFeatureNotification(t *testing.T) {
	featureDecision := &FeatureDecision{Source: FeatureTest, Experiment: entities.Experiment{}, Variation: &entities.Variation{}}

	featureKey := "feature_key"
	userContext := &entities.UserContext{}
	decision := FeatureNotification(featureKey, featureDecision, userContext)

	expectedDecision := &notification.DecisionNotification{Type: "feature", UserContext: entities.UserContext{ID: "", Attributes: map[string]interface{}(nil)},
		DecisionInfo: map[string]interface{}{"feature": map[string]interface{}{"featureEnabled": false, "featureKey": "feature_key", "source": FeatureTest,
			"sourceInfo": map[string]string{"experimentKey": "", "variationKey": ""}}}}
	assert.NotNil(t, decision)
	assert.Equal(t, expectedDecision, decision)
}

func TestFeatureNotificationWithVariables(t *testing.T) {
	featureDecision := &FeatureDecision{Source: FeatureTest, Experiment: entities.Experiment{}, Variation: &entities.Variation{}}

	featureKey := "feature_key"
	userContext := &entities.UserContext{}
	variableMap := map[string]interface{}{
		"variable_key":   "var_key",
		"variable_type":  entities.String,
		"variable_value": "some_value",
	}
	decision := FeatureNotificationWithVariables(featureKey, featureDecision, userContext, variableMap)

	expectedDecision := &notification.DecisionNotification{Type: "feature", UserContext: entities.UserContext{ID: "", Attributes: map[string]interface{}(nil)},
		DecisionInfo: map[string]interface{}{"feature": map[string]interface{}{"featureEnabled": false, "featureKey": "feature_key", "source": FeatureTest,
			"sourceInfo": map[string]string{"experimentKey": "", "variationKey": ""}, "variable_key": "var_key", "variable_type": entities.String, "variable_value": "some_value"}}}
	assert.NotNil(t, decision)
	assert.Equal(t, expectedDecision, decision)
}
