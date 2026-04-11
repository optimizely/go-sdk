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

// Package openfeature provides an OpenFeature-compatible provider for the
// Optimizely Go SDK, allowing feature flag evaluation through the
// OpenFeature standard interface.
package openfeature

import (
	"sync/atomic"

	"github.com/open-feature/go-sdk/openfeature"

	"github.com/optimizely/go-sdk/v2/pkg/client"
)

// Compile-time interface assertions.
var _ openfeature.FeatureProvider = (*Provider)(nil)
var _ openfeature.StateHandler = (*Provider)(nil)
var _ openfeature.EventHandler = (*Provider)(nil)
var _ openfeature.Tracker = (*Provider)(nil)

// Provider is an OpenFeature provider that delegates flag evaluation to an
// Optimizely Go SDK client. It supports two construction modes:
//   - NewProvider: creates and owns an Optimizely client (manages its lifecycle)
//   - NewProviderWithClient: wraps a pre-initialized client (caller manages lifecycle)
type Provider struct {
	client        *client.OptimizelyClient
	ownsClient    bool
	sdkKey        string
	clientOptions []client.OptionFunc
	eventChan     chan openfeature.Event
	ready         atomic.Bool
	stopChan      chan struct{}
}

// providerConfig holds configuration for the provider.
type providerConfig struct {
	clientOptions []client.OptionFunc
}

// ProviderOption configures the provider.
type ProviderOption func(*providerConfig)

// WithClientOptions passes Optimizely factory options to the underlying client
// created by NewProvider. These options are ignored when using NewProviderWithClient.
func WithClientOptions(opts ...client.OptionFunc) ProviderOption {
	return func(cfg *providerConfig) {
		cfg.clientOptions = append(cfg.clientOptions, opts...)
	}
}

// NewProvider creates a provider that creates and owns an Optimizely client
// using the given SDK key. The provider manages the client lifecycle: Init
// creates the client and Shutdown calls Close on it.
func NewProvider(sdkKey string, opts ...ProviderOption) *Provider {
	cfg := &providerConfig{}
	for _, opt := range opts {
		opt(cfg)
	}
	return &Provider{
		sdkKey:        sdkKey,
		ownsClient:    true,
		clientOptions: cfg.clientOptions,
		eventChan:     make(chan openfeature.Event, 5),
		stopChan:      make(chan struct{}),
	}
}

// NewProviderWithClient creates a provider that wraps a pre-initialized
// OptimizelyClient. The provider does NOT own the client — the caller is
// responsible for closing it. Shutdown will not call Close on the client.
func NewProviderWithClient(c *client.OptimizelyClient) *Provider {
	return &Provider{
		client:     c,
		ownsClient: false,
		eventChan:  make(chan openfeature.Event, 5),
		stopChan:   make(chan struct{}),
	}
}

// Metadata returns the provider's metadata.
func (p *Provider) Metadata() openfeature.Metadata {
	return openfeature.Metadata{Name: "Optimizely"}
}

// Hooks returns the provider's hooks. The Optimizely provider does not
// implement any provider-level hooks.
func (p *Provider) Hooks() []openfeature.Hook {
	return nil
}
