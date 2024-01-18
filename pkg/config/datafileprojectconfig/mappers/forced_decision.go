/****************************************************************************
 * Copyright 2021, Optimizely, Inc. and contributors                        *
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

// Package mappers ...
package mappers

import (
	"github.com/optimizely/go-sdk/v2/pkg/entities"
)

// MapFlagVariations all variations for each flag
// datafile does not contain a separate entity for this
// we collect variations used in each rule (experiment rules and delivery rules)
func MapFlagVariations(featureMap map[string]entities.Feature) (flagVariationsMap map[string][]entities.Variation) {
	flagVariationsMap = map[string][]entities.Variation{}
	for _, flag := range featureMap {
		// To track if variation was already added to list
		variationsTracker := map[string]bool{}
		variations := []entities.Variation{}

		allRulesForFlag := []entities.Experiment{}
		allRulesForFlag = append(allRulesForFlag, flag.FeatureExperiments...)
		allRulesForFlag = append(allRulesForFlag, flag.Rollout.Experiments...)

		for _, rule := range allRulesForFlag {
			for _, variation := range rule.Variations {
				if !variationsTracker[variation.ID] {
					variationsTracker[variation.ID] = true
					variations = append(variations, variation)
				}
			}
		}
		flagVariationsMap[flag.Key] = variations
	}
	return flagVariationsMap
}
