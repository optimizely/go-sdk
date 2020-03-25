package bucketer

import (
	"github.com/optimizely/go-sdk/pkg/logging"
	"testing"

	"github.com/optimizely/go-sdk/pkg/decision/reasons"

	"github.com/optimizely/go-sdk/pkg/entities"
	"github.com/stretchr/testify/assert"
)

func TestBucketExclusionGroups(t *testing.T) {
	experiment1 := entities.Experiment{
		ID:  "1886780721",
		Key: "experiment_1",
		Variations: map[string]entities.Variation{
			"22222": entities.Variation{ID: "22222", Key: "exp_1_var_1"},
			"22223": entities.Variation{ID: "22223", Key: "exp_1_var_2"},
		},
		TrafficAllocation: []entities.Range{
			entities.Range{EntityID: "22222", EndOfRange: 4999},
			entities.Range{EntityID: "22223", EndOfRange: 10000},
		},
		GroupID: "1886780722",
	}
	experiment2 := entities.Experiment{
		ID:  "1886780723",
		Key: "experiment_2",
		Variations: map[string]entities.Variation{
			"22224": entities.Variation{ID: "22224", Key: "exp_2_var_1"},
			"22225": entities.Variation{ID: "22225", Key: "exp_2_var_2"},
		},
		TrafficAllocation: []entities.Range{
			entities.Range{EntityID: "22224", EndOfRange: 4999},
			entities.Range{EntityID: "22225", EndOfRange: 10000},
		},
		GroupID: "1886780722",
	}

	exclusionGroup := entities.Group{
		ID:     "1886780722",
		Policy: "random",
		TrafficAllocation: []entities.Range{
			entities.Range{EntityID: "1886780721", EndOfRange: 2500},
			entities.Range{EntityID: "1886780723", EndOfRange: 5000},
		},
	}

	bucketer := NewMurmurhashExperimentBucketer(logging.GetLogger("","TestBucketExclusionGroups" ), DefaultHashSeed)
	// ppid2 + 1886780722 (groupId) will generate bucket value of 2434 which maps to experiment 1
	bucketedVariation, reason, _ := bucketer.Bucket("ppid2", experiment1, exclusionGroup)
	assert.Equal(t, experiment1.Variations["22222"], *bucketedVariation)
	assert.Equal(t, reasons.BucketedIntoVariation, reason)
	// since the bucket value maps to experiment 1, the user will not be bucketed for experiment 2
	bucketedVariation, reason, _ = bucketer.Bucket("ppid2", experiment2, exclusionGroup)
	assert.Nil(t, bucketedVariation)
	assert.Equal(t, reasons.NotBucketedIntoVariation, reason)
}
