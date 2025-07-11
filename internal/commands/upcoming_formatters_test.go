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

func TestUpcomingCommand_Formatters(t *testing.T) {
	tests := []struct {
		testName string
		args     []string
		wantErr  bool
	}{
		{
			testName: "text format",
			args:     []string{"--user", "john@example.com", "--format", "text"},
			wantErr:  false,
		},
		{
			testName: "ical format",
			args:     []string{"--user", "john@example.com", "--format", "ical"},
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			// Setup
			fixture := NewTestFixture()

			// Add a shift for testing
			tomorrow := fixture.Now.Add(24 * time.Hour)
			fixture.MockClient.AddOnCall("USER001", "John Doe", "john@example.com",
				tomorrow, tomorrow.Add(8*time.Hour))

			cmd := NewUpcomingCommand(fixture.Context)

			// Execute
			err := cmd.Execute(tt.args)

			// Verify error expectation
			if (err != nil) != tt.wantErr {
				t.Errorf("UpcomingCommand.Execute() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
