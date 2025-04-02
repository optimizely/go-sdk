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

// Package decision //
package decision

import (
    "errors"
    "fmt"
    "sync"
    "testing"
    "time"

    "github.com/optimizely/go-sdk/v2/pkg/config"
    "github.com/optimizely/go-sdk/v2/pkg/entities"
    "github.com/optimizely/go-sdk/v2/pkg/logging"
    "github.com/stretchr/testify/mock"
    "github.com/stretchr/testify/suite"
)

// MockCmabClient is a mock implementation of CmabClient
type MockCmabClient struct {
    mock.Mock
}

func (m *MockCmabClient) FetchDecision(ruleID, userID string, attributes map[string]interface{}) (CmabDecision, error) {
    args := m.Called(ruleID, userID, attributes)
    return args.Get(0).(CmabDecision), args.Error(1)
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

func (m *MockProjectConfig) GetAttributeByID(id string) (entities.Attribute, error) {
    args := m.Called(id)
    return args.Get(0).(entities.Attribute), args.Error(1)
}

func (m *MockProjectConfig) GetAudienceByID(id string) (entities.Audience, error) {
    args := m.Called(id)
    return args.Get(0).(entities.Audience), args.Error(1)
}

func (m *MockProjectConfig) GetEventByKey(key string) (entities.Event, error) {
    args := m.Called(key)
    return args.Get(0).(entities.Event), args.Error(1)
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

func (m *MockProjectConfig) GetRolloutByID(rolloutID string) (entities.Rollout, error) {
    args := m.Called(rolloutID)
    return args.Get(0).(entities.Rollout), args.Error(1)
}

func (m *MockProjectConfig) GetVariationByKey(experimentKey, variationKey string) (entities.Variation, error) {
    args := m.Called(experimentKey, variationKey)
    return args.Get(0).(entities.Variation), args.Error(1)
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

type CmabServiceTestSuite struct {
    suite.Suite
    mockClient     *MockCmabClient
    mockConfig     *MockProjectConfig
    cmabService    *DefaultCmabService
    testRuleID     string
    testUserID     string
    testAttributes map[string]interface{}
}

func (s *CmabServiceTestSuite) SetupTest() {
    s.mockClient = new(MockCmabClient)
    s.mockConfig = new(MockProjectConfig)
    s.cmabService = NewDefaultCmabService(
        WithCmabLogger(logging.GetLogger("test", "CmabService")),
        WithCmabClient(s.mockClient),
        WithCacheExpiry(10), // 10 seconds for faster testing
    )
    s.testRuleID = "rule-123"
    s.testUserID = "user-456"
    s.testAttributes = map[string]interface{}{
        "age":      30,
        "location": "San Francisco",
    }
}

func (s *CmabServiceTestSuite) TestGetDecision() {
    // Setup mock response
    expectedDecision := CmabDecision{
        RuleID:     s.testRuleID,
        UserID:     s.testUserID,
        VariantID:  "variant-1",
        Attributes: map[string]interface{}{"key": "value"},
    }
    s.mockClient.On("FetchDecision", s.testRuleID, s.testUserID, s.testAttributes).Return(expectedDecision, nil)

    // Create user context
    userContext := entities.UserContext{
        ID:         s.testUserID,
        Attributes: s.testAttributes,
    }

    // Test with default options
    decision, err := s.cmabService.GetDecision(s.mockConfig, userContext, s.testRuleID, nil)
    s.NoError(err)
    s.Equal(expectedDecision, decision)
    s.mockClient.AssertExpectations(s.T())

    // Test with IncludeReasons option
    options := &CmabDecisionOptions{
        IncludeReasons: true,
    }
    decision, err = s.cmabService.GetDecision(s.mockConfig, userContext, s.testRuleID, options)
    s.NoError(err)
    s.Contains(decision.Reasons, "Retrieved from CMAB API")
    s.mockClient.AssertNumberOfCalls(s.T(), "FetchDecision", 1) // Should use cache

    // Test with IgnoreCmabCache option
    options = &CmabDecisionOptions{
        IgnoreCmabCache: true,
        IncludeReasons:  true,
    }
    s.mockClient.On("FetchDecision", s.testRuleID, s.testUserID, s.testAttributes).Return(expectedDecision, nil)
    decision, err = s.cmabService.GetDecision(s.mockConfig, userContext, s.testRuleID, options)
    s.NoError(err)
    s.Contains(decision.Reasons, "Retrieved from CMAB API")
    s.mockClient.AssertNumberOfCalls(s.T(), "FetchDecision", 2) // Should call API again
}

func (s *CmabServiceTestSuite) TestGetDecisionError() {
    // Setup mock error response
    expectedError := errors.New("API error")
    s.mockClient.On("FetchDecision", s.testRuleID, s.testUserID, s.testAttributes).Return(CmabDecision{}, expectedError)

    // Create user context
    userContext := entities.UserContext{
        ID:         s.testUserID,
        Attributes: s.testAttributes,
    }

    // Test error handling
    decision, err := s.cmabService.GetDecision(s.mockConfig, userContext, s.testRuleID, nil)
    s.Error(err)
    s.Equal(CmabDecision{}, decision)
    s.mockClient.AssertExpectations(s.T())
}

func (s *CmabServiceTestSuite) TestCaching() {
    // Setup mock response
    expectedDecision := CmabDecision{
        RuleID:     s.testRuleID,
        UserID:     s.testUserID,
        VariantID:  "variant-1",
        Attributes: map[string]interface{}{"key": "value"},
    }
    s.mockClient.On("FetchDecision", s.testRuleID, s.testUserID, s.testAttributes).Return(expectedDecision, nil).Once()

    // Create user context
    userContext := entities.UserContext{
        ID:         s.testUserID,
        Attributes: s.testAttributes,
    }

    // First call should hit the API
    decision1, err := s.cmabService.GetDecision(s.mockConfig, userContext, s.testRuleID, nil)
    s.NoError(err)
    s.Equal(expectedDecision, decision1)

    // Second call should use cache
    decision2, err := s.cmabService.GetDecision(s.mockConfig, userContext, s.testRuleID, nil)
    s.NoError(err)
    s.Equal(expectedDecision, decision2)

    // Verify API was only called once
    s.mockClient.AssertNumberOfCalls(s.T(), "FetchDecision", 1)

    // Test cache expiration
    s.cmabService.cacheExpiry = 1 // Set to 1 second
    time.Sleep(2 * time.Second)   // Wait for cache to expire

    // Setup mock for second API call
    s.mockClient.On("FetchDecision", s.testRuleID, s.testUserID, s.testAttributes).Return(expectedDecision, nil).Once()

    // Third call should hit API again due to cache expiration
    decision3, err := s.cmabService.GetDecision(s.mockConfig, userContext, s.testRuleID, nil)
    s.NoError(err)
    s.Equal(expectedDecision, decision3)
    s.mockClient.AssertNumberOfCalls(s.T(), "FetchDecision", 2)
}

func (s *CmabServiceTestSuite) TestResetCache() {
    // Setup mock response
    expectedDecision := CmabDecision{
        RuleID:     s.testRuleID,
        UserID:     s.testUserID,
        VariantID:  "variant-1",
        Attributes: map[string]interface{}{"key": "value"},
    }
    s.mockClient.On("FetchDecision", s.testRuleID, s.testUserID, s.testAttributes).Return(expectedDecision, nil)

    // Create user context
    userContext := entities.UserContext{
        ID:         s.testUserID,
        Attributes: s.testAttributes,
    }

    // First call should hit the API
    _, err := s.cmabService.GetDecision(s.mockConfig, userContext, s.testRuleID, nil)
    s.NoError(err)

    // Reset cache
    err = s.cmabService.ResetCache()
    s.NoError(err)

    // Setup mock for second API call
    s.mockClient.On("FetchDecision", s.testRuleID, s.testUserID, s.testAttributes).Return(expectedDecision, nil)

    // Second call should hit API again due to cache reset
    _, err = s.cmabService.GetDecision(s.mockConfig, userContext, s.testRuleID, nil)
    s.NoError(err)
    s.mockClient.AssertNumberOfCalls(s.T(), "FetchDecision", 2)
}

func (s *CmabServiceTestSuite) TestInvalidateUserCache() {
    // Setup mock responses for two different users
    user1Decision := CmabDecision{
        RuleID:     s.testRuleID,
        UserID:     "user-1",
        VariantID:  "variant-1",
        Attributes: map[string]interface{}{"key": "value"},
    }
    user2Decision := CmabDecision{
        RuleID:     s.testRuleID,
        UserID:     "user-2",
        VariantID:  "variant-2",
        Attributes: map[string]interface{}{"key": "value"},
    }

    // Create user contexts
    user1Context := entities.UserContext{
        ID:         "user-1",
        Attributes: s.testAttributes,
    }
    user2Context := entities.UserContext{
        ID:         "user-2",
        Attributes: s.testAttributes,
    }

    // Setup mocks
    s.mockClient.On("FetchDecision", s.testRuleID, "user-1", s.testAttributes).Return(user1Decision, nil)
    s.mockClient.On("FetchDecision", s.testRuleID, "user-2", s.testAttributes).Return(user2Decision, nil)

    // First calls should hit the API
    _, err := s.cmabService.GetDecision(s.mockConfig, user1Context, s.testRuleID, nil)
    s.NoError(err)
    _, err = s.cmabService.GetDecision(s.mockConfig, user2Context, s.testRuleID, nil)
    s.NoError(err)

    // Invalidate user-1 cache
    err = s.cmabService.InvalidateUserCache("user-1")
    s.NoError(err)

    // Setup mock for user-1 second API call
    s.mockClient.On("FetchDecision", s.testRuleID, "user-1", s.testAttributes).Return(user1Decision, nil)

    // user-1 call should hit API again due to cache invalidation
    _, err = s.cmabService.GetDecision(s.mockConfig, user1Context, s.testRuleID, nil)
    s.NoError(err)
    // user-2 call should use cache
    _, err = s.cmabService.GetDecision(s.mockConfig, user2Context, s.testRuleID, nil)
    s.NoError(err)

    // Verify API calls
    s.mockClient.AssertNumberOfCalls(s.T(), "FetchDecision", 3) // 2 initial + 1 after invalidation
}

func (s *CmabServiceTestSuite) TestAsyncBehavior() {
    // Setup mock response
    expectedDecision := CmabDecision{
        RuleID:    s.testRuleID,
        UserID:    s.testUserID,
        VariantID: "variant-1",
    }

    // Create multiple user contexts
    const numUsers = 100
    userContexts := make([]entities.UserContext, numUsers)
    for i := 0; i < numUsers; i++ {
        userID := fmt.Sprintf("user-%d", i)
        userContexts[i] = entities.UserContext{
            ID:         userID,
            Attributes: s.testAttributes,
        }
        // Setup mock for each user
        s.mockClient.On("FetchDecision", s.testRuleID, userID, s.testAttributes).Return(expectedDecision, nil).Maybe()
    }

    var wg sync.WaitGroup
    wg.Add(3)

    // Goroutine 1: Get decisions for first half of users
    go func() {
        defer wg.Done()
        for i := 0; i < numUsers/2; i++ {
            _, _ = s.cmabService.GetDecision(s.mockConfig, userContexts[i], s.testRuleID, nil)
        }
    }()

    // Goroutine 2: Get decisions for second half of users
    go func() {
        defer wg.Done()
        for i := numUsers / 2; i < numUsers; i++ {
            _, _ = s.cmabService.GetDecision(s.mockConfig, userContexts[i], s.testRuleID, nil)
        }
    }()

    // Goroutine 3: Invalidate cache for some users
    go func() {
        defer wg.Done()
        for i := 0; i < numUsers; i += 10 {
            _ = s.cmabService.InvalidateUserCache(userContexts[i].ID)
        }
    }()

    wg.Wait()

    // Now reset cache while getting decisions
    wg.Add(2)

    // Goroutine 1: Get decisions
    go func() {
        defer wg.Done()
        for i := 0; i < numUsers; i += 5 {
            _, _ = s.cmabService.GetDecision(s.mockConfig, userContexts[i], s.testRuleID, nil)
        }
    }()

    // Goroutine 2: Reset cache
    go func() {
        defer wg.Done()
        _ = s.cmabService.ResetCache()
    }()

    wg.Wait()
}

func (s *CmabServiceTestSuite) TestDefaultConstructor() {
    // Test that the default constructor sets reasonable defaults
    service := NewDefaultCmabService()
    s.NotNil(service.logger)
    s.NotNil(service.cache)
    s.Equal(int64(DefaultCacheExpiry), service.cacheExpiry)
    s.NotNil(service.client)
}

func TestCmabServiceTestSuite(t *testing.T) {
    suite.Run(t, new(CmabServiceTestSuite))
}
