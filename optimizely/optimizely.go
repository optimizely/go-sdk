package optimizely

import (
	"fmt"
	"net/url"
)

// Track tracks a conversion event for a user_id
// Logs the conversion
// event_key: goal key representing the event which needs to be recorded
// user_id: ID for user.
// attributes: Dict representing visitor attributes and values which need to be recorded.
// event_value: Value associated with the event. Can be used to represent revenue in cents.
func Track(
	event_key string,
	user_id string,
	attributes []AttributeEntity,
	event_value string,
	project_config ProjectConfig) {

	var Url *url.URL
	Url, err := url.Parse("http://www.example.com")
	if err != nil {
		panic("boom")
	}

	end_user_id := fmt.Sprintf(END_USER_ID_TEMPLATE, user_id)
	goal_id := GetGoalIdFromProjectConfig(event_key, project_config)

	// build string to make GET request with
	parameters := url.Values{}
	parameters.Add(ACCOUNT_ID, project_config.AccountId)
	parameters.Add(PROJECT_ID, project_config.ProjectId)
	parameters.Add(GOAL_NAME, event_key)
	parameters.Add(GOAL_ID, goal_id)
	parameters.Add(END_USER_ID, end_user_id)

	// Set experiment and corresponding variation
	BuildExperimentVariationParams(
		project_config, event_key, project_config.Experiments, user_id, parameters)

	// Set attribute params if any
	if len(attributes) > 0 {
		BuildAttributeParams(project_config, attributes, parameters)
	}

	// Set event_value if set and also append the revenue goal ID
	if len(event_value) != 0 {
		parameters.Add(REVENUE, event_value)
		//parameters.Add(GOAL_ID, fmt.Sprintf("{%v},{%v}", goal_id, GetRevenueGoalFromProjectConfig())
	}

	// Dispatch event
	Url.RawQuery = parameters.Encode()
	tracking_url := Url.String()
	fmt.Print(tracking_url)

}

// Activate buckets visitor and sends impression event to Optimizely
// Activate Logs the impression
// experiment_key: experiment which needs to be activated
// user_id: ID for user
// attributes: optional list representing visitor attributes and values
func Activate(experiment_key string, user_id string, attributes []AttributeEntity) {

}

// GetVariation gets the variation where the visitor will be bucketed
// Experiment_key: experiment which needs to be activated
// User_id: ID for user
func GetVariation(experient_key string, user_id string) {

}
