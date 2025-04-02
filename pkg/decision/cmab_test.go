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
    // Test initialization
    decision := CmabDecision{
        RuleID:     "rule-123",
        UserID:     "user-456",
        VariantID:  "variant-789",
        Attributes: map[string]interface{}{"key1": "value1", "key2": 42},
        Reasons:    []string{"reason1", "reason2"},
    }

    // Verify fields
    assert.Equal(t, "rule-123", decision.RuleID)
    assert.Equal(t, "user-456", decision.UserID)
    assert.Equal(t, "variant-789", decision.VariantID)
    assert.Equal(t, "value1", decision.Attributes["key1"])
    assert.Equal(t, 42, decision.Attributes["key2"])
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
    assert.Equal(t, decision.RuleID, deserializedDecision.RuleID)
    assert.Equal(t, decision.UserID, deserializedDecision.UserID)
    assert.Equal(t, decision.VariantID, deserializedDecision.VariantID)
    assert.Equal(t, decision.Attributes["key1"], deserializedDecision.Attributes["key1"])
    assert.Equal(t, decision.Attributes["key2"], deserializedDecision.Attributes["key2"])
    assert.Equal(t, decision.Reasons, deserializedDecision.Reasons)
}

// TestCmabCacheValueStruct tests the CmabCacheValue struct
func TestCmabCacheValueStruct(t *testing.T) {
    // Create a decision
    decision := CmabDecision{
        RuleID:    "rule-123",
        UserID:    "user-456",
        VariantID: "variant-789",
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

// MockCmabService implements the CmabService interface for testing
type MockCmabService struct {
    mock.Mock
}

func (m *MockCmabService) GetDecision(projectConfig config.ProjectConfig, userContext entities.UserContext, ruleID string, options *CmabDecisionOptions) (CmabDecision, error) {
    args := m.Called(projectConfig, userContext, ruleID, options)
    return args.Get(0).(CmabDecision), args.Error(1)
}

func (m *MockCmabService) ResetCache() error {
    args := m.Called()
    return args.Error(0)
}

func (m *MockCmabService) InvalidateUserCache(userID string) error {
    args := m.Called(userID)
    return args.Error(0)
}

// TestCmabServiceInterface tests that the CmabService interface can be implemented
func TestCmabServiceInterface(t *testing.T) {
    // Create a mock implementation
    mockService := new(MockCmabService)

    // Setup expectations
    expectedDecision := CmabDecision{
        RuleID:    "rule-123",
        UserID:    "user-456",
        VariantID: "variant-789",
    }

    mockProjectConfig := &MockProjectConfig{}
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

// MockCmabClient implements the CmabClient interface for testing
type MockCmabClient struct {
    mock.Mock
}

func (m *MockCmabClient) FetchDecision(ruleID, userID string, attributes map[string]interface{}) (CmabDecision, error) {
    args := m.Called(ruleID, userID, attributes)
    return args.Get(0).(CmabDecision), args.Error(1)
}

// TestCmabClientInterface tests that the CmabClient interface can be implemented
func TestCmabClientInterface(t *testing.T) {
    // Create a mock implementation
    mockClient := new(MockCmabClient)

    // Setup expectations
    expectedDecision := CmabDecision{
        RuleID:    "rule-123",
        UserID:    "user-456",
        VariantID: "variant-789",
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
