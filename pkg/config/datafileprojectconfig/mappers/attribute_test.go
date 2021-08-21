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

	datafileEntities "github.com/optimizely/go-sdk/pkg/config/datafileprojectconfig/entities"
	"github.com/optimizely/go-sdk/pkg/entities"

	"github.com/stretchr/testify/assert"
)

func TestMapAttributesWithEmptyList(t *testing.T) {

	attributeMap, attributeKeyToIDMap, _ := MapAttributes(nil)

	expectedAttributeMap := map[string]entities.Attribute{}
	expectedAttributeKeyToIDMap := map[string]string{}

	assert.Equal(t, attributeMap, expectedAttributeMap)
	assert.Equal(t, attributeKeyToIDMap, expectedAttributeKeyToIDMap)
}
func TestMapAttributes(t *testing.T) {

	attrList := []datafileEntities.Attribute{{ID: "1", Key: "one"}, {ID: "2", Key: "two"},
		{ID: "3", Key: "three"}, {ID: "2", Key: "four"}, {ID: "5", Key: "one"}}

	attributeMap, attributeKeyToIDMap, _ := MapAttributes(attrList)

	expectedAttributeMap := map[string]entities.Attribute{"1": {"1", "one"},
		"2": {"2", "two"}, "3": {"3", "three"}, "5": {"5", "one"}}
	expectedAttributeKeyToIDMap := map[string]string{"one": "5", "three": "3", "two": "2"}

	assert.Equal(t, attributeMap, expectedAttributeMap)
	assert.Equal(t, attributeKeyToIDMap, expectedAttributeKeyToIDMap)
}
