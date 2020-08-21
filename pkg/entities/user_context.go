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

// Package entities //
package entities

import (
	"fmt"

	"github.com/optimizely/go-sdk/pkg/utils"
)

const bucketingIDAttributeName = "$opt_bucketing_id"

// UserContext holds information about a user
type UserContext struct {
	ID         string
	Attributes map[string]interface{}
}

// CheckAttributeExists returns whether the specified attribute name exists in the attributes map.
func (u UserContext) CheckAttributeExists(attrName string) bool {
	if value, ok := u.Attributes[attrName]; ok && value != nil {
		return true
	}

	return false
}

// GetStringAttribute returns the string value for the specified attribute name in the attributes map. Returns error if not found.
func (u UserContext) GetStringAttribute(attrName string) (string, error) {
	if value, ok := u.Attributes[attrName]; ok {
		stringVal, err := utils.GetStringValue(value)
		if err == nil {
			return stringVal, nil
		}
	}

	return "", fmt.Errorf(`no string attribute named "%s"`, attrName)
}

// GetBoolAttribute returns the bool value for the specified attribute name in the attributes map. Returns error if not found.
func (u UserContext) GetBoolAttribute(attrName string) (bool, error) {
	if value, ok := u.Attributes[attrName]; ok {
		boolVal, err := utils.GetBoolValue(value)
		if err == nil {
			return boolVal, nil
		}
	}

	return false, fmt.Errorf(`no bool attribute named "%s"`, attrName)
}

// GetFloatAttribute returns the float64 value for the specified attribute name in the attributes map. Returns error if not found.
func (u UserContext) GetFloatAttribute(attrName string) (float64, error) {
	if value, ok := u.Attributes[attrName]; ok {
		floatVal, err := utils.GetFloatValue(value)
		if err == nil {
			return floatVal, nil
		}
	}

	return 0, fmt.Errorf(`no float attribute named "%s"`, attrName)
}

// GetIntAttribute returns the int64 value for the specified attribute name in the attributes map. Returns error if not found.
func (u UserContext) GetIntAttribute(attrName string) (int64, error) {
	if value, ok := u.Attributes[attrName]; ok {
		intVal, err := utils.GetIntValue(value)
		if err == nil {
			return intVal, nil
		}
	}

	return 0, fmt.Errorf(`no int attribute named "%s"`, attrName)
}

// GetAttribute returns the value for the specified attribute name in the attributes map. Returns error if not found.
func (u UserContext) GetAttribute(attrName string) (interface{}, error) {
	if value, ok := u.Attributes[attrName]; ok {
		return value, nil
	}

	return 0, fmt.Errorf(`no attribute named "%s"`, attrName)
}

// GetBucketingID returns the bucketing ID to use for the given user
func (u UserContext) GetBucketingID() (string, error) {
	// by default
	bucketingID := u.ID

	// If the bucketing ID key is defined in attributes, than use that in place of the user ID
	if value, ok := u.Attributes[bucketingIDAttributeName]; ok {
		customBucketingID, err := u.GetStringAttribute(bucketingIDAttributeName)
		if err != nil {
			return bucketingID, fmt.Errorf(`invalid bucketing ID provided: "%v"`, value)
		}
		bucketingID = customBucketingID
	}

	return bucketingID, nil
}
