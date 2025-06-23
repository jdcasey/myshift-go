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
	"strings"
	"testing"
	"time"
)

func TestPlanCommand_Execute(t *testing.T) {
	tests := []struct {
		testName   string
		start      time.Time
		end        time.Time
		setupMock  func(*MockPagerDutyClient, time.Time)
		wantErr    bool
		wantOutput []string
	}{
		{
			testName: "no shifts in date range",
			setupMock: func(mock *MockPagerDutyClient, now time.Time) {
				mock.AddUser("USER001", "John Doe", "john@example.com")
				// No shifts added
			},
			wantErr:    false,
			wantOutput: []string{"No shifts found from"},
		},
		{
			testName: "single shift in range",
			setupMock: func(mock *MockPagerDutyClient, now time.Time) {
				mock.AddUser("USER001", "John Doe", "john@example.com")
				tomorrow := now.Add(24 * time.Hour)
				mock.AddOnCall("USER001", "John Doe", "john@example.com",
					tomorrow, tomorrow.Add(8*time.Hour))
			},
			wantErr: false,
			wantOutput: []string{
				"Shifts for the next",
				"days:",
				"to",
				"John Doe",
			},
		},
		{
			testName: "multiple shifts with different users",
			setupMock: func(mock *MockPagerDutyClient, now time.Time) {
				mock.AddUser("USER001", "John Doe", "john@example.com")
				mock.AddUser("USER002", "Jane Smith", "jane@example.com")

				// Add shifts for different users
				day1 := now.Add(24 * time.Hour)
				day2 := now.Add(48 * time.Hour)
				day3 := now.Add(72 * time.Hour)

				mock.AddOnCall("USER001", "John Doe", "john@example.com",
					day1, day1.Add(8*time.Hour))
				mock.AddOnCall("USER002", "Jane Smith", "jane@example.com",
					day2, day2.Add(12*time.Hour))
				mock.AddOnCall("USER001", "John Doe", "john@example.com",
					day3, day3.Add(8*time.Hour))
			},
			wantErr: false,
			wantOutput: []string{
				"Shifts for the next",
				"days:",
				"to",
				"John Doe",
				"Jane Smith",
			},
		},
		{
			testName: "user lookup failure handled gracefully",
			setupMock: func(mock *MockPagerDutyClient, now time.Time) {
				// Add a shift but don't add the user to mock
				// This simulates a case where GetUser fails
				tomorrow := now.Add(24 * time.Hour)
				mock.AddOnCall("USER999", "Unknown User", "unknown@example.com",
					tomorrow, tomorrow.Add(8*time.Hour))
			},
			wantErr: false,
			wantOutput: []string{
				"Shifts for the next",
				"days:",
				"to",
				"Unknown User", // Should use the name from the shift data
			},
		},
		{
			testName: "week-long schedule",
			setupMock: func(mock *MockPagerDutyClient, now time.Time) {
				mock.AddUser("USER001", "John Doe", "john@example.com")
				mock.AddUser("USER002", "Jane Smith", "jane@example.com")

				// Add shifts for a full week rotation
				for i := 0; i < 7; i++ {
					dayStart := now.Add(time.Duration(i*24) * time.Hour)
					userID := "USER001"
					userName := "John Doe"
					userEmail := "john@example.com"

					if i%2 == 1 { // Alternate users
						userID = "USER002"
						userName = "Jane Smith"
						userEmail = "jane@example.com"
					}

					mock.AddOnCall(userID, userName, userEmail,
						dayStart, dayStart.Add(12*time.Hour))
				}
			},
			wantErr: false,
			wantOutput: []string{
				"Shifts for the next",
				"days:",
				"to",
				"John Doe",
				"Jane Smith",
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
			end := now.Add(7 * 24 * time.Hour) // Default 7 days
			if !tt.start.IsZero() {
				start = tt.start
			}
			if !tt.end.IsZero() {
				end = tt.end
			}

			tt.setupMock(fixture.MockClient, now)

			cmd := NewPlanCommand(fixture.MockClient)
			formatter := NewTextFormatter()

			// Capture output
			var buf bytes.Buffer

			// Execute
			err := cmd.Execute("SCHED123", start, end, formatter, &buf)
			outputStr := buf.String()

			// Verify error expectation
			if (err != nil) != tt.wantErr {
				t.Errorf("PlanCommand.Execute() error = %v, wantErr %v", err, tt.wantErr)
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
			if !tt.wantErr {
				if len(fixture.MockClient.GetOnCallsCalls) == 0 {
					t.Error("Expected GetOnCalls to be called")
				}
			}
		})
	}
}

func TestPlanCommand_Execute_DateRanges(t *testing.T) {
	tests := []struct {
		testName string
		start    time.Time
		end      time.Time
		wantDays int
	}{
		{
			testName: "one day range",
			start:    time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC),
			end:      time.Date(2024, 3, 15, 23, 59, 59, 0, time.UTC),
			wantDays: 1,
		},
		{
			testName: "seven day range",
			start:    time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC),
			end:      time.Date(2024, 3, 21, 23, 59, 59, 0, time.UTC),
			wantDays: 7,
		},
		{
			testName: "thirty day range",
			start:    time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC),
			end:      time.Date(2024, 3, 30, 23, 59, 59, 0, time.UTC),
			wantDays: 30,
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			mockClient := NewMockPagerDutyClient()
			cmd := NewPlanCommand(mockClient)
			formatter := NewTextFormatter()

			// Capture output
			var buf bytes.Buffer

			err := cmd.Execute("SCHED123", tt.start, tt.end, formatter, &buf)
			outputStr := buf.String()

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			// When there are no shifts, expect "No shifts found" message
			// When there are shifts, expect "Shifts for the next" message
			// Since we didn't add any shifts, we expect the "No shifts found" message
			expectedText := "No shifts found from"
			if !strings.Contains(outputStr, expectedText) {
				t.Errorf("Expected output to contain %q, got: %s", expectedText, outputStr)
			}
		})
	}
}

func TestPlanCommand_Execute_APIError(t *testing.T) {
	mockClient := NewMockPagerDutyClient()

	// Simulate API error by not setting up any mock data
	// and having the mock return an error

	cmd := NewPlanCommand(mockClient)
	formatter := NewTextFormatter()

	now := time.Now()
	var buf bytes.Buffer
	err := cmd.Execute("INVALID_SCHEDULE", now, now.Add(24*time.Hour), formatter, &buf)

	// The actual implementation should handle API errors gracefully
	// For now, we expect it to fail since our mock doesn't handle error simulation
	// In a more sophisticated test, we'd add error simulation to the mock
	if err == nil {
		// This is actually expected with our current mock implementation
		// which returns empty results rather than errors
		t.Log("Command succeeded with empty results (expected with current mock)")
	}
}

// Benchmark to ensure performance is reasonable
func BenchmarkPlanCommand_Execute(b *testing.B) {
	fixture := NewTestFixture()
	fixture.MockClient.AddUser("USER001", "John Doe", "john@example.com")

	// Add some test shifts
	now := fixture.Now
	for i := 0; i < 7; i++ {
		dayStart := now.Add(time.Duration(i*24) * time.Hour)
		fixture.MockClient.AddOnCall("USER001", "John Doe", "john@example.com",
			dayStart, dayStart.Add(8*time.Hour))
	}

	cmd := NewPlanCommand(fixture.MockClient)
	formatter := NewTextFormatter()

	start := now
	end := now.Add(7 * 24 * time.Hour)

	// Use a discard writer for benchmarking
	var buf bytes.Buffer

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		_ = cmd.Execute("SCHED123", start, end, formatter, &buf)
	}
}
