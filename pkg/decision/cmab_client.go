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

// pkg/decision/cmab_client.go
package decision

import (
    "bytes"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "time"

    "github.com/optimizely/go-sdk/v2/pkg/logging"
)

const (
    // DefaultCmabAPIEndpoint is the default endpoint for CMAB API
    DefaultCmabAPIEndpoint = "https://api.optimizely.com/v1/cmab/decisions"
    // DefaultRequestTimeout is the default timeout for API requests in seconds
    DefaultRequestTimeout = 10
)

// CmabRequestPayload represents the request payload for the CMAB API
type CmabRequestPayload struct {
    RuleID     string                 `json:"ruleId"`
    UserID     string                 `json:"userId"`
    Attributes map[string]interface{} `json:"attributes,omitempty"`
}

// CmabResponsePayload represents the response payload from the CMAB API
type CmabResponsePayload struct {
    VariationID string `json:"variationId"`
    CmabUUID    string `json:"cmabUuid"`
}

// DefaultCmabClient implements the CmabClient interface
type DefaultCmabClient struct {
    sdkKey      string
    apiEndpoint string
    httpClient  *http.Client
    logger      logging.OptimizelyLogProducer
}

// CmabClientOption defines functional options for configuring the CMAB client
type CmabClientOption func(*DefaultCmabClient)

// WithAPIEndpoint sets a custom API endpoint
func WithAPIEndpoint(endpoint string) CmabClientOption {
    return func(c *DefaultCmabClient) {
        c.apiEndpoint = endpoint
    }
}

// WithHTTPClient sets a custom HTTP client
func WithHTTPClient(client *http.Client) CmabClientOption {
    return func(c *DefaultCmabClient) {
        c.httpClient = client
    }
}

// WithRequestTimeout sets a custom request timeout in seconds
func WithRequestTimeout(seconds int) CmabClientOption {
    return func(c *DefaultCmabClient) {
        c.httpClient.Timeout = time.Duration(seconds) * time.Second
    }
}

// NewCmabClient creates a new instance of DefaultCmabClient
func NewCmabClient(sdkKey string, options ...CmabClientOption) *DefaultCmabClient {
    client := &DefaultCmabClient{
        sdkKey:      sdkKey,
        apiEndpoint: DefaultCmabAPIEndpoint,
        httpClient: &http.Client{
            Timeout: time.Duration(DefaultRequestTimeout) * time.Second,
        },
        logger: logging.GetLogger(sdkKey, "CmabClient"),
    }

    // Apply options
    for _, opt := range options {
        opt(client)
    }

    return client
}

// FetchDecision fetches a decision from the CMAB API
func (c *DefaultCmabClient) FetchDecision(ruleID string, userID string, attributes map[string]interface{}) (CmabDecision, error) {
    // Prepare request payload
    payload := CmabRequestPayload{
        RuleID:     ruleID,
        UserID:     userID,
        Attributes: attributes,
    }

    // Convert payload to JSON
    payloadBytes, err := json.Marshal(payload)
    if err != nil {
        return CmabDecision{}, fmt.Errorf("failed to marshal request payload: %w", err)
    }

    // Create request
    req, err := http.NewRequest("POST", c.apiEndpoint, bytes.NewBuffer(payloadBytes))
    if err != nil {
        return CmabDecision{}, fmt.Errorf("failed to create request: %w", err)
    }

    // Set headers
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("X-API-Key", c.sdkKey)

    // Send request
    c.logger.Debug(fmt.Sprintf("Sending CMAB API request for rule %s and user %s", ruleID, userID))
    resp, err := c.httpClient.Do(req)
    if err != nil {
        return CmabDecision{}, fmt.Errorf("failed to send request: %w", err)
    }
    defer resp.Body.Close()

    // Read response body
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return CmabDecision{}, fmt.Errorf("failed to read response body: %w", err)
    }

    // Check response status
    if resp.StatusCode != http.StatusOK {
        return CmabDecision{}, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
    }

    // Parse response
    var responsePayload CmabResponsePayload
    if err := json.Unmarshal(body, &responsePayload); err != nil {
        return CmabDecision{}, fmt.Errorf("failed to unmarshal response: %w", err)
    }

    // Create decision
    decision := CmabDecision{
        VariationID: responsePayload.VariationID,
        CmabUUID:    responsePayload.CmabUUID,
    }

    return decision, nil
}
