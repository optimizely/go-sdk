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

package registry

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type ServiceRegistryTestSuite struct {
	suite.Suite
}

func (s *ServiceRegistryTestSuite) TestGetNotificationCenter() {
	// empty state, make sure we get a new notification center
	sdkKey := "sdk_key_1"
	notificationCenter := GetNotificationCenter(sdkKey)
	s.NotNil(notificationCenter)

	notificationCenter2 := GetNotificationCenter(sdkKey)
	s.Equal(notificationCenter, notificationCenter2)
}

func TestServiceRegistryTestSuite(t *testing.T) {
	suite.Run(t, new(ServiceRegistryTestSuite))
}
