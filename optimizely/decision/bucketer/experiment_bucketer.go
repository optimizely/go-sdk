package bucketer

import (
	"math"

	"github.com/optimizely/go-sdk/optimizely/entities"
	"github.com/twmb/murmur3"
)

var maxHashValue = float32(math.Pow(2, 32))

// DefaultHashSeed is the hash seed to use for murmurhash
const DefaultHashSeed = 1
const maxTrafficValue = 10000

// ExperimentBucketer buckets the user
type ExperimentBucketer struct {
	hashSeed uint32
}

// NewExperimentBucketer returns a new instance of the experiment bucketer
func NewExperimentBucketer(hashSeed uint32) *ExperimentBucketer {
	return &ExperimentBucketer{
		hashSeed: hashSeed,
	}
}

// Bucket buckets the user into the given experiment
func (b *ExperimentBucketer) Bucket(bucketingID string, experiment entities.Experiment, group entities.Group) entities.Variation {
	if experiment.GroupID != "" && group.Policy == "random" {
		bucketKey := bucketingID + group.ID
		bucketedExperimentID := b.bucketToEntity(bucketKey, group.TrafficAllocation)
		if bucketedExperimentID == "" || bucketedExperimentID != experiment.ID {
			// User is not bucketed into an experiment in the exclusion group, return an empty variation
			return entities.Variation{}
		}
	}

	bucketKey := bucketingID + experiment.ID
	bucketedVariationID := b.bucketToEntity(bucketKey, experiment.TrafficAllocation)
	if bucketedVariationID == "" {
		// User is not bucketed into a variation in the experiment, return an empty variation
		return entities.Variation{}
	}

	if variation, ok := experiment.Variations[bucketedVariationID]; ok {
		return variation
	}

	return entities.Variation{}
}

func (b *ExperimentBucketer) bucketToEntity(bucketKey string, trafficAllocations []entities.Range) (entityID string) {
	// TODO(mng): return log message re: bucket value
	bucketValue := b.generateBucketValue(bucketKey)
	var currentEndOfRange int
	for _, trafficAllocationRange := range trafficAllocations {
		currentEndOfRange = trafficAllocationRange.EndOfRange
		if bucketValue < currentEndOfRange {
			return trafficAllocationRange.EntityID
		}
	}

	return ""
}

func (b *ExperimentBucketer) generateBucketValue(bucketingKey string) int {
	hasher := murmur3.SeedNew32(b.hashSeed)
	hasher.Write([]byte(bucketingKey))
	hashCode := hasher.Sum32()
	ratio := float32(hashCode) / maxHashValue
	return int(ratio * maxTrafficValue)
}
