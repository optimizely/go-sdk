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
	"time"

	"github.com/optimizely/go-sdk/optimizely/config/datafileProjectConfig"
	"github.com/optimizely/go-sdk/optimizely/requester"
	"github.com/stretchr/testify/assert"
)

func TestNewPollingProjectConfigManager(t *testing.T) {
	URL := "https://cdn.optimizely.com/datafiles/"
	projectConfig, _ := datafileProjectConfig.NewDatafileProjectConfig([]byte{})
	request := requester.New(URL)

	// Bad SDK Key test
	configManager := NewPollingProjectConfigManager(request, "4SLpaJA1r1pgE6T2CoMs9q_bad", []byte{}, 0)
	assert.Equal(t, projectConfig, configManager.GetConfig())
	assert.Equal(t, "[bad_http_request:1, failed_project_config:1]", configManager.GetMetrics())

	// Good SDK Key test
	configManager = NewPollingProjectConfigManager(request, "4SLpaJA1r1pgE6T2CoMs9q", []byte{}, 0)
	newConfig := configManager.GetConfig()

	assert.Equal(t, "", newConfig.GetAccountID())
	assert.Equal(t, 3, len(newConfig.GetAudienceMap()))
	assert.Equal(t, "", configManager.GetMetrics())

}

func TestPollingMetrics(t *testing.T) {
	URL := "https://cdn.optimizely.com/datafiles/"
	request := requester.New(URL)

	// Good SDK Key test -- number of polling
	configManager := NewPollingProjectConfigManager(request, "4SLpaJA1r1pgE6T2CoMs9q", []byte{}, 5*time.Second)
	time.Sleep(14 * time.Second)
	assert.Equal(t, "[polls:3]", configManager.GetMetrics())

}
