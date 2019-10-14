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
