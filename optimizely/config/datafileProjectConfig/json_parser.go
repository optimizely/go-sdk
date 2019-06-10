package datafileProjectConfig

import (
	"encoding/json"

	"github.com/optimizely/go-sdk/optimizely/entities"
)

// FeatureFlag represents a FeatureFlag object from the Optimizely datafile
type FeatureFlag struct {
	ID            string   `json:"id"`
	RolloutID     string   `json:"rolloutId"`
	Key           string   `json:"key"`
	ExperimentIDs []string `json:"experimentIds"`
	Variables     []string `json:"variables"`
}

// Datafile represents the datafile we get from Optimizely
type Datafile struct {
	Version      string        `json:"version"`
	AnonymizeIP  bool          `json:"anonymizeIP"`
	ProjectID    string        `json:"projectId"`
	Variables    []string      `json:"variables"`
	FeatureFlags []FeatureFlag `json:"featureFlags"`
	BotFiltering bool          `json:"botFiltering"`
	AccountID    string        `json:"accountId"`
	Revision     string        `json:"revision"`
}

// DatafileJSONParser implements the DatafileParser interface and parses a JSON-based datafile into a DatafileProjectConfig
type DatafileJSONParser struct {
}

// Parse parses the json datafile
func (parser DatafileJSONParser) Parse(jsonDatafile []byte) (*DatafileProjectConfig, error) {
	datafile := Datafile{}
	projectConfig := &DatafileProjectConfig{}

	err := json.Unmarshal(jsonDatafile, &datafile)
	if err != nil {
		// @TODO(mng): return error
	}

	// convert the Datafile into a ProjectConfig
	projectConfig.features = mapFeatureFlags(datafile.FeatureFlags)
	return projectConfig, nil
}

func mapFeatureFlags(featureFlags []FeatureFlag) map[string]entities.Feature {
	featureMap := make(map[string]entities.Feature)
	for _, featureFlag := range featureFlags {
		featureMap[featureFlag.Key] = entities.Feature{
			Key: featureFlag.Key,
			ID:  featureFlag.ID,
		}
	}
	return featureMap
}
