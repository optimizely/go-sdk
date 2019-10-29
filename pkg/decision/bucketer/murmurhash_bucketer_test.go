package bucketer

import (
	"fmt"
	"testing"

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

	assert.Equal(t, 5254, bucketer.Generate(bucketingKey1))
	assert.Equal(t, 4299, bucketer.Generate(bucketingKey2))
	assert.Equal(t, 2434, bucketer.Generate(bucketingKey3))
	assert.Equal(t, 5439, bucketer.Generate(bucketingKey4))
}
