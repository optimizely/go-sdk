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
	reasons := NewDecisionReasons([]Options{})
	assert.Equal(t, 0, len(reasons.errors))
	assert.Equal(t, 0, len(reasons.logs))
	assert.Equal(t, 0, len(reasons.ToReport()))
}

func TestAddErrorWorksWithEveryOption(t *testing.T) {
	reasons := NewDecisionReasons([]Options{DisableDecisionEvent, EnabledFlagsOnly, IgnoreUserProfileService, ExcludeVariables, IncludeReasons})
	reasons.AddError("error message")
	assert.Equal(t, 1, len(reasons.errors))

	reportedReasons := reasons.ToReport()
	assert.Equal(t, 1, len(reportedReasons))
	assert.Equal(t, "error message", reportedReasons[0])
}

func TestAddInfofOnlyWorksWithIncludeReasonsOption(t *testing.T) {
	reasons := NewDecisionReasons([]Options{DisableDecisionEvent, EnabledFlagsOnly, IgnoreUserProfileService, ExcludeVariables})
	reasons.AddInfof("info message")
	assert.Equal(t, 0, len(reasons.logs))

	reportedReasons := reasons.ToReport()
	assert.Equal(t, 0, len(reportedReasons))

	reasons = NewDecisionReasons([]Options{DisableDecisionEvent, EnabledFlagsOnly, IgnoreUserProfileService, ExcludeVariables, IncludeReasons})
	reasons.AddInfof("info message")
	assert.Equal(t, 1, len(reasons.logs))

	reportedReasons = reasons.ToReport()
	assert.Equal(t, 1, len(reportedReasons))
	assert.Equal(t, "info message", reportedReasons[0])
}

func TestInfoLogsAreOnlyReportedWhenIncludeReasonsOptionIsSet(t *testing.T) {
	reasons := NewDecisionReasons([]Options{DisableDecisionEvent, EnabledFlagsOnly, IgnoreUserProfileService, ExcludeVariables})
	reasons.AddError("error message")
	reasons.AddInfof("info message")

	reportedReasons := reasons.ToReport()
	assert.Equal(t, 1, len(reportedReasons))
	assert.Equal(t, "error message", reportedReasons[0])

	reasons = NewDecisionReasons([]Options{DisableDecisionEvent, EnabledFlagsOnly, IgnoreUserProfileService, ExcludeVariables, IncludeReasons})
	reasons.AddError("error message")
	reasons.AddInfof("info message")

	reportedReasons = reasons.ToReport()
	assert.Equal(t, 2, len(reportedReasons))
	assert.Equal(t, "error message", reportedReasons[0])
	assert.Equal(t, "info message", reportedReasons[1])
}
