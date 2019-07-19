package bucketer

import (
	"fmt"
	"math"

	"github.com/optimizely/go-sdk/optimizely/entities"
	"github.com/optimizely/go-sdk/optimizely/logging"
	"github.com/twmb/murmur3"
)

var logger = logging.GetLogger("ExperimentBucketer")
var maxHashValue = float32(math.Pow(2, 32))

// DefaultHashSeed is the hash seed to use for murmurhash
const DefaultHashSeed = 1
const maxTrafficValue = 10000

// ExperimentBucketer is used to bucket the user into a particular entity in the experiment's traffic alloc range
type ExperimentBucketer interface {
	Bucket(bucketingID string, experiment entities.Experiment, group entities.Group) (entities.Variation, string)
}

// MurmurhashBucketer buckets the user using the mmh3 algorightm
type MurmurhashBucketer struct {
	hashSeed uint32
}

// NewMurmurhashBucketer returns a new instance of the experiment bucketer
func NewMurmurhashBucketer(hashSeed uint32) *MurmurhashBucketer {
	return &MurmurhashBucketer{
		hashSeed: hashSeed,
	}
}

// Bucket buckets the user into the given experiment
func (b MurmurhashBucketer) Bucket(bucketingID string, experiment entities.Experiment, group entities.Group) (entities.Variation, string) {
	if experiment.GroupID != "" && group.Policy == "random" {
		bucketKey := bucketingID + group.ID
		bucketedExperimentID := b.bucketToEntity(bucketKey, group.TrafficAllocation)
		if bucketedExperimentID == "" || bucketedExperimentID != experiment.ID {
			// User is not bucketed into an experiment in the exclusion group, return an empty variation
			reason := fmt.Sprintf(`User "%s" is not in any experiment of group "%s"`, bucketingID, group.ID)
			return entities.Variation{}, reason
		}
	}

	bucketKey := bucketingID + experiment.ID
	bucketedVariationID := b.bucketToEntity(bucketKey, experiment.TrafficAllocation)
	if bucketedVariationID == "" {
		// User is not bucketed into a variation in the experiment, return an empty variation
		reason := fmt.Sprintf(`User "%s" is not in any variation of experiment "%s"`, bucketingID, experiment.Key)
		return entities.Variation{}, reason
	}

	if variation, ok := experiment.Variations[bucketedVariationID]; ok {
		return variation, fmt.Sprintf(`User "%s" is bucketed in variation "%s" of experiment "%s"`, bucketingID, variation.Key, experiment.Key)
	}

	reason := fmt.Sprintf(`Variation with ID "%s" not found for experiment "%s"`, bucketedVariationID, experiment.Key)
	return entities.Variation{}, reason
}

func (b MurmurhashBucketer) bucketToEntity(bucketKey string, trafficAllocations []entities.Range) (entityID string) {
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

func (b MurmurhashBucketer) generateBucketValue(bucketingKey string) int {
	hasher := murmur3.SeedNew32(b.hashSeed)
	hasher.Write([]byte(bucketingKey))
	hashCode := hasher.Sum32()
	ratio := float32(hashCode) / maxHashValue
	return int(ratio * maxTrafficValue)
}
