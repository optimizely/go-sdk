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

// Decision event ID normalization.
//
// These helpers normalize the campaign_id, variation_id, and entity_id fields
// of outgoing decision events so that every decision type (experiment, feature
// test, rollout, holdout) produces uniform, valid wire output regardless of
// what upstream code supplies.
//
// Contract:
//   - campaign_id / entity_id: must be a non-empty string. Any character
//     content is allowed (IDs may be opaque, e.g. "default-12345", "layer_abc").
//     When the value is empty (""), fall back to experiment_id.
//   - variation_id: STRICT numeric-string-only. Must be a non-empty string
//     consisting entirely of decimal digits [0-9]. Anything else (empty,
//     whitespace, non-numeric) becomes nil so it serializes as JSON null.
//   - entity_id (impression events only): same rule as campaign_id, and must
//     equal the normalized campaign_id byte-for-byte for a given impression.
//
// Numeric string definition (for variation_id only): a non-empty string
// consisting entirely of decimal digits [0-9]. Leading zeros are allowed.
// Whitespace, signs, decimals, and exponents are NOT numeric strings.
//
// This normalization MUST NOT log, warn, drop, defer, or fail event dispatch.

// IsNumericIDString reports whether value is a non-empty string consisting
// entirely of decimal digits [0-9]. Leading zeros are permitted. Whitespace,
// signs (+/-), decimals, and exponents are rejected.
//
// This predicate is used for the strict variation_id contract only. The
// campaign_id / entity_id contract has been relaxed to IsNonEmptyString.
func IsNumericIDString(value string) bool {
	if value == "" {
		return false
	}
	for i := 0; i < len(value); i++ {
		c := value[i]
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

// IsNonEmptyString reports whether value has at least one character. It is
// the predicate used for the relaxed campaign_id / entity_id contract: any
// character content is allowed (IDs may be opaque, e.g. "default-12345",
// "layer_abc"). Only the empty string triggers fallback to experiment_id.
func IsNonEmptyString(value string) bool {
	return value != ""
}

// NormalizeCampaignID returns campaignID if it is a non-empty string;
// otherwise it returns experimentID. Any non-empty character content is
// accepted for campaign_id / entity_id — IDs may be opaque (e.g.
// "default-12345", "layer_abc"). Fallback to experiment_id fires ONLY when
// campaignID is the empty string.
//
// This function is used for both decisions[].campaign_id and (for impression
// events) events[].entity_id so both fields are normalized identically and
// remain byte-equivalent.
func NormalizeCampaignID(campaignID, experimentID string) string {
	if IsNonEmptyString(campaignID) {
		return campaignID
	}
	return experimentID
}

// NormalizeVariationID returns a pointer to variationID if it is a numeric
// string; otherwise it returns nil. A nil return value causes the field to
// serialize as JSON null when marshaled via the wire format. The strict
// numeric-string contract is intentional — variation_id remains numeric-only.
func NormalizeVariationID(variationID string) *string {
	if IsNumericIDString(variationID) {
		v := variationID
		return &v
	}
	return nil
}
