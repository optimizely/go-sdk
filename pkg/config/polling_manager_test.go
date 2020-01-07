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
	"reflect"
	"sync/atomic"
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

// assertion method to periodically check target function each tick.
func assertPeriodically(t *testing.T, evaluationMethod func() bool) {
	assert.Eventually(t, func() bool {
		return evaluationMethod()
	}, 500*time.Millisecond, 110*time.Millisecond)
}

func TestNewPollingProjectConfigManagerWithOptions(t *testing.T) {

	mockDatafile := []byte(`{"revision":"42"}`)
	projectConfig, _ := datafileprojectconfig.NewDatafileProjectConfig(mockDatafile)
	mockRequester := new(MockRequester)
	mockRequester.On("Get", []utils.Header(nil)).Return(mockDatafile, http.Header{}, http.StatusOK, nil)

	// Test we fetch using requester
	sdkKey := "test_sdk_key"
	configManager := NewPollingProjectConfigManager(sdkKey, WithRequester(mockRequester))
	mockRequester.AssertExpectations(t)

	actual, err := configManager.GetConfig()
	assert.Nil(t, err)
	assert.NotNil(t, actual)
	assert.Equal(t, projectConfig, actual)
}

func TestNewAsyncPollingProjectConfigManagerWithOptions(t *testing.T) {

	mockDatafile := []byte(`{"revision":"42"}`)
	projectConfig, _ := datafileprojectconfig.NewDatafileProjectConfig(mockDatafile)
	mockRequester := new(MockRequester)
	mockRequester.On("Get", []utils.Header(nil)).Return(mockDatafile, http.Header{}, http.StatusOK, nil)

	// Test we fetch using requester
	sdkKey := "test_sdk_key"

	eg := newExecGroup()
	configManager := NewAsyncPollingProjectConfigManager(sdkKey, WithRequester(mockRequester), WithPollingInterval(100*time.Millisecond))

	// poll after 100ms
	eg.Go(configManager.Start)
	evaluationMethod := func() bool {
		actual, _ := configManager.GetConfig()
		return reflect.DeepEqual(projectConfig, actual)
	}
	assertPeriodically(t, evaluationMethod)
	mockRequester.AssertExpectations(t)
	eg.TerminateAndWait()
}

func TestSyncConfig(t *testing.T) {

	mockDatafile := []byte(`{"revision":"42"}`)
	projectConfig, _ := datafileprojectconfig.NewDatafileProjectConfig(mockDatafile)
	mockRequester := new(MockRequester)
	mockRequester.On("Get", []utils.Header(nil)).Return(mockDatafile, http.Header{}, http.StatusOK, nil)

	// Test we fetch using requester
	configManager := NewAsyncPollingProjectConfigManager("test_sdk_key", WithRequester(mockRequester))
	configManager.SyncConfig()
	mockRequester.AssertCalled(t, "Get", []utils.Header(nil))

	actual, err := configManager.GetConfig()
	assert.Nil(t, err)
	assert.NotNil(t, actual)
	assert.Equal(t, projectConfig, actual)
}

func TestNewPollingProjectConfigManagerWithNull(t *testing.T) {
	mockDatafile := []byte("NOT-VALID")
	mockRequester := new(MockRequester)
	mockRequester.On("Get", []utils.Header(nil)).Return(mockDatafile, http.Header{}, http.StatusOK, nil)

	// Test we fetch using requester
	sdkKey := "test_sdk_key"

	configManager := NewPollingProjectConfigManager(sdkKey, WithRequester(mockRequester))
	mockRequester.AssertExpectations(t)

	_, err := configManager.GetConfig()
	assert.NotNil(t, err)
}

func TestNewAsyncPollingProjectConfigManagerWithNullDatafile(t *testing.T) {
	mockDatafile := []byte("NOT-VALID")
	mockRequester := new(MockRequester)
	mockRequester.On("Get", []utils.Header(nil)).Return(mockDatafile, http.Header{}, http.StatusOK, nil)

	// Test we fetch using requester
	sdkKey := "test_sdk_key"
	eg := newExecGroup()
	configManager := NewAsyncPollingProjectConfigManager(sdkKey, WithRequester(mockRequester), WithPollingInterval(100*time.Millisecond))

	// poll after 100ms
	eg.Go(configManager.Start)
	evaluationMethod := func() bool {
		_, err := configManager.GetConfig()
		return err != nil
	}
	assertPeriodically(t, evaluationMethod)
	mockRequester.AssertExpectations(t)
	eg.TerminateAndWait()
}

func TestNewPollingProjectConfigManagerWithSimilarDatafileRevisions(t *testing.T) {
	mockDatafile1 := []byte(`{"revision":"42","botFiltering":true}`)
	mockDatafile2 := []byte(`{"revision":"42","botFiltering":false}`)
	projectConfig1, _ := datafileprojectconfig.NewDatafileProjectConfig(mockDatafile1)
	mockRequester := new(MockRequester)
	mockRequester.On("Get", []utils.Header(nil)).Return(mockDatafile2, http.Header{}, http.StatusOK, nil)

	sdkKey := "test_sdk_key"
	eg := newExecGroup()
	configManager := NewPollingProjectConfigManager(sdkKey, WithRequester(mockRequester), WithInitialDatafile(mockDatafile1), WithPollingInterval(100*time.Millisecond))

	// initialized with hardcoded datafile
	actual, err := configManager.GetConfig()
	assert.Nil(t, err)
	assert.NotNil(t, actual)
	assert.Equal(t, projectConfig1, actual)

	// poll after 100ms
	eg.Go(configManager.Start)
	evaluationMethod := func() bool {
		actual, _ = configManager.GetConfig()
		return reflect.DeepEqual(projectConfig1, actual)
	}
	assertPeriodically(t, evaluationMethod)
	mockRequester.AssertExpectations(t)
	eg.TerminateAndWait()
}

func TestNewAsyncPollingProjectConfigManagerWithSimilarDatafileRevisions(t *testing.T) {
	mockDatafile1 := []byte(`{"revision":"42","botFiltering":true}`)
	mockDatafile2 := []byte(`{"revision":"42","botFiltering":false}`)
	projectConfig1, _ := datafileprojectconfig.NewDatafileProjectConfig(mockDatafile1)
	mockRequester := new(MockRequester)
	mockRequester.On("Get", []utils.Header(nil)).Return(mockDatafile2, http.Header{}, http.StatusOK, nil)

	sdkKey := "test_sdk_key"

	eg := newExecGroup()
	asyncConfigManager := NewAsyncPollingProjectConfigManager(sdkKey, WithRequester(mockRequester), WithInitialDatafile(mockDatafile1), WithPollingInterval(100*time.Millisecond))

	// initialized with hardcoded datafile
	actual, err := asyncConfigManager.GetConfig()
	assert.Nil(t, err)
	assert.NotNil(t, actual)
	assert.Equal(t, projectConfig1, actual)

	// poll after 100ms
	eg.Go(asyncConfigManager.Start)
	evaluationMethod := func() bool {
		actual, _ = asyncConfigManager.GetConfig()
		return reflect.DeepEqual(projectConfig1, actual)
	}
	assertPeriodically(t, evaluationMethod)
	mockRequester.AssertExpectations(t)
	eg.TerminateAndWait()
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
	configManager := NewPollingProjectConfigManager(sdkKey, WithRequester(mockRequester), WithPollingInterval(100*time.Millisecond))

	// Fetch valid config (initial poll)
	actual, err := configManager.GetConfig()
	assert.Nil(t, err)
	assert.NotNil(t, actual)
	assert.Equal(t, projectConfig1, actual)

	// poll after 100ms to check no changes were made to the previous config because of 304 error code (first poll)
	eg.Go(configManager.Start)
	evaluationMethod := func() bool {
		actual, _ = configManager.GetConfig()
		return reflect.DeepEqual(projectConfig1, actual)
	}
	assertPeriodically(t, evaluationMethod)
	mockRequester.AssertExpectations(t)
	eg.TerminateAndWait()
}

func TestNewAsyncPollingProjectConfigManagerWithLastModifiedDates(t *testing.T) {
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
	configManager := NewAsyncPollingProjectConfigManager(sdkKey, WithRequester(mockRequester), WithPollingInterval(100*time.Millisecond))

	// poll after 100ms to fetch valid config (first poll)
	eg.Go(configManager.Start)
	evaluationMethod := func() bool {
		actual, _ := configManager.GetConfig()
		return reflect.DeepEqual(projectConfig1, actual)
	}
	assertPeriodically(t, evaluationMethod)

	// poll after 100ms to check no changes were made to the previous config because of 304 error code (second poll)
	evaluationMethod = func() bool {
		actual, _ := configManager.GetConfig()
		return reflect.DeepEqual(projectConfig1, actual)
	}
	assertPeriodically(t, evaluationMethod)
	mockRequester.AssertExpectations(t)
	eg.TerminateAndWait()
}

func TestNewPollingProjectConfigManagerWithDifferentDatafileRevisions(t *testing.T) {
	mockDatafile1 := []byte(`{"revision":"42","botFiltering":true}`)
	mockDatafile2 := []byte(`{"revision":"43","botFiltering":false}`)
	projectConfig1, _ := datafileprojectconfig.NewDatafileProjectConfig(mockDatafile1)
	projectConfig2, _ := datafileprojectconfig.NewDatafileProjectConfig(mockDatafile2)
	mockRequester := new(MockRequester)
	mockRequester.On("Get", []utils.Header(nil)).Return(mockDatafile2, http.Header{}, http.StatusOK, nil)

	// Test we fetch using requester
	sdkKey := "test_sdk_key"
	eg := newExecGroup()
	configManager := NewPollingProjectConfigManager(sdkKey, WithRequester(mockRequester), WithInitialDatafile(mockDatafile1), WithPollingInterval(100*time.Millisecond))

	// initialized with hardcoded datafile
	actual, err := configManager.GetConfig()
	assert.Nil(t, err)
	assert.NotNil(t, actual)
	assert.Equal(t, projectConfig1, actual)

	// poll after 100ms
	eg.Go(configManager.Start)
	evaluationMethod := func() bool {
		actual, _ := configManager.GetConfig()
		return reflect.DeepEqual(projectConfig2, actual)
	}
	assertPeriodically(t, evaluationMethod)
	mockRequester.AssertExpectations(t)
	eg.TerminateAndWait()
}

func TestNewAsyncPollingProjectConfigManagerWithDifferentDatafileRevisions(t *testing.T) {
	mockDatafile1 := []byte(`{"revision":"42","botFiltering":true}`)
	mockDatafile2 := []byte(`{"revision":"43","botFiltering":false}`)
	projectConfig1, _ := datafileprojectconfig.NewDatafileProjectConfig(mockDatafile1)
	projectConfig2, _ := datafileprojectconfig.NewDatafileProjectConfig(mockDatafile2)
	mockRequester := new(MockRequester)
	mockRequester.On("Get", []utils.Header(nil)).Return(mockDatafile2, http.Header{}, http.StatusOK, nil)

	// Test we fetch using requester
	sdkKey := "test_sdk_key"
	eg := newExecGroup()
	configManager := NewAsyncPollingProjectConfigManager(sdkKey, WithRequester(mockRequester), WithInitialDatafile(mockDatafile1), WithPollingInterval(100*time.Millisecond))

	// initialized with hardcoded datafile
	actual, err := configManager.GetConfig()
	assert.Nil(t, err)
	assert.NotNil(t, actual)
	assert.Equal(t, projectConfig1, actual)

	// poll after 100ms
	eg.Go(configManager.Start)
	evaluationMethod := func() bool {
		actual, _ := configManager.GetConfig()
		return reflect.DeepEqual(projectConfig2, actual)
	}
	assertPeriodically(t, evaluationMethod)
	mockRequester.AssertExpectations(t)
	eg.TerminateAndWait()
}

func TestNewPollingProjectConfigManagerWithErrorHandling(t *testing.T) {
	mockDatafile1 := []byte("NOT-VALID")
	mockDatafile2 := []byte(`{"revision":"43","botFiltering":false}`)

	projectConfig1, _ := datafileprojectconfig.NewDatafileProjectConfig(mockDatafile1)
	projectConfig2, _ := datafileprojectconfig.NewDatafileProjectConfig(mockDatafile2)
	mockRequester := new(MockRequester)
	mockRequester.On("Get", []utils.Header(nil)).Return(mockDatafile1, http.Header{}, http.StatusOK, nil).Times(1)

	// Test we fetch using requester
	sdkKey := "test_sdk_key"

	configManager := NewPollingProjectConfigManager(sdkKey, WithRequester(mockRequester))
	// verifying initial poll for bad file
	mockRequester.AssertExpectations(t)
	actual, err := configManager.GetConfig()
	assert.NotNil(t, err)
	assert.Nil(t, actual)
	assert.Nil(t, projectConfig1)

	mockRequester.On("Get", []utils.Header(nil)).Return(mockDatafile2, http.Header{}, http.StatusOK, nil).Times(1)
	configManager.SyncConfig() // polling for good file
	mockRequester.AssertExpectations(t)
	actual, err = configManager.GetConfig()
	assert.Nil(t, err)
	assert.Equal(t, projectConfig2, actual)

	mockRequester.On("Get", []utils.Header(nil)).Return(mockDatafile1, http.Header{}, http.StatusOK, nil).Times(1)
	configManager.SyncConfig() // polling for bad file, error not null but good project
	mockRequester.AssertExpectations(t)
	actual, err = configManager.GetConfig()
	assert.Nil(t, err)
	assert.Equal(t, projectConfig2, actual)
}

func TestNewAsyncPollingProjectConfigManagerWithErrorHandling(t *testing.T) {
	mockDatafile1 := []byte("NOT-VALID")
	mockDatafile2 := []byte(`{"revision":"43","botFiltering":false}`)

	projectConfig1, _ := datafileprojectconfig.NewDatafileProjectConfig(mockDatafile1)
	projectConfig2, _ := datafileprojectconfig.NewDatafileProjectConfig(mockDatafile2)
	mockRequester := new(MockRequester)
	mockRequester.On("Get", []utils.Header(nil)).Return(mockDatafile1, http.Header{}, http.StatusOK, nil).Times(1)

	// Test we fetch using requester
	sdkKey := "test_sdk_key"

	asyncConfigManager := NewAsyncPollingProjectConfigManager(sdkKey, WithRequester(mockRequester))

	asyncConfigManager.SyncConfig() // polling for bad file
	mockRequester.AssertExpectations(t)
	actual, err := asyncConfigManager.GetConfig()
	assert.NotNil(t, err)
	assert.Nil(t, actual)
	assert.Nil(t, projectConfig1)

	mockRequester.On("Get", []utils.Header(nil)).Return(mockDatafile2, http.Header{}, http.StatusOK, nil).Times(1)
	asyncConfigManager.SyncConfig() // polling for good file
	mockRequester.AssertExpectations(t)
	actual, err = asyncConfigManager.GetConfig()
	assert.Nil(t, err)
	assert.Equal(t, projectConfig2, actual)

	mockRequester.On("Get", []utils.Header(nil)).Return(mockDatafile1, http.Header{}, http.StatusOK, nil).Times(1)
	asyncConfigManager.SyncConfig() // polling for bad file, error not null but good project
	mockRequester.AssertExpectations(t)
	actual, err = asyncConfigManager.GetConfig()
	assert.Nil(t, err)
	assert.Equal(t, projectConfig2, actual)
}

func TestNewPollingProjectConfigManagerOnDecision(t *testing.T) {
	mockDatafile1 := []byte(`{"revision":"42","botFiltering":true}`)
	mockDatafile2 := []byte(`{"revision":"43","botFiltering":false}`)

	mockRequester := new(MockRequester)
	mockRequester.On("Get", []utils.Header(nil)).Return(mockDatafile2, http.Header{}, http.StatusOK, nil)

	// Test we fetch using requester
	sdkKey := "test_sdk_key"
	eg := newExecGroup()
	configManager := NewPollingProjectConfigManager(sdkKey, WithRequester(mockRequester), WithPollingInterval(100*time.Millisecond), WithInitialDatafile(mockDatafile1))
	var numberOfCalls uint64 = 0
	callback := func(notification notification.ProjectConfigUpdateNotification) {
		atomic.AddUint64(&numberOfCalls, 1)
	}
	id, _ := configManager.OnProjectConfigUpdate(callback)
	assert.NotEqual(t, 0, id)

	// initialized with hardcoded datafile
	config1, err := configManager.GetConfig()
	assert.Nil(t, err)
	assert.NotNil(t, config1)
	assert.Equal(t, uint64(0), atomic.LoadUint64(&numberOfCalls))

	// poll after 100ms
	eg.Go(configManager.Start)
	evaluationMethod := func() bool {
		config2, _ := configManager.GetConfig()
		return !reflect.DeepEqual(config1, config2)
	}
	assertPeriodically(t, evaluationMethod)
	mockRequester.AssertExpectations(t)
	assert.Equal(t, uint64(1), atomic.LoadUint64(&numberOfCalls))
	err = configManager.RemoveOnProjectConfigUpdate(id)
	assert.Nil(t, err)
	eg.TerminateAndWait()
}

func TestNewAsyncPollingProjectConfigManagerOnDecision(t *testing.T) {
	mockDatafile1 := []byte(`{"revision":"42","botFiltering":true}`)

	mockRequester := new(MockRequester)
	mockRequester.On("Get", []utils.Header(nil)).Return(mockDatafile1, http.Header{}, http.StatusOK, nil)

	// Test we fetch using requester
	sdkKey := "test_sdk_key"
	eg := newExecGroup()
	asyncConfigManager := NewAsyncPollingProjectConfigManager(sdkKey, WithRequester(mockRequester), WithPollingInterval(100*time.Millisecond))
	var numberOfCalls uint64 = 0
	callback := func(notification notification.ProjectConfigUpdateNotification) {
		atomic.AddUint64(&numberOfCalls, 1)
	}
	id, _ := asyncConfigManager.OnProjectConfigUpdate(callback)
	assert.NotEqual(t, 0, id)

	// poll after 100ms
	eg.Go(asyncConfigManager.Start)
	evaluationMethod := func() bool {
		actual, _ := asyncConfigManager.GetConfig()
		return actual != nil
	}
	assertPeriodically(t, evaluationMethod)
	mockRequester.AssertExpectations(t)
	assert.Equal(t, uint64(1), atomic.LoadUint64(&numberOfCalls))
	err := asyncConfigManager.RemoveOnProjectConfigUpdate(id)
	assert.Nil(t, err)
	eg.TerminateAndWait()
}

func TestGetOptimizelyConfigForNewPollingProjectConfigManager(t *testing.T) {

	mockDatafile1 := []byte(`{"revision":"42","botFiltering":true}`)
	mockDatafile2 := []byte(`{"revision":"43","botFiltering":false}`)

	mockRequester := new(MockRequester)
	mockRequester.On("Get", []utils.Header(nil)).Return(mockDatafile2, http.Header{}, http.StatusOK, nil)

	// Test we fetch using requester
	sdkKey := "test_sdk_key"
	eg := newExecGroup()
	configManager := NewPollingProjectConfigManager(sdkKey, WithRequester(mockRequester), WithPollingInterval(100*time.Millisecond), WithInitialDatafile(mockDatafile1))

	// initialized with hardcoded datafile
	projectConfig, err := configManager.GetConfig()
	assert.Nil(t, err)
	assert.NotNil(t, projectConfig)
	optimizelyConfig := configManager.GetOptimizelyConfig()
	assert.Equal(t, "42", optimizelyConfig.Revision)

	// poll after 100ms
	eg.Go(configManager.Start)
	evaluationMethod := func() bool {
		optimizelyConfig = configManager.GetOptimizelyConfig()
		return "43" == optimizelyConfig.Revision
	}
	assertPeriodically(t, evaluationMethod)
	mockRequester.AssertExpectations(t)
	eg.TerminateAndWait()
}

func TestGetOptimizelyConfigForNewAsyncPollingProjectConfigManager(t *testing.T) {
	mockDatafile1 := []byte(`{"revision":"42","botFiltering":true}`)
	mockDatafile2 := []byte(`{"revision":"43","botFiltering":false}`)
	mockRequester := new(MockRequester)
	mockRequester.On("Get", []utils.Header(nil)).Return(mockDatafile2, http.Header{}, http.StatusOK, nil)

	// Test we fetch using requester
	sdkKey := "test_sdk_key"

	eg := newExecGroup()
	asyncConfigManager := NewAsyncPollingProjectConfigManager(sdkKey, WithRequester(mockRequester), WithPollingInterval(100*time.Millisecond), WithInitialDatafile(mockDatafile1))

	// initialized with hardcoded datafile
	projectConfig, err := asyncConfigManager.GetConfig()
	assert.Nil(t, err)
	assert.NotNil(t, projectConfig)
	optimizelyConfig := asyncConfigManager.GetOptimizelyConfig()
	assert.Equal(t, "42", optimizelyConfig.Revision)

	// poll after 100ms
	eg.Go(asyncConfigManager.Start)
	evaluationMethod := func() bool {
		optimizelyConfig = asyncConfigManager.GetOptimizelyConfig()
		return "43" == optimizelyConfig.Revision
	}
	assertPeriodically(t, evaluationMethod)
	mockRequester.AssertExpectations(t)
	eg.TerminateAndWait()
}

func TestNewPollingProjectConfigManagerHardcodedDatafile(t *testing.T) {
	mockDatafile1 := []byte(`{"revision":"42"}`)
	mockDatafile2 := []byte(`{"revision":"43"}`)
	sdkKey := "test_sdk_key"

	mockRequester := new(MockRequester)
	mockRequester.On("Get", []utils.Header(nil)).Return(mockDatafile2, http.Header{}, http.StatusOK, nil)

	configManager := NewPollingProjectConfigManager(sdkKey, WithInitialDatafile(mockDatafile1), WithRequester(mockRequester))
	config, err := configManager.GetConfig()

	assert.Nil(t, err)
	assert.NotNil(t, config)
	assert.Equal(t, "42", config.GetRevision())
}

func TestNewAsyncPollingProjectConfigManagerHardcodedDatafile(t *testing.T) {
	mockDatafile1 := []byte(`{"revision":"42"}`)
	mockDatafile2 := []byte(`{"revision":"43"}`)
	sdkKey := "test_sdk_key"

	mockRequester := new(MockRequester)
	mockRequester.On("Get", []utils.Header(nil)).Return(mockDatafile2, http.Header{}, http.StatusOK, nil)

	asyncConfigManager := NewAsyncPollingProjectConfigManager(sdkKey, WithInitialDatafile(mockDatafile1), WithRequester(mockRequester))
	config, err := asyncConfigManager.GetConfig()

	assert.Nil(t, err)
	assert.NotNil(t, config)
	assert.Equal(t, "42", config.GetRevision())
}

func TestNewPollingProjectConfigManagerPullsImmediately(t *testing.T) {
	mockDatafile1 := []byte(`{"revision":"42"}`)
	sdkKey := "test_sdk_key"

	mockRequester := new(MockRequester)
	mockRequester.On("Get", []utils.Header(nil)).Return(mockDatafile1, http.Header{}, http.StatusOK, nil)

	configManager := NewPollingProjectConfigManager(sdkKey, WithRequester(mockRequester))
	config, err := configManager.GetConfig()

	mockRequester.AssertExpectations(t)
	assert.Nil(t, err)
	assert.NotNil(t, config)
	assert.Equal(t, "42", config.GetRevision())
}

func TestNewAsyncPollingProjectConfigManagerDoesNotPullImmediately(t *testing.T) {
	mockDatafile1 := []byte(`{"revision":"42"}`)
	sdkKey := "test_sdk_key"

	mockRequester := new(MockRequester)
	mockRequester.On("Get", []utils.Header(nil)).Return(mockDatafile1, http.Header{}, http.StatusOK, nil)

	asyncConfigManager := NewAsyncPollingProjectConfigManager(sdkKey, WithRequester(mockRequester))
	config, _ := asyncConfigManager.GetConfig()
	assert.Nil(t, config)
}

func TestPollingInterval(t *testing.T) {

	sdkKey := "test_sdk_key"
	configManager := NewPollingProjectConfigManager(sdkKey, WithPollingInterval(5*time.Second))
	asyncConfigManager := NewAsyncPollingProjectConfigManager(sdkKey, WithPollingInterval(5*time.Second))

	assert.Equal(t, configManager.pollingInterval, 5*time.Second)
	assert.Equal(t, asyncConfigManager.pollingInterval, 5*time.Second)
}

func TestInitialDatafile(t *testing.T) {

	sdkKey := "test_sdk_key"
	configManager := NewPollingProjectConfigManager(sdkKey, WithInitialDatafile([]byte("test")))
	asyncConfigManager := NewAsyncPollingProjectConfigManager(sdkKey, WithInitialDatafile([]byte("test")))

	assert.Equal(t, configManager.initDatafile, []byte("test"))
	assert.Equal(t, asyncConfigManager.initDatafile, []byte("test"))
}

func TestDatafileTemplate(t *testing.T) {

	sdkKey := "test_sdk_key"
	datafileTemplate := "https://localhost/v1/%s.json"
	configManager := NewPollingProjectConfigManager(sdkKey, WithDatafileURLTemplate(datafileTemplate))
	asyncConfigManager := NewAsyncPollingProjectConfigManager(sdkKey, WithDatafileURLTemplate(datafileTemplate))

	assert.Equal(t, datafileTemplate, configManager.datafileURLTemplate)
	assert.Equal(t, datafileTemplate, asyncConfigManager.datafileURLTemplate)
}

func TestWithRequester(t *testing.T) {

	sdkKey := "test_sdk_key"
	mockDatafile := []byte(`{"revision":"42"}`)
	mockRequester := new(MockRequester)
	mockRequester.On("Get", []utils.Header(nil)).Return(mockDatafile, http.Header{}, http.StatusOK, nil)
	configManager := NewPollingProjectConfigManager(sdkKey, WithRequester(mockRequester))
	asyncConfigManager := NewAsyncPollingProjectConfigManager(sdkKey, WithRequester(mockRequester))

	assert.Equal(t, mockRequester, configManager.requester)
	assert.Equal(t, mockRequester, asyncConfigManager.requester)
}
