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
	// Create a cache value
	cacheValue := CmabCacheValue{
		AttributesHash: "12345",
		VariationID:    "variation-123",
		CmabUUID:       "uuid-456",
	}

	// Verify fields
	assert.Equal(t, "12345", cacheValue.AttributesHash)
	assert.Equal(t, "variation-123", cacheValue.VariationID)
	assert.Equal(t, "uuid-456", cacheValue.CmabUUID)
}

// TestOptimizelyDecideOptions tests the OptimizelyDecideOptions constants
func TestOptimizelyDecideOptions(t *testing.T) {
	// Test the constants
	assert.Equal(t, OptimizelyDecideOptions("IGNORE_CMAB_CACHE"), IgnoreCMABCache)
	assert.Equal(t, OptimizelyDecideOptions("RESET_CMAB_CACHE"), ResetCMABCache)
	assert.Equal(t, OptimizelyDecideOptions("INVALIDATE_USER_CMAB_CACHE"), InvalidateUserCMABCache)
}

// MockCmabServiceTest implements the CmabService interface for testing
type MockCmabServiceTest struct {
	mock.Mock
}

func (m *MockCmabServiceTest) GetDecision(
	projectConfig config.ProjectConfig,
	userContext entities.UserContext,
	ruleID string,
	options map[OptimizelyDecideOptions]bool,
) (CmabDecision, error) {
	args := m.Called(projectConfig, userContext, ruleID, options)
	return args.Get(0).(CmabDecision), args.Error(1)
}

// MockProjectConfigTest is a mock implementation of config.ProjectConfig
type MockProjectConfigTest struct {
	mock.Mock
}

// GetEnvironmentKey implements config.ProjectConfig.
func (m *MockProjectConfigTest) GetEnvironmentKey() string {
	args := m.Called()
	return args.String(0)
}

// GetEvents implements config.ProjectConfig.
func (m *MockProjectConfigTest) GetEvents() []entities.Event {
	args := m.Called()
	return args.Get(0).([]entities.Event)
}

// GetExperimentList implements config.ProjectConfig.
func (m *MockProjectConfigTest) GetExperimentList() []entities.Experiment {
	args := m.Called()
	return args.Get(0).([]entities.Experiment)
}

// GetFeatureList implements config.ProjectConfig.
func (m *MockProjectConfigTest) GetFeatureList() []entities.Feature {
	args := m.Called()
	return args.Get(0).([]entities.Feature)
}

// GetFlagVariationsMap implements config.ProjectConfig.
func (m *MockProjectConfigTest) GetFlagVariationsMap() map[string][]entities.Variation {
	args := m.Called()
	return args.Get(0).(map[string][]entities.Variation)
}

// GetGroupByID implements config.ProjectConfig.
func (m *MockProjectConfigTest) GetGroupByID(id string) (entities.Group, error) {
	args := m.Called(id)
	return args.Get(0).(entities.Group), args.Error(1)
}

// GetIntegrationList implements config.ProjectConfig.
func (m *MockProjectConfigTest) GetIntegrationList() []entities.Integration {
	args := m.Called()
	return args.Get(0).([]entities.Integration)
}

// GetRolloutList implements config.ProjectConfig.
func (m *MockProjectConfigTest) GetRolloutList() (rolloutList []entities.Rollout) {
	args := m.Called()
	return args.Get(0).([]entities.Rollout)
}

// GetSdkKey implements config.ProjectConfig.
func (m *MockProjectConfigTest) GetSdkKey() string {
	args := m.Called()
	return args.String(0)
}

// GetVariableByKey implements config.ProjectConfig.
func (m *MockProjectConfigTest) GetVariableByKey(featureKey string, variableKey string) (entities.Variable, error) {
	args := m.Called(featureKey, variableKey)
	return args.Get(0).(entities.Variable), args.Error(1)
}

// SendFlagDecisions implements config.ProjectConfig.
func (m *MockProjectConfigTest) SendFlagDecisions() bool {
	args := m.Called()
	return args.Bool(0)
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

func (m *MockProjectConfigTest) GetAttributeKeyByID(id string) (string, error) {
	args := m.Called(id)
	return args.String(0), args.Error(1)
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
	options := map[OptimizelyDecideOptions]bool{
		IgnoreCMABCache: true,
	}

	mockService.On("GetDecision", mockProjectConfig, userContext, "rule-123", options).Return(expectedDecision, nil)

	// Use the interface
	var cmabService CmabService = mockService

	// Call methods through the interface
	decision, err := cmabService.GetDecision(mockProjectConfig, userContext, "rule-123", options)
	assert.NoError(t, err)
	assert.Equal(t, expectedDecision, decision)

	// Verify all expectations were met
	mockService.AssertExpectations(t)
}

// MockCmabClientTest implements the CmabClient interface for testing
type MockCmabClientTest struct {
	mock.Mock
}

func (m *MockCmabClientTest) FetchDecision(
	ruleID string,
	userID string,
	attributes map[string]interface{},
	cmabUUID string,
) (string, error) {
	args := m.Called(ruleID, userID, attributes, cmabUUID)
	return args.String(0), args.Error(1)
}

// TestCmabClientInterface tests that the CmabClient interface can be implemented
func TestCmabClientInterface(t *testing.T) {
	// Create a mock implementation
	mockClient := new(MockCmabClientTest)

	// Setup expectations
	expectedVariationID := "variation-123"
	cmabUUID := "uuid-456"

	attributes := map[string]interface{}{
		"attr1": "value1",
	}

	mockClient.On("FetchDecision", "rule-123", "user-456", attributes, cmabUUID).Return(expectedVariationID, nil)

	// Use the interface
	var cmabClient CmabClient = mockClient

	// Call method through the interface
	variationID, err := cmabClient.FetchDecision("rule-123", "user-456", attributes, cmabUUID)
	assert.NoError(t, err)
	assert.Equal(t, expectedVariationID, variationID)

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

// TestCmabCacheValueMutation tests that modifying a cache value doesn't affect copies
func TestCmabCacheValueMutation(t *testing.T) {
	// Create original cache value
	original := CmabCacheValue{
		AttributesHash: "12345",
		VariationID:    "variation-123",
		CmabUUID:       "uuid-456",
	}

	// Create a copy
	copy := original

	// Modify the copy
	copy.AttributesHash = "67890"
	copy.VariationID = "new-variation"

	// Verify original is unchanged
	assert.Equal(t, "12345", original.AttributesHash)
	assert.Equal(t, "variation-123", original.VariationID)

	// Verify copy has changes
	assert.Equal(t, "67890", copy.AttributesHash)
	assert.Equal(t, "new-variation", copy.VariationID)
}
