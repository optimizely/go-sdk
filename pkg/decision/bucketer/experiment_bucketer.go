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
	"github.com/optimizely/go-sdk/pkg/decision/reasons"
	"github.com/optimizely/go-sdk/pkg/entities"
	"github.com/optimizely/go-sdk/pkg/logging"
)

// ExperimentBucketer is used to bucket the user into a particular entity in the experiment's traffic alloc range
type ExperimentBucketer interface {
	Bucket(bucketingID string, experiment entities.Experiment, group entities.Group) (*entities.Variation, reasons.Reason, error)
}

// MurmurhashExperimentBucketer buckets the user using the mmh3 algorightm
type MurmurhashExperimentBucketer struct {
	bucketer Bucketer
}

// NewMurmurhashExperimentBucketer returns a new instance of the murmurhash experiment bucketer
func NewMurmurhashExperimentBucketer(logger logging.OptimizelyLogProducer, hashSeed uint32) *MurmurhashExperimentBucketer {
	return &MurmurhashExperimentBucketer{
		bucketer: MurmurhashBucketer{hashSeed: hashSeed, logger:logger},
	}
}

// Bucket buckets the user into the given experiment
func (b MurmurhashExperimentBucketer) Bucket(bucketingID string, experiment entities.Experiment, group entities.Group) (*entities.Variation, reasons.Reason, error) {
	if experiment.GroupID != "" && group.Policy == "random" {
		bucketKey := bucketingID + group.ID
		bucketedExperimentID := b.bucketer.BucketToEntity(bucketKey, group.TrafficAllocation)
		if bucketedExperimentID == "" || bucketedExperimentID != experiment.ID {
			// User is not bucketed into provided experiment in mutex group
			return nil, reasons.NotBucketedIntoVariation, nil
		}
	}

	bucketKey := bucketingID + experiment.ID
	bucketedVariationID := b.bucketer.BucketToEntity(bucketKey, experiment.TrafficAllocation)
	if bucketedVariationID == "" {
		// User is not bucketed into a variation in the experiment, return nil variation
		return nil, reasons.NotBucketedIntoVariation, nil
	}

	if variation, ok := experiment.Variations[bucketedVariationID]; ok {
		return &variation, reasons.BucketedIntoVariation, nil
	}

	return nil, reasons.BucketedVariationNotFound, nil
}
