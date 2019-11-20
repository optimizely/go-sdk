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

package optimizely

import (
	"github.com/optimizely/go-sdk/pkg/client"
	"github.com/optimizely/go-sdk/pkg/entities"
)

// Client returns an OptimizelyClient instantitated with the given key and options
func Client(sdkKey string, options ...client.OptionFunc) (*client.OptimizelyClient, error) {
	factory := &client.OptimizelyFactory{
		SDKKey: sdkKey,
	}
	return factory.Client(options...)
}

// UserContext is a helper method for creating a user context
func UserContext(userID string, attributes map[string]interface{}) entities.UserContext {
	return entities.UserContext{
		ID: userID,
		Attributes: attributes,
	}
}
