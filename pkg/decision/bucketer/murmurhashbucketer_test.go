/****************************************************************************
 * Copyright 2019-2020, Optimizely, Inc. and contributors                   *
 *                                                                          *
 * Licensed under the Apache License, Version 2.0 (the "License");          *
 * you may not use this file except in compliance with the License.         *
 * You may obtain a copy of the License at                                  *
 *                                                                          *
 *    http://www.apache.org/licenses/LICENSE-2.0                            *
 *                                                                          *
 * Unless required by applicable law or agreed to in writing, software      *
 * distributed under the License is distributed on an "AS IS" BASIS,        *
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. *
 * See the License for the specific language governing permissions and      *
 * limitations under the License.                                           *
 ***************************************************************************/

// Package bucketer //
package bucketer

import (
	"fmt"
	"testing"

	"github.com/optimizely/go-sdk/pkg/logging"

	"github.com/optimizely/go-sdk/pkg/entities"
	"github.com/stretchr/testify/assert"
)

func TestBucketToEntity(t *testing.T) {
	bucketer := NewMurmurhashBucketer(logging.GetLogger("", "TestBucketToEntity"), DefaultHashSeed)

	experimentID := "1886780721"
	experimentID2 := "1886780722"

	// bucket value 5254
	bucketingID1 := "ppid1"
	// bucket value 4299
	bucketingID2 := "ppid2"
	// bucket value 2434
	bucketingID3 := "ppid3"

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

	assert.Equal(t, variation2, bucketer.BucketToEntity(bucketingID1, experimentID, trafficAlloc))
	assert.Equal(t, variation1, bucketer.BucketToEntity(bucketingID2, experimentID, trafficAlloc))

	// bucket to empty variation range
	assert.Equal(t, "", bucketer.BucketToEntity(bucketingID2, experimentID2, trafficAlloc))

	// bucket outside of range (not in experiment)
	assert.Equal(t, "", bucketer.BucketToEntity(bucketingID3, experimentID, trafficAlloc))
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
