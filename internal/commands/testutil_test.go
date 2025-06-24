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
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/jdcasey/myshift-go/internal/types"
)

// MockPagerDutyClient is a mock implementation of PagerDutyClient for testing.
type MockPagerDutyClient struct {
	Users     map[string]*types.User
	OnCalls   []types.OnCall
	Overrides []types.Override

	// Track method calls for verification
	FindUserByEmailCalls []string
	GetUserCalls         []string
	GetOnCallsCalls      []url.Values
	CreateOverridesCalls []CreateOverridesCall
}

type CreateOverridesCall struct {
	ScheduleID string
	Overrides  []types.Override
}

// NewMockPagerDutyClient creates a new mock client with default test data.
func NewMockPagerDutyClient() *MockPagerDutyClient {
	return &MockPagerDutyClient{
		Users:   make(map[string]*types.User),
		OnCalls: []types.OnCall{},
	}
}

// AddUser adds a user to the mock client's data.
func (m *MockPagerDutyClient) AddUser(id, name, email string) {
	m.Users[email] = &types.User{
		ID:    id,
		Name:  name,
		Email: email,
		Type:  "user",
	}
	m.Users[id] = &types.User{
		ID:    id,
		Name:  name,
		Email: email,
		Type:  "user",
	}
}

// AddOnCall adds an on-call shift to the mock client's data.
func (m *MockPagerDutyClient) AddOnCall(userID, userName, userEmail string, start, end time.Time) {
	onCall := types.OnCall{
		Start: start,
		End:   end,
		User: types.User{
			ID:    userID,
			Name:  userName,
			Email: userEmail,
			Type:  "user",
		},
		Schedule: types.Schedule{
			ID:       "SCHED123",
			Name:     "Test Schedule",
			TimeZone: "UTC",
		},
	}
	m.OnCalls = append(m.OnCalls, onCall)
}

// FindUserByEmail implements PagerDutyClient interface.
func (m *MockPagerDutyClient) FindUserByEmail(email string) (*types.User, error) {
	m.FindUserByEmailCalls = append(m.FindUserByEmailCalls, email)

	user, exists := m.Users[email]
	if !exists {
		return nil, fmt.Errorf("user with email %s not found", email)
	}
	return user, nil
}

// GetUser implements PagerDutyClient interface.
func (m *MockPagerDutyClient) GetUser(userID string) (*types.User, error) {
	m.GetUserCalls = append(m.GetUserCalls, userID)

	user, exists := m.Users[userID]
	if !exists {
		return nil, fmt.Errorf("user with ID %s not found", userID)
	}
	return user, nil
}

// GetOnCalls implements PagerDutyClient interface.
func (m *MockPagerDutyClient) GetOnCalls(params url.Values) ([]types.OnCall, error) {
	m.GetOnCallsCalls = append(m.GetOnCallsCalls, params)

	// Filter by user_ids if specified
	userIDs := params["user_ids[]"]
	if len(userIDs) == 0 {
		return m.OnCalls, nil
	}

	var filtered []types.OnCall
	for _, onCall := range m.OnCalls {
		for _, userID := range userIDs {
			if onCall.User.ID == userID {
				filtered = append(filtered, onCall)
				break
			}
		}
	}

	return filtered, nil
}

// CreateOverrides implements PagerDutyClient interface.
func (m *MockPagerDutyClient) CreateOverrides(scheduleID string, overrides []types.Override) error {
	call := CreateOverridesCall{
		ScheduleID: scheduleID,
		Overrides:  make([]types.Override, len(overrides)),
	}
	copy(call.Overrides, overrides)
	m.CreateOverridesCalls = append(m.CreateOverridesCalls, call)

	m.Overrides = append(m.Overrides, overrides...)
	return nil
}

// TestFixture provides common test data and utilities.
type TestFixture struct {
	MockClient *MockPagerDutyClient
	Now        time.Time
}

// NewTestFixture creates a new test fixture with basic setup but no default shifts.
func NewTestFixture() *TestFixture {
	// Use current time so it matches what NextCommand will use
	now := time.Now()

	mockClient := NewMockPagerDutyClient()

	// Add test users but no default shifts
	// Tests can add their own shifts as needed
	mockClient.AddUser("USER001", "John Doe", "john@example.com")
	mockClient.AddUser("USER002", "Jane Smith", "jane@example.com")

	return &TestFixture{
		MockClient: mockClient,
		Now:        now,
	}
}

// CaptureOutput captures stdout during test execution.
// This is useful for testing commands that print output.
type OutputCapture struct {
	output strings.Builder
}

// NewOutputCapture creates a new output capture.
func NewOutputCapture() *OutputCapture {
	return &OutputCapture{}
}

// String returns the captured output.
func (oc *OutputCapture) String() string {
	return oc.output.String()
}

// Contains checks if the output contains the given string.
func (oc *OutputCapture) Contains(s string) bool {
	return strings.Contains(oc.output.String(), s)
}

// Lines returns the output split by lines.
func (oc *OutputCapture) Lines() []string {
	return strings.Split(strings.TrimSpace(oc.output.String()), "\n")
}
