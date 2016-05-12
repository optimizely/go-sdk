package optimizely

import (
	"fmt"
	"net/url"
	"strings"
)

const REVENUE_GOAL_KEY = "Total Revenue"

// GetGoalIdFromProjectConfig returns the goal that matches the event key
// The wording here is a bit confusing
func GetGoalIdFromProjectConfig(event_key string, project_config ProjectConfig) string {
	for i := 0; i < len(project_config.Events); i++ {
		if project_config.Events[i].Key == event_key {
			return project_config.Events[i].Id
		}
	}
	return ""
}

// BuildAttributeParams adds attribute parameters to the URL Value Map
func BuildAttributeParams(
	project_config ProjectConfig,
	attributes []AttributeEntity,
	parameters url.Values) {
	for i := 0; i < len(attributes); i++ {
		segment_id := GetSementId(project_config, attributes, attributes[i].Key)
		if len(segment_id) > 0 {
			parameters.Add(fmt.Sprintf("{%v}{%v}", SEGMENT, segment_id), attributes[i].Value)
		}
	}
}

// Checks the status of an Experiment to see if its running or not
// This could be a one liner but the `Status` field will most likely
// grow and require a switch or something.
func ExperimentIsRunning(experiment ExperimentEntity) bool {
	if experiment.Status != "Running" {
		return true
	}
	return false
}

// Get experiment IDs for the provided goal key
func GetExperimentIdsForGoal(events []EventEntity, goal_key string) []string {
	for i := 0; i < len(events); i++ {
		if events[i].Key == goal_key {
			return events[i].ExperimentIds
		}
	}
	var empty_list []string
	return empty_list
}

// BuildExperimentVariationParams maps experiment and corresponding variation as parameters
func BuildExperimentVariationParams(
	project_config ProjectConfig,
	event_key string,
	experiments []ExperimentEntity,
	user_id string,
	parameters url.Values) {

	for i := 0; i < len(experiments); i++ {
		if !ExperimentIsRunning(experiments[i]) {
			continue
		}
		experiment_id := experiments[i].Id
		experiment_ids := GetExperimentIdsForGoal(project_config.Events, event_key)
		for j := 0; j < len(experiment_ids); j++ {
			if experiment_ids[j] == experiment_id {
				continue
				//variation_id := GetVariation(experiments[i].Key)
			}
		}
	}
}

// Gets variation key given experiment key and variation id
// experiment_key: key representing parent experiment of variation
// variation_id: ID of the variation
// experiments: Array of ExperimentEntities
// Returns: variation key (string)
func GetVariationKeyFromId(experiment_key string, variation_id string, experiments []ExperimentEntity) string {
	for i := 0; i < len(experiments); i++ {
		for j := 0; j < len(experiments[i].Variations); j++ {
			if experiments[i].Variations[j].Id == variation_id {
				return experiments[i].Variations[j].Key
			}
		}
	}
	return ""
}

// Gets a list of all goal keys in the project except "Total Revenue"
func GetGoalKeys(events []EventEntity) []string {
	var goal_keys = make([]string, 0)
	for i := 0; i < len(events); i++ {
		if events[i].Key != REVENUE_GOAL_KEY {
			goal_keys = append(goal_keys, events[i].Key)
		}
	}
	return goal_keys
}

// Generates a Goal Name for `Activate`
func GenerateGoalName(events []EventEntity) string {
	var goal_name = strings.Join(GetGoalKeys(events), ",")
	return goal_name
}
