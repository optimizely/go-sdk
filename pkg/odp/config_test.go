/****************************************************************************
 * Copyright 2022, Optimizely, Inc. and contributors                        *
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

// Package odp //
package odp

import (
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/suite"
)

type ConfigTestSuite struct {
	suite.Suite
	apiHost, apiKey string
	segmentsToCheck []string
	config          *Config
}

func (c *ConfigTestSuite) SetupTest() {
	c.apiHost = "test-host"
	c.apiKey = "test-api-key"
	c.segmentsToCheck = []string{"a", "b", "c"}
	c.config = NewConfig(c.apiKey, c.apiHost, c.segmentsToCheck)
}

func (c *ConfigTestSuite) TestNewConfigWithValidValues() {
	c.Equal(c.apiHost, c.config.GetAPIHost())
	c.Equal(c.apiKey, c.config.GetAPIKey())
	c.Equal(c.segmentsToCheck, c.config.GetSegmentsToCheck())
	c.True(c.config.IsEventQueueingAllowed())
	c.Equal(NotDetermined, c.config.odpServiceIntegrated)
}

func (c *ConfigTestSuite) TestNewConfigWithEmptyValues() {
	c.config = NewConfig("", "", nil)
	c.Equal("", c.config.GetAPIHost())
	c.Equal("", c.config.GetAPIKey())
	c.Nil(c.config.GetSegmentsToCheck())
	c.True(c.config.IsEventQueueingAllowed())
}

func (c *ConfigTestSuite) TestUpdateWithValidValues() {
	expectedAPIKey := "1"
	expectedAPIHost := "2"
	expectedSegmentsList := []string{"1", "2", "3"}
	c.True(c.config.Update(expectedAPIKey, expectedAPIHost, expectedSegmentsList))
	c.Equal(expectedAPIKey, c.config.GetAPIKey())
	c.Equal(expectedAPIHost, c.config.GetAPIHost())
	c.Equal(expectedSegmentsList, c.config.GetSegmentsToCheck())
	c.Equal(Integrated, c.config.odpServiceIntegrated)
	c.True(c.config.IsEventQueueingAllowed())
}

func (c *ConfigTestSuite) TestUpdateWithEmptyValues() {
	expectedAPIKey := "1"
	expectedAPIHost := ""
	var expectedSegmentsList []string
	c.True(c.config.Update(expectedAPIKey, expectedAPIHost, expectedSegmentsList))
	c.Equal(expectedAPIKey, c.config.GetAPIKey())
	c.Equal(expectedAPIHost, c.config.GetAPIHost())
	c.Equal(expectedSegmentsList, c.config.GetSegmentsToCheck())
	c.Equal(NotIntegrated, c.config.odpServiceIntegrated)
	c.False(c.config.IsEventQueueingAllowed())
}

func (c *ConfigTestSuite) TestUpdateWithSameValues() {
	c.False(c.config.Update(c.apiKey, c.apiHost, c.segmentsToCheck))
	c.Equal(c.apiKey, c.config.GetAPIKey())
	c.Equal(c.apiHost, c.config.GetAPIHost())
	c.Equal(c.segmentsToCheck, c.config.GetSegmentsToCheck())
	c.Equal(Integrated, c.config.odpServiceIntegrated)
	c.True(c.config.IsEventQueueingAllowed())
}

func (c *ConfigTestSuite) TestRaceCondition() {
	wg := sync.WaitGroup{}
	update := func(i int) {
		v := fmt.Sprintf("%d", i)
		c.config.Update(v, v, []string{v})
		wg.Done()
	}

	getAPIKey := func() {
		c.config.GetAPIKey()
		wg.Done()
	}

	getAPIHost := func() {
		c.config.GetAPIHost()
		wg.Done()
	}

	getSegmentsToCheck := func() {
		c.config.GetSegmentsToCheck()
		wg.Done()
	}

	isEventQueueingAllowed := func() {
		c.config.IsEventQueueingAllowed()
		wg.Done()
	}

	iterations := 5
	wg.Add(iterations * 5)

	for i := 0; i < iterations; i++ {
		go update(i)
		go getAPIKey()
		go getAPIHost()
		go getSegmentsToCheck()
		go isEventQueueingAllowed()
	}
	wg.Wait()
}

func TestConfigTestSuite(t *testing.T) {
	suite.Run(t, new(ConfigTestSuite))
}
