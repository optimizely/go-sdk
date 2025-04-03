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
    "testing"
    "time"

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

func (m *MockProjectConfig) GetAttributes() []entities.Attribute {
    args := m.Called()
    return args.Get(0).([]entities.Attribute)
}

func (m *MockProjectConfig) GetAttributeByKey(key string) (entities.Attribute, error) {
    args := m.Called(key)
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

func (m *MockProjectConfig) GetExperimentList() []entities.Experiment {
    args := m.Called()
    return args.Get(0).([]entities.Experiment)
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
    mockConfig     *MockProjectConfig
    cmabService    *DefaultCmabService
    testRuleID     string
    testUserID     string
    testAttributes map[string]interface{}
}

func (s *CmabServiceTestSuite) SetupTest() {
    s.mockClient = new(MockCmabClient)
    s.mockConfig = new(MockProjectConfig)

    // Set up basic expectations for the mock config
    s.mockConfig.On("GetProjectID").Return("project-123").Maybe()
    s.mockConfig.On("GetRevision").Return("1").Maybe()
    s.mockConfig.On("GetAccountID").Return("account-123").Maybe()
    s.mockConfig.On("GetAttributes").Return([]entities.Attribute{}).Maybe()
    s.mockConfig.On("GetEvents").Return([]entities.Event{}).Maybe()
    s.mockConfig.On("GetFeatureList").Return([]entities.Feature{}).Maybe()
    s.mockConfig.On("GetExperimentList").Return([]entities.Experiment{}).Maybe()
    s.mockConfig.On("GetSegmentList").Return([]string{}).Maybe()
    s.mockConfig.On("GetFlagVariationsMap").Return(map[string][]entities.Variation{}).Maybe()

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
        VariationID: "variant-1",
        CmabUUID:    "uuid-123",
        Reasons:     []string{},
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
    s.Equal(expectedDecision.VariationID, decision.VariationID)
    s.Equal(expectedDecision.CmabUUID, decision.CmabUUID)
    s.mockClient.AssertExpectations(s.T())

    // Test with IncludeReasons option
    options := &CmabDecisionOptions{
        IncludeReasons: true,
    }
    decision, err = s.cmabService.GetDecision(s.mockConfig, userContext, s.testRuleID, options)
    s.NoError(err)
    s.Contains(decision.Reasons, "Used cached decision")
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
    s.Equal("", decision.VariationID) // Should be empty
    s.mockClient.AssertExpectations(s.T())
}

func (s *CmabServiceTestSuite) TestCaching() {
    // Setup mock response
    expectedDecision := CmabDecision{
        VariationID: "variant-1",
        CmabUUID:    "uuid-123",
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
    s.Equal(expectedDecision.VariationID, decision1.VariationID)
    s.Equal(expectedDecision.CmabUUID, decision1.CmabUUID)

    // Second call should use cache
    decision2, err := s.cmabService.GetDecision(s.mockConfig, userContext, s.testRuleID, nil)
    s.NoError(err)
    s.Equal(expectedDecision.VariationID, decision2.VariationID)
    s.Equal(expectedDecision.CmabUUID, decision2.CmabUUID)

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
    s.Equal(expectedDecision.VariationID, decision3.VariationID)
    s.Equal(expectedDecision.CmabUUID, decision3.CmabUUID)
    s.mockClient.AssertNumberOfCalls(s.T(), "FetchDecision", 2)
}

func (s *CmabServiceTestSuite) TestResetCache() {
    // Setup mock response
    expectedDecision := CmabDecision{
        VariationID: "variant-1",
        CmabUUID:    "uuid-123",
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
        VariationID: "variant-1",
        CmabUUID:    "uuid-1",
    }
    user2Decision := CmabDecision{
        VariationID: "variant-2",
        CmabUUID:    "uuid-2",
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
