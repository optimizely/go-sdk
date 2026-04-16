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

package entities

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHoldout_IsGlobal(t *testing.T) {
	t.Run("Returns true when IncludedRules is nil", func(t *testing.T) {
		holdout := Holdout{
			ID:            "holdout_1",
			Key:           "test_holdout",
			IncludedRules: nil,
		}

		assert.True(t, holdout.IsGlobal(), "Holdout with nil IncludedRules should be global")
	})

	t.Run("Returns false when IncludedRules is non-nil", func(t *testing.T) {
		holdout := Holdout{
			ID:            "holdout_2",
			Key:           "local_holdout",
			IncludedRules: []string{"rule_1", "rule_2"},
		}

		assert.False(t, holdout.IsGlobal(), "Holdout with non-nil IncludedRules should be local")
	})

	t.Run("Returns false when IncludedRules is empty slice", func(t *testing.T) {
		holdout := Holdout{
			ID:            "holdout_3",
			Key:           "empty_rules_holdout",
			IncludedRules: []string{},
		}

		assert.False(t, holdout.IsGlobal(), "Holdout with empty IncludedRules slice should be local")
	})
}
