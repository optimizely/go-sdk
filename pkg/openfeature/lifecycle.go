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
	"fmt"

	of "github.com/open-feature/go-sdk/openfeature"

	"github.com/optimizely/go-sdk/v2/pkg/client"
	"github.com/optimizely/go-sdk/v2/pkg/notification"
)

// Init initializes the provider. In SDK-key mode, this creates the
// underlying Optimizely client. In pre-initialized client mode, this
// marks the provider as ready immediately.
func (p *Provider) Init(_ of.EvaluationContext) error {
	if p.ownsClient {
		factory := &client.OptimizelyFactory{SDKKey: p.sdkKey}
		c, err := factory.Client(p.clientOptions...)
		if err != nil {
			msg := fmt.Sprintf("failed to initialize Optimizely client: %v", err)
			p.emitEvent(of.ProviderError, msg, of.ProviderFatalCode)
			return &of.ProviderInitError{
				ErrorCode: of.ProviderFatalCode,
				Message:   msg,
			}
		}
		p.client = c
	}

	p.registerConfigUpdateListener()
	p.ready.Store(true)
	p.emitEvent(of.ProviderReady, "", "")
	return nil
}

// Shutdown shuts down the provider. In SDK-key mode (ownsClient=true),
// this calls Close on the Optimizely client. In pre-initialized mode,
// the caller retains lifecycle ownership.
func (p *Provider) Shutdown() {
	p.ready.Store(false)

	// Signal stop to background goroutines
	select {
	case <-p.stopChan:
		// already closed
	default:
		close(p.stopChan)
	}

	if p.ownsClient && p.client != nil {
		p.client.Close()
	}
}

// EventChannel returns the provider's event channel for emitting
// lifecycle events (ready, error, configuration changed).
func (p *Provider) EventChannel() <-chan of.Event {
	return p.eventChan
}

// emitEvent sends a non-blocking event to the event channel. For error events,
// errorCode should be set to the appropriate ErrorCode (e.g., ProviderFatalCode).
// For non-error events, pass an empty string.
func (p *Provider) emitEvent(eventType of.EventType, message string, errorCode of.ErrorCode) {
	evt := of.Event{
		ProviderName: p.Metadata().Name,
		EventType:    eventType,
		ProviderEventDetails: of.ProviderEventDetails{
			Message:   message,
			ErrorCode: errorCode,
		},
	}
	select {
	case p.eventChan <- evt:
	default:
		// Channel full — drop event to avoid blocking
	}
}

// registerConfigUpdateListener subscribes to the Optimizely notification
// center for ProjectConfigUpdate events and forwards them as
// ProviderConfigChange events on the OpenFeature event channel.
func (p *Provider) registerConfigUpdateListener() {
	if p.client == nil {
		return
	}

	nc := p.client.GetNotificationCenter()
	if nc == nil {
		return
	}

	_, _ = nc.AddHandler(notification.ProjectConfigUpdate, func(_ interface{}) {
		select {
		case <-p.stopChan:
			return
		default:
			p.emitEvent(of.ProviderConfigChange, "datafile updated", "")
		}
	})
}
