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
	"net/http"
	"testing"
	"time"

	"github.com/optimizely/go-sdk/pkg/config/datafileprojectconfig"
	"github.com/optimizely/go-sdk/pkg/notification"
	"github.com/optimizely/go-sdk/pkg/utils"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockRequester struct {
	utils.Requester
	mock.Mock
}

func (m *MockRequester) Get(uri string, headers ...utils.Header) (response []byte, responseHeaders http.Header, code int, err error) {
	args := m.Called(headers)
	return args.Get(0).([]byte), args.Get(1).(http.Header), args.Int(2), args.Error(3)
}

func newExecGroup() *utils.ExecGroup {
	return utils.NewExecGroup(context.Background())
}

func TestNewPollingProjectConfigManagerWithOptions(t *testing.T) {

	mockDatafile := []byte(`{"revision":"42"}`)
	projectConfig, _ := datafileprojectconfig.NewDatafileProjectConfig(mockDatafile)
	mockRequester := new(MockRequester)
	mockRequester.On("Get", []utils.Header(nil)).Return(mockDatafile, http.Header{}, http.StatusOK, nil)

	// Test we fetch using requester
	sdkKey := "test_sdk_key"

	eg := newExecGroup()
	configManager := NewPollingProjectConfigManager(sdkKey, WithRequester(mockRequester))
	eg.Go(configManager.Start)
	mockRequester.AssertExpectations(t)

	actual, err := configManager.GetConfig()
	assert.Nil(t, err)
	assert.NotNil(t, actual)
	assert.Equal(t, projectConfig, actual)

	eg.TerminateAndWait() // just sending signal and improving coverage
}

func TestNewPollingProjectConfigManagerWithNull(t *testing.T) {
	mockDatafile := []byte("NOT-VALID")
	mockRequester := new(MockRequester)
	mockRequester.On("Get", []utils.Header(nil)).Return(mockDatafile, http.Header{}, http.StatusOK, nil)

	// Test we fetch using requester
	sdkKey := "test_sdk_key"

	eg := newExecGroup()
	configManager := NewPollingProjectConfigManager(sdkKey, WithRequester(mockRequester))
	eg.Go(configManager.Start)
	mockRequester.AssertExpectations(t)

	_, err := configManager.GetConfig()
	assert.NotNil(t, err)
}

func TestNewPollingProjectConfigManagerWithSimilarDatafileRevisions(t *testing.T) {
	mockDatafile1 := []byte(`{"revision":"42","botFiltering":true}`)
	mockDatafile2 := []byte(`{"revision":"42","botFiltering":false}`)
	projectConfig1, _ := datafileprojectconfig.NewDatafileProjectConfig(mockDatafile1)
	mockRequester := new(MockRequester)
	mockRequester.On("Get", []utils.Header(nil)).Return(mockDatafile1, http.Header{}, http.StatusOK, nil)

	sdkKey := "test_sdk_key"

	eg := newExecGroup()
	configManager := NewPollingProjectConfigManager(sdkKey, WithRequester(mockRequester))
	eg.Go(configManager.Start)
	mockRequester.AssertExpectations(t)

	actual, err := configManager.GetConfig()
	assert.Nil(t, err)
	assert.NotNil(t, actual)
	assert.Equal(t, projectConfig1, actual)

	configManager.SyncConfig(mockDatafile2)
	actual, err = configManager.GetConfig()
	assert.Equal(t, projectConfig1, actual)
}

func TestNewPollingProjectConfigManagerWithLastModifiedDates(t *testing.T) {
	mockDatafile1 := []byte(`{"revision":"42","botFiltering":true}`)
	projectConfig1, _ := datafileprojectconfig.NewDatafileProjectConfig(mockDatafile1)
	mockRequester := new(MockRequester)
	modifiedDate := "Wed, 16 Oct 2019 20:16:45 GMT"
	responseHeaders := http.Header{}
	responseHeaders.Set(LastModified, modifiedDate)

	mockRequester.On("Get", []utils.Header(nil)).Return(mockDatafile1, responseHeaders, http.StatusOK, nil)
	mockRequester.On("Get", []utils.Header{utils.Header{Name: ModifiedSince, Value: modifiedDate}}).Return([]byte{}, responseHeaders, http.StatusNotModified, nil)

	sdkKey := "test_sdk_key"

	eg := newExecGroup()
	configManager := NewPollingProjectConfigManager(sdkKey, WithRequester(mockRequester))
	eg.Go(configManager.Start)

	// Fetch valid config
	actual, err := configManager.GetConfig()
	assert.Nil(t, err)
	assert.NotNil(t, actual)
	assert.Equal(t, projectConfig1, actual)

	// Sync and check no changes were made to the previous config because of 304 error code
	configManager.SyncConfig([]byte{})
	actual, err = configManager.GetConfig()
	assert.Nil(t, err)
	assert.NotNil(t, actual)
	assert.Equal(t, projectConfig1, actual)
	mockRequester.AssertExpectations(t)
}

func TestNewPollingProjectConfigManagerWithDifferentDatafileRevisions(t *testing.T) {
	mockDatafile1 := []byte(`{"revision":"42","botFiltering":true}`)
	mockDatafile2 := []byte(`{"revision":"43","botFiltering":false}`)
	projectConfig1, _ := datafileprojectconfig.NewDatafileProjectConfig(mockDatafile1)
	projectConfig2, _ := datafileprojectconfig.NewDatafileProjectConfig(mockDatafile2)
	mockRequester := new(MockRequester)
	mockRequester.On("Get", []utils.Header(nil)).Return(mockDatafile1, http.Header{}, http.StatusOK, nil)

	// Test we fetch using requester
	sdkKey := "test_sdk_key"

	eg := newExecGroup()
	configManager := NewPollingProjectConfigManager(sdkKey, WithRequester(mockRequester))
	eg.Go(configManager.Start)
	mockRequester.AssertExpectations(t)

	actual, err := configManager.GetConfig()
	assert.Nil(t, err)
	assert.NotNil(t, actual)
	assert.Equal(t, projectConfig1, actual)

	configManager.SyncConfig(mockDatafile2)
	actual, err = configManager.GetConfig()
	assert.Equal(t, projectConfig2, actual)
}

func TestNewPollingProjectConfigManagerWithErrorHandling(t *testing.T) {
	mockDatafile1 := []byte("NOT-VALID")
	mockDatafile2 := []byte(`{"revision":"43","botFiltering":false}`)

	projectConfig1, _ := datafileprojectconfig.NewDatafileProjectConfig(mockDatafile1)
	projectConfig2, _ := datafileprojectconfig.NewDatafileProjectConfig(mockDatafile2)
	mockRequester := new(MockRequester)
	mockRequester.On("Get", []utils.Header(nil)).Return(mockDatafile1, http.Header{}, http.StatusOK, nil)

	// Test we fetch using requester
	sdkKey := "test_sdk_key"

	eg := newExecGroup()
	configManager := NewPollingProjectConfigManager(sdkKey, WithRequester(mockRequester))
	eg.Go(configManager.Start)
	mockRequester.AssertExpectations(t)

	actual, err := configManager.GetConfig() // polling for bad file
	assert.NotNil(t, err)
	assert.Nil(t, actual)
	assert.Nil(t, projectConfig1)

	configManager.SyncConfig(mockDatafile2) // polling for good file
	actual, err = configManager.GetConfig()
	assert.Nil(t, err)
	assert.Equal(t, projectConfig2, actual)

	configManager.SyncConfig(mockDatafile1) // polling for bad file, error not null but good project
	actual, err = configManager.GetConfig()
	assert.Nil(t, err)
	assert.Equal(t, projectConfig2, actual)
}

func TestNewPollingProjectConfigManagerOnDecision(t *testing.T) {
	mockDatafile1 := []byte(`{"revision":"42","botFiltering":true}`)
	mockDatafile2 := []byte(`{"revision":"43","botFiltering":false}`)

	mockRequester := new(MockRequester)
	mockRequester.On("Get", []utils.Header(nil)).Return(mockDatafile1, http.Header{}, http.StatusOK, nil)

	// Test we fetch using requester
	sdkKey := "test_sdk_key"

	eg := newExecGroup()
	configManager := NewPollingProjectConfigManager(sdkKey, WithRequester(mockRequester))
	eg.Go(configManager.Start)

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

	err = configManager.RemoveOnProjectConfigUpdate(id)
	assert.Nil(t, err)
}

func TestPollingInterval(t *testing.T) {

	sdkKey := "test_sdk_key"

	eg := newExecGroup()
	configManager := NewPollingProjectConfigManager(sdkKey, WithPollingInterval(5*time.Second))
	eg.Go(configManager.Start)

	assert.Equal(t, configManager.pollingInterval, 5*time.Second)
}

func TestInitialDatafile(t *testing.T) {

	sdkKey := "test_sdk_key"
	eg := newExecGroup()
	configManager := NewPollingProjectConfigManager(sdkKey, WithInitialDatafile([]byte("test")))
	eg.Go(configManager.Start)

	assert.Equal(t, configManager.initDatafile, []byte("test"))
}

func TestDatafileTemplate(t *testing.T) {

	sdkKey := "test_sdk_key"
	datafileTemplate := "https://localhost/v1/%s.json"
	configManager := NewPollingProjectConfigManager(sdkKey, WithDatafileURLTemplate(datafileTemplate))

	assert.Equal(t, datafileTemplate, configManager.datafileURLTemplate)
}
