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

	"github.com/stretchr/testify/assert"

	"github.com/optimizely/go-sdk/optimizely/config/datafileprojectconfig"
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

func TestNewPollingProjectConfigManagerWithOptions(t *testing.T) {
	mockDatafile := []byte("{ revision: \"42\" }")
	projectConfig, _ := datafileprojectconfig.NewDatafileProjectConfig(mockDatafile)
	mockRequester := new(MockRequester)
	mockRequester.On("Get", []Header(nil)).Return(mockDatafile, 200, nil)

	// Test we fetch using requester
	sdkKey := "test_sdk_key"
	options := PollingProjectConfigManagerOptions{
		Requester: mockRequester,
	}
	configManager := NewPollingProjectConfigManagerWithOptions(context.Background(), sdkKey, options)
	mockRequester.AssertExpectations(t)

	actual, err := configManager.GetConfig()
	assert.NotNil(t, err)
	assert.Equal(t, projectConfig, actual)
}

func TestNewPollingProjectConfigManagerWithNull(t *testing.T) {
	mockDatafile := []byte("NOT-VALID")
	mockRequester := new(MockRequester)
	mockRequester.On("Get", []Header(nil)).Return(mockDatafile, 200, nil)

	// Test we fetch using requester
	sdkKey := "test_sdk_key"
	options := PollingProjectConfigManagerOptions{
		Requester: mockRequester,
	}
	configManager := NewPollingProjectConfigManagerWithOptions(context.Background(), sdkKey, options)
	mockRequester.AssertExpectations(t)

	_, err := configManager.GetConfig()
	assert.NotNil(t, err)
}
