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

package support

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/google/go-cmp/cmp"
	"github.com/optimizely/go-sdk/pkg"

	"gopkg.in/yaml.v3"
)

func sortArrayofMaps(array []map[string]interface{}, sortKey string) []map[string]interface{} {
	sort.Slice(array, func(i, j int) bool {
		// TODO: make it more generic but for now only expecting string comparison
		if val1, ok := array[i][sortKey].(string); ok {
			if val2, ok1 := array[j][sortKey].(string); ok1 {
				if strings.HasPrefix(val1, "$") {
					return false
				}
				return val1 < val2
			}
		}
		// if both are not string let it go.
		return false
	})
	return array
}

func parseYamlArray(s string, config pkg.ProjectConfig) ([]map[string]interface{}, error) {
	var array []map[string]interface{}
	parsedString := parseTemplate(s, config)
	if err := yaml.Unmarshal([]byte(parsedString), &array); err != nil {
		return nil, err
	}
	return array, nil
}

func parseTemplate(s string, config pkg.ProjectConfig) string {

	parsedString := strings.Replace(s, "{{datafile.projectId}}", config.GetProjectID(), -1)

	var replaceExperimentID func()
	replaceExperimentID = func() {
		re := regexp.MustCompile(`{{#expId}}(\S+)*{{/expId}}`)
		matches := re.FindStringSubmatch(parsedString)
		if len(matches) > 1 {
			if exp, err := config.GetExperimentByKey(matches[1]); err == nil {
				parsedString = strings.Replace(parsedString, matches[0], exp.ID, -1)
				replaceExperimentID()
			}
		}
	}

	var replaceEventID func()
	replaceEventID = func() {
		re := regexp.MustCompile(`{{#eventId}}(\S+)*{{/eventId}}`)
		matches := re.FindStringSubmatch(parsedString)
		if len(matches) > 1 {
			if event, err := config.GetEventByKey(matches[1]); err == nil {
				parsedString = strings.Replace(parsedString, matches[0], event.ID, -1)
				replaceEventID()
			}
		}
	}

	var replaceCampaignID func()
	replaceCampaignID = func() {
		re := regexp.MustCompile(`{{#expCampaignId}}(\S+)*{{/expCampaignId}}`)
		matches := re.FindStringSubmatch(parsedString)
		if len(matches) > 1 {
			if exp, err := config.GetExperimentByKey(matches[1]); err == nil {
				parsedString = strings.Replace(parsedString, matches[0], exp.LayerID, -1)
				replaceCampaignID()
			}
		}
	}

	var replaceVariableID func()
	replaceVariableID = func() {
		re := regexp.MustCompile(`{{#varId}}(\S+)*{{/varId}}`)
		matches := re.FindStringSubmatch(parsedString)
		if len(matches) > 1 {
			expVarKey := strings.Split(matches[1], ".")
			if exp, err := config.GetExperimentByKey(expVarKey[0]); err == nil {
				if variation, ok := exp.VariationsKeyMap[expVarKey[1]]; ok {
					parsedString = strings.Replace(parsedString, matches[0], variation.ID, -1)
					replaceVariableID()
				}
			}
		}
	}

	replaceExperimentID()
	replaceEventID()
	replaceCampaignID()
	replaceVariableID()
	return parsedString
}

func getErrorWithDiff(a, b interface{}, message string) error {
	diff := cmp.Diff(a, b)
	return fmt.Errorf("%s:\n--expected\n++actual\n%s", message, diff)
}

// https://stackoverflow.com/a/36000696/4849178
// Unordered string slice comparison
func compareStringSlice(x, y []string) bool {
	if len(x) != len(y) {
		return false
	}
	// create a map of string -> int
	diff := make(map[string]int, len(x))
	for _, _x := range x {
		// 0 value for int is 0, so just increment a counter for the string
		diff[_x]++
	}
	for _, _y := range y {
		// If the string _y is not in diff bail out early
		if _, ok := diff[_y]; !ok {
			return false
		}
		diff[_y]--
		if diff[_y] == 0 {
			delete(diff, _y)
		}
	}
	if len(diff) == 0 {
		return true
	}
	return false
}
