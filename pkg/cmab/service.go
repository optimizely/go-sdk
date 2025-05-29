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
	// Handle cache options first
	if options != nil {
		if options.ResetCMABCache {
			s.cmabCache.Reset()
		}
		if options.InvalidateUserCMABCache {
			cacheKey := s.getCacheKey(userContext.ID, ruleID)
			s.cmabCache.Remove(cacheKey)
		}
	}

	// Check cache first
	if options == nil || !options.IgnoreCMABCache {
		cacheKey := s.getCacheKey(userContext.ID, ruleID)
		if cachedValue := s.cmabCache.Lookup(cacheKey); cachedValue != nil {
			if cv, ok := cachedValue.(CacheValue); ok {
				// Filter attributes to validate cache
				filteredAttrs := s.filterAttributes(projectConfig, userContext, ruleID)
				currentAttrsJSON, _ := s.getAttributesJSON(filteredAttrs)
				hasher := murmur3.SeedNew32(1)
				if _, err := hasher.Write([]byte(currentAttrsJSON)); err != nil {
					s.logger.Debug(fmt.Sprintf("Failed to hash attributes: %v", err))
				} else {
					currentHash := strconv.FormatUint(uint64(hasher.Sum32()), 10)

					if cv.AttributesHash == currentHash {
						// Cache hit with matching attributes
						s.logger.Debug(fmt.Sprintf("Returning cached CMAB decision for rule %s and user %s", ruleID, userContext.ID))
						return Decision{
							VariationID: cv.VariationID,
							CmabUUID:    cv.CmabUUID,
						}, nil
					}
				}

				// Attributes changed, invalidate cache
				s.cmabCache.Remove(cacheKey)
				s.logger.Debug(fmt.Sprintf("Attributes changed for rule %s and user %s, invalidating cache", ruleID, userContext.ID))
			}
		}
	}

	// Filter attributes (experiment is guaranteed to exist by ExperimentCmabService)
	filteredAttrs := s.filterAttributes(projectConfig, userContext, ruleID)

	// Generate UUID for this decision
	cmabUUID := uuid.New().String()

	// Make API call with retry (pass attributes directly)
	decisionResult, err := s.fetchDecisionWithRetry(ruleID, userContext.ID, filteredAttrs)
	if err != nil {
		return Decision{}, fmt.Errorf("CMAB API error: %w", err)
	}

	// Cache the result (if not ignoring cache)
	if options == nil || !options.IgnoreCMABCache {
		cacheKey := s.getCacheKey(userContext.ID, ruleID)

		// Generate hash for caching
		attributesJSON, err := s.getAttributesJSON(filteredAttrs)
		if err != nil {
			s.logger.Debug(fmt.Sprintf("Failed to serialize attributes for caching: %v", err))
		} else {
			hasher := murmur3.SeedNew32(1)
			if _, err := hasher.Write([]byte(attributesJSON)); err != nil {
				s.logger.Debug(fmt.Sprintf("Failed to hash attributes for caching: %v", err))
			} else {
				attributesHash := strconv.FormatUint(uint64(hasher.Sum32()), 10)

				cacheValue := CacheValue{
					AttributesHash: attributesHash,
					VariationID:    decisionResult.VariationID,
					CmabUUID:       cmabUUID,
				}
				s.cmabCache.Save(cacheKey, cacheValue)
				s.logger.Debug(fmt.Sprintf("Cached CMAB decision for rule %s and user %s", ruleID, userContext.ID))
			}
		}
	}

	return Decision{
		VariationID: decisionResult.VariationID,
		CmabUUID:    cmabUUID,
	}, nil
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

// hasOption checks if a specific CMAB option is set
func hasOption(options *decide.Options, option decide.OptimizelyDecideOptions) bool {
	if options == nil {
		return false
	}

	switch option {
	case decide.IgnoreCMABCache:
		return options.IgnoreCMABCache
	case decide.ResetCMABCache:
		return options.ResetCMABCache
	case decide.InvalidateUserCMABCache:
		return options.InvalidateUserCMABCache
	default:
		return false
	}
}
