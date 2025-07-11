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
	"testing"
	"time"
)

func TestUpcomingCommand_Execute(t *testing.T) {
	tests := []struct {
		testName   string
		args       []string
		setupMock  func(*MockPagerDutyClient, time.Time)
		wantErr    bool
		wantOutput []string
	}{
		{
			testName: "successful upcoming display",
			args:     []string{"--user", "john@example.com", "--days", "7"},
			setupMock: func(mock *MockPagerDutyClient, now time.Time) {
				tomorrow := now.Add(24 * time.Hour)
				mock.AddOnCall("USER001", "John Doe", "john@example.com",
					tomorrow, tomorrow.Add(8*time.Hour))
			},
			wantErr: false,
			wantOutput: []string{
				"Shifts for the next 7 days:",
			},
		},
		{
			testName: "no upcoming shifts",
			args:     []string{"--user", "john@example.com", "--days", "7"},
			setupMock: func(mock *MockPagerDutyClient, now time.Time) {
				// No shifts added
			},
			wantErr: false,
			wantOutput: []string{
				"No shifts found",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			// Setup
			fixture := NewTestFixture()
			tt.setupMock(fixture.MockClient, fixture.Now)

			cmd := NewUpcomingCommand(fixture.Context)

			// Execute
			err := cmd.Execute(tt.args)

			// Verify error expectation
			if (err != nil) != tt.wantErr {
				t.Errorf("UpcomingCommand.Execute() error = %v, wantErr %v", err, tt.wantErr)
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

func BenchmarkUpcomingCommand_Execute(b *testing.B) {
	fixture := NewTestFixture()
	cmd := NewUpcomingCommand(fixture.Context)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		fixture.ClearOutput()
		_ = cmd.Execute([]string{"--user", "john@example.com", "--days", "7"})
	}
}
