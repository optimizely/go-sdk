package support

import (
	"encoding/json"
	"regexp"
	"strings"

	"github.com/optimizely/go-sdk/pkg"

	"github.com/optimizely/go-sdk/pkg/event"
	"gopkg.in/yaml.v3"
)

func getDispatchedEventsFromYaml(s string, config pkg.ProjectConfig) ([]event.Batch, error) {
	var eventsArray []map[string]interface{}
	parsedString := parseTemplate(s, config)
	if err := yaml.Unmarshal([]byte(parsedString), &eventsArray); err != nil {
		return nil, err
	}
	jsonString, err := json.Marshal(eventsArray)
	if err != nil {
		return nil, err
	}
	requestedBatchEvents := []event.Batch{}
	if err := json.Unmarshal([]byte(jsonString), &requestedBatchEvents); err != nil {
		return nil, err
	}
	return requestedBatchEvents, nil
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
