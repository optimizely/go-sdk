/****************************************************************************
 * Copyright 2025, Optimizely, Inc. and contributors                        *
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

package mappers

import (
	"strings"
	"sync"
	"testing"

	datafileEntities "github.com/optimizely/go-sdk/v2/pkg/config/datafileprojectconfig/entities"
	"github.com/optimizely/go-sdk/v2/pkg/entities"
	"github.com/optimizely/go-sdk/v2/pkg/logging"
	"github.com/stretchr/testify/assert"
)

// captureLogger is a minimal OptimizelyLogProducer used by the tests to capture
// log messages without depending on the real logger plumbing.
type captureLogger struct {
	mu       sync.Mutex
	errors   []string
	warnings []string
}

func (c *captureLogger) Debug(_ string)             {}
func (c *captureLogger) Info(_ string)              {}
func (c *captureLogger) Warning(message string)     { c.mu.Lock(); defer c.mu.Unlock(); c.warnings = append(c.warnings, message) }
func (c *captureLogger) Error(message string, _ interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.errors = append(c.errors, message)
}

// errorMessages returns a snapshot of captured error messages.
func (c *captureLogger) errorMessages() []string {
	c.mu.Lock()
	defer c.mu.Unlock()
	out := make([]string, len(c.errors))
	copy(out, c.errors)
	return out
}

// Verify captureLogger satisfies the interface at compile time.
var _ logging.OptimizelyLogProducer = (*captureLogger)(nil)

func TestMapHoldoutsEmpty(t *testing.T) {
	holdoutList, holdoutIDMap, globalHoldouts, ruleHoldoutsMap := MapHoldouts(nil, nil, &captureLogger{})

	assert.Empty(t, holdoutList)
	assert.Empty(t, holdoutIDMap)
	assert.Empty(t, globalHoldouts)
	assert.Empty(t, ruleHoldoutsMap)
}

func TestMapHoldoutsGlobalSectionAppliesToAllFlags(t *testing.T) {
	// Running entries in the `holdouts` (global) section are returned in globalHoldouts.
	rawGlobal := []datafileEntities.Holdout{
		{
			ID:     "holdout_1",
			Key:    "running_holdout",
			Status: "Running",
			Variations: []datafileEntities.Variation{
				{ID: "var_1", Key: "variation_1"},
			},
			TrafficAllocation: []datafileEntities.TrafficAllocation{
				{EntityID: "var_1", EndOfRange: 10000},
			},
		},
	}

	holdoutList, holdoutIDMap, globalHoldouts, _ := MapHoldouts(rawGlobal, nil, &captureLogger{})

	assert.Len(t, holdoutList, 1)
	assert.Len(t, holdoutIDMap, 1)
	assert.Equal(t, "holdout_1", holdoutList[0].ID)
	assert.Equal(t, "running_holdout", holdoutList[0].Key)
	assert.Len(t, globalHoldouts, 1)
	assert.Equal(t, "running_holdout", globalHoldouts[0].Key)
	assert.True(t, globalHoldouts[0].IsGlobal())
}

func TestMapHoldoutsNotRunningInGlobalSection(t *testing.T) {
	rawGlobal := []datafileEntities.Holdout{
		{
			ID:     "holdout_1",
			Key:    "paused_holdout",
			Status: "Paused",
			Variations: []datafileEntities.Variation{
				{ID: "var_1", Key: "variation_1"},
			},
		},
	}

	holdoutList, holdoutIDMap, globalHoldouts, _ := MapHoldouts(rawGlobal, nil, &captureLogger{})

	assert.Empty(t, holdoutList)
	assert.Empty(t, holdoutIDMap)
	assert.Empty(t, globalHoldouts)
}

func TestMapHoldoutsMultipleGlobalHoldouts(t *testing.T) {
	rawGlobal := []datafileEntities.Holdout{
		{
			ID:     "holdout_1",
			Key:    "holdout_1",
			Status: "Running",
			Variations: []datafileEntities.Variation{
				{ID: "var_1", Key: "variation_1"},
			},
			TrafficAllocation: []datafileEntities.TrafficAllocation{
				{EntityID: "var_1", EndOfRange: 5000},
			},
		},
		{
			ID:     "holdout_2",
			Key:    "holdout_2",
			Status: "Running",
			Variations: []datafileEntities.Variation{
				{ID: "var_2", Key: "variation_2"},
			},
			TrafficAllocation: []datafileEntities.TrafficAllocation{
				{EntityID: "var_2", EndOfRange: 10000},
			},
		},
	}

	holdoutList, holdoutIDMap, globalHoldouts, _ := MapHoldouts(rawGlobal, nil, &captureLogger{})

	assert.Len(t, holdoutList, 2)
	assert.Len(t, holdoutIDMap, 2)
	assert.Len(t, globalHoldouts, 2)
	assert.Equal(t, "holdout_1", globalHoldouts[0].Key)
	assert.Equal(t, "holdout_2", globalHoldouts[1].Key)
}

func TestMapHoldoutsWithAudienceConditions(t *testing.T) {
	rawGlobal := []datafileEntities.Holdout{
		{
			ID:                 "holdout_1",
			Key:                "holdout_with_audience",
			Status:             "Running",
			AudienceIds:        []string{"audience_1", "audience_2"},
			AudienceConditions: []interface{}{"or", "audience_1", "audience_2"},
			Variations: []datafileEntities.Variation{
				{ID: "var_1", Key: "variation_1"},
			},
			TrafficAllocation: []datafileEntities.TrafficAllocation{
				{EntityID: "var_1", EndOfRange: 10000},
			},
		},
	}

	holdoutList, _, _, _ := MapHoldouts(rawGlobal, nil, &captureLogger{})

	assert.Len(t, holdoutList, 1)
	assert.Equal(t, []string{"audience_1", "audience_2"}, holdoutList[0].AudienceIds)
	assert.NotNil(t, holdoutList[0].AudienceConditionTree)
}

func TestMapHoldoutsVariationsMapping(t *testing.T) {
	rawGlobal := []datafileEntities.Holdout{
		{
			ID:     "holdout_1",
			Key:    "holdout_variations",
			Status: "Running",
			Variations: []datafileEntities.Variation{
				{
					ID:             "var_1",
					Key:            "variation_1",
					FeatureEnabled: true,
					Variables: []datafileEntities.VariationVariable{
						{ID: "var_var_1", Value: "value_1"},
					},
				},
				{
					ID:             "var_2",
					Key:            "variation_2",
					FeatureEnabled: false,
				},
			},
			TrafficAllocation: []datafileEntities.TrafficAllocation{
				{EntityID: "var_1", EndOfRange: 5000},
				{EntityID: "var_2", EndOfRange: 10000},
			},
		},
	}

	holdoutList, _, _, _ := MapHoldouts(rawGlobal, nil, &captureLogger{})

	assert.Len(t, holdoutList, 1)
	assert.Len(t, holdoutList[0].Variations, 2)
	assert.Contains(t, holdoutList[0].Variations, "var_1")
	assert.Contains(t, holdoutList[0].Variations, "var_2")

	assert.Len(t, holdoutList[0].TrafficAllocation, 2)
	assert.Equal(t, "var_1", holdoutList[0].TrafficAllocation[0].EntityID)
	assert.Equal(t, 5000, holdoutList[0].TrafficAllocation[0].EndOfRange)
	assert.Equal(t, "var_2", holdoutList[0].TrafficAllocation[1].EntityID)
	assert.Equal(t, 10000, holdoutList[0].TrafficAllocation[1].EndOfRange)
}

// ---------------------------------------------------------------------------
// FSSDK-12760 — backward-compatible localHoldouts section tests
// ---------------------------------------------------------------------------

func TestMapHoldoutsLocalSectionRegistersPerRule(t *testing.T) {
	// Local section entries must be registered under each rule in IncludedRules.
	includedRules := []string{"rule_id_1", "rule_id_2"}
	rawLocal := []datafileEntities.Holdout{
		{
			ID:            "holdout_local",
			Key:           "local_holdout",
			Status:        "Running",
			IncludedRules: &includedRules,
			Variations: []datafileEntities.Variation{
				{ID: "var_1", Key: "variation_1"},
			},
			TrafficAllocation: []datafileEntities.TrafficAllocation{
				{EntityID: "var_1", EndOfRange: 10000},
			},
		},
	}

	holdoutList, _, globalHoldouts, ruleHoldoutsMap := MapHoldouts(nil, rawLocal, &captureLogger{})

	assert.Len(t, holdoutList, 1)
	assert.False(t, holdoutList[0].IsGlobal())
	assert.NotNil(t, holdoutList[0].IncludedRules)
	// Local section entry must NOT appear in globalHoldouts.
	assert.Empty(t, globalHoldouts)
	assert.Contains(t, ruleHoldoutsMap, "rule_id_1")
	assert.Contains(t, ruleHoldoutsMap, "rule_id_2")
	assert.Len(t, ruleHoldoutsMap["rule_id_1"], 1)
	assert.Len(t, ruleHoldoutsMap["rule_id_2"], 1)
}

func TestMapHoldoutsLocalSectionEmptyIncludedRulesIsValidButTargetsNoRules(t *testing.T) {
	// An empty (non-nil) IncludedRules slice is valid: the entity is created
	// and tracked, but does not register against any rule. Not invalid, not global.
	emptyRules := []string{}
	rawLocal := []datafileEntities.Holdout{
		{
			ID:            "holdout_empty_local",
			Key:           "empty_local_holdout",
			Status:        "Running",
			IncludedRules: &emptyRules,
			Variations: []datafileEntities.Variation{
				{ID: "var_1", Key: "variation_1"},
			},
			TrafficAllocation: []datafileEntities.TrafficAllocation{
				{EntityID: "var_1", EndOfRange: 10000},
			},
		},
	}

	logger := &captureLogger{}
	holdoutList, holdoutIDMap, globalHoldouts, ruleHoldoutsMap := MapHoldouts(nil, rawLocal, logger)

	assert.Len(t, holdoutList, 1)
	assert.False(t, holdoutList[0].IsGlobal())
	assert.Empty(t, globalHoldouts)
	assert.Empty(t, ruleHoldoutsMap)
	// Entity is tracked in the id map (not invalid)
	assert.Contains(t, holdoutIDMap, "holdout_empty_local")
	// And no error was logged.
	assert.Empty(t, logger.errorMessages())
}

func TestMapHoldoutsLocalSectionMissingIncludedRulesIsInvalid(t *testing.T) {
	// Per FSSDK-12760: local entries WITHOUT IncludedRules are invalid — logged and skipped.
	rawLocal := []datafileEntities.Holdout{
		{
			ID:     "holdout_invalid",
			Key:    "invalid_local",
			Status: "Running",
			// IncludedRules: nil — invalid in the localHoldouts section.
			Variations: []datafileEntities.Variation{
				{ID: "var_1", Key: "variation_1"},
			},
			TrafficAllocation: []datafileEntities.TrafficAllocation{
				{EntityID: "var_1", EndOfRange: 10000},
			},
		},
	}

	logger := &captureLogger{}
	holdoutList, holdoutIDMap, globalHoldouts, ruleHoldoutsMap := MapHoldouts(nil, rawLocal, logger)

	// Invalid entry must be excluded across all outputs — NEVER promoted to global.
	assert.Empty(t, holdoutList)
	assert.Empty(t, holdoutIDMap)
	assert.Empty(t, globalHoldouts)
	assert.Empty(t, ruleHoldoutsMap)

	// Error must be logged and reference both the holdout key and "includedRules".
	errors := logger.errorMessages()
	assert.Len(t, errors, 1)
	assert.Contains(t, errors[0], "invalid_local")
	assert.Contains(t, errors[0], "includedRules")
}

func TestMapHoldoutsLocalSectionInvalidEntryDoesNotAffectValidEntries(t *testing.T) {
	// Mixing one invalid local entry with a valid one: invalid is skipped, valid is processed.
	validRules := []string{"rule_x"}
	rawLocal := []datafileEntities.Holdout{
		{
			ID:     "holdout_invalid",
			Key:    "invalid_local",
			Status: "Running",
			// nil IncludedRules — invalid.
		},
		{
			ID:            "holdout_valid",
			Key:           "valid_local",
			Status:        "Running",
			IncludedRules: &validRules,
			Variations: []datafileEntities.Variation{
				{ID: "var_v", Key: "v"},
			},
			TrafficAllocation: []datafileEntities.TrafficAllocation{
				{EntityID: "var_v", EndOfRange: 10000},
			},
		},
	}

	logger := &captureLogger{}
	holdoutList, holdoutIDMap, _, ruleHoldoutsMap := MapHoldouts(nil, rawLocal, logger)

	assert.Len(t, holdoutList, 1)
	assert.Equal(t, "holdout_valid", holdoutList[0].ID)
	assert.Contains(t, holdoutIDMap, "holdout_valid")
	assert.NotContains(t, holdoutIDMap, "holdout_invalid")
	assert.Contains(t, ruleHoldoutsMap, "rule_x")
	assert.Len(t, logger.errorMessages(), 1)
}

func TestMapHoldoutsGlobalSectionStripsIncludedRules(t *testing.T) {
	// If a global-section entry accidentally has IncludedRules, it must be stripped
	// at parse time and the entity must still be classified as global. The stray
	// rule id must NOT appear in the per-rule map.
	strayRules := []string{"rule_should_be_ignored"}
	rawGlobal := []datafileEntities.Holdout{
		{
			ID:            "holdout_stray",
			Key:           "stray_global",
			Status:        "Running",
			IncludedRules: &strayRules,
			Variations: []datafileEntities.Variation{
				{ID: "var_1", Key: "variation_1"},
			},
			TrafficAllocation: []datafileEntities.TrafficAllocation{
				{EntityID: "var_1", EndOfRange: 10000},
			},
		},
	}

	holdoutList, _, globalHoldouts, ruleHoldoutsMap := MapHoldouts(rawGlobal, nil, &captureLogger{})

	assert.Len(t, holdoutList, 1)
	assert.True(t, holdoutList[0].IsGlobal(), "entry from global section must always be global")
	assert.Nil(t, holdoutList[0].IncludedRules, "IncludedRules must be stripped on global entries")

	assert.Len(t, globalHoldouts, 1)
	assert.True(t, globalHoldouts[0].IsGlobal())

	// The stray rule id must not have leaked into the per-rule map.
	assert.NotContains(t, ruleHoldoutsMap, "rule_should_be_ignored")
	assert.Empty(t, ruleHoldoutsMap)
}

func TestMapHoldoutsBothSectionsPartitionCorrectly(t *testing.T) {
	// When both sections are present, scope is enforced strictly by section
	// membership — entries never cross over.
	localRulesA := []string{"rule_a"}
	localRulesB := []string{"rule_b"}
	rawGlobal := []datafileEntities.Holdout{
		{
			ID:     "g1",
			Key:    "global_1",
			Status: "Running",
			Variations: []datafileEntities.Variation{
				{ID: "var_g1", Key: "var_g1"},
			},
			TrafficAllocation: []datafileEntities.TrafficAllocation{
				{EntityID: "var_g1", EndOfRange: 10000},
			},
		},
		{
			ID:     "g2",
			Key:    "global_2",
			Status: "Running",
			Variations: []datafileEntities.Variation{
				{ID: "var_g2", Key: "var_g2"},
			},
			TrafficAllocation: []datafileEntities.TrafficAllocation{
				{EntityID: "var_g2", EndOfRange: 10000},
			},
		},
	}
	rawLocal := []datafileEntities.Holdout{
		{
			ID:            "l1",
			Key:           "local_1",
			Status:        "Running",
			IncludedRules: &localRulesA,
			Variations: []datafileEntities.Variation{
				{ID: "var_l1", Key: "var_l1"},
			},
			TrafficAllocation: []datafileEntities.TrafficAllocation{
				{EntityID: "var_l1", EndOfRange: 10000},
			},
		},
		{
			ID:            "l2",
			Key:           "local_2",
			Status:        "Running",
			IncludedRules: &localRulesB,
			Variations: []datafileEntities.Variation{
				{ID: "var_l2", Key: "var_l2"},
			},
			TrafficAllocation: []datafileEntities.TrafficAllocation{
				{EntityID: "var_l2", EndOfRange: 10000},
			},
		},
	}

	holdoutList, holdoutIDMap, globalHoldouts, ruleHoldoutsMap := MapHoldouts(rawGlobal, rawLocal, &captureLogger{})

	assert.Len(t, holdoutList, 4)
	assert.Len(t, holdoutIDMap, 4)

	globalIDs := map[string]bool{}
	for _, h := range globalHoldouts {
		globalIDs[h.ID] = true
	}
	assert.Equal(t, map[string]bool{"g1": true, "g2": true}, globalIDs)

	assert.Len(t, ruleHoldoutsMap["rule_a"], 1)
	assert.Equal(t, "l1", ruleHoldoutsMap["rule_a"][0].ID)
	assert.Len(t, ruleHoldoutsMap["rule_b"], 1)
	assert.Equal(t, "l2", ruleHoldoutsMap["rule_b"][0].ID)

	// Global ids must NOT show up in any per-rule entry.
	for _, ruleHoldouts := range ruleHoldoutsMap {
		for _, h := range ruleHoldouts {
			assert.NotContains(t, []string{"g1", "g2"}, h.ID)
		}
	}
}

func TestMapHoldoutsBackwardCompatNoLocalSection(t *testing.T) {
	// Old datafiles without a `localHoldouts` section continue to work unchanged:
	// every entry in `holdouts` is global, no errors, no log noise.
	rawGlobal := []datafileEntities.Holdout{
		{
			ID:     "holdout_old",
			Key:    "old_global_holdout",
			Status: "Running",
			Variations: []datafileEntities.Variation{
				{ID: "var_1", Key: "variation_1"},
			},
			TrafficAllocation: []datafileEntities.TrafficAllocation{
				{EntityID: "var_1", EndOfRange: 10000},
			},
		},
	}

	logger := &captureLogger{}
	// localHoldoutsRaw is nil — simulates a pre-FSSDK-12760 datafile.
	holdoutList, _, globalHoldouts, ruleHoldoutsMap := MapHoldouts(rawGlobal, nil, logger)

	assert.Len(t, holdoutList, 1)
	assert.True(t, holdoutList[0].IsGlobal())
	assert.Len(t, globalHoldouts, 1)
	assert.Equal(t, "old_global_holdout", globalHoldouts[0].Key)
	assert.Empty(t, ruleHoldoutsMap)
	assert.Empty(t, logger.errorMessages())
}

func TestMapHoldoutsLocalSectionNonRunningExcluded(t *testing.T) {
	// Non-Running entries in `localHoldouts` are filtered out before validation —
	// they must not be evaluated and must not produce error logs.
	includedRules := []string{"rule_x"}
	rawLocal := []datafileEntities.Holdout{
		{
			ID:            "holdout_draft",
			Key:           "draft_local",
			Status:        "Draft",
			IncludedRules: &includedRules,
		},
	}

	logger := &captureLogger{}
	holdoutList, holdoutIDMap, _, ruleHoldoutsMap := MapHoldouts(nil, rawLocal, logger)

	assert.Empty(t, holdoutList)
	assert.Empty(t, holdoutIDMap)
	assert.Empty(t, ruleHoldoutsMap)
	assert.Empty(t, logger.errorMessages())
}

func TestMapHoldoutsLocalSectionMissingIncludedRulesUsesIDWhenKeyAbsent(t *testing.T) {
	// When an invalid local holdout has no Key, the log message must still
	// identify the entry — fall back to the ID.
	rawLocal := []datafileEntities.Holdout{
		{
			ID:     "id_only_holdout",
			Status: "Running",
			// Key empty, IncludedRules nil — invalid.
		},
	}

	logger := &captureLogger{}
	MapHoldouts(nil, rawLocal, logger)

	errors := logger.errorMessages()
	assert.Len(t, errors, 1)
	assert.Contains(t, errors[0], "id_only_holdout")
}

func TestMapHoldoutsNilLoggerWithInvalidLocalEntryDoesNotPanic(t *testing.T) {
	// Defensive: callers may pass a nil logger; the mapper must not panic.
	rawLocal := []datafileEntities.Holdout{
		{
			ID:     "holdout_invalid",
			Key:    "invalid_local",
			Status: "Running",
			// IncludedRules: nil — invalid; would normally produce a log.
		},
	}

	assert.NotPanics(t, func() {
		holdoutList, _, _, _ := MapHoldouts(nil, rawLocal, nil)
		assert.Empty(t, holdoutList)
	})
}

func TestMapHoldoutsIsGlobalProperty(t *testing.T) {
	// Verify the entity-level IsGlobal property is unchanged by the section split.
	nilRules := (*[]string)(nil)
	emptyRules := []string{}
	ruleIDs := []string{"rule_1"}

	globalHoldout := entities.Holdout{IncludedRules: nilRules}
	localHoldoutEmpty := entities.Holdout{IncludedRules: &emptyRules}
	localHoldoutWithRules := entities.Holdout{IncludedRules: &ruleIDs}

	assert.True(t, globalHoldout.IsGlobal(), "nil IncludedRules should be global")
	assert.False(t, localHoldoutEmpty.IsGlobal(), "empty non-nil IncludedRules should NOT be global")
	assert.False(t, localHoldoutWithRules.IsGlobal(), "non-nil IncludedRules with rules should NOT be global")
}

func TestMapHoldoutsErrorMessageWording(t *testing.T) {
	// The error message wording is part of the API surface for operators — ensure
	// it remains user-actionable across refactors.
	rawLocal := []datafileEntities.Holdout{
		{
			ID:     "h_x",
			Key:    "x_local",
			Status: "Running",
		},
	}

	logger := &captureLogger{}
	MapHoldouts(nil, rawLocal, logger)

	errors := logger.errorMessages()
	assert.Len(t, errors, 1)
	// Message must say what is missing and what the consequence is.
	assert.True(t, strings.Contains(errors[0], "missing"), "message should mention 'missing'")
	assert.True(t, strings.Contains(errors[0], "excluded"), "message should mention 'excluded'")
}
