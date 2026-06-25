/****************************************************************************
 * Copyright 2026, Optimizely, Inc. and contributors                        *
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

// Package event //
package event

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/optimizely/go-sdk/v2/pkg/config"
	decisionPkg "github.com/optimizely/go-sdk/v2/pkg/decision"
	"github.com/optimizely/go-sdk/v2/pkg/entities"
)

// Decision-event ID normalization tests.
//
// These tests verify the cross-SDK contract for outgoing decision events:
//   - campaign_id / entity_id: non-empty string (any character content;
//     opaque IDs allowed). Fallback to experiment_id ONLY when empty.
//   - variation_id: STRICT non-empty numeric string OR JSON null.
//   - entity_id (impression events) mirrors campaign_id byte-for-byte.
//   - rules apply uniformly to every decision type.
//   - no logging, no warning, no dropping of events on the normalization path.

// --- IsNumericIDString -------------------------------------------------------

func TestIsNumericIDString(t *testing.T) {
	cases := []struct {
		name  string
		value string
		want  bool
	}{
		{"plain digits", "12345", true},
		{"single digit", "0", true},
		{"all zeros", "0000", true},
		{"leading zero numeric", "0123456789", true},
		{"all nines", "9999999999", true},
		{"large numeric id", "15402980349", true},

		{"empty string is invalid", "", false},
		{"whitespace only", "   ", false},
		{"leading space", " 123", false},
		{"trailing space", "123 ", false},
		{"interior space", "12 34", false},
		{"alpha placeholder", "holdout_var_123", false},
		{"alpha only", "abc", false},
		{"mixed digits and letters", "12a34", false},
		{"negative sign", "-123", false},
		{"positive sign", "+123", false},
		{"decimal point", "12.34", false},
		{"exponent notation", "1e5", false},
		{"tab character", "12\t34", false},
		{"newline character", "12\n34", false},
		{"unicode digits", "१२३", false}, // devanagari digits are not [0-9]
		{"hex prefix", "0x123", false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, IsNumericIDString(tc.value))
		})
	}
}

// --- IsNonEmptyString --------------------------------------------------------

func TestIsNonEmptyString(t *testing.T) {
	// Relaxed campaign_id / entity_id predicate. Only the empty string is
	// rejected — any other character content is accepted, including
	// whitespace-only, opaque IDs, and non-numeric content.
	cases := []struct {
		name  string
		value string
		want  bool
	}{
		{"plain digits", "12345", true},
		{"opaque layer id", "layer_abc", true},
		{"opaque default id", "default-12345", true},
		{"alpha placeholder", "holdout_123", true},
		{"single character", "a", true},
		{"whitespace only", "   ", true},
		{"leading zero numeric", "0123456789", true},
		{"unicode", "१२३", true},

		{"empty string is invalid", "", false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, IsNonEmptyString(tc.value))
		})
	}
}

// --- NormalizeCampaignID -----------------------------------------------------

func TestNormalizeCampaignID_ReturnsCampaignIDWhenNumeric(t *testing.T) {
	got := NormalizeCampaignID("15399420423", "15402980349")
	assert.Equal(t, "15399420423", got)
}

func TestNormalizeCampaignID_SubstitutesExperimentIDWhenCampaignEmpty(t *testing.T) {
	// FR-001/FR-002 (relaxed): empty campaign_id is replaced with experiment_id.
	// Fallback fires ONLY for the empty string.
	got := NormalizeCampaignID("", "15402980349")
	assert.Equal(t, "15402980349", got)
}

func TestNormalizeCampaignID_PassesThroughOpaqueCampaignID(t *testing.T) {
	// Relaxed spec: opaque, non-numeric campaign IDs pass through unchanged.
	// IDs may be of the form "layer_abc", "default-12345", etc.
	got := NormalizeCampaignID("layer_abc", "15402980349")
	assert.Equal(t, "layer_abc", got)
}

func TestNormalizeCampaignID_PassesThroughHoldoutPlaceholder(t *testing.T) {
	// Relaxed spec: alphanumeric placeholders like "holdout_123" are valid
	// campaign IDs and pass through; experiment_id substitution does NOT fire.
	got := NormalizeCampaignID("holdout_123", "15402980349")
	assert.Equal(t, "holdout_123", got)
}

func TestNormalizeCampaignID_PassesThroughWhitespaceCampaign(t *testing.T) {
	// Relaxed spec: whitespace-only is non-empty, so it passes through.
	// This is intentionally permissive — we no longer try to "fix" content.
	got := NormalizeCampaignID("   ", "15402980349")
	assert.Equal(t, "   ", got)
}

func TestNormalizeCampaignID_AllowsLeadingZeros(t *testing.T) {
	got := NormalizeCampaignID("0123", "9999")
	assert.Equal(t, "0123", got)
}

func TestNormalizeCampaignID_PreservesExperimentIDFallback(t *testing.T) {
	// When campaign_id is empty, fall back to experiment_id verbatim — even
	// when experiment_id itself is non-numeric. Downstream behavior for
	// malformed datafiles is unchanged.
	got := NormalizeCampaignID("", "exp_42")
	assert.Equal(t, "exp_42", got)
}

// --- NormalizeVariationID ----------------------------------------------------

func TestNormalizeVariationID_ReturnsPointerWhenNumeric(t *testing.T) {
	got := NormalizeVariationID("15410990633")
	if assert.NotNil(t, got) {
		assert.Equal(t, "15410990633", *got)
	}
}

func TestNormalizeVariationID_ReturnsNilWhenEmpty(t *testing.T) {
	// FR-003/FR-004: empty variation_id becomes JSON null.
	got := NormalizeVariationID("")
	assert.Nil(t, got)
}

func TestNormalizeVariationID_ReturnsNilWhenNonNumeric(t *testing.T) {
	got := NormalizeVariationID("variation_a")
	assert.Nil(t, got)
}

func TestNormalizeVariationID_ReturnsNilForWhitespace(t *testing.T) {
	got := NormalizeVariationID("   ")
	assert.Nil(t, got)
}

func TestNormalizeVariationID_AllowsLeadingZeros(t *testing.T) {
	got := NormalizeVariationID("0042")
	if assert.NotNil(t, got) {
		assert.Equal(t, "0042", *got)
	}
}

func TestNormalizeVariationID_NilSerializesToJSONNull(t *testing.T) {
	// FR-003/FR-004: variation_id with nil pointer must marshal to JSON null,
	// matching the cross-SDK wire contract.
	decision := Decision{
		CampaignID:   "15402980349",
		ExperimentID: "15402980349",
		VariationID:  NormalizeVariationID(""), // nil
	}

	b, err := json.Marshal(decision)
	assert.NoError(t, err)

	var raw map[string]interface{}
	assert.NoError(t, json.Unmarshal(b, &raw))

	// json.Unmarshal of null into interface{} produces a nil value, and the
	// key must still be present.
	got, ok := raw["variation_id"]
	assert.True(t, ok, "variation_id key must be present in JSON output")
	assert.Nil(t, got, "variation_id must serialize as JSON null when nil")
}

func TestNormalizeVariationID_NumericSerializesToJSONString(t *testing.T) {
	decision := Decision{
		CampaignID:   "15402980349",
		ExperimentID: "15402980349",
		VariationID:  NormalizeVariationID("15410990633"),
	}

	b, err := json.Marshal(decision)
	assert.NoError(t, err)

	var raw map[string]interface{}
	assert.NoError(t, json.Unmarshal(b, &raw))
	assert.Equal(t, "15410990633", raw["variation_id"])
}

// --- Integration: createImpressionEvent + createImpressionVisitor ------------

// testNormConfig is a minimal ProjectConfig stub for normalization integration
// tests. It is intentionally separate from TestConfig in factory_test.go so
// these tests do not depend on shared global state.
type testNormConfig struct {
	config.ProjectConfig
}

func (testNormConfig) GetAttributeByKey(string) (entities.Attribute, error) {
	return entities.Attribute{ID: "100000", Key: "sample_attribute"}, nil
}
func (testNormConfig) GetProjectID() string         { return "15389410617" }
func (testNormConfig) GetRevision() string          { return "7" }
func (testNormConfig) GetAccountID() string         { return "8362480420" }
func (testNormConfig) GetAnonymizeIP() bool         { return true }
func (testNormConfig) GetAttributeID(string) string { return "" }
func (testNormConfig) GetBotFiltering() bool        { return false }
func (testNormConfig) GetClientName() string        { return "go-sdk" }
func (testNormConfig) GetClientVersion() string     { return "1.0.0" }
func (testNormConfig) SendFlagDecisions() bool      { return true }
func (testNormConfig) GetRegion() string            { return "US" }

func newNormUserContext() entities.UserContext {
	return entities.UserContext{ID: "user-1", Attributes: map[string]interface{}{}}
}

func numericExperiment() entities.Experiment {
	return entities.Experiment{Key: "exp_key", LayerID: "15399420423", ID: "15402980349"}
}

func numericVariation() entities.Variation {
	return entities.Variation{Key: "variation_a", ID: "15410990633"}
}

func holdoutExperiment() entities.Experiment {
	// Holdouts ship with no LayerID. Normalizer must fall back to
	// ExperimentID for both campaign_id and entity_id.
	return entities.Experiment{Key: "holdout_key", LayerID: "", ID: "9876543210"}
}

// TestImpressionEvent_NormalizesCampaignAndEntityIDsForHoldout verifies
// FR-001/FR-002 and FR-009 in the impression event flow for holdouts.
func TestImpressionEvent_NormalizesCampaignAndEntityIDsForHoldout(t *testing.T) {
	tc := testNormConfig{}
	exp := holdoutExperiment()
	variation := numericVariation()

	userEvent, ok := CreateImpressionUserEvent(tc, exp, &variation, newNormUserContext(),
		"flag_key", exp.Key, decisionPkg.Holdout, false, nil)
	assert.True(t, ok)

	// FR-001/FR-002: empty LayerID is substituted with ExperimentID.
	assert.Equal(t, exp.ID, userEvent.Impression.CampaignID,
		"holdout campaign_id must fall back to experiment_id when LayerID is empty")
	// FR-009: entity_id mirrors campaign_id byte-for-byte.
	assert.Equal(t, exp.ID, userEvent.Impression.EntityID,
		"holdout entity_id must equal campaign_id byte-for-byte")

	// Same invariant must hold in the wire visitor / decision payload.
	visitor := createVisitorFromUserEvent(userEvent)
	assert.Len(t, visitor.Snapshots, 1)
	assert.Len(t, visitor.Snapshots[0].Decisions, 1)
	assert.Equal(t, exp.ID, visitor.Snapshots[0].Decisions[0].CampaignID)
	assert.Equal(t, exp.ID, visitor.Snapshots[0].Events[0].EntityID)
	// Byte-equivalent (FR-009).
	assert.Equal(t, visitor.Snapshots[0].Decisions[0].CampaignID,
		visitor.Snapshots[0].Events[0].EntityID,
		"decisions[].campaign_id and events[].entity_id must be byte-equivalent")
}

// TestImpressionEvent_PassesThroughOpaqueLayerID verifies the relaxed spec:
// a non-numeric but non-empty LayerID (e.g. "layer_abc", "default-12345")
// passes through to campaign_id and entity_id unchanged. Fallback to
// experiment_id fires ONLY when LayerID is the empty string.
func TestImpressionEvent_PassesThroughOpaqueLayerID(t *testing.T) {
	tc := testNormConfig{}
	exp := entities.Experiment{Key: "exp_key", LayerID: "default-12345", ID: "15402980349"}
	variation := numericVariation()

	userEvent, ok := CreateImpressionUserEvent(tc, exp, &variation, newNormUserContext(),
		"flag_key", exp.Key, decisionPkg.FeatureTest, true, nil)
	assert.True(t, ok)

	// Opaque LayerID passes through under the relaxed spec — NOT substituted
	// with experiment_id.
	assert.Equal(t, "default-12345", userEvent.Impression.CampaignID,
		"opaque LayerID must pass through (relaxed contract)")
	assert.Equal(t, "default-12345", userEvent.Impression.EntityID,
		"entity_id must mirror campaign_id (FR-009)")

	visitor := createVisitorFromUserEvent(userEvent)
	assert.Equal(t, "default-12345", visitor.Snapshots[0].Decisions[0].CampaignID)
	assert.Equal(t, "default-12345", visitor.Snapshots[0].Events[0].EntityID)
	assert.Equal(t, visitor.Snapshots[0].Decisions[0].CampaignID,
		visitor.Snapshots[0].Events[0].EntityID)
}

// TestImpressionEvent_PassesThroughNumericCampaignID verifies happy path for
// non-holdout decisions where LayerID is already a numeric ID.
func TestImpressionEvent_PassesThroughNumericCampaignID(t *testing.T) {
	tc := testNormConfig{}
	exp := numericExperiment()
	variation := numericVariation()

	userEvent, ok := CreateImpressionUserEvent(tc, exp, &variation, newNormUserContext(),
		"flag_key", exp.Key, decisionPkg.FeatureTest, true, nil)
	assert.True(t, ok)

	assert.Equal(t, exp.LayerID, userEvent.Impression.CampaignID)
	assert.Equal(t, exp.LayerID, userEvent.Impression.EntityID)

	visitor := createVisitorFromUserEvent(userEvent)
	assert.Equal(t, exp.LayerID, visitor.Snapshots[0].Decisions[0].CampaignID)
	assert.Equal(t, exp.LayerID, visitor.Snapshots[0].Events[0].EntityID)
}

// TestImpressionEvent_NormalizesVariationIDToJSONNull verifies FR-003/FR-004:
// a non-numeric variation ID must become JSON null on the wire.
func TestImpressionEvent_NormalizesVariationIDToJSONNull(t *testing.T) {
	tc := testNormConfig{}
	exp := numericExperiment()
	badVariation := entities.Variation{Key: "variation_a", ID: "variation_a"} // non-numeric ID

	userEvent, ok := CreateImpressionUserEvent(tc, exp, &badVariation, newNormUserContext(),
		"flag_key", exp.Key, decisionPkg.FeatureTest, true, nil)
	assert.True(t, ok)

	visitor := createVisitorFromUserEvent(userEvent)
	assert.Nil(t, visitor.Snapshots[0].Decisions[0].VariationID,
		"non-numeric variation_id must be normalized to nil so it serializes as JSON null")

	// Verify on-the-wire JSON shape.
	b, err := json.Marshal(visitor.Snapshots[0].Decisions[0])
	assert.NoError(t, err)
	var raw map[string]interface{}
	assert.NoError(t, json.Unmarshal(b, &raw))
	got, ok := raw["variation_id"]
	assert.True(t, ok)
	assert.Nil(t, got, "variation_id must marshal as JSON null when normalized")
}

// TestImpressionEvent_KeepsNumericVariationID verifies happy path for
// variation IDs that are already valid numeric strings.
func TestImpressionEvent_KeepsNumericVariationID(t *testing.T) {
	tc := testNormConfig{}
	exp := numericExperiment()
	variation := numericVariation()

	userEvent, ok := CreateImpressionUserEvent(tc, exp, &variation, newNormUserContext(),
		"flag_key", exp.Key, decisionPkg.FeatureTest, true, nil)
	assert.True(t, ok)

	visitor := createVisitorFromUserEvent(userEvent)
	if assert.NotNil(t, visitor.Snapshots[0].Decisions[0].VariationID) {
		assert.Equal(t, variation.ID, *visitor.Snapshots[0].Decisions[0].VariationID)
	}
}

// TestImpressionEvent_NormalizationAppliesUniformlyAcrossRuleTypes verifies
// FR-005: identical normalization is applied for every decision type. No
// per-type branching is permitted in the normalization path.
func TestImpressionEvent_NormalizationAppliesUniformlyAcrossRuleTypes(t *testing.T) {
	tc := testNormConfig{}
	exp := holdoutExperiment() // empty LayerID forces normalization
	variation := numericVariation()

	ruleTypes := []string{
		decisionPkg.FeatureTest,
		decisionPkg.Holdout,
		decisionPkg.Rollout,
		"experiment",
		"anything-else",
	}

	for _, rt := range ruleTypes {
		t.Run(rt, func(t *testing.T) {
			userEvent, ok := CreateImpressionUserEvent(tc, exp, &variation, newNormUserContext(),
				"flag_key", exp.Key, rt, true, nil)
			if !ok {
				t.Skipf("rule type %q does not emit impressions in this config", rt)
				return
			}

			visitor := createVisitorFromUserEvent(userEvent)
			d := visitor.Snapshots[0].Decisions[0]
			e := visitor.Snapshots[0].Events[0]

			// Same normalization for every rule type.
			assert.Equal(t, exp.ID, d.CampaignID, "rule type %q: campaign_id", rt)
			assert.Equal(t, exp.ID, e.EntityID, "rule type %q: entity_id", rt)
			assert.Equal(t, d.CampaignID, e.EntityID, "rule type %q: byte-equivalent", rt)
			if assert.NotNil(t, d.VariationID, "rule type %q: variation_id pointer", rt) {
				assert.Equal(t, variation.ID, *d.VariationID, "rule type %q: variation_id value", rt)
			}
		})
	}
}

// TestImpressionEvent_NormalizationIsByteEquivalentOnTheWire is a full
// wire-output regression. It marshals a complete Batch and asserts the
// normalized campaign_id and entity_id are identical strings in the JSON.
func TestImpressionEvent_NormalizationIsByteEquivalentOnTheWire(t *testing.T) {
	tc := testNormConfig{}
	exp := holdoutExperiment()
	variation := numericVariation()

	userEvent, ok := CreateImpressionUserEvent(tc, exp, &variation, newNormUserContext(),
		"flag_key", exp.Key, decisionPkg.Holdout, false, nil)
	assert.True(t, ok)

	batch := createBatchEvent(userEvent, createVisitorFromUserEvent(userEvent))
	b, err := json.Marshal(batch)
	assert.NoError(t, err)

	// Decode opaquely and inspect the relevant fields by path. Avoids
	// coupling the test to internal Go struct layout.
	var raw map[string]interface{}
	assert.NoError(t, json.Unmarshal(b, &raw))

	visitors, _ := raw["visitors"].([]interface{})
	if !assert.Len(t, visitors, 1) {
		return
	}
	snapshots, _ := visitors[0].(map[string]interface{})["snapshots"].([]interface{})
	if !assert.Len(t, snapshots, 1) {
		return
	}
	decisions, _ := snapshots[0].(map[string]interface{})["decisions"].([]interface{})
	events, _ := snapshots[0].(map[string]interface{})["events"].([]interface{})
	if !assert.Len(t, decisions, 1) || !assert.Len(t, events, 1) {
		return
	}

	campaignID, _ := decisions[0].(map[string]interface{})["campaign_id"].(string)
	entityID, _ := events[0].(map[string]interface{})["entity_id"].(string)

	assert.Equal(t, exp.ID, campaignID, "wire campaign_id must equal normalized experiment id")
	assert.Equal(t, exp.ID, entityID, "wire entity_id must equal normalized campaign id")
	assert.Equal(t, campaignID, entityID, "wire campaign_id and entity_id must be byte-equivalent (FR-009)")
}

// TestConversionEvent_EntityIDIsUnchanged verifies FR-010: conversion events
// derive entity_id from the event definition (event.ID), not from a
// decision. Normalization must NOT touch the conversion path.
func TestConversionEvent_EntityIDIsUnchanged(t *testing.T) {
	tc := testNormConfig{}
	conv := entities.Event{ID: "15368860886", Key: "sample_conversion"}

	userEvent := CreateConversionUserEvent(tc, conv, newNormUserContext(), map[string]interface{}{})
	visitor := createVisitorFromUserEvent(userEvent)

	assert.Equal(t, conv.ID, visitor.Snapshots[0].Events[0].EntityID,
		"conversion entity_id must be the event ID untouched (FR-010)")
	// And conversion visitors do not emit a decisions array, so no
	// campaign_id / variation_id normalization happens here.
	assert.Empty(t, visitor.Snapshots[0].Decisions)
}
