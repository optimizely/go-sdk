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
	"errors"
	"testing"
	"time"

	of "github.com/open-feature/go-sdk/openfeature"
	"github.com/stretchr/testify/assert"

	"github.com/optimizely/go-sdk/v2/pkg/client"
	"github.com/optimizely/go-sdk/v2/pkg/config"
)

// newStaticConfigManager creates a StaticProjectConfigManager for testing.
func newStaticConfigManager(t *testing.T, datafile []byte) config.ProjectConfigManager {
	t.Helper()
	return config.NewStaticProjectConfigManagerWithOptions("", config.WithInitialDatafile(datafile))
}

func TestInitWithPreInitializedClient(t *testing.T) {
	factory := client.OptimizelyFactory{
		Datafile: []byte(`{"revision":"1","version":"4"}`),
	}
	c, err := factory.StaticClient()
	assert.NoError(t, err)
	defer c.Close()

	p := NewProviderWithClient(c)
	err = p.Init(of.EvaluationContext{})
	assert.NoError(t, err)
	assert.True(t, p.ready.Load())

	// Should emit a ready event
	select {
	case evt := <-p.EventChannel():
		assert.Equal(t, of.ProviderReady, evt.EventType)
	case <-time.After(time.Second):
		t.Fatal("expected ProviderReady event")
	}
}

func TestInitWithSDKKey(t *testing.T) {
	// Use a datafile-based approach via factory option
	datafile := []byte(`{"revision":"1","version":"4"}`)
	p := NewProvider("test-sdk-key",
		WithClientOptions(
			client.WithConfigManager(
				newStaticConfigManager(t, datafile),
			),
		),
	)

	err := p.Init(of.EvaluationContext{})
	assert.NoError(t, err)
	assert.True(t, p.ready.Load())
	assert.NotNil(t, p.client)

	// Should emit a ready event
	select {
	case evt := <-p.EventChannel():
		assert.Equal(t, of.ProviderReady, evt.EventType)
	case <-time.After(time.Second):
		t.Fatal("expected ProviderReady event")
	}

	p.Shutdown()
}

func TestShutdownWithOwnedClient(t *testing.T) {
	datafile := []byte(`{"revision":"1","version":"4"}`)
	p := NewProvider("test-sdk-key",
		WithClientOptions(
			client.WithConfigManager(
				newStaticConfigManager(t, datafile),
			),
		),
	)

	err := p.Init(of.EvaluationContext{})
	assert.NoError(t, err)
	// Drain ready event
	<-p.EventChannel()

	assert.True(t, p.ready.Load())
	p.Shutdown()
	assert.False(t, p.ready.Load())
}

func TestShutdownWithPreInitializedClientDoesNotClose(t *testing.T) {
	factory := client.OptimizelyFactory{
		Datafile: []byte(`{"revision":"1","version":"4"}`),
	}
	c, err := factory.StaticClient()
	assert.NoError(t, err)
	// We control lifecycle — don't defer Close here

	p := NewProviderWithClient(c)
	err = p.Init(of.EvaluationContext{})
	assert.NoError(t, err)
	// Drain ready event
	<-p.EventChannel()

	p.Shutdown()
	assert.False(t, p.ready.Load())

	// Client should still be usable after provider shutdown
	userCtx := c.CreateUserContext("user-1", nil)
	assert.NotNil(t, userCtx)

	c.Close()
}

func TestDoubleShutdownDoesNotPanic(t *testing.T) {
	datafile := []byte(`{"revision":"1","version":"4"}`)
	p := NewProvider("test-sdk-key",
		WithClientOptions(
			client.WithConfigManager(
				newStaticConfigManager(t, datafile),
			),
		),
	)

	err := p.Init(of.EvaluationContext{})
	assert.NoError(t, err)
	// Drain ready event
	<-p.EventChannel()

	assert.NotPanics(t, func() {
		p.Shutdown()
		p.Shutdown()
	})
}

func TestInitFailureReturnsProviderInitError(t *testing.T) {
	// Init with no config manager — will fail
	p := NewProvider("",
		WithClientOptions(
			client.WithConfigManager(nil),
		),
	)

	err := p.Init(of.EvaluationContext{})
	assert.Error(t, err)

	var initErr *of.ProviderInitError
	assert.True(t, errors.As(err, &initErr), "error must be *ProviderInitError")
	assert.Equal(t, of.ProviderFatalCode, initErr.ErrorCode,
		"error code must be PROVIDER_FATAL")
}

func TestInitFailureEmitsErrorEventWithFatalCode(t *testing.T) {
	// Init with no config manager and no SDK key — will fail
	p := NewProvider("",
		WithClientOptions(
			client.WithConfigManager(nil),
		),
	)

	_ = p.Init(of.EvaluationContext{})

	select {
	case evt := <-p.EventChannel():
		assert.Equal(t, of.ProviderError, evt.EventType)
		assert.Equal(t, of.ProviderFatalCode, evt.ProviderEventDetails.ErrorCode,
			"error event must carry PROVIDER_FATAL error code")
	case <-time.After(time.Second):
		t.Fatal("expected ProviderError event")
	}
}

func TestInitSuccessEmitsReadyEventWithNoErrorCode(t *testing.T) {
	datafile := []byte(`{"revision":"1","version":"4"}`)
	p := NewProvider("test-sdk-key",
		WithClientOptions(
			client.WithConfigManager(
				newStaticConfigManager(t, datafile),
			),
		),
	)

	err := p.Init(of.EvaluationContext{})
	assert.NoError(t, err)
	defer p.Shutdown()

	select {
	case evt := <-p.EventChannel():
		assert.Equal(t, of.ProviderReady, evt.EventType)
		assert.Equal(t, of.ErrorCode(""), evt.ProviderEventDetails.ErrorCode,
			"ready event must not carry an error code")
	case <-time.After(time.Second):
		t.Fatal("expected ProviderReady event")
	}
}

func TestEmitEventWithErrorCode(t *testing.T) {
	p := NewProvider("test-key")

	p.emitEvent(of.ProviderError, "init failed", of.ProviderFatalCode)

	select {
	case evt := <-p.EventChannel():
		assert.Equal(t, of.ProviderError, evt.EventType)
		assert.Equal(t, of.ProviderFatalCode, evt.ProviderEventDetails.ErrorCode)
		assert.Equal(t, "init failed", evt.ProviderEventDetails.Message)
	case <-time.After(time.Second):
		t.Fatal("expected error event")
	}
}

func TestEmitEventWithoutErrorCode(t *testing.T) {
	p := NewProvider("test-key")

	p.emitEvent(of.ProviderReady, "", "")

	select {
	case evt := <-p.EventChannel():
		assert.Equal(t, of.ProviderReady, evt.EventType)
		assert.Equal(t, of.ErrorCode(""), evt.ProviderEventDetails.ErrorCode)
	case <-time.After(time.Second):
		t.Fatal("expected ready event")
	}
}

func TestEventChannelReturnsChannel(t *testing.T) {
	p := NewProvider("test-key")
	ch := p.EventChannel()
	assert.NotNil(t, ch)
}
