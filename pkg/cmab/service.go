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

// Package cmab provides contextual multi-armed bandit functionality
package cmab

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/optimizely/go-sdk/v2/pkg/cache"
	"github.com/optimizely/go-sdk/v2/pkg/config"
	"github.com/optimizely/go-sdk/v2/pkg/decide"
	"github.com/optimizely/go-sdk/v2/pkg/entities"
	"github.com/optimizely/go-sdk/v2/pkg/logging"
	"github.com/twmb/murmur3"
)

// DefaultCmabService implements the CmabService interface
type DefaultCmabService struct {
	cmabCache  cache.CacheWithRemove
	cmabClient Client
	logger     logging.OptimizelyLogProducer
}

// ServiceOptions defines options for creating a CMAB service
type ServiceOptions struct {
	Logger     logging.OptimizelyLogProducer
	CmabCache  cache.CacheWithRemove
	CmabClient Client
}

// NewDefaultCmabService creates a new instance of DefaultCmabService
func NewDefaultCmabService(options ServiceOptions) *DefaultCmabService {
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
	options *decide.Options,
) (Decision, error) {
	// Handle cache bypass early
	if s.shouldIgnoreCache(options) {
		filteredAttrs := s.filterAttributes(projectConfig, userContext, ruleID)
		return s.fetchDecisionWithRetry(ruleID, userContext.ID, filteredAttrs)
	}

	// Handle cache management options
	s.handleCacheOptions(options, userContext.ID, ruleID)

	// Try cache lookup
	if cachedDecision, found := s.tryGetCachedDecision(projectConfig, userContext, ruleID); found {
		return cachedDecision, nil
	}

	// Make fresh decision and cache it
	return s.makeFreshDecision(projectConfig, userContext, ruleID, options)
}

func (s *DefaultCmabService) shouldIgnoreCache(options *decide.Options) bool {
	return options != nil && options.IgnoreCMABCache
}

func (s *DefaultCmabService) handleCacheOptions(options *decide.Options, userID, ruleID string) {
	if options == nil {
		return
	}

	if options.ResetCMABCache {
		s.cmabCache.Reset()
	}

	if options.InvalidateUserCMABCache {
		cacheKey := s.getCacheKey(userID, ruleID)
		s.cmabCache.Remove(cacheKey)
	}
}

func (s *DefaultCmabService) tryGetCachedDecision(
	projectConfig config.ProjectConfig,
	userContext entities.UserContext,
	ruleID string,
) (Decision, bool) {
	cacheKey := s.getCacheKey(userContext.ID, ruleID)
	cachedValue := s.cmabCache.Lookup(cacheKey)

	if cachedValue == nil {
		return Decision{}, false
	}

	cv, ok := cachedValue.(CacheValue)
	if !ok {
		return Decision{}, false
	}

	// Validate cache with current attributes
	if s.isCacheValid(projectConfig, userContext, ruleID, cv) {
		s.logger.Debug(fmt.Sprintf("Returning cached CMAB decision for rule %s and user %s", ruleID, userContext.ID))
		return Decision{
			VariationID: cv.VariationID,
			CmabUUID:    cv.CmabUUID,
		}, true
	}

	// Cache invalid, remove it
	s.cmabCache.Remove(cacheKey)
	s.logger.Debug(fmt.Sprintf("Attributes changed for rule %s and user %s, invalidating cache", ruleID, userContext.ID))
	return Decision{}, false
}

func (s *DefaultCmabService) isCacheValid(
	projectConfig config.ProjectConfig,
	userContext entities.UserContext,
	ruleID string,
	cv CacheValue,
) bool {
	filteredAttrs := s.filterAttributes(projectConfig, userContext, ruleID)
	currentAttrsJSON, _ := s.getAttributesJSON(filteredAttrs)

	hasher := murmur3.SeedNew32(1)
	if _, err := hasher.Write([]byte(currentAttrsJSON)); err != nil {
		s.logger.Debug(fmt.Sprintf("Failed to hash attributes: %v", err))
		return false
	}

	currentHash := strconv.FormatUint(uint64(hasher.Sum32()), 10)
	return cv.AttributesHash == currentHash
}

func (s *DefaultCmabService) makeFreshDecision(
	projectConfig config.ProjectConfig,
	userContext entities.UserContext,
	ruleID string,
	options *decide.Options,
) (Decision, error) {
	filteredAttrs := s.filterAttributes(projectConfig, userContext, ruleID)
	cmabUUID := uuid.New().String()

	// Make API call
	decisionResult, err := s.fetchDecisionWithRetry(ruleID, userContext.ID, filteredAttrs)
	if err != nil {
		return Decision{}, fmt.Errorf("CMAB API error: %w", err)
	}

	// Cache the result if not ignoring cache
	if !s.shouldIgnoreCache(options) {
		s.cacheDecision(userContext.ID, ruleID, filteredAttrs, decisionResult.VariationID, cmabUUID)
	}

	return Decision{
		VariationID: decisionResult.VariationID,
		CmabUUID:    cmabUUID,
	}, nil
}

func (s *DefaultCmabService) cacheDecision(userID, ruleID string, filteredAttrs map[string]interface{}, variationID, cmabUUID string) {
	cacheKey := s.getCacheKey(userID, ruleID)

	attributesJSON, err := s.getAttributesJSON(filteredAttrs)
	if err != nil {
		s.logger.Debug(fmt.Sprintf("Failed to serialize attributes for caching: %v", err))
		return
	}

	hasher := murmur3.SeedNew32(1)
	if _, err := hasher.Write([]byte(attributesJSON)); err != nil {
		s.logger.Debug(fmt.Sprintf("Failed to hash attributes for caching: %v", err))
		return
	}

	attributesHash := strconv.FormatUint(uint64(hasher.Sum32()), 10)
	cacheValue := CacheValue{
		AttributesHash: attributesHash,
		VariationID:    variationID,
		CmabUUID:       cmabUUID,
	}

	s.cmabCache.Save(cacheKey, cacheValue)
	s.logger.Debug(fmt.Sprintf("Cached CMAB decision for rule %s and user %s", ruleID, userID))
}

// fetchDecisionWithRetry fetches a decision from the CMAB API with retry logic
func (s *DefaultCmabService) fetchDecisionWithRetry(
	ruleID string,
	userID string,
	attributes map[string]interface{},
) (Decision, error) {
	cmabUUID := uuid.New().String()
	reasons := []string{}

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
			reasons = append(reasons, fmt.Sprintf("Retry attempt %d/%d after backoff", attempt+1, maxRetries))
		}

		s.logger.Debug(fmt.Sprintf("Fetching CMAB decision for rule %s and user %s (attempt %d/%d)",
			ruleID, userID, attempt+1, maxRetries))

		variationID, err := s.cmabClient.FetchDecision(ruleID, userID, attributes, cmabUUID)
		if err == nil {
			reasons = append(reasons, fmt.Sprintf("Successfully fetched CMAB decision on attempt %d/%d", attempt+1, maxRetries))
			return Decision{
				VariationID: variationID,
				CmabUUID:    cmabUUID,
				Reasons:     reasons,
			}, nil
		}

		lastErr = err
		s.logger.Warning(fmt.Sprintf("CMAB API request failed (attempt %d/%d): %v",
			attempt+1, maxRetries, err))
	}

	reasons = append(reasons, fmt.Sprintf("Failed to fetch CMAB decision after %d attempts", maxRetries))
	return Decision{Reasons: reasons}, fmt.Errorf("failed to fetch CMAB decision after %d attempts: %w",
		maxRetries, lastErr)
}

// filterAttributes filters user attributes based on CMAB configuration
func (s *DefaultCmabService) filterAttributes(
	projectConfig config.ProjectConfig,
	userContext entities.UserContext,
	ruleID string,
) map[string]interface{} {
	filteredAttributes := make(map[string]interface{})

	// Get experiment by ID directly using the interface method
	targetExperiment, err := projectConfig.GetExperimentByID(ruleID)
	if err != nil || targetExperiment.Cmab == nil {
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

// getAttributesJSON serializes attributes to a JSON string
func (s *DefaultCmabService) getAttributesJSON(attributes map[string]interface{}) (string, error) {
	// Serialize to JSON - json.Marshal already sorts map keys alphabetically
	jsonBytes, err := json.Marshal(attributes)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

// getCacheKey generates a cache key for the user and rule
func (s *DefaultCmabService) getCacheKey(userID, ruleID string) string {
	// Include length of userID to avoid ambiguity when IDs contain the separator
	return fmt.Sprintf("%d:%s:%s", len(userID), userID, ruleID)
}
