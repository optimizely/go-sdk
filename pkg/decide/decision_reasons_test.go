/****************************************************************************
 * Copyright 2020, Optimizely, Inc. and contributors                        *
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

package decide

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewDecisionReasonsWithEmptyOptions(t *testing.T) {
	reasons := NewDecisionReasons(&Options{})
	assert.Equal(t, 0, len(reasons.ToReport()))
}

func TestAddErrorWorksWithEveryOption(t *testing.T) {
	options := &Options{
		DisableDecisionEvent:     true,
		EnabledFlagsOnly:         true,
		IgnoreUserProfileService: true,
		ExcludeVariables:         true,
		IncludeReasons:           true,
	}
	reasons := NewDecisionReasons(options)
	reasons.AddError("error message")

	reportedReasons := reasons.ToReport()
	assert.Equal(t, 1, len(reportedReasons))
	assert.Equal(t, "error message", reportedReasons[0])
}

func TestInfoLogsAreOnlyReportedWhenIncludeReasonsOptionIsSet(t *testing.T) {
	options := &Options{
		DisableDecisionEvent:     true,
		EnabledFlagsOnly:         true,
		IgnoreUserProfileService: true,
		ExcludeVariables:         true,
	}
	reasons := NewDecisionReasons(options)
	reasons.AddError("error message")
	reasons.AddError("error message: code %d", 121)
	reasons.AddInfo("info message")
	reasons.AddInfo("info message: %s", "unexpected string")

	reportedReasons := reasons.ToReport()
	assert.Equal(t, 2, len(reportedReasons))
	assert.Equal(t, "error message", reportedReasons[0])
	assert.Equal(t, "error message: code 121", reportedReasons[1])

	options.IncludeReasons = true
	reasons = NewDecisionReasons(options)
	reasons.AddError("error message")
	reasons.AddError("error message: code %d", 121)
	reasons.AddInfo("info message")
	reasons.AddInfo("info message: %s", "unexpected string")

	reportedReasons = reasons.ToReport()
	assert.Equal(t, 4, len(reportedReasons))
	assert.Equal(t, "error message", reportedReasons[0])
	assert.Equal(t, "error message: code 121", reportedReasons[1])
	assert.Equal(t, "info message", reportedReasons[2])
	assert.Equal(t, "info message: unexpected string", reportedReasons[3])
}

func TestAppend(t *testing.T) {
	options := &Options{
		DisableDecisionEvent:     true,
		EnabledFlagsOnly:         true,
		IgnoreUserProfileService: true,
		ExcludeVariables:         true,
		IncludeReasons:           true,
	}
	reasons1 := NewDecisionReasons(options)
	reasons1.AddError("error message")
	reasons1.AddError("error message: code %d", 121)
	reasons1.AddInfo("info message")
	reasons1.AddInfo("info message: %s", "unexpected string")

	// Shouldn't append info logs if include reasons is set to false
	reasons2 := NewDecisionReasons(nil)
	reasons2.Append(reasons1)
	reportedReasons := reasons2.ToReport()
	assert.Equal(t, 2, len(reportedReasons))
	assert.Equal(t, "error message", reportedReasons[0])
	assert.Equal(t, "error message: code 121", reportedReasons[1])

	// Should append info logs if include reasons is set to true
	options.IncludeReasons = true
	reasons2 = NewDecisionReasons(options)
	reasons2.Append(reasons1)
	reportedReasons = reasons2.ToReport()
	assert.Equal(t, 4, len(reportedReasons))
	assert.Equal(t, "error message", reportedReasons[0])
	assert.Equal(t, "error message: code 121", reportedReasons[1])
	assert.Equal(t, "info message", reportedReasons[2])
	assert.Equal(t, "info message: unexpected string", reportedReasons[3])
}
