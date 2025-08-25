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

// Package cmab provides contextual multi-armed bandit functionality
package cmab

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
// var CMABPredictionEndpoint = "https://prediction.cmab.optimizely.com/predict/%s"		// prod
var CMABPredictionEndpoint = "https://prep.prediction.cmab.optimizely.com/predict/%s" // rc

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

// Attribute represents an attribute in a CMAB request
type Attribute struct {
	ID    string      `json:"id"`
	Value interface{} `json:"value"`
	Type  string      `json:"type"`
}

// Instance represents an instance in a CMAB request
type Instance struct {
	VisitorID    string      `json:"visitorId"`
	ExperimentID string      `json:"experimentId"`
	Attributes   []Attribute `json:"attributes"`
	CmabUUID     string      `json:"cmabUUID"`
}

// Request represents a request to the CMAB API
type Request struct {
	Instances []Instance `json:"instances"`
}

// Prediction represents a prediction in a CMAB response
type Prediction struct {
	VariationID string `json:"variation_id"`
}

// Response represents a response from the CMAB API
type Response struct {
	Predictions []Prediction `json:"predictions"`
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

// ClientOptions defines options for creating a CMAB client
type ClientOptions struct {
	HTTPClient  *http.Client
	RetryConfig *RetryConfig
	Logger      logging.OptimizelyLogProducer
}

// NewDefaultCmabClient creates a new instance of DefaultCmabClient
func NewDefaultCmabClient(options ClientOptions) *DefaultCmabClient {
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

	// Log the URL being called
	c.logger.Debug(fmt.Sprintf("CMAB Prediction URL: %s", url))

	// Convert attributes to CMAB format
	cmabAttributes := make([]Attribute, 0, len(attributes))
	for key, value := range attributes {
		cmabAttributes = append(cmabAttributes, Attribute{
			ID:    key,
			Value: value,
			Type:  "custom_attribute",
		})
	}

	// Create the request body
	requestBody := Request{
		Instances: []Instance{
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

	// Log the request body
	c.logger.Debug(fmt.Sprintf("CMAB request body: %s", string(bodyBytes)))

	// If no retry config, just do a single fetch
	if c.retryConfig == nil {
		return c.doFetch(context.Background(), url, bodyBytes)
	}

	// Retry with exponential backoff
	var lastErr error
	for i := 0; i <= c.retryConfig.MaxRetries; i++ {
		// Make the request
		result, err := c.doFetch(context.Background(), url, bodyBytes)
		if err == nil {
			return result, nil
		}

		lastErr = err
		c.logger.Warning(fmt.Sprintf("CMAB API request failed (attempt %d/%d): %v",
			i+1, c.retryConfig.MaxRetries+1, err))

		// Don't wait after the last attempt
		if i < c.retryConfig.MaxRetries {
			// Calculate backoff duration with exponential backoff
			backoffDuration := c.retryConfig.InitialBackoff * time.Duration(math.Pow(c.retryConfig.BackoffMultiplier, float64(i)))
			if backoffDuration > c.retryConfig.MaxBackoff {
				backoffDuration = c.retryConfig.MaxBackoff
			}
			c.logger.Debug(fmt.Sprintf("CMAB request retry with backoff: %v", backoffDuration))
			time.Sleep(backoffDuration)
		}
	}

	return "", fmt.Errorf("failed to fetch CMAB decision after %d attempts: %w", c.retryConfig.MaxRetries, lastErr)
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

	// Log the raw response
	c.logger.Debug(fmt.Sprintf("CMAB raw response: %s", string(respBody)))

	// Parse response
	var cmabResponse Response
	if err := json.Unmarshal(respBody, &cmabResponse); err != nil {
		return "", fmt.Errorf("failed to unmarshal CMAB response: %w", err)
	}

	// Validate response
	if !c.validateResponse(cmabResponse) {
		return "", fmt.Errorf("invalid CMAB response: missing predictions or variation_id")
	}

	// Log the parsed variation ID
	variationID := cmabResponse.Predictions[0].VariationID
	c.logger.Debug(fmt.Sprintf("CMAB parsed variation ID: %s", variationID))

	// Return the variation ID
	return variationID, nil
}

// validateResponse validates the CMAB response
func (c *DefaultCmabClient) validateResponse(response Response) bool {
	return len(response.Predictions) > 0 && response.Predictions[0].VariationID != ""
}
