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
	"io"
	"os"
	"strings"
	"testing"
	"time"
)

func TestNextCommand_Execute(t *testing.T) {
	tests := []struct {
		testName   string
		userEmail  string
		days       int
		setupMock  func(*MockPagerDutyClient, time.Time)
		wantErr    bool
		wantOutput []string
	}{
		{
			testName:  "user not found",
			userEmail: "nonexistent@example.com",
			days:      90,
			setupMock: func(mock *MockPagerDutyClient, now time.Time) {
				// No users added
			},
			wantErr: true,
		},
		{
			testName:  "no upcoming shifts",
			userEmail: "john@example.com",
			days:      90,
			setupMock: func(mock *MockPagerDutyClient, now time.Time) {
				mock.AddUser("USER001", "John Doe", "john@example.com")
				// No shifts added
			},
			wantErr:    false,
			wantOutput: []string{"No upcoming shifts found"},
		},
		{
			testName:  "next shift in future",
			userEmail: "john@example.com",
			days:      90,
			setupMock: func(mock *MockPagerDutyClient, now time.Time) {
				mock.AddUser("USER001", "John Doe", "john@example.com")
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
			testName:  "currently on call",
			userEmail: "john@example.com",
			days:      90,
			setupMock: func(mock *MockPagerDutyClient, now time.Time) {
				mock.AddUser("USER001", "John Doe", "john@example.com")
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
			testName:  "multiple shifts - finds earliest",
			userEmail: "john@example.com",
			days:      90,
			setupMock: func(mock *MockPagerDutyClient, now time.Time) {
				mock.AddUser("USER001", "John Doe", "john@example.com")
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
			testName:  "empty email",
			userEmail: "",
			days:      90,
			setupMock: func(mock *MockPagerDutyClient, now time.Time) {},
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			// Setup
			fixture := NewTestFixture()
			tt.setupMock(fixture.MockClient, fixture.Now)

			cmd := NewNextCommand(fixture.MockClient)

			// Capture output
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			// Execute
			err := cmd.Execute("SCHED123", tt.userEmail, tt.days)

			// Restore stdout and read output
			w.Close()
			os.Stdout = oldStdout
			output, _ := io.ReadAll(r)
			outputStr := string(output)

			// Verify error expectation
			if (err != nil) != tt.wantErr {
				t.Errorf("NextCommand.Execute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Verify output if no error expected
			if !tt.wantErr {
				for _, expectedLine := range tt.wantOutput {
					if !strings.Contains(outputStr, expectedLine) {
						t.Errorf("Expected output to contain %q, got:\n%s", expectedLine, outputStr)
					}
				}
			}

			// Verify API calls were made correctly
			if tt.userEmail != "" && !tt.wantErr {
				if len(fixture.MockClient.FindUserByEmailCalls) == 0 {
					t.Error("Expected FindUserByEmail to be called")
				} else if fixture.MockClient.FindUserByEmailCalls[0] != tt.userEmail {
					t.Errorf("Expected FindUserByEmail to be called with %q, got %q",
						tt.userEmail, fixture.MockClient.FindUserByEmailCalls[0])
				}
			}
		})
	}
}

func TestNextCommand_Execute_TimeHandling(t *testing.T) {
	// Test that demonstrates time handling - this test focuses on the logic, not exact times
	mockClient := NewMockPagerDutyClient()
	mockClient.AddUser("USER001", "John Doe", "john@example.com")

	// Add a shift that's currently active relative to when the command runs
	now := time.Now()
	shiftStart := now.Add(-1 * time.Hour) // Started 1 hour ago
	shiftEnd := now.Add(2 * time.Hour)    // Ends in 2 hours
	mockClient.AddOnCall("USER001", "John Doe", "john@example.com", shiftStart, shiftEnd)

	cmd := NewNextCommand(mockClient)

	// Capture output
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := cmd.Execute("SCHED123", "john@example.com", 90)

	w.Close()
	os.Stdout = oldStdout
	output, _ := io.ReadAll(r)
	outputStr := string(output)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if !strings.Contains(outputStr, "Currently on call") {
		t.Errorf("Expected 'Currently on call' in output, got: %s", outputStr)
	}

	if !strings.Contains(outputStr, "Shift ends:") {
		t.Errorf("Expected 'Shift ends:' in output, got: %s", outputStr)
	}
}

func TestNextCommand_Execute_ValidationErrors(t *testing.T) {
	tests := []struct {
		testName       string
		userEmail      string
		expectedErrMsg string
	}{
		{
			testName:       "empty email",
			userEmail:      "",
			expectedErrMsg: "user email is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			mockClient := NewMockPagerDutyClient()
			cmd := NewNextCommand(mockClient)

			err := cmd.Execute("SCHED123", tt.userEmail, 90)

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
	cmd := NewNextCommand(fixture.MockClient)

	// Silence output for benchmarking
	oldStdout := os.Stdout
	os.Stdout, _ = os.Open(os.DevNull)
	defer func() { os.Stdout = oldStdout }()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = cmd.Execute("SCHED123", "john@example.com", 90)
	}
}
