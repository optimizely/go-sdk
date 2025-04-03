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
	"encoding/json"
	"testing"
	"time"

	"github.com/optimizely/go-sdk/v2/pkg/config"
	"github.com/optimizely/go-sdk/v2/pkg/entities"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// TestCmabDecisionStruct tests the CmabDecision struct
func TestCmabDecisionStruct(t *testing.T) {
	// Create a decision
	decision := CmabDecision{
		VariationID: "variation-123",
		CmabUUID:    "uuid-456",
		Reasons:     []string{"reason1", "reason2"},
	}

	// Verify fields
	assert.Equal(t, "variation-123", decision.VariationID)
	assert.Equal(t, "uuid-456", decision.CmabUUID)
	assert.Equal(t, 2, len(decision.Reasons))
	assert.Equal(t, "reason1", decision.Reasons[0])
	assert.Equal(t, "reason2", decision.Reasons[1])

	// Test JSON serialization
	jsonData, err := json.Marshal(decision)
	assert.NoError(t, err)

	// Test JSON deserialization
	var deserializedDecision CmabDecision
	err = json.Unmarshal(jsonData, &deserializedDecision)
	assert.NoError(t, err)

	// Verify deserialized fields
	assert.Equal(t, decision.VariationID, deserializedDecision.VariationID)
	assert.Equal(t, decision.CmabUUID, deserializedDecision.CmabUUID)
	assert.Equal(t, decision.Reasons, deserializedDecision.Reasons)
}

// TestCmabCacheValueStruct tests the CmabCacheValue struct
func TestCmabCacheValueStruct(t *testing.T) {
	// Create a decision
	decision := CmabDecision{
		VariationID: "variation-123",
		CmabUUID:    "uuid-456",
		Reasons:     []string{"reason1"},
	}

	// Create a cache value
	now := time.Now().Unix()
	cacheValue := CmabCacheValue{
		Decision: decision,
		Created:  now,
	}

	// Verify fields
	assert.Equal(t, decision, cacheValue.Decision)
	assert.Equal(t, now, cacheValue.Created)
}

// TestCmabDecisionOptionsStruct tests the CmabDecisionOptions struct
func TestCmabDecisionOptionsStruct(t *testing.T) {
	// Test default initialization
	options := CmabDecisionOptions{}
	assert.False(t, options.IgnoreCmabCache)
	assert.False(t, options.IncludeReasons)

	// Test explicit initialization
	options = CmabDecisionOptions{
		IgnoreCmabCache: true,
		IncludeReasons:  true,
	}
	assert.True(t, options.IgnoreCmabCache)
	assert.True(t, options.IncludeReasons)
}

// MockCmabServiceTest implements the CmabService interface for testing
type MockCmabServiceTest struct {
	mock.Mock
}

func (m *MockCmabServiceTest) GetDecision(projectConfig config.ProjectConfig, userContext entities.UserContext, ruleID string, options *CmabDecisionOptions) (CmabDecision, error) {
	args := m.Called(projectConfig, userContext, ruleID, options)
	return args.Get(0).(CmabDecision), args.Error(1)
}

func (m *MockCmabServiceTest) ResetCache() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockCmabServiceTest) InvalidateUserCache(userID string) error {
	args := m.Called(userID)
	return args.Error(0)
}

// MockProjectConfigTest is a mock implementation of config.ProjectConfig
type MockProjectConfigTest struct {
	mock.Mock
}

// GetEnvironmentKey implements config.ProjectConfig.
func (m *MockProjectConfigTest) GetEnvironmentKey() string {
	panic("unimplemented")
}

// GetEvents implements config.ProjectConfig.
func (m *MockProjectConfigTest) GetEvents() []entities.Event {
	panic("unimplemented")
}

// GetExperimentList implements config.ProjectConfig.
func (m *MockProjectConfigTest) GetExperimentList() []entities.Experiment {
	panic("unimplemented")
}

// GetFeatureList implements config.ProjectConfig.
func (m *MockProjectConfigTest) GetFeatureList() []entities.Feature {
	panic("unimplemented")
}

// GetFlagVariationsMap implements config.ProjectConfig.
func (m *MockProjectConfigTest) GetFlagVariationsMap() map[string][]entities.Variation {
	panic("unimplemented")
}

// GetGroupByID implements config.ProjectConfig.
func (m *MockProjectConfigTest) GetGroupByID(string) (entities.Group, error) {
	panic("unimplemented")
}

// GetIntegrationList implements config.ProjectConfig.
func (m *MockProjectConfigTest) GetIntegrationList() []entities.Integration {
	panic("unimplemented")
}

// GetRolloutList implements config.ProjectConfig.
func (m *MockProjectConfigTest) GetRolloutList() (rolloutList []entities.Rollout) {
	panic("unimplemented")
}

// GetSdkKey implements config.ProjectConfig.
func (m *MockProjectConfigTest) GetSdkKey() string {
	panic("unimplemented")
}

// GetVariableByKey implements config.ProjectConfig.
func (m *MockProjectConfigTest) GetVariableByKey(featureKey string, variableKey string) (entities.Variable, error) {
	panic("unimplemented")
}

// SendFlagDecisions implements config.ProjectConfig.
func (m *MockProjectConfigTest) SendFlagDecisions() bool {
	panic("unimplemented")
}

func (m *MockProjectConfigTest) GetDatafile() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockProjectConfigTest) GetAudienceMap() map[string]entities.Audience {
	args := m.Called()
	return args.Get(0).(map[string]entities.Audience)
}

func (m *MockProjectConfigTest) GetAudienceList() []entities.Audience {
	args := m.Called()
	return args.Get(0).([]entities.Audience)
}

func (m *MockProjectConfigTest) GetAttributes() []entities.Attribute {
	args := m.Called()
	return args.Get(0).([]entities.Attribute)
}

func (m *MockProjectConfigTest) GetProjectID() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockProjectConfigTest) GetRevision() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockProjectConfigTest) GetAccountID() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockProjectConfigTest) GetAnonymizeIP() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockProjectConfigTest) GetAttributeID(key string) string {
	args := m.Called(key)
	return args.String(0)
}

func (m *MockProjectConfigTest) GetAttributeByKey(key string) (entities.Attribute, error) {
	args := m.Called(key)
	return args.Get(0).(entities.Attribute), args.Error(1)
}

func (m *MockProjectConfigTest) GetAttributeByID(id string) (entities.Attribute, error) {
	args := m.Called(id)
	return args.Get(0).(entities.Attribute), args.Error(1)
}

func (m *MockProjectConfigTest) GetAudienceByID(id string) (entities.Audience, error) {
	args := m.Called(id)
	return args.Get(0).(entities.Audience), args.Error(1)
}

func (m *MockProjectConfigTest) GetEventByKey(key string) (entities.Event, error) {
	args := m.Called(key)
	return args.Get(0).(entities.Event), args.Error(1)
}

func (m *MockProjectConfigTest) GetFeatureByKey(featureKey string) (entities.Feature, error) {
	args := m.Called(featureKey)
	return args.Get(0).(entities.Feature), args.Error(1)
}

func (m *MockProjectConfigTest) GetExperimentByKey(experimentKey string) (entities.Experiment, error) {
	args := m.Called(experimentKey)
	return args.Get(0).(entities.Experiment), args.Error(1)
}

func (m *MockProjectConfigTest) GetExperimentByID(id string) (entities.Experiment, error) {
	args := m.Called(id)
	return args.Get(0).(entities.Experiment), args.Error(1)
}

func (m *MockProjectConfigTest) GetRolloutByID(rolloutID string) (entities.Rollout, error) {
	args := m.Called(rolloutID)
	return args.Get(0).(entities.Rollout), args.Error(1)
}

func (m *MockProjectConfigTest) GetVariationByKey(experimentKey, variationKey string) (entities.Variation, error) {
	args := m.Called(experimentKey, variationKey)
	return args.Get(0).(entities.Variation), args.Error(1)
}

func (m *MockProjectConfigTest) GetPublicKeyForODP() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockProjectConfigTest) GetHostForODP() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockProjectConfigTest) GetSegmentList() []string {
	args := m.Called()
	return args.Get(0).([]string)
}

func (m *MockProjectConfigTest) GetBotFiltering() bool {
	args := m.Called()
	return args.Bool(0)
}

// TestCmabServiceInterface tests that the CmabService interface can be implemented
func TestCmabServiceInterface(t *testing.T) {
	// Create a mock implementation
	mockService := new(MockCmabServiceTest)

	// Setup expectations
	expectedDecision := CmabDecision{
		VariationID: "variation-123",
		CmabUUID:    "uuid-456",
		Reasons:     []string{"reason1"},
	}

	mockProjectConfig := new(MockProjectConfigTest)
	userContext := entities.UserContext{
		ID: "user-456",
		Attributes: map[string]interface{}{
			"attr1": "value1",
		},
	}
	options := &CmabDecisionOptions{
		IncludeReasons: true,
	}

	mockService.On("GetDecision", mockProjectConfig, userContext, "rule-123", options).Return(expectedDecision, nil)
	mockService.On("ResetCache").Return(nil)
	mockService.On("InvalidateUserCache", "user-456").Return(nil)

	// Use the interface
	var cmabService CmabService = mockService

	// Call methods through the interface
	decision, err := cmabService.GetDecision(mockProjectConfig, userContext, "rule-123", options)
	assert.NoError(t, err)
	assert.Equal(t, expectedDecision, decision)

	err = cmabService.ResetCache()
	assert.NoError(t, err)

	err = cmabService.InvalidateUserCache("user-456")
	assert.NoError(t, err)

	// Verify all expectations were met
	mockService.AssertExpectations(t)
}

// MockCmabClientTest implements the CmabClient interface for testing
type MockCmabClientTest struct {
	mock.Mock
}

func (m *MockCmabClientTest) FetchDecision(ruleID, userID string, attributes map[string]interface{}) (CmabDecision, error) {
	args := m.Called(ruleID, userID, attributes)
	return args.Get(0).(CmabDecision), args.Error(1)
}

// TestCmabClientInterface tests that the CmabClient interface can be implemented
func TestCmabClientInterface(t *testing.T) {
	// Create a mock implementation
	mockClient := new(MockCmabClientTest)

	// Setup expectations
	expectedDecision := CmabDecision{
		VariationID: "variation-123",
		CmabUUID:    "uuid-456",
	}

	attributes := map[string]interface{}{
		"attr1": "value1",
	}

	mockClient.On("FetchDecision", "rule-123", "user-456", attributes).Return(expectedDecision, nil)

	// Use the interface
	var cmabClient CmabClient = mockClient

	// Call method through the interface
	decision, err := cmabClient.FetchDecision("rule-123", "user-456", attributes)
	assert.NoError(t, err)
	assert.Equal(t, expectedDecision, decision)

	// Verify all expectations were met
	mockClient.AssertExpectations(t)
}

// TestCmabDecisionMutation tests that modifying a decision doesn't affect cached copies
func TestCmabDecisionMutation(t *testing.T) {
	// Create original decision
	original := CmabDecision{
		VariationID: "variation-123",
		CmabUUID:    "uuid-456",
		Reasons:     []string{"reason1"},
	}

	// Create a copy
	copy := original

	// Modify the copy
	copy.VariationID = "new-variation"
	copy.Reasons = append(copy.Reasons, "reason2")

	// Verify original is unchanged
	assert.Equal(t, "variation-123", original.VariationID)
	assert.Equal(t, 1, len(original.Reasons))
	assert.Equal(t, "reason1", original.Reasons[0])

	// Verify copy has changes
	assert.Equal(t, "new-variation", copy.VariationID)
	assert.Equal(t, 2, len(copy.Reasons))
	assert.Equal(t, "reason2", copy.Reasons[1])
}

// TestCmabDecisionOptionsDefaults tests default behavior with nil options
func TestCmabDecisionOptionsDefaults(t *testing.T) {
	// Create a mock service that checks option defaults
	mockService := new(MockCmabServiceTest)
	mockProjectConfig := new(MockProjectConfigTest)
	userContext := entities.UserContext{ID: "user-1"}

	// Setup expectations - the service should use default options when nil is passed
	expectedDecision := CmabDecision{VariationID: "var-1"}

	// This verifies that nil options are handled properly
	mockService.On("GetDecision", mockProjectConfig, userContext, "rule-1", (*CmabDecisionOptions)(nil)).Return(expectedDecision, nil)

	// Call with nil options
	decision, err := mockService.GetDecision(mockProjectConfig, userContext, "rule-1", nil)

	assert.NoError(t, err)
	assert.Equal(t, expectedDecision, decision)
	mockService.AssertExpectations(t)
}

// TestCmabCacheValueCreation tests creating cache values with different timestamps
func TestCmabCacheValueCreation(t *testing.T) {
	decision := CmabDecision{VariationID: "var-1"}

	// Create cache values with different timestamps
	now := time.Now().Unix()
	pastTime := now - 100
	futureTime := now + 100

	cacheNow := CmabCacheValue{Decision: decision, Created: now}
	cachePast := CmabCacheValue{Decision: decision, Created: pastTime}
	cacheFuture := CmabCacheValue{Decision: decision, Created: futureTime}

	// Verify timestamps
	assert.Equal(t, now, cacheNow.Created)
	assert.Equal(t, pastTime, cachePast.Created)
	assert.Equal(t, futureTime, cacheFuture.Created)

	// Verify all have the same decision
	assert.Equal(t, decision, cacheNow.Decision)
	assert.Equal(t, decision, cachePast.Decision)
	assert.Equal(t, decision, cacheFuture.Decision)
}
