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

// pkg/decision/cmab_service.go
package decision

import (
    "fmt"
    "sync"
    "time"

    "github.com/optimizely/go-sdk/v2/pkg/config"
    "github.com/optimizely/go-sdk/v2/pkg/entities"
    "github.com/optimizely/go-sdk/v2/pkg/logging"
)

const (
    // DefaultCacheExpiry is the default time in seconds for cache expiration
    DefaultCacheExpiry = 600 // 10 minutes
)

// DefaultCmabService implements the CmabService interface
type DefaultCmabService struct {
    client       CmabClient
    logger       logging.OptimizelyLogProducer
    cache        map[string]CmabCacheValue
    cacheMutex   sync.RWMutex
    cacheExpiry  int64
}

// CmabServiceOption defines functional options for configuring the CMAB service
type CmabServiceOption func(*DefaultCmabService)

// WithCmabLogger sets a custom logger for the CMAB service
func WithCmabLogger(logger logging.OptimizelyLogProducer) CmabServiceOption {
    return func(s *DefaultCmabService) {
        s.logger = logger
    }
}

// WithCmabClient sets a custom CMAB client
func WithCmabClient(client CmabClient) CmabServiceOption {
    return func(s *DefaultCmabService) {
        s.client = client
    }
}

// WithCacheExpiry sets a custom cache expiry time in seconds
func WithCacheExpiry(seconds int64) CmabServiceOption {
    return func(s *DefaultCmabService) {
        s.cacheExpiry = seconds
    }
}

// NewDefaultCmabService creates a new instance of DefaultCmabService with the given options
func NewDefaultCmabService(options ...CmabServiceOption) *DefaultCmabService {
    service := &DefaultCmabService{
        logger:      logging.GetLogger("", "DefaultCmabService"),
        cache:       make(map[string]CmabCacheValue),
        cacheExpiry: DefaultCacheExpiry,
        cacheMutex:  sync.RWMutex{},
    }

    // Apply options
    for _, opt := range options {
        opt(service)
    }

    // If no client was provided, create a default one
    if service.client == nil {
        service.client = NewCmabClient("")
    }

    return service
}

// GetDecision returns a CMAB decision for the given rule and user context
func (s *DefaultCmabService) GetDecision(
    projectConfig config.ProjectConfig,
    userContext entities.UserContext,
    ruleID string,
    options *CmabDecisionOptions,
) (CmabDecision, error) {
    // Initialize empty decision
    decision := CmabDecision{}

    // Handle nil options
    if options == nil {
        options = &CmabDecisionOptions{}
    }

    // Create cache key
    cacheKey := fmt.Sprintf("%s:%s", ruleID, userContext.ID)

    // Check cache if not explicitly ignored
    if !options.IgnoreCmabCache {
        cachedDecision, found := s.getCachedDecision(cacheKey)
        if found {
            s.logger.Debug(fmt.Sprintf("Returning cached CMAB decision for rule %s and user %s", ruleID, userContext.ID))
            if options.IncludeReasons {
                cachedDecision.Reasons = append(cachedDecision.Reasons, "Used cached decision")
            }
            return cachedDecision, nil
        }
    }

    // Prepare attributes for API call
    attributes := make(map[string]interface{})
    for key, value := range userContext.Attributes {
        attributes[key] = value
    }

    // Call CMAB API
	s.logger.Debug(fmt.Sprintf("Fetching CMAB decision for rule %s and user %s", ruleID, userContext.ID))
	apiDecision, err := s.client.FetchDecision(ruleID, userContext.ID, attributes)
	if err != nil {
		s.logger.Error(fmt.Sprintf("Error fetching CMAB decision: %v", err), nil)
		return decision, fmt.Errorf("failed to fetch CMAB decision: %w", err)
	}

    // Add reasons if requested
    if options.IncludeReasons {
        apiDecision.Reasons = append(apiDecision.Reasons, "Retrieved from CMAB API")
    }

    // Cache the decision
    s.cacheDecision(cacheKey, apiDecision)

    return apiDecision, nil
}

// getCachedDecision retrieves a decision from the cache if it exists and is not expired
func (s *DefaultCmabService) getCachedDecision(cacheKey string) (CmabDecision, bool) {
    s.cacheMutex.RLock()
    defer s.cacheMutex.RUnlock()

    cachedValue, exists := s.cache[cacheKey]
    if !exists {
        return CmabDecision{}, false
    }

    // Check if cache entry is expired
    now := time.Now().Unix()
    if now-cachedValue.Created > s.cacheExpiry {
        return CmabDecision{}, false
    }

    return cachedValue.Decision, true
}

// cacheDecision stores a decision in the cache
func (s *DefaultCmabService) cacheDecision(cacheKey string, decision CmabDecision) {
    s.cacheMutex.Lock()
    defer s.cacheMutex.Unlock()

    s.cache[cacheKey] = CmabCacheValue{
        Decision: decision,
        Created:  time.Now().Unix(),
    }
}

// ResetCache resets the entire CMAB cache
func (s *DefaultCmabService) ResetCache() error {
    s.cacheMutex.Lock()
    defer s.cacheMutex.Unlock()

    s.cache = make(map[string]CmabCacheValue)
    return nil
}

// InvalidateUserCache invalidates cache entries for a specific user
func (s *DefaultCmabService) InvalidateUserCache(userID string) error {
    s.cacheMutex.Lock()
    defer s.cacheMutex.Unlock()

    for key := range s.cache {
        if key[len(key)-len(userID):] == userID { // Check if key ends with userID
            delete(s.cache, key)
        }
    }
    return nil
}
