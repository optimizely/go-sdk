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

// Package cmab //
package cmab

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"testing"

	"github.com/optimizely/go-sdk/v2/pkg/decide"
	"github.com/optimizely/go-sdk/v2/pkg/entities"
	"github.com/optimizely/go-sdk/v2/pkg/logging"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"github.com/twmb/murmur3"
)

// MockCmabClient is a mock implementation of CmabClient
type MockCmabClient struct {
	mock.Mock
}

func (m *MockCmabClient) FetchDecision(ruleID, userID string, attributes map[string]interface{}, cmabUUID string) (string, error) {
	args := m.Called(ruleID, userID, attributes, cmabUUID)
	return args.String(0), args.Error(1)
}

// MockCache is a mock implementation of cache.CacheWithRemove
type MockCache struct {
	mock.Mock
}

func (m *MockCache) Save(key string, value interface{}) {
	m.Called(key, value)
}

func (m *MockCache) Lookup(key string) interface{} {
	args := m.Called(key)
	return args.Get(0)
}

func (m *MockCache) Reset() {
	m.Called()
}

func (m *MockCache) Remove(key string) {
	m.Called(key)
}

// MockProjectConfig is a mock implementation of config.ProjectConfig
type MockProjectConfig struct {
	mock.Mock
}

func (m *MockProjectConfig) GetProjectID() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockProjectConfig) GetRevision() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockProjectConfig) GetAccountID() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockProjectConfig) GetAnonymizeIP() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockProjectConfig) GetAttributeID(key string) string {
	args := m.Called(key)
	return args.String(0)
}

func (m *MockProjectConfig) GetAttributes() []entities.Attribute {
	args := m.Called()
	return args.Get(0).([]entities.Attribute)
}

func (m *MockProjectConfig) GetAttributeByKey(key string) (entities.Attribute, error) {
	args := m.Called(key)
	return args.Get(0).(entities.Attribute), args.Error(1)
}

func (m *MockProjectConfig) GetAttributeKeyByID(id string) (string, error) {
	args := m.Called(id)
	return args.String(0), args.Error(1)
}

func (m *MockProjectConfig) GetAudienceByID(id string) (entities.Audience, error) {
	args := m.Called(id)
	return args.Get(0).(entities.Audience), args.Error(1)
}

func (m *MockProjectConfig) GetEventByKey(key string) (entities.Event, error) {
	args := m.Called(key)
	return args.Get(0).(entities.Event), args.Error(1)
}

func (m *MockProjectConfig) GetEvents() []entities.Event {
	args := m.Called()
	return args.Get(0).([]entities.Event)
}

func (m *MockProjectConfig) GetFeatureByKey(featureKey string) (entities.Feature, error) {
	args := m.Called(featureKey)
	return args.Get(0).(entities.Feature), args.Error(1)
}

func (m *MockProjectConfig) GetExperimentByKey(experimentKey string) (entities.Experiment, error) {
	args := m.Called(experimentKey)
	return args.Get(0).(entities.Experiment), args.Error(1)
}

func (m *MockProjectConfig) GetExperimentByID(id string) (entities.Experiment, error) {
	args := m.Called(id)
	return args.Get(0).(entities.Experiment), args.Error(1)
}

func (m *MockProjectConfig) GetExperimentList() []entities.Experiment {
	args := m.Called()
	return args.Get(0).([]entities.Experiment)
}

func (m *MockProjectConfig) GetPublicKeyForODP() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockProjectConfig) GetHostForODP() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockProjectConfig) GetSegmentList() []string {
	args := m.Called()
	return args.Get(0).([]string)
}

func (m *MockProjectConfig) GetBotFiltering() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockProjectConfig) GetSdkKey() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockProjectConfig) GetEnvironmentKey() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockProjectConfig) GetVariableByKey(featureKey, variableKey string) (entities.Variable, error) {
	args := m.Called(featureKey, variableKey)
	return args.Get(0).(entities.Variable), args.Error(1)
}

func (m *MockProjectConfig) GetFeatureList() []entities.Feature {
	args := m.Called()
	return args.Get(0).([]entities.Feature)
}

func (m *MockProjectConfig) GetIntegrationList() []entities.Integration {
	args := m.Called()
	return args.Get(0).([]entities.Integration)
}

func (m *MockProjectConfig) GetRolloutList() []entities.Rollout {
	args := m.Called()
	return args.Get(0).([]entities.Rollout)
}

func (m *MockProjectConfig) GetAudienceList() []entities.Audience {
	args := m.Called()
	return args.Get(0).([]entities.Audience)
}

func (m *MockProjectConfig) GetAudienceMap() map[string]entities.Audience {
	args := m.Called()
	return args.Get(0).(map[string]entities.Audience)
}

func (m *MockProjectConfig) GetGroupByID(groupID string) (entities.Group, error) {
	args := m.Called(groupID)
	return args.Get(0).(entities.Group), args.Error(1)
}

func (m *MockProjectConfig) SendFlagDecisions() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockProjectConfig) GetFlagVariationsMap() map[string][]entities.Variation {
	args := m.Called()
	return args.Get(0).(map[string][]entities.Variation)
}

func (m *MockProjectConfig) GetDatafile() string {
	args := m.Called()
	return args.String(0)
}

type CmabServiceTestSuite struct {
	suite.Suite
	mockClient     *MockCmabClient
	mockCache      *MockCache
	mockConfig     *MockProjectConfig
	cmabService    *DefaultCmabService
	testRuleID     string
	testUserID     string
	testAttributes map[string]interface{}
}

func (s *CmabServiceTestSuite) SetupTest() {
	s.mockClient = new(MockCmabClient)
	s.mockCache = new(MockCache)
	s.mockConfig = new(MockProjectConfig)

	// Set up the CMAB service
	s.cmabService = NewDefaultCmabService(CmabServiceOptions{
		Logger:     logging.GetLogger("test", "CmabService"),
		CmabCache:  s.mockCache,
		CmabClient: s.mockClient,
	})

	s.testRuleID = "rule-123"
	s.testUserID = "user-456"
	s.testAttributes = map[string]interface{}{
		"age":      30,
		"location": "San Francisco",
	}
}

func (s *CmabServiceTestSuite) TestGetDecision() {
	// Setup mock experiment with CMAB configuration
	experiment := entities.Experiment{
		ID: s.testRuleID,
		Cmab: &entities.Cmab{
			AttributeIds: []string{"attr1", "attr2"},
		},
	}

	// Setup mock config
	s.mockConfig.On("GetAttributeKeyByID", "attr1").Return("age", nil)
	s.mockConfig.On("GetAttributeKeyByID", "attr2").Return("location", nil)
	s.mockConfig.On("GetExperimentByID", s.testRuleID).Return(experiment, nil)

	// Create user context
	userContext := entities.UserContext{
		ID:         s.testUserID,
		Attributes: s.testAttributes,
	}

	// Setup cache key
	cacheKey := s.cmabService.getCacheKey(s.testUserID, s.testRuleID)

	// Setup cache lookup - return nil to simulate cache miss
	s.mockCache.On("Lookup", cacheKey).Return(nil)

	// Setup mock API response
	expectedVariationID := "variant-1"
	s.mockClient.On("FetchDecision", s.testRuleID, s.testUserID, mock.Anything, mock.Anything).Return(expectedVariationID, nil)

	// Setup cache save
	s.mockCache.On("Save", cacheKey, mock.Anything).Return()

	// Test with no options
	decision, err := s.cmabService.GetDecision(s.mockConfig, userContext, s.testRuleID, nil)
	s.NoError(err)
	s.Equal(expectedVariationID, decision.VariationID)
	s.NotEmpty(decision.CmabUUID)

	// Verify expectations
	s.mockConfig.AssertExpectations(s.T())
	s.mockCache.AssertExpectations(s.T())
	s.mockClient.AssertExpectations(s.T())
}

func (s *CmabServiceTestSuite) TestGetDecisionWithCache() {
	// Setup mock experiment with CMAB configuration
	experiment := entities.Experiment{
		ID: s.testRuleID,
		Cmab: &entities.Cmab{
			AttributeIds: []string{"attr1", "attr2"},
		},
	}

	// Setup mock config
	s.mockConfig.On("GetAttributeKeyByID", "attr1").Return("age", nil)
	s.mockConfig.On("GetAttributeKeyByID", "attr2").Return("location", nil)
	s.mockConfig.On("GetExperimentByID", s.testRuleID).Return(experiment, nil)

	// Create user context
	userContext := entities.UserContext{
		ID:         s.testUserID,
		Attributes: s.testAttributes,
	}

	// Setup cache key
	cacheKey := s.cmabService.getCacheKey(s.testUserID, s.testRuleID)

	// Calculate attributes hash using murmur3 as in your implementation
	attributesJSON, _ := s.cmabService.getAttributesJSON(s.testAttributes)
	hasher := murmur3.SeedNew32(1)
	hasher.Write([]byte(attributesJSON))
	attributesHash := strconv.FormatUint(uint64(hasher.Sum32()), 10)

	// Setup cache hit with matching attributes hash
	cachedValue := CacheValue{
		AttributesHash: attributesHash,
		VariationID:    "cached-variant",
		CmabUUID:       "cached-uuid",
	}
	s.mockCache.On("Lookup", cacheKey).Return(cachedValue)

	// Test with cache hit
	decision, err := s.cmabService.GetDecision(s.mockConfig, userContext, s.testRuleID, nil)
	s.NoError(err)
	s.Equal("cached-variant", decision.VariationID)
	s.Equal("cached-uuid", decision.CmabUUID)

	// Verify API was not called
	s.mockClient.AssertNotCalled(s.T(), "FetchDecision")
}

func (s *CmabServiceTestSuite) TestGetDecisionWithIgnoreCache() {
	// Setup mock experiment with CMAB configuration
	experiment := entities.Experiment{
		ID: s.testRuleID,
		Cmab: &entities.Cmab{
			AttributeIds: []string{"attr1", "attr2"},
		},
	}

	// Setup mock config
	s.mockConfig.On("GetAttributeKeyByID", "attr1").Return("age", nil)
	s.mockConfig.On("GetAttributeKeyByID", "attr2").Return("location", nil)
	s.mockConfig.On("GetExperimentByID", s.testRuleID).Return(experiment, nil)

	// Create user context
	userContext := entities.UserContext{
		ID:         s.testUserID,
		Attributes: s.testAttributes,
	}

	// Setup cache key
	cacheKey := s.cmabService.getCacheKey(s.testUserID, s.testRuleID)

	// Setup mock API response
	expectedVariationID := "variant-1"
	s.mockClient.On("FetchDecision", s.testRuleID, s.testUserID, mock.Anything, mock.Anything).Return(expectedVariationID, nil)

	// Test with IgnoreCMABCache option
	options := &decide.Options{
		IgnoreCMABCache: true,
	}

	decision, err := s.cmabService.GetDecision(s.mockConfig, userContext, s.testRuleID, options)
	s.NoError(err)
	s.Equal(expectedVariationID, decision.VariationID)

	// Verify API was called (cache was ignored)
	s.mockClient.AssertCalled(s.T(), "FetchDecision", s.testRuleID, s.testUserID, mock.Anything, mock.Anything)

	// Verify cache lookup was not called
	s.mockCache.AssertNotCalled(s.T(), "Lookup", cacheKey)

	// Verify cache save was not called
	s.mockCache.AssertNotCalled(s.T(), "Save", cacheKey, mock.Anything)
}

func (s *CmabServiceTestSuite) TestGetDecisionWithResetCache() {
	// Setup mock experiment with CMAB configuration
	experiment := entities.Experiment{
		ID: s.testRuleID,
		Cmab: &entities.Cmab{
			AttributeIds: []string{"attr1", "attr2"},
		},
	}

	// Setup mock config
	s.mockConfig.On("GetAttributeKeyByID", "attr1").Return("age", nil)
	s.mockConfig.On("GetAttributeKeyByID", "attr2").Return("location", nil)
	s.mockConfig.On("GetExperimentByID", s.testRuleID).Return(experiment, nil)

	// Create user context
	userContext := entities.UserContext{
		ID:         s.testUserID,
		Attributes: s.testAttributes,
	}

	// Setup cache reset
	s.mockCache.On("Reset").Return()

	// Setup cache key
	cacheKey := s.cmabService.getCacheKey(s.testUserID, s.testRuleID)

	// Setup cache lookup after reset
	s.mockCache.On("Lookup", cacheKey).Return(nil)

	// Setup mock API response
	expectedVariationID := "variant-1"
	s.mockClient.On("FetchDecision", s.testRuleID, s.testUserID, mock.Anything, mock.Anything).Return(expectedVariationID, nil)

	// Setup cache save
	s.mockCache.On("Save", cacheKey, mock.Anything).Return()

	// Test with ResetCMABCache option
	options := &decide.Options{
		ResetCMABCache: true,
	}

	decision, err := s.cmabService.GetDecision(s.mockConfig, userContext, s.testRuleID, options)
	s.NoError(err)
	s.Equal(expectedVariationID, decision.VariationID)

	// Verify cache was reset
	s.mockCache.AssertCalled(s.T(), "Reset")
}

func (s *CmabServiceTestSuite) TestGetDecisionWithInvalidateUserCache() {
	// Setup mock experiment with CMAB configuration
	experiment := entities.Experiment{
		ID: s.testRuleID,
		Cmab: &entities.Cmab{
			AttributeIds: []string{"attr1", "attr2"},
		},
	}

	// Setup mock config
	s.mockConfig.On("GetAttributeKeyByID", "attr1").Return("age", nil)
	s.mockConfig.On("GetAttributeKeyByID", "attr2").Return("location", nil)
	s.mockConfig.On("GetExperimentByID", s.testRuleID).Return(experiment, nil)

	// Create user context
	userContext := entities.UserContext{
		ID:         s.testUserID,
		Attributes: s.testAttributes,
	}

	// Setup cache key
	cacheKey := s.cmabService.getCacheKey(s.testUserID, s.testRuleID)

	// Setup cache remove
	s.mockCache.On("Remove", cacheKey).Return()

	// Setup cache lookup after remove
	s.mockCache.On("Lookup", cacheKey).Return(nil)

	// Setup mock API response
	expectedVariationID := "variant-1"
	s.mockClient.On("FetchDecision", s.testRuleID, s.testUserID, mock.Anything, mock.Anything).Return(expectedVariationID, nil)

	// Setup cache save
	s.mockCache.On("Save", cacheKey, mock.Anything).Return()

	// Test with InvalidateUserCMABCache option
	options := &decide.Options{
		InvalidateUserCMABCache: true,
	}

	decision, err := s.cmabService.GetDecision(s.mockConfig, userContext, s.testRuleID, options)
	s.NoError(err)
	s.Equal(expectedVariationID, decision.VariationID)

	// Verify user cache was invalidated
	s.mockCache.AssertCalled(s.T(), "Remove", cacheKey)
}

func (s *CmabServiceTestSuite) TestGetDecisionError() {
	// Setup mock experiment with CMAB configuration
	experiment := entities.Experiment{
		ID: s.testRuleID,
		Cmab: &entities.Cmab{
			AttributeIds: []string{"attr1", "attr2"},
		},
	}

	// Setup mock config
	s.mockConfig.On("GetAttributeKeyByID", "attr1").Return("age", nil)
	s.mockConfig.On("GetAttributeKeyByID", "attr2").Return("location", nil)
	s.mockConfig.On("GetExperimentByID", s.testRuleID).Return(experiment, nil)

	// Create user context
	userContext := entities.UserContext{
		ID:         s.testUserID,
		Attributes: s.testAttributes,
	}

	// Setup cache key
	cacheKey := s.cmabService.getCacheKey(s.testUserID, s.testRuleID)

	// Setup cache miss
	s.mockCache.On("Lookup", cacheKey).Return(nil)

	// Setup mock API error
	expectedError := errors.New("API error")
	s.mockClient.On("FetchDecision", s.testRuleID, s.testUserID, mock.Anything, mock.Anything).Return("", expectedError)

	// Test error handling
	decision, err := s.cmabService.GetDecision(s.mockConfig, userContext, s.testRuleID, nil)
	s.Error(err)
	s.Equal("", decision.VariationID) // Should be empty
}

func (s *CmabServiceTestSuite) TestFilterAttributes() {
	// Setup mock experiment with CMAB configuration
	experiment := entities.Experiment{
		ID: s.testRuleID,
		Cmab: &entities.Cmab{
			AttributeIds: []string{"attr1", "attr2", "attr3"},
		},
	}

	// Setup mock config
	s.mockConfig.On("GetAttributeKeyByID", "attr1").Return("age", nil)
	s.mockConfig.On("GetAttributeKeyByID", "attr2").Return("location", nil)
	s.mockConfig.On("GetAttributeKeyByID", "attr3").Return("", errors.New("attribute not found"))
	s.mockConfig.On("GetExperimentByID", s.testRuleID).Return(experiment, nil)

	// Create user context with extra attributes that should be filtered out
	userContext := entities.UserContext{
		ID: s.testUserID,
		Attributes: map[string]interface{}{
			"age":       30,
			"location":  "San Francisco",
			"extra_key": "should be filtered out",
		},
	}

	// Call filterAttributes directly
	filteredAttrs := s.cmabService.filterAttributes(s.mockConfig, userContext, s.testRuleID)

	// Verify only the configured attributes are included
	s.Equal(2, len(filteredAttrs))
	s.Equal(30, filteredAttrs["age"])
	s.Equal("San Francisco", filteredAttrs["location"])
	s.NotContains(filteredAttrs, "extra_key")
}

func (s *CmabServiceTestSuite) TestOnlyFilteredAttributesPassedToClient() {
	// Setup mock experiment with CMAB configuration
	experiment := entities.Experiment{
		ID: s.testRuleID,
		Cmab: &entities.Cmab{
			AttributeIds: []string{"attr1", "attr2"},
		},
	}

	// Setup mock config
	s.mockConfig.On("GetAttributeKeyByID", "attr1").Return("age", nil)
	s.mockConfig.On("GetAttributeKeyByID", "attr2").Return("location", nil)
	s.mockConfig.On("GetExperimentByID", s.testRuleID).Return(experiment, nil)

	// Create user context with extra attributes that should be filtered out
	userContext := entities.UserContext{
		ID: s.testUserID,
		Attributes: map[string]interface{}{
			"age":       30,
			"location":  "San Francisco",
			"extra_key": "should be filtered out",
		},
	}

	// Expected filtered attributes
	expectedFilteredAttrs := map[string]interface{}{
		"age":      30,
		"location": "San Francisco",
	}

	// Setup cache key
	cacheKey := s.cmabService.getCacheKey(s.testUserID, s.testRuleID)

	// Setup cache lookup
	s.mockCache.On("Lookup", cacheKey).Return(nil)

	// Setup mock API response with attribute verification
	expectedVariationID := "variant-1"
	s.mockClient.On("FetchDecision", s.testRuleID, s.testUserID, mock.MatchedBy(func(attrs map[string]interface{}) bool {
		// Verify only the filtered attributes are passed
		if len(attrs) != 2 {
			return false
		}
		if attrs["age"] != 30 {
			return false
		}
		if attrs["location"] != "San Francisco" {
			return false
		}
		if _, exists := attrs["extra_key"]; exists {
			return false
		}
		return true
	}), mock.Anything).Return(expectedVariationID, nil)

	// Setup cache save
	s.mockCache.On("Save", cacheKey, mock.Anything).Return()

	// Call GetDecision
	decision, err := s.cmabService.GetDecision(s.mockConfig, userContext, s.testRuleID, nil)
	s.NoError(err)
	s.Equal(expectedVariationID, decision.VariationID)

	// Verify client was called with the filtered attributes
	s.mockClient.AssertCalled(s.T(), "FetchDecision", s.testRuleID, s.testUserID, mock.MatchedBy(func(attrs map[string]interface{}) bool {
		return reflect.DeepEqual(attrs, expectedFilteredAttrs)
	}), mock.Anything)
}

func (s *CmabServiceTestSuite) TestCacheInvalidatedWhenAttributesChange() {
	// Setup mock experiment with CMAB configuration
	experiment := entities.Experiment{
		ID: s.testRuleID,
		Cmab: &entities.Cmab{
			AttributeIds: []string{"attr1", "attr2"},
		},
	}

	// Setup mock config
	s.mockConfig.On("GetAttributeKeyByID", "attr1").Return("age", nil)
	s.mockConfig.On("GetAttributeKeyByID", "attr2").Return("location", nil)
	s.mockConfig.On("GetExperimentByID", s.testRuleID).Return(experiment, nil)

	// Create user context
	userContext := entities.UserContext{
		ID: s.testUserID,
		Attributes: map[string]interface{}{
			"age":      30,
			"location": "San Francisco",
		},
	}

	// Setup cache key
	cacheKey := s.cmabService.getCacheKey(s.testUserID, s.testRuleID)

	// First, create a cached value with a different attributes hash
	oldAttributesHash := "old-hash"
	cachedValue := CacheValue{
		AttributesHash: oldAttributesHash,
		VariationID:    "cached-variant",
		CmabUUID:       "cached-uuid",
	}

	// Setup cache lookup to return the cached value
	s.mockCache.On("Lookup", cacheKey).Return(cachedValue)

	// Setup cache remove (should be called when attributes change)
	s.mockCache.On("Remove", cacheKey).Return()

	// Setup mock API response (should be called when attributes change)
	expectedVariationID := "new-variant"
	s.mockClient.On("FetchDecision", s.testRuleID, s.testUserID, mock.Anything, mock.Anything).Return(expectedVariationID, nil)

	// Setup cache save for the new decision
	s.mockCache.On("Save", cacheKey, mock.Anything).Return()

	// Call GetDecision
	decision, err := s.cmabService.GetDecision(s.mockConfig, userContext, s.testRuleID, nil)
	s.NoError(err)
	s.Equal(expectedVariationID, decision.VariationID)

	// Verify cache was looked up
	s.mockCache.AssertCalled(s.T(), "Lookup", cacheKey)

	// Verify cache entry was removed due to attribute change
	s.mockCache.AssertCalled(s.T(), "Remove", cacheKey)

	// Verify API was called to get a new decision
	s.mockClient.AssertCalled(s.T(), "FetchDecision", s.testRuleID, s.testUserID, mock.Anything, mock.Anything)

	// Verify new decision was cached
	s.mockCache.AssertCalled(s.T(), "Save", cacheKey, mock.MatchedBy(func(value CacheValue) bool {
		return value.VariationID == expectedVariationID && value.AttributesHash != oldAttributesHash
	}))
}

func (s *CmabServiceTestSuite) TestGetAttributesJSON() {
	// Test with empty attributes
	emptyJSON, err := s.cmabService.getAttributesJSON(map[string]interface{}{})
	s.NoError(err)
	s.Equal("{}", emptyJSON)

	// Test with attributes
	attributes := map[string]interface{}{
		"c": 3,
		"a": 1,
		"b": 2,
	}
	json, err := s.cmabService.getAttributesJSON(attributes)
	s.NoError(err)
	// Keys should be sorted alphabetically
	s.Equal(`{"a":1,"b":2,"c":3}`, json)
}

func (s *CmabServiceTestSuite) TestGetCacheKey() {
	// Update the expected format to match the new implementation
	expected := fmt.Sprintf("%d:%s:%s", len("user123"), "user123", "rule456")
	actual := s.cmabService.getCacheKey("user123", "rule456")
	s.Equal(expected, actual)
}

func (s *CmabServiceTestSuite) TestNewDefaultCmabService() {
	// Test with default options
	service := NewDefaultCmabService(CmabServiceOptions{})

	// Only check that the service is created, not the specific fields
	s.NotNil(service)
}

func TestCmabServiceTestSuite(t *testing.T) {
	suite.Run(t, new(CmabServiceTestSuite))
}
