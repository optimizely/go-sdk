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
	"errors"
	"testing"

	"github.com/optimizely/go-sdk/pkg/config/datafileprojectconfig"
	"github.com/optimizely/go-sdk/pkg/logging"
	"github.com/optimizely/go-sdk/pkg/notification"

	"github.com/stretchr/testify/assert"
)

var tlogger = logging.GetLogger("", "staticManager")

func TestNewStaticProjectConfigManager(t *testing.T) {
	projectConfig := datafileprojectconfig.DatafileProjectConfig{}
	configManager := NewStaticProjectConfigManager(projectConfig, tlogger)

	actual, _ := configManager.GetConfig()
	assert.Equal(t, projectConfig, actual)
}

func TestNewStaticProjectConfigManagerFromPayload(t *testing.T) {

	mockDatafile := []byte(`{"accountId":"42","projectId":"123""}`)
	configManager, err := NewStaticProjectConfigManagerFromPayload(tlogger, mockDatafile)
	assert.Error(t, err)

	mockDatafile = []byte(`{"accountId":"42","projectId":"123",}`)
	configManager, err = NewStaticProjectConfigManagerFromPayload(tlogger, mockDatafile)
	assert.Error(t, err)

	mockDatafile = []byte(`{"accountId":"42","projectId":"123","version":"4"}`)
	configManager, err = NewStaticProjectConfigManagerFromPayload(tlogger, mockDatafile)
	assert.Nil(t, err)

	assert.Nil(t, configManager.optimizelyConfig)

	actual, _ := configManager.GetConfig()
	assert.NotNil(t, actual)
}

func TestStaticGetOptimizelyConfig(t *testing.T) {

	mockDatafile := []byte(`{"accountId":"42","projectId":"123","version":"4"}`)
	configManager, err := NewStaticProjectConfigManagerFromPayload(tlogger, mockDatafile)
	assert.Nil(t, err)

	assert.Nil(t, configManager.optimizelyConfig)

	optimizelyConfig := configManager.GetOptimizelyConfig()
	assert.NotNil(t, configManager.optimizelyConfig)
	assert.Equal(t, &OptimizelyConfig{ExperimentsMap: map[string]OptimizelyExperiment{},
		FeaturesMap: map[string]OptimizelyFeature{}}, optimizelyConfig)
}
func TestNewStaticProjectConfigManagerFromURL(t *testing.T) {

	configManager, err := NewStaticProjectConfigManagerFromURL("no_key_exists")
	assert.Error(t, err)
	assert.Nil(t, configManager)
}

func TestNewStaticProjectConfigManagerOnDecision(t *testing.T) {
	mockDatafile := []byte(`{"accountId":"42","projectId":"123","version":"4"}`)
	configManager, err := NewStaticProjectConfigManagerFromPayload(tlogger, mockDatafile)
	assert.Nil(t, err)

	callback := func(notification notification.ProjectConfigUpdateNotification) {

	}
	id, err := configManager.OnProjectConfigUpdate(callback)

	assert.Error(t, err)
	assert.Equal(t, err, errors.New("method OnProjectConfigUpdate does not have any effect on StaticProjectConfigManager"))
	assert.Equal(t, id, 0)

	err = configManager.RemoveOnProjectConfigUpdate(id)
	assert.Error(t, err)
	assert.Equal(t, err, errors.New("method RemoveOnProjectConfigUpdate does not have any effect on StaticProjectConfigManager"))

}
