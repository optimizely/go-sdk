package optimizely

// GetOptimizelyClient returns a client that can be used to interface
// with Optimizely
func GetOptimizelyClient(account_id string) (*OptimizelyClient, error) {
	client := OptimizelyClient{}
	project_config, err := FetchProjectConfig(account_id)
	if err != nil {
		return &client, err
	}
	client.account_id = account_id
	client.project_config = project_config
	return &client, nil

}

// Track tracks a conversion event for a user_id
// Logs the conversion
// event_key: goal key representing the event which needs to be recorded
// user_id: ID for user.
// attributes: Dict representing visitor attributes and values which need to be recorded.
// event_value: Value associated with the event. Can be used to represent revenue in cents.
func (client *OptimizelyClient) Track(
	event_key string,
	user_id string,
	attributes []AttributeEntity,
	event_value string) {

}

// Activate buckets visitor and sends impression event to Optimizely
// Activate Logs the impression
// experiment_key: experiment which needs to be activated
// user_id: ID for user
// attributes: optional list representing visitor attributes and values
func (client *OptimizelyClient) Activate(experiment_key string, user_id string, attributes []AttributeEntity) {
	var valid_experiment = false
	for i := 0; i < len(client.project_config.Experiments); i++ {
		if client.project_config.Experiments[i].Key == experiment_key {
			if ExperimentIsRunning(client.project_config.Experiments[i]) {
				valid_experiment = true
			}
		}
	}

	if !valid_experiment {
		return
	}

	variation_id := ""
	impression_event := CreateImpressionEvent(
		client, experiment_key, variation_id, user_id, attributes)
	impression_event_url := GetUrlForImpressionEvent(client.project_config.ProjectId)

}

// GetVariation gets the variation key where the visitor will be bucketed
// Experiment_key: experiment which needs to be activated
// User_id: ID for user
// Returns: variation key representing the variation the user will be
// bucketed into
func (client *OptimizelyClient) GetVariation(experiment_key string, user_id string) string {
	variation_id := client.Bucket(experiment_key, user_id)
	variation_key := GetVariationKeyFromId(experiment_key, variation_id, client.project_config.Experiments)
	return variation_key
}
