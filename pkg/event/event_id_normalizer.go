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

// Decision event ID normalization (FSSDK-12813).
//
// These helpers normalize the campaign_id, variation_id, and entity_id fields
// of outgoing decision events so that every decision type (experiment, feature
// test, rollout, holdout) produces uniform, valid wire output regardless of
// what upstream code supplies.
//
// Numeric string definition: a non-empty string consisting entirely of decimal
// digits [0-9]. Leading zeros are allowed. Whitespace, signs, decimals, and
// exponents are NOT numeric strings.
//
// Normalization rules:
//   - campaign_id: must be a numeric string. If empty / non-numeric, substitute
//     experiment_id (which already comes from datafile and is expected numeric).
//   - variation_id: must be a numeric string OR JSON null. If empty /
//     non-numeric, substitute nil so it serializes as JSON null.
//   - entity_id (impression events only): same rule as campaign_id, and must
//     equal the normalized campaign_id byte-for-byte for a given impression.
//
// This normalization MUST NOT log, warn, drop, defer, or fail event dispatch.

// IsNumericIDString reports whether value is a non-empty string consisting
// entirely of decimal digits [0-9]. Leading zeros are permitted. Whitespace,
// signs (+/-), decimals, and exponents are rejected.
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

// NormalizeCampaignID returns campaignID if it is a numeric string; otherwise
// it returns experimentID. The caller is responsible for ensuring experimentID
// is itself a valid numeric ID (it comes from the datafile and is expected to
// be numeric). If experimentID is also non-numeric, it is returned as-is so
// downstream behavior is unchanged for malformed datafiles.
//
// This function is used for both decisions[].campaign_id and (for impression
// events) events[].entity_id so both fields are normalized identically and
// remain byte-equivalent.
func NormalizeCampaignID(campaignID, experimentID string) string {
	if IsNumericIDString(campaignID) {
		return campaignID
	}
	return experimentID
}

// NormalizeVariationID returns a pointer to variationID if it is a numeric
// string; otherwise it returns nil. A nil return value causes the field to
// serialize as JSON null when marshaled via the wire format.
func NormalizeVariationID(variationID string) *string {
	if IsNumericIDString(variationID) {
		v := variationID
		return &v
	}
	return nil
}
