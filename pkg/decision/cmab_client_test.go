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
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Mock logger for testing
type mockLogger struct {
	debugFn   func(message string)
	warningFn func(message string)
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
func (m *mockLogger) Error(message string, err interface{}) {}

func TestDefaultCmabClient_FetchDecision(t *testing.T) {
	// Setup test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request method
		assert.Equal(t, http.MethodPost, r.Method)

		// Verify content type
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		// Parse request body
		var requestBody CMABRequest
		err := json.NewDecoder(r.Body).Decode(&requestBody)
		assert.NoError(t, err)

		// Verify request structure
		assert.Len(t, requestBody.Instances, 1)
		instance := requestBody.Instances[0]
		assert.Equal(t, "user123", instance.VisitorID)
		assert.Equal(t, "rule456", instance.ExperimentID)
		assert.Equal(t, "test-uuid", instance.CmabUUID)

		// Verify attributes
		assert.Len(t, instance.Attributes, 2)
		// Attributes order is not guaranteed, so we need to check both
		for _, attr := range instance.Attributes {
			if attr.ID == "browser" {
				assert.Equal(t, "chrome", attr.Value)
			} else if attr.ID == "isMobile" {
				assert.Equal(t, true, attr.Value)
			} else {
				t.Errorf("Unexpected attribute ID: %s", attr.ID)
			}
			assert.Equal(t, "custom_attribute", attr.Type)
		}

		// Return response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		response := CMABResponse{
			Predictions: []CMABPrediction{
				{
					VariationID: "var123",
				},
			},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Create client with custom endpoint
	client := NewDefaultCmabClient(CmabClientOptions{
		HTTPClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	})

	// Override the endpoint for testing
	originalEndpoint := CMABPredictionEndpoint
	CMABPredictionEndpoint = server.URL + "/%s"
	defer func() { CMABPredictionEndpoint = originalEndpoint }()

	// Test fetch decision
	attributes := map[string]interface{}{
		"browser":  "chrome",
		"isMobile": true,
	}
	variationID, err := client.FetchDecision("rule456", "user123", attributes, "test-uuid")

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

		var requestBody CMABRequest
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
		response := CMABResponse{
			Predictions: []CMABPrediction{
				{
					VariationID: "var123",
				},
			},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Create client with custom endpoint and retry config
	client := NewDefaultCmabClient(CmabClientOptions{
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
	variationID, err := client.FetchDecision("rule456", "user123", attributes, "test-uuid")
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
	client := NewDefaultCmabClient(CmabClientOptions{
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

	variationID, err := client.FetchDecision("rule456", "user123", attributes, "test-uuid")

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
	client := NewDefaultCmabClient(CmabClientOptions{
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

	_, err := client.FetchDecision("rule456", "user123", attributes, "test-uuid")

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
			client := NewDefaultCmabClient(CmabClientOptions{
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

			_, err := client.FetchDecision("rule456", "user123", attributes, "test-uuid")

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
	client := NewDefaultCmabClient(CmabClientOptions{
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

	_, err := client.FetchDecision("rule456", "user123", attributes, "test-uuid")

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
		response := CMABResponse{
			Predictions: []CMABPrediction{
				{
					VariationID: "var123",
				},
			},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Create client with custom endpoint and specific retry config
	client := NewDefaultCmabClient(CmabClientOptions{
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

	variationID, err := client.FetchDecision("rule456", "user123", attributes, "test-uuid")

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

func TestDefaultCmabClient_RequestValidation(t *testing.T) {
	// Setup test server that validates the request format
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Read and validate the request body
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)

		// Check that the body is valid JSON
		var requestBody map[string]interface{}
		err = json.Unmarshal(body, &requestBody)
		require.NoError(t, err)

		// Check that the required fields are present
		instances, ok := requestBody["instances"].([]interface{})
		require.True(t, ok, "Request should have 'instances' array")
		require.Len(t, instances, 1, "Request should have exactly one instance")

		instance := instances[0].(map[string]interface{})
		require.Contains(t, instance, "visitorId", "Instance should have visitorId")
		require.Contains(t, instance, "experimentId", "Instance should have experimentId")
		require.Contains(t, instance, "attributes", "Instance should have attributes")
		require.Contains(t, instance, "cmabUUID", "Instance should have cmabUUID")

		// Return success response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"predictions":[{"variation_id":"var123"}]}`)
	}))
	defer server.Close()

	// Create client with custom endpoint
	client := NewDefaultCmabClient(CmabClientOptions{
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

	variationID, err := client.FetchDecision("rule456", "user123", attributes, "test-uuid")

	// Verify results
	assert.NoError(t, err)
	assert.Equal(t, "var123", variationID)
}

func TestNewDefaultCmabClient_DefaultValues(t *testing.T) {
	// Test with empty options
	client := NewDefaultCmabClient(CmabClientOptions{})

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
	client := NewDefaultCmabClient(CmabClientOptions{
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

	_, err := client.FetchDecision("rule456", "user123", attributes, "test-uuid")
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
