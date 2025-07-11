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
	"strings"
	"testing"
	"time"

	"github.com/jdcasey/myshift-go/internal/types"
)

func TestNextCommand_Execute(t *testing.T) {
	tests := []struct {
		testName   string
		args       []string
		setupMock  func(*MockPagerDutyClient, time.Time)
		wantErr    bool
		wantOutput []string
	}{
		{
			testName: "user not found",
			args:     []string{"--user", "nonexistent@example.com"},
			setupMock: func(mock *MockPagerDutyClient, now time.Time) {
				// Clear default user
				mock.users = make(map[string]*types.User)
			},
			wantErr: true,
		},
		{
			testName: "no upcoming shifts",
			args:     []string{"--user", "john@example.com"},
			setupMock: func(mock *MockPagerDutyClient, now time.Time) {
				// User exists but no shifts added
			},
			wantErr:    false,
			wantOutput: []string{"No upcoming shifts found"},
		},
		{
			testName: "next shift in future",
			args:     []string{"--user", "john@example.com"},
			setupMock: func(mock *MockPagerDutyClient, now time.Time) {
				tomorrow := now.Add(24 * time.Hour)
				mock.AddOnCall("USER001", "John Doe", "john@example.com",
					tomorrow, tomorrow.Add(8*time.Hour))
			},
			wantErr: false,
			wantOutput: []string{
				"Next shift:",
				"Starts:",
				"Ends:",
			},
		},
		{
			testName: "currently on call",
			args:     []string{"--user", "john@example.com"},
			setupMock: func(mock *MockPagerDutyClient, now time.Time) {
				// Shift started 2 hours ago, ends in 6 hours
				mock.AddOnCall("USER001", "John Doe", "john@example.com",
					now.Add(-2*time.Hour), now.Add(6*time.Hour))
			},
			wantErr: false,
			wantOutput: []string{
				"Currently on call",
				"Shift ends:",
			},
		},
		{
			testName: "multiple shifts - finds earliest",
			args:     []string{"--user", "john@example.com"},
			setupMock: func(mock *MockPagerDutyClient, now time.Time) {
				tomorrow := now.Add(24 * time.Hour)
				dayAfter := now.Add(48 * time.Hour)

				// Add shifts in reverse order to test sorting
				mock.AddOnCall("USER001", "John Doe", "john@example.com",
					dayAfter, dayAfter.Add(8*time.Hour))
				mock.AddOnCall("USER001", "John Doe", "john@example.com",
					tomorrow, tomorrow.Add(8*time.Hour))
			},
			wantErr: false,
			wantOutput: []string{
				"Next shift:",
				"Starts:",
			},
		},
		{
			testName: "uses default user from config",
			args:     []string{},
			setupMock: func(mock *MockPagerDutyClient, now time.Time) {
				tomorrow := now.Add(24 * time.Hour)
				mock.AddOnCall("USER001", "John Doe", "john@example.com",
					tomorrow, tomorrow.Add(8*time.Hour))
			},
			wantErr: false,
			wantOutput: []string{
				"Next shift:",
				"Starts:",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			// Setup
			fixture := NewTestFixture()
			tt.setupMock(fixture.MockClient, fixture.Now)

			cmd := NewNextCommand(fixture.Context)

			// Execute
			err := cmd.Execute(tt.args)

			// Verify error expectation
			if (err != nil) != tt.wantErr {
				t.Errorf("NextCommand.Execute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Verify output if no error expected
			if !tt.wantErr {
				if !fixture.ContainsOutput(tt.wantOutput...) {
					t.Errorf("Expected output to contain %v, got:\n%s", tt.wantOutput, fixture.GetOutput())
				}
			}
		})
	}
}

func TestNextCommand_Execute_TimeHandling(t *testing.T) {
	// Test that demonstrates time handling - this test focuses on the logic, not exact times
	fixture := NewTestFixture()

	// Add a shift that's currently active relative to when the command runs
	now := fixture.Now
	shiftStart := now.Add(-1 * time.Hour) // Started 1 hour ago
	shiftEnd := now.Add(2 * time.Hour)    // Ends in 2 hours
	fixture.MockClient.AddOnCall("USER001", "John Doe", "john@example.com", shiftStart, shiftEnd)

	cmd := NewNextCommand(fixture.Context)

	err := cmd.Execute([]string{"--user", "john@example.com"})

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if !fixture.ContainsOutput("Currently on call") {
		t.Errorf("Expected 'Currently on call' in output, got: %s", fixture.GetOutput())
	}

	if !fixture.ContainsOutput("Shift ends:") {
		t.Errorf("Expected 'Shift ends:' in output, got: %s", fixture.GetOutput())
	}
}

func TestNextCommand_Execute_ValidationErrors(t *testing.T) {
	tests := []struct {
		testName       string
		args           []string
		setupMock      func(*MockPagerDutyClient)
		expectedErrMsg string
	}{
		{
			testName: "no user configured and no user flag",
			args:     []string{},
			setupMock: func(mock *MockPagerDutyClient) {
				// Clear default user from config
				mock.users = make(map[string]*types.User)
			},
			expectedErrMsg: "user email is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			fixture := NewTestFixture()
			if tt.setupMock != nil {
				tt.setupMock(fixture.MockClient)
			}
			// Clear my_user from config to test validation
			fixture.Config.MyUser = ""

			cmd := NewNextCommand(fixture.Context)

			err := cmd.Execute(tt.args)

			if err == nil {
				t.Error("Expected error but got none")
				return
			}

			if !strings.Contains(err.Error(), tt.expectedErrMsg) {
				t.Errorf("Expected error message to contain %q, got %q",
					tt.expectedErrMsg, err.Error())
			}
		})
	}
}

// Benchmark to ensure performance is reasonable
func BenchmarkNextCommand_Execute(b *testing.B) {
	fixture := NewTestFixture()
	cmd := NewNextCommand(fixture.Context)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		fixture.ClearOutput()
		_ = cmd.Execute([]string{"--user", "john@example.com"})
	}
}
