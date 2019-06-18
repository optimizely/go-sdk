package client

import (
	"fmt"
	"testing"
)

const datafileStr = `{"version": "4", "rollouts": [{"experiments": [{"status": "Not started", "audienceIds": [], "variations": [{"variables": [], "id": "13158010388", "key": "13158010388", "featureEnabled": false}], "id": "13180720413", "key": "13180720413", "layerId": "13177380550", "trafficAllocation": [{"entityId": "13158010388", "endOfRange": 0}], "forcedVariations": {}}], "id": "13177380550"}], "typedAudiences": [{"id": "15227890301", "conditions": ["and", ["or", ["or", {"value": "California", "type": "custom_attribute", "name": "country", "match": "exact"}]], ["or", ["or", {"value": true, "type": "custom_attribute", "name": "likes_cupcake", "match": "exact"}], ["or", {"value": true, "type": "custom_attribute", "name": "likes_donuts", "match": "exact"}]]], "name": "California Sweet Tooths"}, {"id": "15245420535", "conditions": ["and", ["or", ["or", {"value": "Russia", "type": "custom_attribute", "name": "country", "match": "exact"}]], ["or", ["or", {"value": true, "type": "custom_attribute", "name": "likes_donuts", "match": "exact"}]]], "name": "Russian Donut Eaters"}], "anonymizeIP": false, "projectId": "13175460377", "variables": [], "featureFlags": [{"experimentIds": ["13198330290"], "rolloutId": "13177380550", "variables": [], "id": "13161700299", "key": "binary_feature"}], "experiments": [{"status": "Launched", "audienceConditions": ["or", "15245420535", "15227890301"], "audienceIds": ["15245420535", "15227890301"], "variations": [{"variables": [], "id": "13182560234", "key": "variation_1", "featureEnabled": false}, {"variables": [], "id": "13165740399", "key": "variation_2", "featureEnabled": true}], "id": "13198330290", "key": "binary_feature_test", "layerId": "13194270270", "trafficAllocation": [{"entityId": "13165740399", "endOfRange": 5000}, {"entityId": "13182560234", "endOfRange": 10000}], "forcedVariations": {}}], "audiences": [{"id": "15227890301", "conditions": "[\"or\", {\"match\": \"exact\", \"name\": \"$opt_dummy_attribute\", \"type\": \"custom_attribute\", \"value\": \"$opt_dummy_value\"}]", "name": "California Sweet Tooths"}, {"id": "15245420535", "conditions": "[\"or\", {\"match\": \"exact\", \"name\": \"$opt_dummy_attribute\", \"type\": \"custom_attribute\", \"value\": \"$opt_dummy_value\"}]", "name": "Russian Donut Eaters"}, {"conditions": "[\"or\", {\"match\": \"exact\", \"name\": \"$opt_dummy_attribute\", \"type\": \"custom_attribute\", \"value\": \"$opt_dummy_value\"}]", "id": "$opt_dummy_audience", "name": "Optimizely-Generated Audience for Backwards Compatibility"}], "groups": [], "attributes": [{"id": "15249570426", "key": "country"}, {"id": "15251490714", "key": "likes_cupcake"}, {"id": "15255450342", "key": "likes_donuts"}], "botFiltering": false, "accountId": "2002030909", "events": [{"experimentIds": ["13198330290"], "id": "13204200069", "key": "my_precious_event"}], "revision": "49"}`

func BenchmarkIsFeatureEnabled(b *testing.B) {
	optimizelyFactory := &OptimizelyFactory{
		Datafile: []byte(datafileStr),
	}

	for i := 0; i < b.N; i++ {
		user := fmt.Sprintf("user_%d", i)
		client := optimizelyFactory.Client()
		client.IsFeatureEnabled("binary_feature", user, nil)
	}
}
