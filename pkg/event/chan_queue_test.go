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

// Package event //
package event

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestChanQueue_Add_Size_Remove(t *testing.T) {
	q := NewChanQueue(100)

	impression := BuildTestImpressionEvent()
	conversion := BuildTestConversionEvent()

	q.Add(impression)
	q.Add(impression)
	q.Add(conversion)

	time.Sleep(100 * time.Millisecond)

	items1 := q.Get(2)

	assert.Equal(t, 2, len(items1))

	q.Remove(1)

	items2 := q.Get(1)

	assert.True(t, len(items2) != 0)

	allItems := q.Remove(3)

	assert.True(t,len(allItems) > 0)

	assert.Equal(t, 0, q.Size())
}
