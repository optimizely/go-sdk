/****************************************************************************
 * Copyright 2026, Optimizely, Inc. and contributors                       *
 *                                                                          *
 * Licensed under the Apache License, Version 2.0 (the "License");          *
 * you may not use this file except in compliance with the License.         *
 * You may obtain a copy of the License at                                  *
 *                                                                          *
 *    https://www.apache.org/licenses/LICENSE-2.0                           *
 *                                                                          *
 * Unless required by applicable law or agreed to in writing, software      *
 * distributed under the License is distributed on an "AS IS" BASIS,        *
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. *
 * See the License for the specific language governing permissions and      *
 * limitations under the License.                                           *
 ***************************************************************************/

package openfeature

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/optimizely/go-sdk/v2/pkg/client"
)

func TestMetadata(t *testing.T) {
	p := NewProvider("test-key")
	meta := p.Metadata()
	assert.Equal(t, "Optimizely", meta.Name)
}

func TestHooksReturnsNil(t *testing.T) {
	p := NewProvider("test-key")
	hooks := p.Hooks()
	assert.Nil(t, hooks)
}

func TestNewProvider(t *testing.T) {
	p := NewProvider("sdk-key-123")
	assert.Equal(t, "sdk-key-123", p.sdkKey)
	assert.True(t, p.ownsClient)
	assert.Nil(t, p.client)
	assert.NotNil(t, p.eventChan)
	assert.NotNil(t, p.stopChan)
}

func TestNewProviderWithClientOptions(t *testing.T) {
	opt := client.WithBatchEventProcessor(10, 100, 0)
	p := NewProvider("sdk-key", WithClientOptions(opt))
	assert.Len(t, p.clientOptions, 1)
}

func TestNewProviderWithClient(t *testing.T) {
	// Create a real static client
	factory := client.OptimizelyFactory{
		Datafile: []byte(`{"revision":"1","version":"4"}`),
	}
	c, err := factory.StaticClient()
	assert.NoError(t, err)
	defer c.Close()

	p := NewProviderWithClient(c)
	assert.Equal(t, c, p.client)
	assert.False(t, p.ownsClient)
	assert.NotNil(t, p.eventChan)
}
