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

func TestUpcomingCommand_Execute(t *testing.T) {
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
			days:      28,
			setupMock: func(mock *MockPagerDutyClient, now time.Time) {
				// No users added
			},
			wantErr: true,
		},
		{
			testName:  "no upcoming shifts",
			userEmail: "john@example.com",
			days:      28,
			setupMock: func(mock *MockPagerDutyClient, now time.Time) {
				mock.AddUser("USER001", "John Doe", "john@example.com")
				// No shifts added
			},
			wantErr:    false,
			wantOutput: []string{"No shifts found"},
		},
		{
			testName:  "single upcoming shift",
			userEmail: "john@example.com",
			days:      28,
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
			},
		},
		{
			testName:  "multiple upcoming shifts",
			userEmail: "john@example.com",
			days:      7,
			setupMock: func(mock *MockPagerDutyClient, now time.Time) {
				mock.AddUser("USER001", "John Doe", "john@example.com")

				// Add multiple shifts over the next week
				day1 := now.Add(24 * time.Hour)
				day3 := now.Add(72 * time.Hour)
				day5 := now.Add(120 * time.Hour)

				mock.AddOnCall("USER001", "John Doe", "john@example.com",
					day1, day1.Add(8*time.Hour))
				mock.AddOnCall("USER001", "John Doe", "john@example.com",
					day3, day3.Add(12*time.Hour))
				mock.AddOnCall("USER001", "John Doe", "john@example.com",
					day5, day5.Add(8*time.Hour))
			},
			wantErr: false,
			wantOutput: []string{
				"Shifts for the next",
				"days:",
				"to",
			},
		},
		{
			testName:  "custom days parameter",
			userEmail: "john@example.com",
			days:      14,
			setupMock: func(mock *MockPagerDutyClient, now time.Time) {
				mock.AddUser("USER001", "John Doe", "john@example.com")
				nextWeek := now.Add(7 * 24 * time.Hour)
				mock.AddOnCall("USER001", "John Doe", "john@example.com",
					nextWeek, nextWeek.Add(8*time.Hour))
			},
			wantErr: false,
			wantOutput: []string{
				"Shifts for the next",
				"days:",
				"to",
			},
		},
		{
			testName:  "empty email",
			userEmail: "",
			days:      28,
			setupMock: func(mock *MockPagerDutyClient, now time.Time) {},
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			// Setup
			fixture := NewTestFixture()
			tt.setupMock(fixture.MockClient, fixture.Now)

			cmd := NewUpcomingCommand(fixture.MockClient)
			formatter := NewTextFormatter()

			// Capture output
			var buf bytes.Buffer

			// Execute
			err := cmd.Execute("SCHED123", tt.userEmail, tt.days, formatter, &buf)
			outputStr := buf.String()

			// Verify error expectation
			if (err != nil) != tt.wantErr {
				t.Errorf("UpcomingCommand.Execute() error = %v, wantErr %v", err, tt.wantErr)
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

func TestUpcomingCommand_Execute_ValidationErrors(t *testing.T) {
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
			cmd := NewUpcomingCommand(mockClient)
			formatter := NewTextFormatter()

			var buf bytes.Buffer
			err := cmd.Execute("SCHED123", tt.userEmail, 28, formatter, &buf)

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
func BenchmarkUpcomingCommand_Execute(b *testing.B) {
	fixture := NewTestFixture()
	fixture.MockClient.AddUser("USER001", "John Doe", "john@example.com")

	cmd := NewUpcomingCommand(fixture.MockClient)
	formatter := NewTextFormatter()

	// Use a discard writer for benchmarking
	var buf bytes.Buffer

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		_ = cmd.Execute("SCHED123", "john@example.com", 28, formatter, &buf)
	}
}
