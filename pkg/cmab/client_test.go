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
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/optimizely/go-sdk/v2/pkg/entities"
	"github.com/optimizely/go-sdk/v2/pkg/event"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Create a simple test implementation that satisfies the interface
type testProjectConfig struct {
	// Map to store attributes by key for GetAttributeByKey
	attributesByKey map[string]entities.Attribute

	// Functions that can be overridden for testing
	getExperimentByIDFn func(id string) (entities.Experiment, error)
	getFeatureListFn    func() []entities.Feature
}

// NewTestProjectConfig creates a new test project config
func NewTestProjectConfig() *testProjectConfig {
	return &testProjectConfig{
		attributesByKey: make(map[string]entities.Attribute),
		getExperimentByIDFn: func(id string) (entities.Experiment, error) {
			if id == "rule-id" {
				return entities.Experiment{ID: id, Key: "rule-key"}, nil
			}
			return entities.Experiment{}, fmt.Errorf("experiment not found")
		},
		getFeatureListFn: func() []entities.Feature {
			return []entities.Feature{
				{
					Key: "feature-key",
					FeatureExperiments: []entities.Experiment{
						{ID: "rule-id"},
					},
				},
			}
		},
	}
}

// GetExperimentByID returns a mock experiment
func (c *testProjectConfig) GetExperimentByID(id string) (entities.Experiment, error) {
	if c.getExperimentByIDFn != nil {
		return c.getExperimentByIDFn(id)
	}
	return entities.Experiment{}, fmt.Errorf("experiment not found")
}

// GetFeatureList returns mock features
func (c *testProjectConfig) GetFeatureList() []entities.Feature {
	if c.getFeatureListFn != nil {
		return c.getFeatureListFn()
	}
	return []entities.Feature{}
}

// GetAttributeByKey returns a mock attribute
func (c *testProjectConfig) GetAttributeByKey(key string) (entities.Attribute, error) {
	if attr, ok := c.attributesByKey[key]; ok {
		return attr, nil
	}
	return entities.Attribute{Key: key}, nil
}

// Implement all other required methods with minimal functionality
func (c *testProjectConfig) GetProjectID() string             { return "mock-project" }
func (c *testProjectConfig) GetRevision() string              { return "1" }
func (c *testProjectConfig) GetAccountID() string             { return "mock-account" }
func (c *testProjectConfig) GetAnonymizeIP() bool             { return false }
func (c *testProjectConfig) GetAttributeID(key string) string { return "attr-id" }
func (c *testProjectConfig) GetAttributeKeyByID(id string) (string, error) {
	return "attr-key", nil
}
func (c *testProjectConfig) GetAttributeByID(id string) (entities.Attribute, error) {
	return entities.Attribute{ID: id, Key: "attr-key"}, nil
}
func (c *testProjectConfig) GetAudienceByID(id string) (entities.Audience, error) {
	return entities.Audience{ID: id}, nil
}
func (c *testProjectConfig) GetAudienceMap() map[string]entities.Audience {
	return map[string]entities.Audience{"audience-id": {ID: "audience-id"}}
}
func (c *testProjectConfig) GetBotFiltering() bool    { return false }
func (c *testProjectConfig) GetClientName() string    { return "mock-client" }
func (c *testProjectConfig) GetClientVersion() string { return "1.0.0" }
func (c *testProjectConfig) GetDatafile() string      { return "{}" }
func (c *testProjectConfig) GetEventByKey(key string) (entities.Event, error) {
	return entities.Event{Key: key}, nil
}
func (c *testProjectConfig) GetFeatureByID(featureID string) (entities.Feature, error) {
	return entities.Feature{ID: featureID}, nil
}
func (c *testProjectConfig) GetFeatureByKey(featureKey string) (entities.Feature, error) {
	return entities.Feature{Key: featureKey}, nil
}
func (c *testProjectConfig) GetExperimentByKey(key string) (entities.Experiment, error) {
	return entities.Experiment{Key: key}, nil
}
func (c *testProjectConfig) GetExperimentFeatureMap() map[string][]entities.Feature {
	return map[string][]entities.Feature{}
}
func (c *testProjectConfig) GetFlagVariationsMap() map[string][]entities.Variation {
	return map[string][]entities.Variation{}
}
func (c *testProjectConfig) GetGroupByID(id string) (entities.Group, error) {
	return entities.Group{ID: id}, nil
}
func (c *testProjectConfig) GetRolloutByID(id string) (entities.Rollout, error) {
	return entities.Rollout{ID: id}, nil
}
func (c *testProjectConfig) GetRolloutIDFromExperimentID(experimentID string) string {
	return "rollout-id"
}
func (c *testProjectConfig) GetVariationFromKey(experimentKey, variationKey string) (entities.Variation, error) {
	return entities.Variation{Key: variationKey}, nil
}
func (c *testProjectConfig) GetVariationFromID(experimentID, variationID string) (entities.Variation, error) {
	return entities.Variation{ID: variationID}, nil
}
func (c *testProjectConfig) GetSdkKey() string         { return "mock-sdk-key" }
func (c *testProjectConfig) GetEnvironmentKey() string { return "mock-env-key" }
func (c *testProjectConfig) SendFlagDecisions() bool   { return false }

// GetAttributes returns all attributes
func (c *testProjectConfig) GetAttributes() []entities.Attribute {
	return []entities.Attribute{
		{ID: "attr-id", Key: "attr-key"},
	}
}

// GetAudienceList returns all audiences
func (c *testProjectConfig) GetAudienceList() []entities.Audience {
	return []entities.Audience{
		{ID: "audience-id", Name: "Test Audience"},
	}
}

// GetEvents returns all events
func (c *testProjectConfig) GetEvents() []entities.Event {
	return []entities.Event{
		{ID: "event-id", Key: "test_event"},
	}
}

// GetHostForODP returns the host for ODP
func (c *testProjectConfig) GetHostForODP() string {
	return "https://odp.example.com"
}

// GetPublicKeyForODP returns the public key for ODP
func (c *testProjectConfig) GetPublicKeyForODP() string {
	return "mock-public-key"
}

// GetExperimentList returns all experiments
func (c *testProjectConfig) GetExperimentList() []entities.Experiment {
	return []entities.Experiment{
		{ID: "rule-id", Key: "rule-key"},
	}
}

// GetSegmentList returns all segments
func (c *testProjectConfig) GetSegmentList() []string {
	return []string{"segment1", "segment2"}
}

// GetIntegrationList returns all integrations
func (c *testProjectConfig) GetIntegrationList() []entities.Integration {
	return []entities.Integration{}
}

// GetRolloutList returns all rollouts
func (c *testProjectConfig) GetRolloutList() []entities.Rollout {
	return []entities.Rollout{
		{ID: "rollout-id"},
	}
}

// GetVariableByKey returns a variable by key
func (c *testProjectConfig) GetVariableByKey(featureKey string, variableKey string) (entities.Variable, error) {
	return entities.Variable{
		ID:   "var-id",
		Key:  variableKey,
		Type: "string",
	}, nil
}

// Mock logger for testing
type mockLogger struct {
	debugFn   func(message string)
	warningFn func(message string)
	errorFn   func(message string, err interface{})
}

func (m *mockLogger) Debug(message string) {
	if m.debugFn != nil {
		m.debugFn(message)
	}
}

func (m *mockLogger) Info(message string) {}

func (m *mockLogger) Warning(message string) {
	if m.warningFn != nil {
		m.warningFn(message)
	}
}

// Update the Error method to match the expected interface
func (m *mockLogger) Error(message string, err interface{}) {
	if m.errorFn != nil {
		m.errorFn(message, err)
	}
}

func TestDefaultCmabClient_FetchDecision(t *testing.T) {
	// Setup test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request method
		assert.Equal(t, http.MethodPost, r.Method)

		// Verify content type
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		// Parse request body
		var requestBody Request
		err := json.NewDecoder(r.Body).Decode(&requestBody)
		assert.NoError(t, err)

		// Verify request structure
		assert.Len(t, requestBody.Instances, 1)
		instance := requestBody.Instances[0]
		assert.Equal(t, "user123", instance.VisitorID)
		assert.Equal(t, "rule456", instance.ExperimentID)
		assert.Equal(t, "test-uuid", instance.CmabUUID)

		// Verify attributes - check for various types
		assert.Len(t, instance.Attributes, 5)

		// Create a map for easier attribute checking
		attrMap := make(map[string]Attribute)
		for _, attr := range instance.Attributes {
			attrMap[attr.ID] = attr
			assert.Equal(t, "custom_attribute", attr.Type)
		}

		// Check string attribute
		assert.Contains(t, attrMap, "string_attr")
		assert.Equal(t, "string value", attrMap["string_attr"].Value)

		// Check int attribute
		assert.Contains(t, attrMap, "int_attr")
		assert.Equal(t, float64(42), attrMap["int_attr"].Value) // JSON numbers are float64

		// Check float attribute
		assert.Contains(t, attrMap, "float_attr")
		assert.Equal(t, 3.14, attrMap["float_attr"].Value)

		// Check bool attribute
		assert.Contains(t, attrMap, "bool_attr")
		assert.Equal(t, true, attrMap["bool_attr"].Value)

		// Check null attribute
		assert.Contains(t, attrMap, "null_attr")
		assert.Nil(t, attrMap["null_attr"].Value)

		// Return response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		response := Response{
			Predictions: []Prediction{
				{
					VariationID: "var123",
				},
			},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Create client with custom endpoint
	client := NewDefaultCmabClient(ClientOptions{
		HTTPClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	})

	// Override the endpoint for testing
	originalEndpoint := CMABPredictionEndpoint
	CMABPredictionEndpoint = server.URL + "/%s"
	defer func() { CMABPredictionEndpoint = originalEndpoint }()

	// Test with various attribute types
	attributes := map[string]interface{}{
		"string_attr": "string value",
		"int_attr":    42,
		"float_attr":  3.14,
		"bool_attr":   true,
		"null_attr":   nil,
	}

	// Create a context for the request
	ctx := context.Background()

	variationID, err := client.FetchDecision(ctx, "rule456", "user123", attributes, "test-uuid")

	// Verify results
	assert.NoError(t, err)
	assert.Equal(t, "var123", variationID)
}

func TestDefaultCmabClient_FetchDecision_WithRetry(t *testing.T) {
	// Setup counter for tracking request attempts
	requestCount := 0

	// Setup test server that fails initially then succeeds
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++

		// Verify request method and content type
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		// Parse request body to verify it's consistent across retries
		body, err := io.ReadAll(r.Body)
		assert.NoError(t, err)

		var requestBody Request
		err = json.Unmarshal(body, &requestBody)
		assert.NoError(t, err)

		// Verify request structure is consistent
		assert.Len(t, requestBody.Instances, 1)
		instance := requestBody.Instances[0]
		assert.Equal(t, "user123", instance.VisitorID)
		assert.Equal(t, "rule456", instance.ExperimentID)
		assert.Equal(t, "test-uuid", instance.CmabUUID)

		// First two requests fail, third succeeds
		if requestCount <= 2 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// Return success response on third attempt
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		response := Response{
			Predictions: []Prediction{
				{
					VariationID: "var123",
				},
			},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Create client with custom endpoint and retry config
	client := NewDefaultCmabClient(ClientOptions{
		HTTPClient: &http.Client{
			Timeout: 5 * time.Second,
		},
		RetryConfig: &RetryConfig{
			MaxRetries:        5,
			InitialBackoff:    10 * time.Millisecond, // Short backoff for testing
			MaxBackoff:        100 * time.Millisecond,
			BackoffMultiplier: 2.0,
		},
	})

	// Override the endpoint for testing
	originalEndpoint := CMABPredictionEndpoint
	CMABPredictionEndpoint = server.URL + "/%s"
	defer func() { CMABPredictionEndpoint = originalEndpoint }()

	// Test fetch decision with retry
	attributes := map[string]interface{}{
		"browser":  "chrome",
		"isMobile": true,
	}

	startTime := time.Now()
	variationID, err := client.FetchDecision(context.Background(), "rule456", "user123", attributes, "test-uuid")
	duration := time.Since(startTime)

	// Verify results
	assert.NoError(t, err)
	assert.Equal(t, "var123", variationID)
	assert.Equal(t, 3, requestCount, "Expected 3 request attempts")

	// Verify that backoff was applied (at least some delay between requests)
	assert.True(t, duration >= 30*time.Millisecond, "Expected some backoff delay between requests")
}

func TestDefaultCmabClient_FetchDecision_ExhaustedRetries(t *testing.T) {
	// Setup counter for tracking request attempts
	requestCount := 0

	// Setup test server that always fails
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	// Create client with custom endpoint and retry config
	client := NewDefaultCmabClient(ClientOptions{
		HTTPClient: &http.Client{
			Timeout: 5 * time.Second,
		},
		RetryConfig: &RetryConfig{
			MaxRetries:        2, // Allow 2 retries (3 total attempts)
			InitialBackoff:    10 * time.Millisecond,
			MaxBackoff:        100 * time.Millisecond,
			BackoffMultiplier: 2.0,
		},
	})

	// Override the endpoint for testing
	originalEndpoint := CMABPredictionEndpoint
	CMABPredictionEndpoint = server.URL + "/%s"
	defer func() { CMABPredictionEndpoint = originalEndpoint }()

	// Test fetch decision with exhausted retries
	attributes := map[string]interface{}{
		"browser":  "chrome",
		"isMobile": true,
	}

	variationID, err := client.FetchDecision(context.Background(), "rule456", "user123", attributes, "test-uuid")

	// Verify results
	assert.Error(t, err)
	assert.Equal(t, "", variationID)
	assert.Equal(t, 3, requestCount, "Expected 3 request attempts (initial + 2 retries)")
	assert.Contains(t, err.Error(), "failed to fetch CMAB decision after 2 attempts")
	assert.Contains(t, err.Error(), "non-success status code: 500")
}

func TestDefaultCmabClient_FetchDecision_NoRetryConfig(t *testing.T) {
	// Setup counter for tracking request attempts
	requestCount := 0

	// Setup test server that fails
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	// Create client with custom endpoint but no retry config
	client := NewDefaultCmabClient(ClientOptions{
		HTTPClient: &http.Client{
			Timeout: 5 * time.Second,
		},
		RetryConfig: nil, // Explicitly set to nil to override default
	})

	// Override the endpoint for testing
	originalEndpoint := CMABPredictionEndpoint
	CMABPredictionEndpoint = server.URL + "/%s"
	defer func() { CMABPredictionEndpoint = originalEndpoint }()

	// Test fetch decision without retry config
	attributes := map[string]interface{}{
		"browser": "chrome",
	}

	_, err := client.FetchDecision(context.Background(), "rule456", "user123", attributes, "test-uuid")

	// Verify results
	assert.Error(t, err)
	assert.Equal(t, 1, requestCount, "Expected only 1 request attempt without retry config")
}

func TestDefaultCmabClient_FetchDecision_InvalidResponse(t *testing.T) {
	// Test cases for invalid responses
	testCases := []struct {
		name           string
		responseBody   string
		expectedErrMsg string
	}{
		{
			name:           "Empty predictions array",
			responseBody:   `{"predictions": []}`,
			expectedErrMsg: "invalid CMAB response",
		},
		{
			name:           "Missing variation_id",
			responseBody:   `{"predictions": [{"some_field": "value"}]}`,
			expectedErrMsg: "invalid CMAB response",
		},
		{
			name:           "Empty variation_id",
			responseBody:   `{"predictions": [{"variation_id": ""}]}`,
			expectedErrMsg: "invalid CMAB response",
		},
		{
			name:           "Invalid JSON",
			responseBody:   `{invalid json`,
			expectedErrMsg: "failed to unmarshal CMAB response",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup test server that returns the test case response
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(tc.responseBody))
			}))
			defer server.Close()

			// Create client with custom endpoint
			client := NewDefaultCmabClient(ClientOptions{
				HTTPClient: &http.Client{
					Timeout: 5 * time.Second,
				},
			})

			// Override the endpoint for testing
			originalEndpoint := CMABPredictionEndpoint
			CMABPredictionEndpoint = server.URL + "/%s"
			defer func() { CMABPredictionEndpoint = originalEndpoint }()

			// Test fetch decision with invalid response
			attributes := map[string]interface{}{
				"browser": "chrome",
			}

			_, err := client.FetchDecision(context.Background(), "rule456", "user123", attributes, "test-uuid")

			// Verify results
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tc.expectedErrMsg)
		})
	}
}

func TestDefaultCmabClient_FetchDecision_NetworkErrors(t *testing.T) {
	// Create a custom logger that captures log messages to verify retries
	retryAttempted := false
	mockLogger := &mockLogger{
		warningFn: func(message string) {
			if strings.Contains(message, "CMAB API request failed (attempt") {
				retryAttempted = true
			}
		},
	}

	// Create client with non-existent server to simulate network errors
	client := NewDefaultCmabClient(ClientOptions{
		HTTPClient: &http.Client{
			Timeout: 100 * time.Millisecond, // Short timeout to fail quickly
		},
		RetryConfig: &RetryConfig{
			MaxRetries:        1,
			InitialBackoff:    10 * time.Millisecond,
			MaxBackoff:        100 * time.Millisecond,
			BackoffMultiplier: 2.0,
		},
		Logger: mockLogger,
	})

	// Set endpoint to a non-existent server
	originalEndpoint := CMABPredictionEndpoint
	CMABPredictionEndpoint = "http://non-existent-server.example.com/%s"
	defer func() { CMABPredictionEndpoint = originalEndpoint }()

	// Test fetch decision with network error
	attributes := map[string]interface{}{
		"browser": "chrome",
	}

	_, err := client.FetchDecision(context.Background(), "rule456", "user123", attributes, "test-uuid")

	// Verify results
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to fetch CMAB decision after 1 attempts")

	// Verify that retry was attempted by checking if the warning log was produced
	assert.True(t, retryAttempted, "Expected retry to be attempted")
}

func TestDefaultCmabClient_ExponentialBackoff(t *testing.T) {
	// Setup test server that tracks request times
	requestTimes := []time.Time{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestTimes = append(requestTimes, time.Now())

		// First 3 requests fail, 4th succeeds
		if len(requestTimes) < 4 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// Return success response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		response := Response{
			Predictions: []Prediction{
				{
					VariationID: "var123",
				},
			},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Create client with custom endpoint and specific retry config
	client := NewDefaultCmabClient(ClientOptions{
		HTTPClient: &http.Client{
			Timeout: 5 * time.Second,
		},
		RetryConfig: &RetryConfig{
			MaxRetries:        5,
			InitialBackoff:    50 * time.Millisecond,
			MaxBackoff:        1 * time.Second,
			BackoffMultiplier: 2.0,
		},
	})

	// Override the endpoint for testing
	originalEndpoint := CMABPredictionEndpoint
	CMABPredictionEndpoint = server.URL + "/%s"
	defer func() { CMABPredictionEndpoint = originalEndpoint }()

	// Test fetch decision with exponential backoff
	attributes := map[string]interface{}{
		"browser": "chrome",
	}

	variationID, err := client.FetchDecision(context.Background(), "rule456", "user123", attributes, "test-uuid")

	// Verify results
	require.NoError(t, err)
	assert.Equal(t, "var123", variationID)
	assert.Equal(t, 4, len(requestTimes), "Expected 4 request attempts")

	// Verify exponential backoff intervals
	// First request happens immediately, then we should see increasing intervals
	if len(requestTimes) >= 4 {
		interval1 := requestTimes[1].Sub(requestTimes[0])
		interval2 := requestTimes[2].Sub(requestTimes[1])
		interval3 := requestTimes[3].Sub(requestTimes[2])

		// Each interval should be approximately double the previous one
		// Allow some margin for test execution timing variations
		assert.True(t, interval1 >= 50*time.Millisecond, "First backoff should be at least initialBackoff")
		assert.True(t, interval2 >= 100*time.Millisecond, "Second backoff should be at least 2x initialBackoff")
		assert.True(t, interval3 >= 200*time.Millisecond, "Third backoff should be at least 4x initialBackoff")

		// Verify increasing pattern
		assert.True(t, interval2 > interval1, "Backoff intervals should increase")
		assert.True(t, interval3 > interval2, "Backoff intervals should increase")
	}
}

func TestNewDefaultCmabClient_DefaultValues(t *testing.T) {
	// Test with empty options
	client := NewDefaultCmabClient(ClientOptions{})

	// Verify default values
	assert.NotNil(t, client.httpClient)
	assert.Nil(t, client.retryConfig) // retryConfig should be nil by default
	assert.NotNil(t, client.logger)
}

func TestDefaultCmabClient_LoggingBehavior(t *testing.T) {
	// Create a custom logger that captures log messages
	logMessages := []string{}
	mockLogger := &mockLogger{
		debugFn: func(message string) {
			logMessages = append(logMessages, "DEBUG: "+message)
		},
		warningFn: func(message string) {
			logMessages = append(logMessages, "WARNING: "+message)
		},
	}

	// Setup test server that fails then succeeds
	requestCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++

		if requestCount == 1 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"predictions":[{"variation_id":"var123"}]}`)
	}))
	defer server.Close()

	// Create client with custom logger
	client := NewDefaultCmabClient(ClientOptions{
		HTTPClient: &http.Client{
			Timeout: 5 * time.Second,
		},
		RetryConfig: &RetryConfig{
			MaxRetries:        1,
			InitialBackoff:    10 * time.Millisecond,
			MaxBackoff:        100 * time.Millisecond,
			BackoffMultiplier: 2.0,
		},
		Logger: mockLogger,
	})

	// Override the endpoint for testing
	originalEndpoint := CMABPredictionEndpoint
	CMABPredictionEndpoint = server.URL + "/%s"
	defer func() { CMABPredictionEndpoint = originalEndpoint }()

	// Test fetch decision
	attributes := map[string]interface{}{
		"browser": "chrome",
	}

	_, err := client.FetchDecision(context.Background(), "rule456", "user123", attributes, "test-uuid")
	assert.NoError(t, err)

	// Verify log messages
	assert.True(t, len(logMessages) >= 2, "Expected at least 2 log messages")

	// Check for retry warning
	foundRetryWarning := false
	foundBackoffDebug := false
	for _, msg := range logMessages {
		if strings.Contains(msg, "WARNING") && strings.Contains(msg, "CMAB API request failed") {
			foundRetryWarning = true
		}
		if strings.Contains(msg, "DEBUG") && strings.Contains(msg, "CMAB request retry") {
			foundBackoffDebug = true
		}
	}

	assert.True(t, foundRetryWarning, "Expected warning log about API request failure")
	assert.True(t, foundBackoffDebug, "Expected debug log about retry backoff")
}

func TestDefaultCmabClient_NonSuccessStatusCode(t *testing.T) {
	// Setup test server that returns different non-2xx status codes
	testCases := []struct {
		name       string
		statusCode int
		statusText string
	}{
		{"BadRequest", http.StatusBadRequest, "Bad Request"},
		{"Unauthorized", http.StatusUnauthorized, "Unauthorized"},
		{"Forbidden", http.StatusForbidden, "Forbidden"},
		{"NotFound", http.StatusNotFound, "Not Found"},
		{"InternalServerError", http.StatusInternalServerError, "Internal Server Error"},
		{"ServiceUnavailable", http.StatusServiceUnavailable, "Service Unavailable"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tc.statusCode)
				w.Write([]byte(tc.statusText))
			}))
			defer server.Close()

			// Create client with custom endpoint and no retries
			client := NewDefaultCmabClient(ClientOptions{
				HTTPClient: &http.Client{
					Timeout: 5 * time.Second,
				},
				// No retry config to simplify the test
			})

			// Override the endpoint for testing
			originalEndpoint := CMABPredictionEndpoint
			CMABPredictionEndpoint = server.URL + "/%s"
			defer func() { CMABPredictionEndpoint = originalEndpoint }()

			// Test fetch decision
			attributes := map[string]interface{}{
				"browser": "chrome",
			}

			// Create a context for the request
			ctx := context.Background()

			variationID, err := client.FetchDecision(ctx, "rule456", "user123", attributes, "test-uuid")

			// Verify results
			assert.Error(t, err, "Expected error for non-success status code")
			assert.Equal(t, "", variationID, "Expected empty variation ID for error response")
			assert.Contains(t, err.Error(), "non-success status code")
			assert.Contains(t, err.Error(), fmt.Sprintf("%d", tc.statusCode))
		})
	}
}

func TestDefaultCmabClient_FetchDecision_ContextCancellation(t *testing.T) {
	// Setup test server that delays response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Sleep to simulate a slow response
		time.Sleep(500 * time.Millisecond)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		response := Response{
			Predictions: []Prediction{
				{
					VariationID: "var123",
				},
			},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Create client with custom endpoint
	client := NewDefaultCmabClient(ClientOptions{
		HTTPClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	})

	// Override the endpoint for testing
	originalEndpoint := CMABPredictionEndpoint
	CMABPredictionEndpoint = server.URL + "/%s"
	defer func() { CMABPredictionEndpoint = originalEndpoint }()

	// Create a context with a short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// Test fetch decision with a context that will time out
	attributes := map[string]interface{}{
		"browser": "chrome",
	}

	_, err := client.FetchDecision(ctx, "rule456", "user123", attributes, "test-uuid")

	// Verify that we got a context deadline exceeded error
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "context")
	assert.Contains(t, err.Error(), "deadline exceeded")
}

func TestLogImpression(t *testing.T) {
	// Create a client with minimal options
	client := NewDefaultCmabClient(ClientOptions{})

	// Test the unimplemented method
	err := client.LogImpression(
		context.Background(),
		nil,
		nil,
		"rule-id",
		"user-id",
		"variation-key",
		"variation-id",
		nil,
		"cmab-uuid",
	)

	// Verify we get the expected error
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not implemented")
}

// Mock event processor for testing
type mockEventProcessor struct {
	processEventFn func(userEvent event.UserEvent) bool
}

func (m *mockEventProcessor) Process(userEvent interface{}) bool {
	if m.processEventFn != nil {
		if userEventTyped, ok := userEvent.(event.UserEvent); ok {
			return m.processEventFn(userEventTyped)
		}
	}
	return true
}

func (m *mockEventProcessor) ProcessEvent(userEvent event.UserEvent) bool {
	if m.processEventFn != nil {
		return m.processEventFn(userEvent)
	}
	return true
}

func (m *mockEventProcessor) OnEventDispatch(callback func(event.LogEvent)) (int, error) {
	return 1, nil // Return a dummy notification ID and no error
}

func (m *mockEventProcessor) RemoveOnEventDispatch(id int) error {
	return nil // Return no error
}

// Mock project config for testing
type mockProjectConfig struct {
	getExperimentByIDFn func(id string) (entities.Experiment, error)
	getFeatureListFn    func() []entities.Feature
}

func (m *mockProjectConfig) GetExperimentByID(id string) (entities.Experiment, error) {
	if m.getExperimentByIDFn != nil {
		return m.getExperimentByIDFn(id)
	}
	return entities.Experiment{}, fmt.Errorf("experiment not found")
}

func (m *mockProjectConfig) GetAttributeByKey(key string) (entities.Attribute, error) {
	return entities.Attribute{}, nil
}

func (m *mockProjectConfig) GetAttributeKeyByID(id string) (string, error) {
	return "attr-key", nil
}

func (m *mockProjectConfig) GetFeatureList() []entities.Feature {
	if m.getFeatureListFn != nil {
		return m.getFeatureListFn()
	}
	return []entities.Feature{}
}

// Implement other required methods of the ProjectConfig interface with empty implementations
func (m *mockProjectConfig) GetProjectID() string             { return "mock-project" }
func (m *mockProjectConfig) GetRevision() string              { return "1" }
func (m *mockProjectConfig) GetAccountID() string             { return "mock-account" }
func (m *mockProjectConfig) GetAnonymizeIP() bool             { return false }
func (m *mockProjectConfig) GetAttributeID(key string) string { return "" }
func (m *mockProjectConfig) GetAttributeByID(id string) (entities.Attribute, error) {
	return entities.Attribute{}, nil
}
func (m *mockProjectConfig) GetAudienceByID(id string) (entities.Audience, error) {
	return entities.Audience{}, nil
}
func (m *mockProjectConfig) GetAudienceMap() map[string]entities.Audience { return nil }
func (m *mockProjectConfig) GetBotFiltering() bool                        { return false }
func (m *mockProjectConfig) GetClientName() string                        { return "mock-client" }
func (m *mockProjectConfig) GetClientVersion() string                     { return "1.0.0" }
func (m *mockProjectConfig) GetDatafile() string                          { return "" }
func (m *mockProjectConfig) GetEventByKey(key string) (entities.Event, error) {
	return entities.Event{}, nil
}
func (m *mockProjectConfig) GetFeatureByID(featureID string) (entities.Feature, error) {
	return entities.Feature{}, nil
}
func (m *mockProjectConfig) GetFeatureByKey(featureKey string) (entities.Feature, error) {
	return entities.Feature{}, nil
}
func (m *mockProjectConfig) GetExperimentByKey(key string) (entities.Experiment, error) {
	return entities.Experiment{}, nil
}
func (m *mockProjectConfig) GetExperimentFeatureMap() map[string][]entities.Feature { return nil }
func (m *mockProjectConfig) GetFlagVariationsMap() map[string][]entities.Variation  { return nil }
func (m *mockProjectConfig) GetGroupByID(id string) (entities.Group, error) {
	return entities.Group{}, nil
}
func (m *mockProjectConfig) GetRolloutByID(id string) (entities.Rollout, error) {
	return entities.Rollout{}, nil
}
func (m *mockProjectConfig) GetRolloutIDFromExperimentID(experimentID string) string { return "" }
func (m *mockProjectConfig) GetVariationFromKey(experimentKey, variationKey string) (entities.Variation, error) {
	return entities.Variation{}, nil
}
func (m *mockProjectConfig) GetVariationFromID(experimentID, variationID string) (entities.Variation, error) {
	return entities.Variation{}, nil
}
func (m *mockProjectConfig) GetSdkKey() string         { return "mock-sdk-key" }
func (m *mockProjectConfig) GetEnvironmentKey() string { return "mock-env-key" }
func (m *mockProjectConfig) SendFlagDecisions() bool   { return false }

// GetAttributes returns all attributes
func (m *mockProjectConfig) GetAttributes() []entities.Attribute {
	return []entities.Attribute{
		{ID: "attr-id", Key: "attr-key"},
	}
}

// GetAudienceList returns all audiences
func (m *mockProjectConfig) GetAudienceList() []entities.Audience {
	return []entities.Audience{
		{ID: "audience-id", Name: "Test Audience"},
	}
}

// GetEvents returns all events
func (m *mockProjectConfig) GetEvents() []entities.Event {
	return []entities.Event{
		{ID: "event-id", Key: "test_event"},
	}
}

// GetHostForODP returns the host for ODP
func (m *mockProjectConfig) GetHostForODP() string {
	return "https://odp.example.com"
}

// GetPublicKeyForODP returns the public key for ODP
func (m *mockProjectConfig) GetPublicKeyForODP() string {
	return "mock-public-key"
}

// GetExperimentList returns all experiments
func (m *mockProjectConfig) GetExperimentList() []entities.Experiment {
	return []entities.Experiment{
		{ID: "rule-id", Key: "rule-key"},
	}
}

// GetSegmentList returns all segments
func (m *mockProjectConfig) GetSegmentList() []string {
	return []string{"segment1", "segment2"}
}

// GetIntegrationList returns all integrations
func (m *mockProjectConfig) GetIntegrationList() []entities.Integration {
	return []entities.Integration{}
}

// GetRolloutList returns all rollouts
func (m *mockProjectConfig) GetRolloutList() []entities.Rollout {
	return []entities.Rollout{
		{ID: "rollout-id"},
	}
}

// GetVariableByKey returns a variable by key
func (m *mockProjectConfig) GetVariableByKey(featureKey string, variableKey string) (entities.Variable, error) {
	return entities.Variable{
		ID:   "var-id",
		Key:  variableKey,
		Type: "string",
	}, nil
}

func TestTrackCMABDecision(t *testing.T) {
	// Test 1: With nil event processor and project config
	t.Run("with nil dependencies", func(t *testing.T) {
		debugMessages := []string{}
		mockLogger := &mockLogger{
			debugFn: func(message string) {
				debugMessages = append(debugMessages, message)
			},
		}

		client := NewDefaultCmabClient(ClientOptions{
			Logger: mockLogger,
		})

		client.TrackCMABDecision(
			"rule-id",
			"user-id",
			"variation-id",
			"variation-key",
			nil,
			"cmab-uuid",
		)

		// Verify log message was produced
		assert.Len(t, debugMessages, 1)
		assert.Contains(t, debugMessages[0], "Event processor or project config not available")
	})

	// Test 2: With error getting experiment
	t.Run("with error getting experiment", func(t *testing.T) {
		errorMessages := []string{}
		mockLogger := &mockLogger{
			errorFn: func(message string, err interface{}) {
				errorMessages = append(errorMessages, message)
			},
		}

		// Create a test config with a custom GetExperimentByID function
		testConfig := NewTestProjectConfig()
		testConfig.getExperimentByIDFn = func(id string) (entities.Experiment, error) {
			return entities.Experiment{}, fmt.Errorf("experiment not found")
		}

		client := NewDefaultCmabClient(ClientOptions{
			Logger:        mockLogger,
			ProjectConfig: testConfig,
		})

		client.TrackCMABDecision(
			"rule-id",
			"user-id",
			"variation-id",
			"variation-key",
			nil,
			"cmab-uuid",
		)

		// This test will pass even without checking logs because we're testing that
		// the function doesn't panic when experiment is not found
	})

	// Test 3: With valid event processor and project config
	t.Run("with valid dependencies", func(t *testing.T) {
		processedEvents := []event.UserEvent{}
		mockEventProcessor := &mockEventProcessor{
			processEventFn: func(userEvent event.UserEvent) bool {
				processedEvents = append(processedEvents, userEvent)
				return true
			},
		}

		// Setup mock experiment
		mockExperiment := entities.Experiment{
			ID:  "rule-id",
			Key: "rule-key",
		}

		// Setup mock feature
		mockFeature := entities.Feature{
			Key: "feature-key",
			FeatureExperiments: []entities.Experiment{
				{
					ID: "rule-id",
				},
			},
		}

		mockProjectConfig := &mockProjectConfig{
			getExperimentByIDFn: func(id string) (entities.Experiment, error) {
				if id == "rule-id" {
					return mockExperiment, nil
				}
				return entities.Experiment{}, fmt.Errorf("experiment not found")
			},
			getFeatureListFn: func() []entities.Feature {
				return []entities.Feature{mockFeature}
			},
		}

		client := NewDefaultCmabClient(ClientOptions{
			EventProcessor: mockEventProcessor,
			ProjectConfig:  mockProjectConfig,
		})

		client.TrackCMABDecision(
			"rule-id",
			"user-id",
			"variation-id",
			"variation-key",
			map[string]interface{}{"attr1": "value1"},
			"cmab-uuid",
		)

		// Verify the event processor was called
		assert.Len(t, processedEvents, 1, "Expected one event to be processed")
	})
}
