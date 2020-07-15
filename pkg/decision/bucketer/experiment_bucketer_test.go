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

	"github.com/optimizely/go-sdk/pkg/decision/reasons"
	"github.com/optimizely/go-sdk/pkg/logging"

	"github.com/optimizely/go-sdk/pkg/entities"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockLogger struct {
	mock.Mock
}

func (m *MockLogger) Debug(message string) {
	m.Called(message)
}

func (m *MockLogger) Info(message string) {
	m.Called(message)
}

func (m *MockLogger) Warning(message string) {
	m.Called(message)
}

func (m *MockLogger) Error(message string, err interface{}) {
	m.Called(message, err)
}

func TestBucketExclusionGroups(t *testing.T) {
	mockLogger := MockLogger{}
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

	bucketer := NewMurmurhashExperimentBucketer(&mockLogger, DefaultHashSeed)
	// ppid2 + 1886780722 (groupId) will generate bucket value of 2434 which maps to experiment 1
	mockLogger.On("Debug", fmt.Sprintf(logging.UserAssignedToBucketValue.String(), 2434, "ppid2"))
	mockLogger.On("Info", fmt.Sprintf(logging.UserBucketedIntoExperimentInGroup.String(), "ppid2", "experiment_1", "1886780722"))
	mockLogger.On("Debug", fmt.Sprintf(logging.UserAssignedToBucketValue.String(), 4299, "ppid2"))
	mockLogger.On("Info", fmt.Sprintf(logging.UserNotBucketedIntoExperimentInGroup.String(), "ppid2", "experiment_2", "1886780722"))

	bucketedVariation, reason, _ := bucketer.Bucket("ppid2", experiment1, exclusionGroup)
	assert.Equal(t, experiment1.Variations["22222"], *bucketedVariation)
	assert.Equal(t, reasons.BucketedIntoVariation, reason)
	// since the bucket value maps to experiment 1, the user will not be bucketed for experiment 2
	bucketedVariation, reason, _ = bucketer.Bucket("ppid2", experiment2, exclusionGroup)
	assert.Nil(t, bucketedVariation)
	assert.Equal(t, reasons.NotBucketedIntoVariation, reason)
}
