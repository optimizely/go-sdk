package optimizely

import (
	"fmt"
	"hash/fnv"
	"math"
)

var BUCKETING_ID_TEMPLATE = "{%v}{%v}"
var MAX_HASH_VALUE = math.Exp2(32)

// generate_bucket_value generates an unit32 value for a bucket
func generate_bucket_value(bucketing_id string) uint32 {
	h := fnv.New32a()
	bucket_slice := []byte(bucketing_id)
	h.Write(bucket_slice)
	return h.Sum32()
}

// Given an experiment key, returns the traffic allocation within that experiment
// experiment_key: key of experiment
// experiments: array of ExperimentEntities
// Returns: pointer to array of TrafficAllocationEntities
func get_traffic_allocations(experiment_key string, experiments []ExperimentEntity) []TrafficAllocationEntity {
	for i := 0; i < len(experiments); i++ {
		if experiments[i].Key == experiment_key { // Id or Key?
			return experiments[i].TrafficAllocation
		}
	}
	return nil
}

// Bucket determines for a given experiment key and bucketing ID determines ID of variation to be shown to visitor.
// experiment_key: Key representing experiment for which visitor is to be bucketed.
// user_id: ID for user.
// Returns: Variation ID for variation in which the visitor with ID user_id will be put in
// If no variation, return empty string (TODO: move to error)
func (client *OptimizelyClient) Bucket(experiment_key string, user_id string) string {
	var experiment_id = ""
	for j := 0; j < len(client.project_config.Experiments); j++ {
		if client.project_config.Experiments[j].Key == experiment_id {
			experiment_id = client.project_config.Experiments[j].Id
		}
	}
	var bucketing_id = fmt.Sprintf(BUCKETING_ID_TEMPLATE, user_id, experiment_id)
	var bucketing_number = generate_bucket_value(bucketing_id)
	var traffic_allocations = get_traffic_allocations(experiment_key, client.project_config.Experiments)
	for i := 0; i < len(traffic_allocations); i++ {
		current_end_of_range := traffic_allocations[i].EndOfRange
		if bucketing_number <= uint32(current_end_of_range) {
			return traffic_allocations[i].EntityId
		}
	}
	return ""

}
