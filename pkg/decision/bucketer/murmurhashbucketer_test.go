package bucketer

import (
	"fmt"
	"testing"

	"github.com/optimizely/go-sdk/v2/pkg/logging"

	"github.com/optimizely/go-sdk/v2/pkg/entities"
	"github.com/stretchr/testify/assert"
)

func TestBucketToEntity(t *testing.T) {
	bucketer := NewMurmurhashBucketer(logging.GetLogger("", "TestBucketToEntity"), DefaultHashSeed)

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
	bucketer := NewMurmurhashBucketer(logging.GetLogger("", "TestGenerateBucketValue"), DefaultHashSeed)

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

// TestFloat64PrecisionEdgeCase tests the fix for FSSDK-11917
// This test ensures that users with hash values near the maximum uint32 value
// are bucketed correctly (bucket value 9999) instead of getting bucket value 10000
// which would exclude them from the "Everyone Else" rollout rule.
func TestFloat64PrecisionEdgeCase(t *testing.T) {
	bucketer := NewMurmurhashBucketer(logging.GetLogger("", "TestFloat64PrecisionEdgeCase"), DefaultHashSeed)

	// The main test: ensure all bucket values are in valid range [0, 9999]
	// This catches any case where float32 precision would cause bucket value 10000
	for i := 0; i < 10000; i++ {
		bucketingKey := fmt.Sprintf("user_%d_rule_12345", i)
		bucket := bucketer.Generate(bucketingKey)
		assert.GreaterOrEqual(t, bucket, 0, "Bucket value should be >= 0 for key %s", bucketingKey)
		assert.LessOrEqual(t, bucket, 9999, "Bucket value should be <= 9999 for key %s (float32 bug would produce 10000)", bucketingKey)
	}

	// Test with a bucketing key that Mark Biesheuvel identified
	// The issue occurs with specific user ID + rule ID combinations
	// Test a range of user IDs to find edge cases
	for userID := 15580841; userID <= 15580851; userID++ {
		t.Run(fmt.Sprintf("UserID_%d", userID), func(t *testing.T) {
			// Test with different rule IDs that might trigger the edge case
			ruleIDs := []string{"default-rollout-386456-20914452332", "rule_123", "everyone_else"}
			for _, ruleID := range ruleIDs {
				bucketingKey := fmt.Sprintf("%d%s", userID, ruleID)
				bucket := bucketer.Generate(bucketingKey)
				assert.GreaterOrEqual(t, bucket, 0, "Bucket should be >= 0")
				assert.LessOrEqual(t, bucket, 9999, "Bucket should be <= 9999 (float32 would produce 10000 for some edge cases)")
			}
		})
	}
}

// TestBucketingWithHighHashValues tests that users with hash values close to max uint32
// are correctly bucketed into rollout rules
func TestBucketingWithHighHashValues(t *testing.T) {
	bucketer := NewMurmurhashBucketer(logging.GetLogger("", "TestBucketingWithHighHashValues"), DefaultHashSeed)

	// Create a traffic allocation for "Everyone Else" (0-10000 range)
	everyoneElseAllocation := []entities.Range{
		{
			EntityID:   "variation_1",
			EndOfRange: 10000,
		},
	}

	// Test various users to ensure they are properly bucketed
	// The bucketing key needs to match the format used in actual bucketing (userID + experimentID or ruleID)
	testBucketingKey := "15580846default-rollout-386456-20914452332"
	bucketValue := bucketer.Generate(testBucketingKey)

	// The important test: bucket value should be in valid range [0, 9999]
	assert.GreaterOrEqual(t, bucketValue, 0, "Bucket value should be >= 0")
	assert.LessOrEqual(t, bucketValue, 9999, "Bucket value should be <= 9999")

	// Test that this user is bucketed into the "Everyone Else" variation
	entityID := bucketer.BucketToEntity(testBucketingKey, everyoneElseAllocation)
	assert.Equal(t, "variation_1", entityID, "User should be bucketed into 'Everyone Else' variation")

	// Test that the fix doesn't break normal bucketing behavior
	// Create a more complex traffic allocation
	complexAllocation := []entities.Range{
		{
			EntityID:   "",
			EndOfRange: 5000, // 50% get no variation
		},
		{
			EntityID:   "variation_a",
			EndOfRange: 10000, // 50% get variation_a
		},
	}

	// Test with the specific bucketing key
	entityID = bucketer.BucketToEntity(testBucketingKey, complexAllocation)
	// The bucket value will determine which variation is returned
	if bucketValue < 5000 {
		assert.Equal(t, "", entityID, "User with bucket < 5000 should get no variation")
	} else {
		assert.Equal(t, "variation_a", entityID, "User with bucket >= 5000 should get variation_a")
	}
}
