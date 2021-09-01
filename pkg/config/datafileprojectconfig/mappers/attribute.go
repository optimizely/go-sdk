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

// Package mappers  ...
package mappers

import (
	datafileEntities "github.com/optimizely/go-sdk/pkg/config/datafileprojectconfig/entities"
	"github.com/optimizely/go-sdk/pkg/entities"
)

// MapAttributes maps the raw datafile attribute entities to SDK Attribute entities
func MapAttributes(attributes []datafileEntities.Attribute) (attributesList []entities.Attribute, attributeMap map[string]entities.Attribute, attributeKeyToIDMap map[string]string) {

	attributesList = []entities.Attribute{}
	attributeMap = make(map[string]entities.Attribute)
	attributeKeyToIDMap = make(map[string]string)
	for _, attribute := range attributes {
		_, ok := attributeMap[attribute.ID]
		if !ok {
			tmpAttribute := entities.Attribute{
				ID:  attribute.ID,
				Key: attribute.Key,
			}
			attributesList = append(attributesList, tmpAttribute)
			attributeMap[attribute.ID] = tmpAttribute
			attributeKeyToIDMap[attribute.Key] = attribute.ID
		}
	}
	return attributesList, attributeMap, attributeKeyToIDMap
}
