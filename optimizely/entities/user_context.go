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

package entities

import (
	"fmt"
	"reflect"
)

const bucketingIDAttributeName = "$opt_bucketing_id"

// UserContext holds information about a user
type UserContext struct {
	ID         string
	Attributes UserAttributes
}

// UserAttributes holds information about the user's attributes
type UserAttributes struct {
	Attributes map[string]interface{}
}

var floatType = reflect.TypeOf(float64(0))

// GetString returns the string value for the specified attribute name in the attributes map. Returns error if not found.
func (u UserAttributes) GetString(attrName string) (string, error) {
	if value, ok := u.Attributes[attrName]; ok {
		v := reflect.ValueOf(value)
		if v.Type().String() == "string" {
			return v.String(), nil
		}
	}

	return "", fmt.Errorf(`No string attribute named "%s"`, attrName)
}

// GetBool returns the bool value for the specified attribute name in the attributes map. Returns error if not found.
func (u UserAttributes) GetBool(attrName string) (bool, error) {
	if value, ok := u.Attributes[attrName]; ok {
		v := reflect.ValueOf(value)
		if v.Type().String() == "bool" {
			return v.Bool(), nil
		}
	}

	return false, fmt.Errorf(`No bool attribute named "%s"`, attrName)
}

// GetFloat returns the float64 value for the specified attribute name in the attributes map. Returns error if not found.
func (u UserAttributes) GetFloat(attrName string) (float64, error) {
	if value, ok := u.Attributes[attrName]; ok {
		v := reflect.ValueOf(value)
		if v.Type().String() == "float64" || v.Type().ConvertibleTo(floatType) {
			floatValue := v.Convert(floatType).Float()
			return floatValue, nil
		}
	}

	return 0, fmt.Errorf(`No float attribute named "%s"`, attrName)
}

// GetBucketingID returns the bucketing ID to use for the given user
func (u UserContext) GetBucketingID() (string, error) {
	// by default
	bucketingID := u.ID

	// If the bucketing ID key is defined in attributes, than use that in place of the user ID
	if value, ok := u.Attributes.Attributes[bucketingIDAttributeName]; ok {
		customBucketingID, err := u.Attributes.GetString(bucketingIDAttributeName)
		if err != nil {
			return bucketingID, fmt.Errorf(`Invalid bucketing ID provided: "%s"`, value)
		} else {
			bucketingID = customBucketingID
		}
	}

	return bucketingID, nil
}
