/****************************************************************************
 * Copyright 2020, Optimizely, Inc. and contributors                        *
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

// Package event //
package event

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSnapshotHasOptionalDecisionsAndNonOptionalEvents(t *testing.T) {
	snapshot := Snapshot{
		Decisions: []Decision{
			Decision{
				VariationID: "1",
			},
		},
		Events: []SnapshotEvent{
			SnapshotEvent{
				EntityID: "1",
			},
		},
	}

	// Check with decisions and events
	jsonValue, err := json.Marshal(snapshot)
	assert.Nil(t, err)

	dict := map[string]interface{}{}
	err = json.Unmarshal(jsonValue, &dict)
	assert.Nil(t, err)
	_, ok := dict["decisions"]
	assert.True(t, ok)
	_, ok = dict["events"]
	assert.True(t, ok)

	// Check without decisions and events
	snapshot.Decisions = nil
	snapshot.Events = nil
	jsonValue, err = json.Marshal(snapshot)
	assert.Nil(t, err)

	dict2 := map[string]interface{}{}
	err = json.Unmarshal(jsonValue, &dict2)
	assert.Nil(t, err)
	_, ok = dict2["decisions"]
	assert.False(t, ok)
	_, ok = dict2["events"]
	assert.True(t, ok)
}
