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

package commands

import (
	"bytes"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/jdcasey/myshift-go/internal/types"
)

// MockPagerDutyClient is a mock implementation of PagerDutyClient for testing
type MockPagerDutyClient struct {
	users                  map[string]*types.User
	shifts                 []types.OnCall
	FindUserByEmailCalls   []string
	GetUserCalls           []string
	GetOnCallsCalls        []url.Values
	CreateOverridesCalls   []types.Override
	shouldErrorOnUser      bool
	shouldErrorOnOnCalls   bool
	shouldErrorOnOverrides bool
}

// NewMockPagerDutyClient creates a new mock PagerDuty client
func NewMockPagerDutyClient() *MockPagerDutyClient {
	return &MockPagerDutyClient{
		users:                make(map[string]*types.User),
		shifts:               []types.OnCall{},
		FindUserByEmailCalls: []string{},
		GetUserCalls:         []string{},
		GetOnCallsCalls:      []url.Values{},
		CreateOverridesCalls: []types.Override{},
	}
}

// AddUser adds a user to the mock client
func (m *MockPagerDutyClient) AddUser(id, name, email string) {
	m.users[email] = &types.User{
		ID:    id,
		Name:  name,
		Email: email,
		Type:  "user",
	}
}

// AddOnCall adds an on-call shift to the mock client
func (m *MockPagerDutyClient) AddOnCall(userID, userName, userEmail string, start, end time.Time) {
	m.shifts = append(m.shifts, types.OnCall{
		Start: start,
		End:   end,
		User: types.User{
			ID:    userID,
			Name:  userName,
			Email: userEmail,
			Type:  "user",
		},
		Schedule: types.Schedule{
			ID:          "SCHED123",
			Name:        "Test Schedule",
			Description: "Test schedule for testing",
			TimeZone:    "UTC",
		},
	})
}

// SetErrorOnUser makes the mock return an error for user operations
func (m *MockPagerDutyClient) SetErrorOnUser(shouldError bool) {
	m.shouldErrorOnUser = shouldError
}

// SetErrorOnOnCalls makes the mock return an error for on-call operations
func (m *MockPagerDutyClient) SetErrorOnOnCalls(shouldError bool) {
	m.shouldErrorOnOnCalls = shouldError
}

// SetErrorOnOverrides makes the mock return an error for override operations
func (m *MockPagerDutyClient) SetErrorOnOverrides(shouldError bool) {
	m.shouldErrorOnOverrides = shouldError
}

// FindUserByEmail implements PagerDutyClient interface
func (m *MockPagerDutyClient) FindUserByEmail(email string) (*types.User, error) {
	m.FindUserByEmailCalls = append(m.FindUserByEmailCalls, email)

	if m.shouldErrorOnUser {
		return nil, fmt.Errorf("mock error finding user")
	}

	user, exists := m.users[email]
	if !exists {
		return nil, fmt.Errorf("user with email %s not found", email)
	}

	return user, nil
}

// GetUser implements PagerDutyClient interface
func (m *MockPagerDutyClient) GetUser(userID string) (*types.User, error) {
	m.GetUserCalls = append(m.GetUserCalls, userID)

	if m.shouldErrorOnUser {
		return nil, fmt.Errorf("mock error getting user")
	}

	for _, user := range m.users {
		if user.ID == userID {
			return user, nil
		}
	}

	return nil, fmt.Errorf("user with ID %s not found", userID)
}

// GetOnCalls implements PagerDutyClient interface
func (m *MockPagerDutyClient) GetOnCalls(params url.Values) ([]types.OnCall, error) {
	m.GetOnCallsCalls = append(m.GetOnCallsCalls, params)

	if m.shouldErrorOnOnCalls {
		return nil, fmt.Errorf("mock error getting on-calls")
	}

	// Filter shifts based on parameters
	var filteredShifts []types.OnCall
	userIDs := params["user_ids[]"]
	scheduleIDs := params["schedule_ids[]"]

	for _, shift := range m.shifts {
		// Check user filter
		if len(userIDs) > 0 {
			userMatch := false
			for _, userID := range userIDs {
				if shift.User.ID == userID {
					userMatch = true
					break
				}
			}
			if !userMatch {
				continue
			}
		}

		// Check schedule filter
		if len(scheduleIDs) > 0 {
			scheduleMatch := false
			for _, scheduleID := range scheduleIDs {
				if shift.Schedule.ID == scheduleID {
					scheduleMatch = true
					break
				}
			}
			if !scheduleMatch {
				continue
			}
		}

		// Check time range
		if sinceStr := params.Get("since"); sinceStr != "" {
			since, err := time.Parse(time.RFC3339, sinceStr)
			if err == nil && shift.End.Before(since) {
				continue
			}
		}

		if untilStr := params.Get("until"); untilStr != "" {
			until, err := time.Parse(time.RFC3339, untilStr)
			if err == nil && shift.Start.After(until) {
				continue
			}
		}

		filteredShifts = append(filteredShifts, shift)
	}

	return filteredShifts, nil
}

// CreateOverrides implements PagerDutyClient interface
func (m *MockPagerDutyClient) CreateOverrides(scheduleID string, overrides []types.Override) error {
	if m.shouldErrorOnOverrides {
		return fmt.Errorf("mock error creating overrides")
	}

	m.CreateOverridesCalls = append(m.CreateOverridesCalls, overrides...)

	return nil
}

// TestFixture provides a complete test setup
type TestFixture struct {
	MockClient *MockPagerDutyClient
	Config     *types.Config
	Context    *CommandContext
	Buffer     *bytes.Buffer
	Now        time.Time
}

// NewTestFixture creates a complete test fixture
func NewTestFixture() *TestFixture {
	mockClient := NewMockPagerDutyClient()
	config := &types.Config{
		PagerDutyToken: "test-token",
		ScheduleID:     "SCHED123",
		MyUser:         "john@example.com",
	}

	buffer := &bytes.Buffer{}
	context := NewCommandContext(mockClient, config, buffer)

	// Add default test user
	mockClient.AddUser("USER001", "John Doe", "john@example.com")

	return &TestFixture{
		MockClient: mockClient,
		Config:     config,
		Context:    context,
		Buffer:     buffer,
		Now:        time.Now(),
	}
}

// GetOutput returns the captured output as a string
func (f *TestFixture) GetOutput() string {
	return f.Buffer.String()
}

// ContainsOutput checks if the output contains the expected strings
func (f *TestFixture) ContainsOutput(expected ...string) bool {
	output := f.GetOutput()
	for _, exp := range expected {
		if !strings.Contains(output, exp) {
			return false
		}
	}
	return true
}

// ClearOutput clears the captured output
func (f *TestFixture) ClearOutput() {
	f.Buffer.Reset()
}
