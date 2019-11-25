package bucketer

import (
	"fmt"
	"testing"

	"github.com/optimizely/go-sdk/pkg/entities"
	"github.com/stretchr/testify/assert"
)

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

	assert.Equal(t, variation2, bucketer.BucketToEntity(bucketingKey1, trafficAlloc))
	assert.Equal(t, variation1, bucketer.BucketToEntity(bucketingKey2, trafficAlloc))

	// bucket to empty variation range
	assert.Equal(t, "", bucketer.BucketToEntity(bucketingKey3, trafficAlloc))

	// bucket outside of range (not in experiment)
	assert.Equal(t, "", bucketer.BucketToEntity(bucketingKey4, trafficAlloc))
}

func TestGenerateBucketValue(t *testing.T) {
	bucketer := NewMurmurhashBucketer(DefaultHashSeed)

	// copied from unit tests in the other SDKs
	experimentID := "1886780721"
	experimentID2 := "1886780722"
	bucketingKey1 := fmt.Sprintf("%s%s", "ppid1", experimentID)
	bucketingKey2 := fmt.Sprintf("%s%s", "ppid2", experimentID)
	bucketingKey3 := fmt.Sprintf("%s%s", "ppid2", experimentID2)
	bucketingKey4 := fmt.Sprintf("%s%s", "ppid3", experimentID)

	assert.Equal(t, 5254, bucketer.Generate(bucketingKey1))
	assert.Equal(t, 4299, bucketer.Generate(bucketingKey2))
	assert.Equal(t, 2434, bucketer.Generate(bucketingKey3))
	assert.Equal(t, 5439, bucketer.Generate(bucketingKey4))
}
