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

// Package decision provides CMAB decision service implementation
package decision

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/optimizely/go-sdk/v2/pkg/cache"
	"github.com/optimizely/go-sdk/v2/pkg/config"
	"github.com/optimizely/go-sdk/v2/pkg/entities"
	"github.com/optimizely/go-sdk/v2/pkg/logging"
	"github.com/twmb/murmur3"
)

// DefaultCmabService implements the CmabService interface
type DefaultCmabService struct {
	cmabCache  cache.CacheWithRemove
	cmabClient CmabClient
	logger     logging.OptimizelyLogProducer
}

// CmabServiceOptions defines options for creating a CMAB service
type CmabServiceOptions struct {
	Logger     logging.OptimizelyLogProducer
	CmabCache  cache.CacheWithRemove
	CmabClient CmabClient
}

// NewDefaultCmabService creates a new instance of DefaultCmabService
func NewDefaultCmabService(options CmabServiceOptions) *DefaultCmabService {
	logger := options.Logger
	if logger == nil {
		logger = logging.GetLogger("", "DefaultCmabService")
	}

	return &DefaultCmabService{
		cmabCache:  options.CmabCache,
		cmabClient: options.CmabClient,
		logger:     logger,
	}
}

// GetDecision returns a CMAB decision for the given rule and user context
func (s *DefaultCmabService) GetDecision(
	projectConfig config.ProjectConfig,
	userContext entities.UserContext,
	ruleID string,
	options map[OptimizelyDecideOptions]bool,
) (CmabDecision, error) {
	// Filter attributes based on CMAB configuration
	filteredAttributes := s.filterAttributes(projectConfig, userContext, ruleID)

	// Check if we should ignore the cache
	if options[IgnoreCMABCache] {
		return s.fetchDecisionWithRetry(ruleID, userContext.ID, filteredAttributes)
	}

	// Reset cache if requested
	if options[ResetCMABCache] {
		s.cmabCache.Reset()
	}

	// Create cache key
	cacheKey := s.getCacheKey(userContext.ID, ruleID)

	// Invalidate user cache if requested
	if options[InvalidateUserCMABCache] {
		s.cmabCache.Remove(cacheKey)
	}

	// Generate attributes hash for cache validation
	attributesJSON, err := s.getAttributesJSON(filteredAttributes)
	if err != nil {
		return CmabDecision{}, fmt.Errorf("failed to serialize attributes: %w", err)
	}
	hasher := murmur3.SeedNew32(1) // Use seed 1 for consistency
	_, err = hasher.Write([]byte(attributesJSON))
	if err != nil {
		return CmabDecision{}, fmt.Errorf("failed to hash attributes: %w", err)
	}
	attributesHash := strconv.FormatUint(uint64(hasher.Sum32()), 10)

	// Try to get from cache
	cachedValue := s.cmabCache.Lookup(cacheKey)
	if cachedValue != nil {
		// Need to type assert since Lookup returns interface{}
		if cacheVal, ok := cachedValue.(CmabCacheValue); ok {
			// Check if attributes have changed
			if cacheVal.AttributesHash == attributesHash {
				s.logger.Debug(fmt.Sprintf("Returning cached CMAB decision for rule %s and user %s", ruleID, userContext.ID))
				return CmabDecision{
					VariationID: cacheVal.VariationID,
					CmabUUID:    cacheVal.CmabUUID,
				}, nil
			}

			// Attributes changed, remove from cache
			s.cmabCache.Remove(cacheKey)
		}
	}

	// Fetch new decision
	decision, err := s.fetchDecisionWithRetry(ruleID, userContext.ID, filteredAttributes)
	if err != nil {
		return CmabDecision{}, err
	}

	// Cache the decision
	cacheValue := CmabCacheValue{
		AttributesHash: attributesHash,
		VariationID:    decision.VariationID,
		CmabUUID:       decision.CmabUUID,
	}

	s.cmabCache.Save(cacheKey, cacheValue)

	return decision, nil
}

// fetchDecisionWithRetry fetches a decision from the CMAB API with retry logic
func (s *DefaultCmabService) fetchDecisionWithRetry(
	ruleID string,
	userID string,
	attributes map[string]interface{},
) (CmabDecision, error) {
	cmabUUID := uuid.New().String()

	// Retry configuration
	maxRetries := 3
	backoffFactor := 2
	initialBackoff := 100 * time.Millisecond

	var lastErr error

	for attempt := 0; attempt < maxRetries; attempt++ {
		// Exponential backoff if this is a retry
		if attempt > 0 {
			backoffDuration := initialBackoff * time.Duration(backoffFactor^attempt)
			time.Sleep(backoffDuration)
		}

		s.logger.Debug(fmt.Sprintf("Fetching CMAB decision for rule %s and user %s (attempt %d/%d)",
			ruleID, userID, attempt+1, maxRetries))

		variationID, err := s.cmabClient.FetchDecision(ruleID, userID, attributes, cmabUUID)
		if err == nil {
			return CmabDecision{
				VariationID: variationID,
				CmabUUID:    cmabUUID,
			}, nil
		}

		lastErr = err
		s.logger.Warning(fmt.Sprintf("CMAB API request failed (attempt %d/%d): %v",
			attempt+1, maxRetries, err))
	}

	return CmabDecision{}, fmt.Errorf("failed to fetch CMAB decision after %d attempts: %w",
		maxRetries, lastErr)
}

// filterAttributes filters user attributes based on CMAB configuration
func (s *DefaultCmabService) filterAttributes(
	projectConfig config.ProjectConfig,
	userContext entities.UserContext,
	ruleID string,
) map[string]interface{} {
	filteredAttributes := make(map[string]interface{})

	// Get all experiments and find the one with matching ID
	experimentList := projectConfig.GetExperimentList()
	var targetExperiment entities.Experiment
	found := false

	for _, experiment := range experimentList {
		if experiment.ID == ruleID {
			targetExperiment = experiment
			found = true
			break
		}
	}

	if !found || targetExperiment.Cmab == nil {
		return filteredAttributes
	}

	// Get attribute IDs from CMAB configuration
	cmabAttributeIDs := targetExperiment.Cmab.AttributeIds

	// Filter attributes based on CMAB configuration
	for _, attributeID := range cmabAttributeIDs {
		// Get the attribute key for this ID
		attributeKey, err := projectConfig.GetAttributeKeyByID(attributeID)
		if err != nil {
			s.logger.Debug(fmt.Sprintf("Attribute with ID %s not found in project config: %v", attributeID, err))
			continue
		}

		if value, exists := userContext.Attributes[attributeKey]; exists {
			filteredAttributes[attributeKey] = value
		}
	}

	return filteredAttributes
}

// getAttributesJSON serializes attributes to a deterministic JSON string
func (s *DefaultCmabService) getAttributesJSON(attributes map[string]interface{}) (string, error) {
	// Get sorted keys for deterministic serialization
	keys := make([]string, 0, len(attributes))
	for k := range attributes {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Create a map with sorted keys
	sortedMap := make(map[string]interface{})
	for _, k := range keys {
		sortedMap[k] = attributes[k]
	}

	// Serialize to JSON
	jsonBytes, err := json.Marshal(sortedMap)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

// getCacheKey generates a cache key for the user and rule
func (s *DefaultCmabService) getCacheKey(userID, ruleID string) string {
	return fmt.Sprintf("%s:%s", userID, ruleID)
}
