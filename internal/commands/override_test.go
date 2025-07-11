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
)

func TestOverrideCommand_Execute(t *testing.T) {
	tests := []struct {
		testName   string
		args       []string
		setupMock  func(*MockPagerDutyClient, time.Time)
		wantErr    bool
		wantOutput []string
		wantErrMsg string
	}{
		{
			testName: "successful single override",
			args:     []string{"--user", "john@example.com", "--target", "jane@example.com", "--start", "2024-03-15 09:00", "--end", "2024-03-15 17:00"},
			setupMock: func(mock *MockPagerDutyClient, now time.Time) {
				mock.AddUser("USER001", "John Doe", "john@example.com")
				mock.AddUser("USER002", "Jane Smith", "jane@example.com")

				// Add a shift for the target user
				start, _ := time.Parse("2006-01-02 15:04", "2024-03-15 09:00")
				end, _ := time.Parse("2006-01-02 15:04", "2024-03-15 17:00")
				mock.AddOnCall("USER002", "Jane Smith", "jane@example.com", start, end)
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
			testName: "missing required flags",
			args:     []string{"--user", "john@example.com"},
			setupMock: func(mock *MockPagerDutyClient, now time.Time) {
				mock.AddUser("USER001", "John Doe", "john@example.com")
			},
			wantErr:    true,
			wantErrMsg: "are all required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			// Setup
			fixture := NewTestFixture()
			tt.setupMock(fixture.MockClient, fixture.Now)

			cmd := NewOverrideCommand(fixture.Context)

			// Execute
			err := cmd.Execute(tt.args)

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
				if !fixture.ContainsOutput(tt.wantOutput...) {
					t.Errorf("Expected output to contain %v, got:\n%s", tt.wantOutput, fixture.GetOutput())
				}
			}
		})
	}
}

// Simple benchmark to ensure performance is reasonable
func BenchmarkOverrideCommand_Execute(b *testing.B) {
	fixture := NewTestFixture()
	fixture.MockClient.AddUser("USER002", "Jane Smith", "jane@example.com")

	// Add a shift for the target user
	start, _ := time.Parse("2006-01-02 15:04", "2024-03-15 09:00")
	end, _ := time.Parse("2006-01-02 15:04", "2024-03-15 17:00")
	fixture.MockClient.AddOnCall("USER002", "Jane Smith", "jane@example.com", start, end)

	cmd := NewOverrideCommand(fixture.Context)

	args := []string{"--user", "john@example.com", "--target", "jane@example.com", "--start", "2024-03-15 09:00", "--end", "2024-03-15 17:00"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		fixture.ClearOutput()
		_ = cmd.Execute(args)
	}
}
