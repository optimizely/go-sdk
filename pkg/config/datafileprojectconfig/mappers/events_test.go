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

package mappers

import (
	"testing"

	"github.com/goccy/go-json"
	datafileEntities "github.com/optimizely/go-sdk/pkg/config/datafileprojectconfig/entities"
	"github.com/optimizely/go-sdk/pkg/entities"
	"github.com/stretchr/testify/assert"
)

func TestMapEvents(t *testing.T) {
	const testEventString = `{
		"id": "some_id",
		"key": "event1",
		"experimentIds": ["11111", "11112"]
	 }`

	var rawEvent datafileEntities.Event
	json.Unmarshal([]byte(testEventString), &rawEvent)

	rawEvents := []datafileEntities.Event{rawEvent}
	eventsMap := MapEvents(rawEvents)
	expectedEventMap := map[string]entities.Event{
		"event1": {
			ID:            "some_id",
			Key:           "event1",
			ExperimentIds: []string{"11111", "11112"},
		},
	}

	assert.Equal(t, expectedEventMap, eventsMap)
}
