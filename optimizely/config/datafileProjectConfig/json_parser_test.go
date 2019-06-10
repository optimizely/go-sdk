package datafileProjectConfig

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseDatafilePasses(t *testing.T) {
	testFeatureKey := "feature_test_1"
	testFeatureID := "feature_id_123"
	datafileString := fmt.Sprintf(`{
		"projectId": "1337",
		"featureFlags": [
			{
				"key": "%s",
				"id" : "%s"
			}
		]
	}`, testFeatureKey, testFeatureID)

	rawDatafile := []byte(datafileString)
	parser := DatafileJSONParser{}
	projectConfig, err := parser.Parse(rawDatafile)
	if err != nil {
		assert.Fail(t, err.Error())
	}
	feature, err := projectConfig.GetFeatureByKey("feature_test_1")
	if err != nil {
		assert.Fail(t, err.Error())
	}

	assert.Equal(t, "feature_id_123", feature.ID)
}
