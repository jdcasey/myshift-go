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

	"github.com/jdcasey/myshift-go/pkg/myshift"
)

const (
	// BaseURL is the base URL for the PagerDuty API
	BaseURL = "https://api.pagerduty.com"
	// UserAgent is the user agent string for requests
	UserAgent = "myshift-go/" + myshift.Version
)

// Client represents a PagerDuty API client.
type Client struct {
	apiToken   string
	baseURL    string
	httpClient *http.Client
}

// NewClient creates a new PagerDuty API client.
func NewClient(apiToken string) *Client {
	return &Client{
		apiToken: apiToken,
		baseURL:  BaseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// APIResponse represents a generic API response wrapper.
type APIResponse struct {
	Limit  int  `json:"limit,omitempty"`
	Offset int  `json:"offset,omitempty"`
	More   bool `json:"more,omitempty"`
	Total  int  `json:"total,omitempty"`
}

// UsersResponse represents the response from the users API.
type UsersResponse struct {
	APIResponse
	Users []myshift.User `json:"users"`
}

// OnCallsResponse represents the response from the oncalls API.
type OnCallsResponse struct {
	APIResponse
	OnCalls []myshift.OnCall `json:"oncalls"`
}

// OverridesResponse represents the response from creating overrides.
type OverridesResponse struct {
	APIResponse
	Overrides []myshift.Override `json:"overrides"`
}

// makeRequest makes an HTTP request to the PagerDuty API.
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

// FindUserByEmail finds a user by their email address.
func (c *Client) FindUserByEmail(email string) (*myshift.User, error) {
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

// GetUser gets a user by their ID.
func (c *Client) GetUser(userID string) (*myshift.User, error) {
	resp, err := c.makeRequest("GET", "/users/"+userID, nil, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		User myshift.User `json:"user"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}

	return &result.User, nil
}

// GetOnCalls gets on-call shifts with the specified parameters.
func (c *Client) GetOnCalls(params url.Values) ([]myshift.OnCall, error) {
	var allOnCalls []myshift.OnCall
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
			resp.Body.Close()
			return nil, fmt.Errorf("error decoding response: %w", err)
		}
		resp.Body.Close()

		allOnCalls = append(allOnCalls, onCallsResp.OnCalls...)

		if !onCallsResp.More || len(onCallsResp.OnCalls) < limit {
			break
		}

		offset += limit
	}

	return allOnCalls, nil
}

// CreateOverrides creates schedule overrides.
func (c *Client) CreateOverrides(scheduleID string, overrides []myshift.Override) error {
	requestBody := struct {
		Overrides []myshift.Override `json:"overrides"`
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
