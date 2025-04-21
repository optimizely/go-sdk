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

// Package decision provides CMAB client implementation
package decision

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"time"

	"github.com/optimizely/go-sdk/v2/pkg/logging"
)

// CMABPredictionEndpoint is the endpoint for CMAB predictions
var CMABPredictionEndpoint = "https://prediction.cmab.optimizely.com/predict/%s"

const (
	// DefaultMaxRetries is the default number of retries for CMAB requests
	DefaultMaxRetries = 3
	// DefaultInitialBackoff is the default initial backoff duration
	DefaultInitialBackoff = 100 * time.Millisecond
	// DefaultMaxBackoff is the default maximum backoff duration
	DefaultMaxBackoff = 10 * time.Second
	// DefaultBackoffMultiplier is the default multiplier for exponential backoff
	DefaultBackoffMultiplier = 2.0
)

// CMABAttribute represents an attribute in a CMAB request
type CMABAttribute struct {
	ID    string      `json:"id"`
	Value interface{} `json:"value"`
	Type  string      `json:"type"`
}

// CMABInstance represents an instance in a CMAB request
type CMABInstance struct {
	VisitorID    string          `json:"visitorId"`
	ExperimentID string          `json:"experimentId"`
	Attributes   []CMABAttribute `json:"attributes"`
	CmabUUID     string          `json:"cmabUUID"`
}

// CMABRequest represents a request to the CMAB API
type CMABRequest struct {
	Instances []CMABInstance `json:"instances"`
}

// CMABPrediction represents a prediction in a CMAB response
type CMABPrediction struct {
	VariationID string `json:"variation_id"`
}

// CMABResponse represents a response from the CMAB API
type CMABResponse struct {
	Predictions []CMABPrediction `json:"predictions"`
}

// RetryConfig defines configuration for retry behavior
type RetryConfig struct {
	// MaxRetries is the maximum number of retry attempts
	MaxRetries int
	// InitialBackoff is the initial backoff duration
	InitialBackoff time.Duration
	// MaxBackoff is the maximum backoff duration
	MaxBackoff time.Duration
	// BackoffMultiplier is the multiplier for exponential backoff
	BackoffMultiplier float64
}

// DefaultCmabClient implements the CmabClient interface
type DefaultCmabClient struct {
	httpClient  *http.Client
	retryConfig *RetryConfig
	logger      logging.OptimizelyLogProducer
}

// CmabClientOptions defines options for creating a CMAB client
type CmabClientOptions struct {
	HTTPClient  *http.Client
	RetryConfig *RetryConfig
	Logger      logging.OptimizelyLogProducer
}

// NewDefaultCmabClient creates a new instance of DefaultCmabClient
func NewDefaultCmabClient(options CmabClientOptions) *DefaultCmabClient {
	httpClient := options.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{
			Timeout: 10 * time.Second,
		}
	}

	// retry is optional:
	// retryConfig can be nil - in that case, no retries will be performed
	retryConfig := options.RetryConfig

	logger := options.Logger
	if logger == nil {
		logger = logging.GetLogger("", "DefaultCmabClient")
	}

	return &DefaultCmabClient{
		httpClient:  httpClient,
		retryConfig: retryConfig,
		logger:      logger,
	}
}

// FetchDecision fetches a decision from the CMAB API
func (c *DefaultCmabClient) FetchDecision(
	ruleID string,
	userID string,
	attributes map[string]interface{},
	cmabUUID string,
) (string, error) {
	// Create the URL
	url := fmt.Sprintf(CMABPredictionEndpoint, ruleID)

	// Convert attributes to CMAB format
	cmabAttributes := make([]CMABAttribute, 0, len(attributes))
	for key, value := range attributes {
		cmabAttributes = append(cmabAttributes, CMABAttribute{
			ID:    key,
			Value: value,
			Type:  "custom_attribute",
		})
	}

	// Create the request body
	requestBody := CMABRequest{
		Instances: []CMABInstance{
			{
				VisitorID:    userID,
				ExperimentID: ruleID,
				Attributes:   cmabAttributes,
				CmabUUID:     cmabUUID,
			},
		},
	}

	// Serialize the request body
	bodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal CMAB request: %w", err)
	}

	// Create context for cancellation
	ctx := context.Background()

	// If no retry config, just do a single fetch
	if c.retryConfig == nil {
		return c.doFetch(ctx, url, bodyBytes)
	}

	// Retry sending request with exponential backoff
	for i := 0; i <= c.retryConfig.MaxRetries; i++ {
		// Make the request
		result, err := c.doFetch(ctx, url, bodyBytes)
		if err == nil {
			return result, nil
		}

		// If this is the last retry, return the error
		if i == c.retryConfig.MaxRetries {
			return "", fmt.Errorf("failed to fetch CMAB decision after %d attempts: %w",
				c.retryConfig.MaxRetries, err)
		}

		// Calculate backoff duration
		backoffDuration := c.retryConfig.InitialBackoff * time.Duration(math.Pow(c.retryConfig.BackoffMultiplier, float64(i)))
		if backoffDuration > c.retryConfig.MaxBackoff {
			backoffDuration = c.retryConfig.MaxBackoff
		}

		c.logger.Debug(fmt.Sprintf("CMAB request retry %d/%d, backing off for %v",
			i+1, c.retryConfig.MaxRetries, backoffDuration))

		// Wait for backoff duration
		time.Sleep(backoffDuration)

		c.logger.Warning(fmt.Sprintf("CMAB API request failed (attempt %d/%d): %v",
			i+1, c.retryConfig.MaxRetries, err))
	}

	// This should never be reached due to the return in the loop above
	return "", fmt.Errorf("unexpected error in retry loop")
}

// doFetch performs a single fetch operation to the CMAB API
func (c *DefaultCmabClient) doFetch(ctx context.Context, url string, bodyBytes []byte) (string, error) {
	// Create the request
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return "", fmt.Errorf("failed to create CMAB request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")

	// Execute the request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("CMAB request failed: %w", err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("CMAB API returned non-success status code: %d", resp.StatusCode)
	}

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read CMAB response body: %w", err)
	}

	// Parse response
	var cmabResponse CMABResponse
	if err := json.Unmarshal(respBody, &cmabResponse); err != nil {
		return "", fmt.Errorf("failed to unmarshal CMAB response: %w", err)
	}

	// Validate response
	if !c.validateResponse(cmabResponse) {
		return "", fmt.Errorf("invalid CMAB response: missing predictions or variation_id")
	}

	// Return the variation ID
	return cmabResponse.Predictions[0].VariationID, nil
}

// validateResponse validates the CMAB response
func (c *DefaultCmabClient) validateResponse(response CMABResponse) bool {
	return len(response.Predictions) > 0 && response.Predictions[0].VariationID != ""
}
