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

func TestOverrideCommand_Execute(t *testing.T) {
	tests := []struct {
		testName    string
		userEmail   string
		targetEmail string
		start       time.Time
		end         time.Time
		setupMock   func(*MockPagerDutyClient, time.Time)
		wantErr     bool
		wantOutput  []string
		wantErrMsg  string
	}{
		{
			testName:    "user not found",
			userEmail:   "nonexistent@example.com",
			targetEmail: "target@example.com",
			setupMock: func(mock *MockPagerDutyClient, now time.Time) {
				// No users added
			},
			wantErr:    true,
			wantErrMsg: "error finding user nonexistent@example.com",
		},
		{
			testName:    "target user not found",
			userEmail:   "john@example.com",
			targetEmail: "nonexistent@example.com",
			setupMock: func(mock *MockPagerDutyClient, now time.Time) {
				mock.AddUser("USER001", "John Doe", "john@example.com")
			},
			wantErr:    true,
			wantErrMsg: "error finding target user nonexistent@example.com",
		},
		{
			testName:    "no target shifts found",
			userEmail:   "john@example.com",
			targetEmail: "jane@example.com",
			setupMock: func(mock *MockPagerDutyClient, now time.Time) {
				mock.AddUser("USER001", "John Doe", "john@example.com")
				mock.AddUser("USER002", "Jane Smith", "jane@example.com")
				// No shifts added for target user
			},
			wantErr:    true,
			wantErrMsg: "no shifts found for target user jane@example.com",
		},
		{
			testName:    "successful single override",
			userEmail:   "john@example.com",
			targetEmail: "jane@example.com",
			setupMock: func(mock *MockPagerDutyClient, now time.Time) {
				mock.AddUser("USER001", "John Doe", "john@example.com")
				mock.AddUser("USER002", "Jane Smith", "jane@example.com")

				// Add a shift for the target user
				tomorrow := now.Add(24 * time.Hour)
				mock.AddOnCall("USER002", "Jane Smith", "jane@example.com",
					tomorrow, tomorrow.Add(8*time.Hour))
			},
			wantErr: false,
			wantOutput: []string{
				"Successfully created 1 override(s) for John Doe",
				"Override 1:",
				"Start:",
				"End:",
			},
		},
		{
			testName:    "successful multiple overrides",
			userEmail:   "john@example.com",
			targetEmail: "jane@example.com",
			setupMock: func(mock *MockPagerDutyClient, now time.Time) {
				mock.AddUser("USER001", "John Doe", "john@example.com")
				mock.AddUser("USER002", "Jane Smith", "jane@example.com")

				// Add multiple shifts for the target user
				day1 := now.Add(24 * time.Hour)
				day2 := now.Add(48 * time.Hour)
				day3 := now.Add(72 * time.Hour)

				mock.AddOnCall("USER002", "Jane Smith", "jane@example.com",
					day1, day1.Add(8*time.Hour))
				mock.AddOnCall("USER002", "Jane Smith", "jane@example.com",
					day2, day2.Add(12*time.Hour))
				mock.AddOnCall("USER002", "Jane Smith", "jane@example.com",
					day3, day3.Add(8*time.Hour))
			},
			wantErr: false,
			wantOutput: []string{
				"Successfully created 3 override(s) for John Doe",
				"Override 1:",
				"Override 2:",
				"Override 3:",
				"Start:",
				"End:",
			},
		},
		{
			testName:    "empty user email",
			userEmail:   "",
			targetEmail: "jane@example.com",
			setupMock:   func(mock *MockPagerDutyClient, now time.Time) {},
			wantErr:     true,
			wantErrMsg:  "user email is required",
		},
		{
			testName:    "empty target email",
			userEmail:   "john@example.com",
			targetEmail: "",
			setupMock:   func(mock *MockPagerDutyClient, now time.Time) {},
			wantErr:     true,
			wantErrMsg:  "target user email is required",
		},
		{
			testName:    "time range filtering",
			userEmail:   "john@example.com",
			targetEmail: "jane@example.com",
			setupMock: func(mock *MockPagerDutyClient, now time.Time) {
				mock.AddUser("USER001", "John Doe", "john@example.com")
				mock.AddUser("USER002", "Jane Smith", "jane@example.com")

				// Add shifts - some in range, some out of range
				inRange := now.Add(12 * time.Hour)        // Within the test time range
				outOfRange := now.Add(5 * 24 * time.Hour) // Outside the test time range

				mock.AddOnCall("USER002", "Jane Smith", "jane@example.com",
					inRange, inRange.Add(8*time.Hour))
				mock.AddOnCall("USER002", "Jane Smith", "jane@example.com",
					outOfRange, outOfRange.Add(8*time.Hour))
			},
			wantErr: false,
			wantOutput: []string{
				"Successfully created",
				"override(s) for John Doe",
				"Override 1:",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			// Setup
			fixture := NewTestFixture()
			now := fixture.Now

			// Use default time range if not specified
			start := now
			end := now.Add(48 * time.Hour) // Default 2 days
			if !tt.start.IsZero() {
				start = tt.start
			}
			if !tt.end.IsZero() {
				end = tt.end
			}

			tt.setupMock(fixture.MockClient, now)

			cmd := NewOverrideCommand(fixture.MockClient)

			// Capture output
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			// Execute
			err := cmd.Execute("SCHED123", tt.userEmail, tt.targetEmail, start, end)

			// Restore stdout and read output
			w.Close()
			os.Stdout = oldStdout
			output, _ := io.ReadAll(r)
			outputStr := string(output)

			// Verify error expectation
			if (err != nil) != tt.wantErr {
				t.Errorf("OverrideCommand.Execute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Verify error message if error expected
			if tt.wantErr && tt.wantErrMsg != "" {
				if !strings.Contains(err.Error(), tt.wantErrMsg) {
					t.Errorf("Expected error message to contain %q, got %q",
						tt.wantErrMsg, err.Error())
				}
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
			if tt.userEmail != "" && tt.targetEmail != "" && !tt.wantErr {
				if len(fixture.MockClient.FindUserByEmailCalls) < 2 {
					t.Error("Expected FindUserByEmail to be called at least twice")
				}
				if len(fixture.MockClient.CreateOverridesCalls) == 0 {
					t.Error("Expected CreateOverrides to be called")
				}
			}
		})
	}
}

func TestOverrideCommand_Execute_ValidationErrors(t *testing.T) {
	tests := []struct {
		testName       string
		userEmail      string
		targetEmail    string
		expectedErrMsg string
	}{
		{
			testName:       "empty user email",
			userEmail:      "",
			targetEmail:    "target@example.com",
			expectedErrMsg: "user email is required",
		},
		{
			testName:       "empty target email",
			userEmail:      "user@example.com",
			targetEmail:    "",
			expectedErrMsg: "target user email is required",
		},
		{
			testName:       "both emails empty",
			userEmail:      "",
			targetEmail:    "",
			expectedErrMsg: "user email is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			mockClient := NewMockPagerDutyClient()
			cmd := NewOverrideCommand(mockClient)

			now := time.Now()
			start := now
			end := now.Add(24 * time.Hour)

			err := cmd.Execute("SCHED123", tt.userEmail, tt.targetEmail, start, end)

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

func TestOverrideCommand_Execute_OverrideCreation(t *testing.T) {
	// Test that verifies the actual override data is correct
	fixture := NewTestFixture()
	mockClient := fixture.MockClient

	// Setup users
	mockClient.AddUser("USER001", "John Doe", "john@example.com")
	mockClient.AddUser("USER002", "Jane Smith", "jane@example.com")

	// Add a shift for the target user
	now := fixture.Now
	shiftStart := now.Add(24 * time.Hour)
	shiftEnd := shiftStart.Add(8 * time.Hour)
	mockClient.AddOnCall("USER002", "Jane Smith", "jane@example.com", shiftStart, shiftEnd)

	cmd := NewOverrideCommand(mockClient)

	// Capture output but we're more interested in the API calls
	oldStdout := os.Stdout
	os.Stdout, _ = os.Open(os.DevNull)
	defer func() { os.Stdout = oldStdout }()

	start := now
	end := now.Add(48 * time.Hour)

	err := cmd.Execute("SCHED123", "john@example.com", "jane@example.com", start, end)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Verify that CreateOverrides was called with correct data
	if len(mockClient.CreateOverridesCalls) != 1 {
		t.Fatalf("Expected 1 CreateOverrides call, got %d", len(mockClient.CreateOverridesCalls))
	}

	call := mockClient.CreateOverridesCalls[0]
	if call.ScheduleID != "SCHED123" {
		t.Errorf("Expected schedule ID 'SCHED123', got %q", call.ScheduleID)
	}

	if len(call.Overrides) != 1 {
		t.Fatalf("Expected 1 override, got %d", len(call.Overrides))
	}

	override := call.Overrides[0]
	if override.User.ID != "USER001" {
		t.Errorf("Expected override user ID 'USER001', got %q", override.User.ID)
	}

	if override.User.Type != "user_reference" {
		t.Errorf("Expected user type 'user_reference', got %q", override.User.Type)
	}

	if !override.Start.Equal(shiftStart) {
		t.Errorf("Expected override start time %v, got %v", shiftStart, override.Start)
	}

	if !override.End.Equal(shiftEnd) {
		t.Errorf("Expected override end time %v, got %v", shiftEnd, override.End)
	}
}

// Benchmark to ensure performance is reasonable
func BenchmarkOverrideCommand_Execute(b *testing.B) {
	fixture := NewTestFixture()
	fixture.MockClient.AddUser("USER001", "John Doe", "john@example.com")
	fixture.MockClient.AddUser("USER002", "Jane Smith", "jane@example.com")

	// Add a shift for the target user
	now := fixture.Now
	tomorrow := now.Add(24 * time.Hour)
	fixture.MockClient.AddOnCall("USER002", "Jane Smith", "jane@example.com",
		tomorrow, tomorrow.Add(8*time.Hour))

	cmd := NewOverrideCommand(fixture.MockClient)

	// Silence output for benchmarking
	oldStdout := os.Stdout
	os.Stdout, _ = os.Open(os.DevNull)
	defer func() { os.Stdout = oldStdout }()

	start := now
	end := now.Add(48 * time.Hour)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Reset the mock calls for each iteration
		fixture.MockClient.CreateOverridesCalls = nil
		_ = cmd.Execute("SCHED123", "john@example.com", "jane@example.com", start, end)
	}
}
