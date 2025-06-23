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

func TestUpcomingCommand_TextFormatter(t *testing.T) {
	// Setup mock client
	mockClient := NewMockPagerDutyClient()
	mockClient.AddUser("USER001", "John Doe", "john@example.com")

	// Add some upcoming shifts
	now := time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC)
	shift1 := now.Add(24 * time.Hour)
	shift2 := now.Add(72 * time.Hour)

	mockClient.AddOnCall("USER001", "John Doe", "john@example.com",
		shift1, shift1.Add(8*time.Hour))
	mockClient.AddOnCall("USER001", "John Doe", "john@example.com",
		shift2, shift2.Add(12*time.Hour))

	cmd := NewUpcomingCommand(mockClient)
	formatter := NewTextFormatter()

	var buf bytes.Buffer
	err := cmd.Execute("SCHED123", "john@example.com", 7, formatter, &buf)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	output := buf.String()

	// Verify text format output
	expectedStrings := []string{
		"Shifts for the next",
		"days:",
		"2024-03-16",
		"2024-03-18",
		"to",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(output, expected) {
			t.Errorf("Expected output to contain %q, got:\n%s", expected, output)
		}
	}
}

func TestUpcomingCommand_ICalFormatter(t *testing.T) {
	// Setup mock client
	mockClient := NewMockPagerDutyClient()
	mockClient.AddUser("USER001", "John Doe", "john@example.com")

	// Add an upcoming shift
	shift1 := time.Date(2024, 3, 16, 9, 0, 0, 0, time.UTC)

	mockClient.AddOnCall("USER001", "John Doe", "john@example.com",
		shift1, shift1.Add(8*time.Hour))

	cmd := NewUpcomingCommand(mockClient)
	formatter := NewICalFormatter()

	var buf bytes.Buffer
	err := cmd.Execute("SCHED123", "john@example.com", 7, formatter, &buf)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	output := buf.String()

	// Verify iCal format output
	expectedStrings := []string{
		"BEGIN:VCALENDAR",
		"VERSION:2.0",
		"PRODID:-//myshift-go//ON-CALL SCHEDULE//EN",
		"BEGIN:VEVENT",
		"UID:oncall-0-USER001@myshift-go",
		"DTSTART:20240316T090000Z",
		"DTEND:20240316T170000Z",
		"SUMMARY:On-Call: John Doe",
		"CATEGORIES:ON-CALL",
		"END:VEVENT",
		"END:VCALENDAR",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(output, expected) {
			t.Errorf("Expected iCal output to contain %q, got:\n%s", expected, output)
		}
	}

	// Verify proper line endings (iCal requires CRLF)
	if !strings.Contains(output, "\r\n") {
		t.Error("iCal output should use CRLF line endings")
	}
}

func TestUpcomingCommand_EmptyShifts_TextFormatter(t *testing.T) {
	// Setup mock client with no shifts
	mockClient := NewMockPagerDutyClient()
	mockClient.AddUser("USER001", "John Doe", "john@example.com")

	cmd := NewUpcomingCommand(mockClient)
	formatter := NewTextFormatter()

	var buf bytes.Buffer
	err := cmd.Execute("SCHED123", "john@example.com", 7, formatter, &buf)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	output := buf.String()

	// Should show "No shifts found" message
	if !strings.Contains(output, "No shifts found") {
		t.Errorf("Expected 'No shifts found' message, got: %s", output)
	}
}

func TestUpcomingCommand_EmptyShifts_ICalFormatter(t *testing.T) {
	// Setup mock client with no shifts
	mockClient := NewMockPagerDutyClient()
	mockClient.AddUser("USER001", "John Doe", "john@example.com")

	cmd := NewUpcomingCommand(mockClient)
	formatter := NewICalFormatter()

	var buf bytes.Buffer
	err := cmd.Execute("SCHED123", "john@example.com", 7, formatter, &buf)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	output := buf.String()

	// Should still have valid iCal structure even with no events
	expectedStrings := []string{
		"BEGIN:VCALENDAR",
		"VERSION:2.0",
		"END:VCALENDAR",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(output, expected) {
			t.Errorf("Expected iCal output to contain %q, got:\n%s", expected, output)
		}
	}

	// Should not contain any events
	if strings.Contains(output, "BEGIN:VEVENT") {
		t.Error("Expected no VEVENT blocks in empty calendar")
	}
}

func TestUpcomingCommand_UserMapGeneration(t *testing.T) {
	// This test verifies that the user map is correctly generated
	// from the user data fetched during the upcoming command execution

	mockClient := NewMockPagerDutyClient()
	mockClient.AddUser("USER001", "John Doe", "john@example.com")

	// Add a shift
	now := time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC)
	shift1 := now.Add(24 * time.Hour)

	mockClient.AddOnCall("USER001", "John Doe", "john@example.com",
		shift1, shift1.Add(8*time.Hour))

	cmd := NewUpcomingCommand(mockClient)
	formatter := NewTextFormatter()

	var buf bytes.Buffer
	err := cmd.Execute("SCHED123", "john@example.com", 7, formatter, &buf)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	output := buf.String()

	// The user name should appear in the output, proving the user map was created correctly
	if !strings.Contains(output, "John Doe") {
		t.Errorf("Expected user name 'John Doe' to appear in output, got:\n%s", output)
	}

	// Verify the user lookup was called
	if len(mockClient.FindUserByEmailCalls) == 0 {
		t.Error("Expected FindUserByEmail to be called")
	} else if mockClient.FindUserByEmailCalls[0] != "john@example.com" {
		t.Errorf("Expected FindUserByEmail to be called with 'john@example.com', got %q",
			mockClient.FindUserByEmailCalls[0])
	}
}
