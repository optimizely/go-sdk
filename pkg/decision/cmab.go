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

// pkg/decision/cmab.go
package decision

import (
    "github.com/optimizely/go-sdk/v2/pkg/config"
    "github.com/optimizely/go-sdk/v2/pkg/entities"
)

// Contains options for CMAB decision requests
type CmabDecisionOptions struct {
    IncludeReasons bool
    IgnoreCmabCache bool
}

// Represents a decision from the CMAB service
type CmabDecision struct {
    VariationID string
    CmabUUID    string
    Reasons     []string
}

// Represents a cached CMAB decision
type CmabCacheValue struct {
    Decision CmabDecision
    Created  int64
}

// Defines the interface for CMAB decision service
type CmabService interface {
    GetDecision(projectConfig config.ProjectConfig, userContext entities.UserContext, ruleID string, options *CmabDecisionOptions) (CmabDecision, error)
    ResetCache() error
    InvalidateUserCache(userID string) error
}

// Defines the interface for CMAB API client
type CmabClient interface {
    FetchDecision(ruleID string, userID string, attributes map[string]interface{}) (CmabDecision, error)
}
