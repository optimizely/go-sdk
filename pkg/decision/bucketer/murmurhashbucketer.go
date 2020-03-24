/****************************************************************************
 * Copyright 2019, Optimizely, Inc. and contributors                        *
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
	"math"

	"github.com/optimizely/go-sdk/pkg/entities"
	"github.com/optimizely/go-sdk/pkg/logging"

	"github.com/twmb/murmur3"
)

var maxHashValue = float32(math.Pow(2, 32))

// DefaultHashSeed is the hash seed to use for murmurhash
const DefaultHashSeed = 1
const maxTrafficValue = 10000

// Bucketer is used to generate bucket value using bucketing key
type Bucketer interface {
	Generate(bucketingKey string) int
	BucketToEntity(bucketKey string, trafficAllocations []entities.Range) (entityID string)
}

// MurmurhashBucketer generates the bucketing value using the mmh3 algorightm
type MurmurhashBucketer struct {
	hashSeed uint32
	logger   logging.OptimizelyLogProducer
}

// NewMurmurhashBucketer returns a new instance of the murmurhash bucketer
func NewMurmurhashBucketer(logger logging.OptimizelyLogProducer, hashSeed uint32) *MurmurhashBucketer {
	return &MurmurhashBucketer{
		hashSeed: hashSeed,
		logger: logger,
	}
}

// Generate returns a bucketing value for bucketing key
func (b MurmurhashBucketer) Generate(bucketingKey string) int {
	hasher := murmur3.SeedNew32(b.hashSeed)
	if _, err := hasher.Write([]byte(bucketingKey)); err != nil {
		b.logger.Error(fmt.Sprintf("Unable to generate a hash for the bucketing key=%s", bucketingKey), err)
	}
	hashCode := hasher.Sum32()
	ratio := float32(hashCode) / maxHashValue
	return int(ratio * maxTrafficValue)
}

// BucketToEntity buckets into a traffic against given bucketKey
func (b MurmurhashBucketer) BucketToEntity(bucketKey string, trafficAllocations []entities.Range) (entityID string) {
	bucketValue := b.Generate(bucketKey)

	var currentEndOfRange int
	for _, trafficAllocationRange := range trafficAllocations {
		currentEndOfRange = trafficAllocationRange.EndOfRange
		if bucketValue < currentEndOfRange {
			return trafficAllocationRange.EntityID
		}
	}

	return ""
}
