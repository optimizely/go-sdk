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
	"context"
	"testing"

	of "github.com/open-feature/go-sdk/openfeature"
	"github.com/stretchr/testify/assert"

	"github.com/optimizely/go-sdk/v2/pkg/entities"
	"github.com/optimizely/go-sdk/v2/pkg/event"
)

func TestTrackForwardsRevenueFromTrackingDetails(t *testing.T) {
	p, c := newTestProvider(t)
	defer c.Close()

	var capturedTags map[string]interface{}
	c.OnTrack(func(eventKey string, userContext entities.UserContext, eventTags map[string]interface{}, _ event.ConversionEvent) {
		capturedTags = eventTags
	})

	evalCtx := of.NewEvaluationContext("user-123", nil)
	details := of.NewTrackingEventDetails(49.99)
	p.Track(context.Background(), "purchase", evalCtx, details)

	assert.NotNil(t, capturedTags)
	assert.Equal(t, 49.99, capturedTags["revenue"])
}

func TestTrackMergesCustomAttributes(t *testing.T) {
	p, c := newTestProvider(t)
	defer c.Close()

	var capturedTags map[string]interface{}
	c.OnTrack(func(_ string, _ entities.UserContext, eventTags map[string]interface{}, _ event.ConversionEvent) {
		capturedTags = eventTags
	})

	evalCtx := of.NewEvaluationContext("user-123", map[string]interface{}{
		"plan": "enterprise",
	})
	details := of.NewTrackingEventDetails(10.0).Add("currency", "USD").Add("item_count", 3)
	p.Track(context.Background(), "purchase", evalCtx, details)

	assert.NotNil(t, capturedTags)
	assert.Equal(t, "enterprise", capturedTags["plan"], "context attributes should pass through")
	assert.Equal(t, "USD", capturedTags["currency"])
	assert.Equal(t, 3, capturedTags["item_count"])
	assert.Equal(t, 10.0, capturedTags["revenue"])
}

func TestTrackDetailsOverrideContextAttributes(t *testing.T) {
	p, c := newTestProvider(t)
	defer c.Close()

	var capturedTags map[string]interface{}
	c.OnTrack(func(_ string, _ entities.UserContext, eventTags map[string]interface{}, _ event.ConversionEvent) {
		capturedTags = eventTags
	})

	evalCtx := of.NewEvaluationContext("user-123", map[string]interface{}{
		"plan": "enterprise",
	})
	// Tracking details with conflicting key "plan"
	details := of.NewTrackingEventDetails(0).Add("plan", "premium")
	p.Track(context.Background(), "purchase", evalCtx, details)

	assert.NotNil(t, capturedTags)
	assert.Equal(t, "premium", capturedTags["plan"], "tracking details should override context attributes")
}

func TestTrackWithZeroValueOmitsRevenue(t *testing.T) {
	p, c := newTestProvider(t)
	defer c.Close()

	var capturedTags map[string]interface{}
	c.OnTrack(func(_ string, _ entities.UserContext, eventTags map[string]interface{}, _ event.ConversionEvent) {
		capturedTags = eventTags
	})

	evalCtx := of.NewEvaluationContext("user-123", nil)
	details := of.NewTrackingEventDetails(0.0)
	p.Track(context.Background(), "purchase", evalCtx, details)

	assert.NotNil(t, capturedTags)
	_, hasRevenue := capturedTags["revenue"]
	assert.False(t, hasRevenue, "zero value should not set revenue tag")
}

func TestTrackWithZeroValuePreservesContextAttributes(t *testing.T) {
	p, c := newTestProvider(t)
	defer c.Close()

	var capturedTags map[string]interface{}
	c.OnTrack(func(_ string, _ entities.UserContext, eventTags map[string]interface{}, _ event.ConversionEvent) {
		capturedTags = eventTags
	})

	evalCtx := of.NewEvaluationContext("user-123", map[string]interface{}{
		"plan": "enterprise",
	})
	details := of.NewTrackingEventDetails(0)
	p.Track(context.Background(), "purchase", evalCtx, details)

	assert.NotNil(t, capturedTags)
	assert.Equal(t, "enterprise", capturedTags["plan"], "context attributes should pass through")
	_, hasRevenue := capturedTags["revenue"]
	assert.False(t, hasRevenue, "zero-value revenue should be omitted")
}

func TestTrackWithNotReadyProvider(t *testing.T) {
	p := NewProvider("fake-key")
	evalCtx := of.NewEvaluationContext("user-123", nil)
	assert.NotPanics(t, func() {
		p.Track(context.Background(), "purchase", evalCtx, of.NewTrackingEventDetails(0))
	})
}

func TestTrackWithMissingTargetingKey(t *testing.T) {
	p, c := newTestProvider(t)
	defer c.Close()

	evalCtx := of.NewEvaluationContext("", nil)
	assert.NotPanics(t, func() {
		p.Track(context.Background(), "purchase", evalCtx, of.NewTrackingEventDetails(0))
	})
}
