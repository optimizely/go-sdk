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
	"regexp"
	"strings"

	"github.com/optimizely/go-sdk/pkg"

	"gopkg.in/yaml.v3"
)

func getDispatchedEventsMapFromYaml(s string, config pkg.ProjectConfig) ([]map[string]interface{}, error) {
	var eventsArray []map[string]interface{}
	parsedString := parseTemplate(s, config)
	if err := yaml.Unmarshal([]byte(parsedString), &eventsArray); err != nil {
		return nil, err
	}
	return eventsArray, nil
}

func parseTemplate(s string, config pkg.ProjectConfig) string {

	parsedString := strings.Replace(s, "{{datafile.projectId}}", config.GetProjectID(), -1)

	re := regexp.MustCompile(`{{#expId}}(\S+)*{{/expId}}`)
	matches := re.FindStringSubmatch(parsedString)
	if len(matches) > 1 {
		if exp, err := config.GetExperimentByKey(matches[1]); err == nil {
			parsedString = strings.Replace(parsedString, matches[0], exp.ID, -1)
		}
	}

	re = regexp.MustCompile(`{{#eventId}}(\S+)*{{/eventId}}`)
	matches = re.FindStringSubmatch(parsedString)
	if len(matches) > 1 {
		if event, err := config.GetEventByKey(matches[1]); err == nil {
			parsedString = strings.Replace(parsedString, matches[0], event.ID, -1)
		}
	}

	re = regexp.MustCompile(`{{#expCampaignId}}(\S+)*{{/expCampaignId}}`)
	matches = re.FindStringSubmatch(parsedString)
	if len(matches) > 1 {
		if exp, err := config.GetExperimentByKey(matches[1]); err == nil {
			parsedString = strings.Replace(parsedString, matches[0], exp.LayerID, -1)
		}
	}

	re = regexp.MustCompile(`{{#varId}}(\S+)*{{/varId}}`)
	matches = re.FindStringSubmatch(parsedString)
	if len(matches) > 1 {
		expVarKey := strings.Split(matches[1], ".")
		if exp, err := config.GetExperimentByKey(expVarKey[0]); err == nil {
			for _, variation := range exp.Variations {
				if variation.Key == expVarKey[1] {
					parsedString = strings.Replace(parsedString, matches[0], variation.ID, -1)
					break
				}
			}
		}
	}
	return parsedString
}
