package bucketer

import (
	"fmt"
	"testing"

	"github.com/optimizely/go-sdk/pkg/decision/reasons"

	"github.com/optimizely/go-sdk/pkg/entities"
	"github.com/stretchr/testify/assert"
)

func TestGenerateBucketValue(t *testing.T) {
	bucketer := NewMurmurhashBucketer(DefaultHashSeed)

	// copied from unit tests in the other SDKs
	experimentID := "1886780721"
	experimentID2 := "1886780722"
	bucketingKey1 := fmt.Sprintf("%s%s", "ppid1", experimentID)
	bucketingKey2 := fmt.Sprintf("%s%s", "ppid2", experimentID)
	bucketingKey3 := fmt.Sprintf("%s%s", "ppid2", experimentID2)
	bucketingKey4 := fmt.Sprintf("%s%s", "ppid3", experimentID)

	assert.Equal(t, 5254, bucketer.generateBucketValue(bucketingKey1))
	assert.Equal(t, 4299, bucketer.generateBucketValue(bucketingKey2))
	assert.Equal(t, 2434, bucketer.generateBucketValue(bucketingKey3))
	assert.Equal(t, 5439, bucketer.generateBucketValue(bucketingKey4))
}

func TestBucketToEntity(t *testing.T) {
	bucketer := NewMurmurhashBucketer(DefaultHashSeed)

	experimentID := "1886780721"
	experimentID2 := "1886780722"

	// bucket value 5254
	bucketingKey1 := fmt.Sprintf("%s%s", "ppid1", experimentID)
	// bucket value 4299
	bucketingKey2 := fmt.Sprintf("%s%s", "ppid2", experimentID)
	// bucket value 2434
	bucketingKey3 := fmt.Sprintf("%s%s", "ppid2", experimentID2)
	// bucket value 5439
	bucketingKey4 := fmt.Sprintf("%s%s", "ppid3", experimentID)

	variation1 := "1234567123"
	variation2 := "5949300123"
	trafficAlloc := []entities.Range{
		entities.Range{
			EntityID:   "",
			EndOfRange: 2500,
		},
		entities.Range{
			EntityID:   variation1,
			EndOfRange: 4999,
		},
		entities.Range{
			EntityID:   variation2,
			EndOfRange: 5399,
		},
	}

	assert.Equal(t, variation2, bucketer.bucketToEntity(bucketingKey1, trafficAlloc))
	assert.Equal(t, variation1, bucketer.bucketToEntity(bucketingKey2, trafficAlloc))

	// bucket to empty variation range
	assert.Equal(t, "", bucketer.bucketToEntity(bucketingKey3, trafficAlloc))

	// bucket outside of range (not in experiment)
	assert.Equal(t, "", bucketer.bucketToEntity(bucketingKey4, trafficAlloc))
}

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

	bucketer := NewMurmurhashBucketer(DefaultHashSeed)
	// ppid2 + 1886780722 (groupId) will generate bucket value of 2434 which maps to experiment 1
	bucketedVariation, reason, _ := bucketer.Bucket("ppid2", experiment1, exclusionGroup)
	assert.Equal(t, experiment1.Variations["22222"], *bucketedVariation)
	assert.Equal(t, reasons.BucketedIntoVariation, reason)
	// since the bucket value maps to experiment 1, the user will not be bucketed for experiment 2
	bucketedVariation, reason, _ = bucketer.Bucket("ppid2", experiment2, exclusionGroup)
	assert.Nil(t, bucketedVariation)
	assert.Equal(t, reasons.NotBucketedIntoVariation, reason)
}
