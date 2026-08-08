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

	of "github.com/open-feature/go-sdk/openfeature"

	"github.com/optimizely/go-sdk/v2/pkg/entities"
)

// Track tracks a conversion event via the Optimizely SDK. The evaluation
// context's targeting key maps to the Optimizely user ID. Event tags are
// built by merging context attributes with tracking event details: the
// numeric value from TrackingEventDetails is set as the "revenue" tag,
// and custom attributes from TrackingEventDetails override context
// attributes on key conflict.
func (p *Provider) Track(_ context.Context, trackingEventName string, evaluationContext of.EvaluationContext, details of.TrackingEventDetails) {
	if !p.ready.Load() || p.client == nil {
		return
	}

	targetingKey := evaluationContext.TargetingKey()
	if targetingKey == "" {
		return
	}

	attrs := evaluationContext.Attributes()
	userContext := entities.UserContext{
		ID:         targetingKey,
		Attributes: attrs,
	}

	// Build event tags: start with context attributes as base
	eventTags := make(map[string]interface{})
	for k, v := range attrs {
		eventTags[k] = v
	}

	// Only set revenue when the caller provided a non-zero value, so
	// non-revenue events (e.g. page views) don't get a spurious "revenue: 0".
	if revenue := details.Value(); revenue != 0 {
		eventTags["revenue"] = revenue
	}

	// Merge custom attributes from tracking details (overrides context on conflict)
	for k, v := range details.Attributes() {
		eventTags[k] = v
	}

	p.client.Track(trackingEventName, userContext, eventTags)
}
