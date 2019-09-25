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
	"testing"

	"github.com/optimizely/go-sdk/optimizely/config/datafileprojectconfig"
	"github.com/optimizely/go-sdk/optimizely/notification"
	"github.com/optimizely/go-sdk/optimizely/utils"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockRequester struct {
	utils.Requester
	mock.Mock
}

func (m *MockRequester) Get(headers ...utils.Header) (response []byte, code int, err error) {
	args := m.Called(headers)
	return args.Get(0).([]byte), args.Int(1), args.Error(2)
}

func TestNewPollingProjectConfigManagerWithOptions(t *testing.T) {

	mockDatafile := []byte(`{"revision":"42"}`)
	projectConfig, _ := datafileprojectconfig.NewDatafileProjectConfig(mockDatafile)
	mockRequester := new(MockRequester)
	mockRequester.On("Get", []utils.Header(nil)).Return(mockDatafile, 200, nil)

	// Test we fetch using requester
	sdkKey := "test_sdk_key"

	exeCtx := utils.NewCancelableExecutionCtx()
	configManager := NewPollingProjectConfigManager(exeCtx, sdkKey, SetRequester(mockRequester))
	mockRequester.AssertExpectations(t)

	actual, err := configManager.GetConfig()
	assert.Nil(t, err)
	assert.NotNil(t, actual)
	assert.Equal(t, projectConfig, actual)
}

func TestNewPollingProjectConfigManagerWithNull(t *testing.T) {
	mockDatafile := []byte("NOT-VALID")
	mockRequester := new(MockRequester)
	mockRequester.On("Get", []utils.Header(nil)).Return(mockDatafile, 200, nil)

	// Test we fetch using requester
	sdkKey := "test_sdk_key"

	exeCtx := utils.NewCancelableExecutionCtx()
	configManager := NewPollingProjectConfigManager(exeCtx, sdkKey, SetRequester(mockRequester))
	mockRequester.AssertExpectations(t)

	_, err := configManager.GetConfig()
	assert.NotNil(t, err)
}

func TestNewPollingProjectConfigManagerWithSimilarDatafileRevisions(t *testing.T) {
	mockDatafile1 := []byte(`{"revision":"42","botFiltering":true}`)
	mockDatafile2 := []byte(`{"revision":"42","botFiltering":false}`)
	projectConfig1, _ := datafileprojectconfig.NewDatafileProjectConfig(mockDatafile1)
	mockRequester := new(MockRequester)
	mockRequester.On("Get", []utils.Header(nil)).Return(mockDatafile1, 200, nil)

	sdkKey := "test_sdk_key"

	exeCtx := utils.NewCancelableExecutionCtx()
	configManager := NewPollingProjectConfigManager(exeCtx, sdkKey, SetRequester(mockRequester))
	mockRequester.AssertExpectations(t)

	actual, err := configManager.GetConfig()
	assert.Nil(t, err)
	assert.NotNil(t, actual)
	assert.Equal(t, projectConfig1, actual)

	configManager.SyncConfig(mockDatafile2)
	actual, err = configManager.GetConfig()
	assert.Equal(t, projectConfig1, actual)
}

func TestNewPollingProjectConfigManagerWithDifferentDatafileRevisions(t *testing.T) {
	mockDatafile1 := []byte(`{"revision":"42","botFiltering":true}`)
	mockDatafile2 := []byte(`{"revision":"43","botFiltering":false}`)
	projectConfig1, _ := datafileprojectconfig.NewDatafileProjectConfig(mockDatafile1)
	projectConfig2, _ := datafileprojectconfig.NewDatafileProjectConfig(mockDatafile2)
	mockRequester := new(MockRequester)
	mockRequester.On("Get", []utils.Header(nil)).Return(mockDatafile1, 200, nil)

	// Test we fetch using requester
	sdkKey := "test_sdk_key"

	exeCtx := utils.NewCancelableExecutionCtx()
	configManager := NewPollingProjectConfigManager(exeCtx, sdkKey, SetRequester(mockRequester))
	mockRequester.AssertExpectations(t)

	actual, err := configManager.GetConfig()
	assert.Nil(t, err)
	assert.NotNil(t, actual)
	assert.Equal(t, projectConfig1, actual)

	configManager.SyncConfig(mockDatafile2)
	actual, err = configManager.GetConfig()
	assert.Equal(t, projectConfig2, actual)
}

func TestNewPollingProjectConfigManagerOnDecision(t *testing.T) {
	mockDatafile1 := []byte(`{"revision":"42","botFiltering":true}`)
	mockDatafile2 := []byte(`{"revision":"43","botFiltering":false}`)

	mockRequester := new(MockRequester)
	mockRequester.On("Get", []utils.Header(nil)).Return(mockDatafile1, 200, nil)

	// Test we fetch using requester
	sdkKey := "test_sdk_key"

	exeCtx := utils.NewCancelableExecutionCtx()
	configManager := NewPollingProjectConfigManager(exeCtx, sdkKey, SetRequester(mockRequester), SetNotification(notification.NewNotificationCenter()))

	var numberOfCalls = 0
	callback := func(notification notification.ProjectConfigUpdateNotification) {
		numberOfCalls++
	}
	id, _ := configManager.OnProjectConfigUpdate(callback)
	mockRequester.AssertExpectations(t)

	actual, err := configManager.GetConfig()
	assert.Nil(t, err)
	assert.NotNil(t, actual)

	configManager.SyncConfig(mockDatafile2)
	actual, err = configManager.GetConfig()
	assert.Nil(t, err)
	assert.NotNil(t, actual)

	assert.NotEqual(t, id, 0)
	assert.Equal(t, numberOfCalls, 1)

	err = configManager.RemoveOnProjectConfigUpdate(id)
	assert.Nil(t, err)
}
