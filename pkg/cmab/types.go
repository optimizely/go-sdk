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

// Package cmab provides functionality for Contextual Multi-Armed Bandit (CMAB)
// decision-making, including client and service implementations for making and
// handling CMAB requests and responses.
package cmab

import (
	"github.com/optimizely/go-sdk/v2/pkg/config"
	"github.com/optimizely/go-sdk/v2/pkg/decide"
	"github.com/optimizely/go-sdk/v2/pkg/entities"
)

// Decision represents a decision from the CMAB service
type Decision struct {
	VariationID string
	CmabUUID    string
	Reasons     []string
}

// CacheValue represents a cached CMAB decision with attribute hash
type CacheValue struct {
	AttributesHash string
	VariationID    string
	CmabUUID       string
}

// Service defines the interface for CMAB decision services
type Service interface {
	// GetDecision returns a CMAB decision for the given rule and user context
	GetDecision(
		projectConfig config.ProjectConfig,
		userContext entities.UserContext,
		ruleID string,
		options *decide.Options,
	) (Decision, error)
}

// Client defines the interface for CMAB API clients
type Client interface {
	// FetchDecision fetches a decision from the CMAB API
	FetchDecision(
		ruleID string,
		userID string,
		attributes map[string]interface{},
		cmabUUID string,
	) (string, error)
}
