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

// Package decision provides CMAB decision service interfaces and types
package decision

import (
	"github.com/optimizely/go-sdk/v2/pkg/config"
	"github.com/optimizely/go-sdk/v2/pkg/entities"
)

// OptimizelyDecideOptions represents options for the Decide method
type OptimizelyDecideOptions string

// CMAB-specific decide options
const (
	// IgnoreCMABCache ignores the CMAB cache and forces a new API request
	IgnoreCMABCache OptimizelyDecideOptions = "IGNORE_CMAB_CACHE"
	// ResetCMABCache resets the entire CMAB cache
	ResetCMABCache OptimizelyDecideOptions = "RESET_CMAB_CACHE"
	// InvalidateUserCMABCache invalidates cache entries for the current user
	InvalidateUserCMABCache OptimizelyDecideOptions = "INVALIDATE_USER_CMAB_CACHE"
)

// CmabDecision represents a decision from the CMAB service
type CmabDecision struct {
	VariationID string
	CmabUUID    string
	Reasons     []string
}

// CmabCacheValue represents a cached CMAB decision with attribute hash
type CmabCacheValue struct {
	AttributesHash string
	VariationID    string
	CmabUUID       string
}

// CmabService defines the interface for CMAB decision services
type CmabService interface {
	// GetDecision returns a CMAB decision for the given rule and user context
	GetDecision(
		projectConfig config.ProjectConfig,
		userContext entities.UserContext,
		ruleID string,
		options map[OptimizelyDecideOptions]bool,
	) (CmabDecision, error)
}

// CmabClient defines the interface for CMAB API clients
type CmabClient interface {
	// FetchDecision fetches a decision from the CMAB API
	FetchDecision(
		ruleID string,
		userID string,
		attributes map[string]interface{},
		cmabUUID string,
	) (string, error)
}
