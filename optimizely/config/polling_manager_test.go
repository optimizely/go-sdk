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

package config

import (
	"context"
	"testing"

	"github.com/optimizely/go-sdk/optimizely/config/datafileprojectconfig"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockRequester struct {
	Requester
	mock.Mock
}

func (m *MockRequester) Get(headers ...Header) (response []byte, code int, err error) {
	args := m.Called(headers)
	return args.Get(0).([]byte), args.Int(1), args.Error(2)
}

func TestNewPollingProjectConfigManager(t *testing.T) {
	projectConfig, _ := datafileprojectconfig.NewDatafileProjectConfig([]byte{})
	request := new(MockRequester)

	// Bad SDK Key test
	configManager := NewPollingProjectConfigManager(context.Background(), request, []byte{}, 0)
	assert.Equal(t, projectConfig, configManager.GetConfig())

	// Good SDK Key test
	configManager = NewPollingProjectConfigManager(context.Background(), request, []byte{}, 0)
	newConfig := configManager.GetConfig()

	assert.Equal(t, "", newConfig.GetAccountID())
	assert.Equal(t, 4, len(newConfig.GetAudienceMap()))
}
