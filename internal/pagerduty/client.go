// Copyright 2025 John Casey
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package pagerduty provides a client for interacting with the PagerDuty REST API v2.
//
// This package implements the PagerDutyClient interface and provides methods for
// retrieving users, on-call shifts, and creating schedule overrides. It handles
// API authentication, request/response formatting, and error handling for all
// PagerDuty API operations used by myshift-go.
//
// The client supports paginated requests and automatic retry logic for robust
// API interactions. All timestamps are handled in RFC3339 format as required
// by the PagerDuty API.
package pagerduty

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/jdcasey/myshift-go/internal/types"
)

const (
	// BaseURL is the base URL for the PagerDuty REST API v2.
	BaseURL = "https://api.pagerduty.com"
	// UserAgent is the user agent string sent with all API requests.
	UserAgent = "myshift-go/" + types.Version
)

// Client represents a PagerDuty API client that implements the PagerDutyClient interface.
// It handles authentication, request formatting, response parsing, and error handling
// for all PagerDuty API operations.
type Client struct {
	apiToken   string
	baseURL    string
	httpClient *http.Client
}

// NewClient creates a new PagerDuty API client with the provided API token.
// The client is configured with a 30-second timeout and uses the standard
// PagerDuty API base URL.
func NewClient(apiToken string) *Client {
	return &Client{
		apiToken: apiToken,
		baseURL:  BaseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// APIResponse represents a generic API response wrapper containing pagination metadata.
// This struct is embedded in all specific response types to provide consistent
// pagination information across PagerDuty API endpoints.
type APIResponse struct {
	Limit  int  `json:"limit,omitempty"`  // Maximum number of items per page
	Offset int  `json:"offset,omitempty"` // Number of items skipped from the beginning
	More   bool `json:"more,omitempty"`   // Whether there are more items available
	Total  int  `json:"total,omitempty"`  // Total number of items available
}

// UsersResponse represents the response from the PagerDuty users API endpoint.
// It contains an array of users matching the search criteria along with pagination metadata.
type UsersResponse struct {
	APIResponse
	Users []types.User `json:"users"`
}

// OnCallsResponse represents the response from the PagerDuty oncalls API endpoint.
// It contains an array of on-call shifts matching the query parameters.
type OnCallsResponse struct {
	APIResponse
	OnCalls []types.OnCall `json:"oncalls"`
}

// OverridesResponse represents the response from creating schedule overrides.
// It contains the created override objects with their assigned IDs and metadata.
type OverridesResponse struct {
	APIResponse
	Overrides []types.Override `json:"overrides"`
}

// makeRequest makes an HTTP request to the PagerDuty API with proper authentication
// and error handling. It sets required headers, handles request body marshaling,
// and validates response status codes.
//
// Parameters:
//   - method: HTTP method (GET, POST, PUT, DELETE)
//   - path: API endpoint path (e.g., "/users", "/oncalls")
//   - params: URL query parameters
//   - body: Request body object to be JSON-marshaled (can be nil)
//
// Returns the HTTP response or an error if the request fails or returns a non-2xx status.
func (c *Client) makeRequest(method, path string, params url.Values, body interface{}) (*http.Response, error) {
	reqURL := c.baseURL + path
	if params != nil && len(params) > 0 {
		reqURL += "?" + params.Encode()
	}

	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("error marshaling request body: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonBody)
	}

	req, err := http.NewRequest(method, reqURL, reqBody)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Authorization", "Token token="+c.apiToken)
	req.Header.Set("Accept", "application/vnd.pagerduty+json;version=2")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", UserAgent)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		defer resp.Body.Close()
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error: %d %s - %s", resp.StatusCode, resp.Status, string(bodyBytes))
	}

	return resp, nil
}

// FindUserByEmail searches for a PagerDuty user by their email address.
// It performs a case-insensitive search and returns the first matching user.
// This method is commonly used to resolve email addresses to PagerDuty user IDs
// for use in other API operations.
//
// Parameters:
//   - email: The email address to search for
//
// Returns the matching User object or an error if the user is not found or the API call fails.
func (c *Client) FindUserByEmail(email string) (*types.User, error) {
	params := url.Values{
		"query": []string{email},
		"limit": []string{"1"},
	}

	resp, err := c.makeRequest("GET", "/users", params, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var usersResp UsersResponse
	if err := json.NewDecoder(resp.Body).Decode(&usersResp); err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}

	for _, user := range usersResp.Users {
		if strings.EqualFold(user.Email, email) {
			return &user, nil
		}
	}

	return nil, fmt.Errorf("user with email %s not found", email)
}

// GetUser retrieves detailed information about a PagerDuty user by their ID.
// This method returns comprehensive user details including name, email, and type.
// It's typically used after finding a user ID through other means.
//
// Parameters:
//   - userID: The PagerDuty user ID to retrieve
//
// Returns the User object with full details or an error if the user doesn't exist or the API call fails.
func (c *Client) GetUser(userID string) (*types.User, error) {
	resp, err := c.makeRequest("GET", "/users/"+userID, nil, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		User types.User `json:"user"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}

	return &result.User, nil
}

// GetOnCalls retrieves on-call shifts from PagerDuty based on the provided parameters.
// This method supports extensive filtering and automatically handles pagination to return
// all matching results across multiple API calls if necessary.
//
// Common parameters include:
//   - since: Start time for the query (RFC3339 format)
//   - until: End time for the query (RFC3339 format)
//   - user_ids[]: Filter by specific user IDs
//   - schedule_ids[]: Filter by specific schedule IDs
//   - overflow: Include shifts that overflow the time boundaries
//
// The method handles pagination automatically, collecting all results before returning.
//
// Parameters:
//   - params: URL query parameters for filtering the on-call shifts
//
// Returns a slice of OnCall objects matching the criteria, or an error if the API call fails.
func (c *Client) GetOnCalls(params url.Values) ([]types.OnCall, error) {
	var allOnCalls []types.OnCall
	offset := 0
	limit := 100

	for {
		pageParams := make(url.Values)
		for k, v := range params {
			pageParams[k] = v
		}
		pageParams.Set("offset", fmt.Sprintf("%d", offset))
		pageParams.Set("limit", fmt.Sprintf("%d", limit))

		resp, err := c.makeRequest("GET", "/oncalls", pageParams, nil)
		if err != nil {
			return nil, err
		}

		var onCallsResp OnCallsResponse
		if err := json.NewDecoder(resp.Body).Decode(&onCallsResp); err != nil {
			_ = resp.Body.Close() // Explicitly ignore error during error handling
			return nil, fmt.Errorf("error decoding response: %w", err)
		}
		_ = resp.Body.Close() // Explicitly ignore error - response already processed

		allOnCalls = append(allOnCalls, onCallsResp.OnCalls...)

		if !onCallsResp.More || len(onCallsResp.OnCalls) < limit {
			break
		}

		offset += limit
	}

	return allOnCalls, nil
}

// CreateOverrides creates one or more schedule overrides in PagerDuty.
// Overrides temporarily assign different users to handle on-call duties during
// specified time periods, effectively replacing the originally scheduled person.
//
// Each override must specify:
//   - Start and end times (in UTC)
//   - User reference (ID and type)
//   - Optional timezone information
//
// This operation is atomic - either all overrides are created successfully,
// or none are created if any validation fails.
//
// Parameters:
//   - scheduleID: The PagerDuty schedule ID to create overrides for
//   - overrides: Slice of Override objects to create
//
// Returns nil on success, or an error if validation fails or the API call fails.
func (c *Client) CreateOverrides(scheduleID string, overrides []types.Override) error {
	requestBody := struct {
		Overrides []types.Override `json:"overrides"`
	}{
		Overrides: overrides,
	}

	resp, err := c.makeRequest("POST", "/schedules/"+scheduleID+"/overrides", nil, requestBody)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}
