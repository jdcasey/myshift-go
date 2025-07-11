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
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/jdcasey/myshift-go/internal/types"
)

func TestTextFormatter_Format(t *testing.T) {
	formatter := NewTextFormatter()

	// Test data
	start := time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, 3, 22, 0, 0, 0, 0, time.UTC)

	shifts := []types.OnCall{
		{
			Start: time.Date(2024, 3, 15, 9, 0, 0, 0, time.UTC),
			End:   time.Date(2024, 3, 15, 17, 0, 0, 0, time.UTC),
			User: types.User{
				ID:   "USER001",
				Name: "John Doe",
			},
			Schedule: types.Schedule{
				ID:   "SCHED001",
				Name: "Primary Schedule",
			},
		},
		{
			Start: time.Date(2024, 3, 16, 9, 0, 0, 0, time.UTC),
			End:   time.Date(2024, 3, 16, 17, 0, 0, 0, time.UTC),
			User: types.User{
				ID:   "USER002",
				Name: "Jane Smith",
			},
			Schedule: types.Schedule{
				ID:   "SCHED001",
				Name: "Primary Schedule",
			},
		},
	}

	userMap := map[string]string{
		"USER001": "John Doe",
		"USER002": "Jane Smith",
	}

	var buf bytes.Buffer
	err := formatter.Format(&buf, shifts, userMap, start, end)

	if err != nil {
		t.Fatalf("TextFormatter.Format() failed: %v", err)
	}

	output := buf.String()

	// Verify output contains expected elements
	expectedStrings := []string{
		"Shifts for the next",
		"days:",
		"John Doe",
		"Jane Smith",
		"2024-03-15 09:00",
		"2024-03-16 09:00",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(output, expected) {
			t.Errorf("Expected output to contain %q, got:\n%s", expected, output)
		}
	}
}

func TestTextFormatter_Format_NoShifts(t *testing.T) {
	formatter := NewTextFormatter()

	start := time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, 3, 22, 0, 0, 0, 0, time.UTC)

	var buf bytes.Buffer
	err := formatter.Format(&buf, []types.OnCall{}, map[string]string{}, start, end)

	if err != nil {
		t.Fatalf("TextFormatter.Format() failed: %v", err)
	}

	output := buf.String()

	if !strings.Contains(output, "No shifts found") {
		t.Errorf("Expected output to contain 'No shifts found', got: %s", output)
	}
}

func TestICalFormatter_Format(t *testing.T) {
	formatter := NewICalFormatter()

	// Test data
	start := time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, 3, 22, 0, 0, 0, 0, time.UTC)

	shifts := []types.OnCall{
		{
			Start: time.Date(2024, 3, 15, 9, 0, 0, 0, time.UTC),
			End:   time.Date(2024, 3, 15, 17, 0, 0, 0, time.UTC),
			User: types.User{
				ID:   "USER001",
				Name: "John Doe",
			},
			Schedule: types.Schedule{
				ID:   "SCHED001",
				Name: "Primary Schedule",
			},
		},
	}

	userMap := map[string]string{
		"USER001": "John Doe",
	}

	var buf bytes.Buffer
	err := formatter.Format(&buf, shifts, userMap, start, end)

	if err != nil {
		t.Fatalf("ICalFormatter.Format() failed: %v", err)
	}

	output := buf.String()

	// Verify iCal structure
	expectedStrings := []string{
		"BEGIN:VCALENDAR",
		"VERSION:2.0",
		"PRODID:-//myshift-go//ON-CALL SCHEDULE//EN",
		"CALSCALE:GREGORIAN",
		"BEGIN:VEVENT",
		"UID:oncall-0-USER001@myshift-go",
		"DTSTART:20240315T090000Z",
		"DTEND:20240315T170000Z",
		"SUMMARY:On-Call: John Doe",
		"DESCRIPTION:On-call shift for John Doe\\nSchedule: Primary Schedule",
		"CATEGORIES:ON-CALL",
		"STATUS:CONFIRMED",
		"TRANSP:OPAQUE",
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

func TestICalFormatter_Format_EmptyShifts(t *testing.T) {
	formatter := NewICalFormatter()

	start := time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, 3, 22, 0, 0, 0, 0, time.UTC)

	var buf bytes.Buffer
	err := formatter.Format(&buf, []types.OnCall{}, map[string]string{}, start, end)

	if err != nil {
		t.Fatalf("ICalFormatter.Format() failed: %v", err)
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

func TestICalFormatter_EscapeText(t *testing.T) {
	formatter := NewICalFormatter()

	tests := []struct {
		input    string
		expected string
	}{
		{"Simple text", "Simple text"},
		{"Text with, comma", "Text with\\, comma"},
		{"Text with; semicolon", "Text with\\; semicolon"},
		{"Text with\nnewline", "Text with\\nnewline"},
		{"Text with\rcarriage return", "Text with\\rcarriage return"},
		{"Text with\\backslash", "Text with\\\\backslash"},
		{"Complex, text; with\nmultiple\rspecial characters", "Complex\\, text\\; with\\nmultiple\\rspecial characters"},
	}

	for _, tt := range tests {
		result := formatter.escapeICalText(tt.input)
		if result != tt.expected {
			t.Errorf("escapeICalText(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestGetFormatter(t *testing.T) {
	tests := []struct {
		format      string
		wantType    string
		shouldError bool
	}{
		{"text", "*commands.TextFormatter", false},
		{"txt", "*commands.TextFormatter", false},
		{"TEXT", "*commands.TextFormatter", false},
		{"ical", "*commands.ICalFormatter", false},
		{"ics", "*commands.ICalFormatter", false},
		{"ICAL", "*commands.ICalFormatter", false},
		{"invalid", "", true},
		{"json", "", true},
		{"", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.format, func(t *testing.T) {
			formatter, err := GetFormatter(tt.format)

			if tt.shouldError {
				if err == nil {
					t.Errorf("GetFormatter(%q) should have returned an error", tt.format)
				}
				return
			}

			if err != nil {
				t.Errorf("GetFormatter(%q) returned unexpected error: %v", tt.format, err)
				return
			}

			if formatter == nil {
				t.Errorf("GetFormatter(%q) returned nil formatter", tt.format)
				return
			}

			// Check that we can use the formatter (basic interface test)
			var buf bytes.Buffer
			testErr := formatter.Format(&buf, []types.OnCall{}, map[string]string{}, time.Now(), time.Now())
			if testErr != nil {
				t.Errorf("Formatter returned by GetFormatter(%q) failed to format: %v", tt.format, testErr)
			}
		})
	}
}

func TestDeduplicateOnCalls(t *testing.T) {
	// Test data setup
	now := time.Now()
	user1 := types.User{ID: "USER001", Name: "John Doe", Email: "john@example.com"}
	user2 := types.User{ID: "USER002", Name: "Jane Smith", Email: "jane@example.com"}
	schedule1 := types.Schedule{ID: "SCHED001", Name: "Test Schedule 1"}
	schedule2 := types.Schedule{ID: "SCHED002", Name: "Test Schedule 2"}

	tests := []struct {
		name     string
		input    []types.OnCall
		expected int // expected number of unique shifts
	}{
		{
			name:     "empty slice",
			input:    []types.OnCall{},
			expected: 0,
		},
		{
			name: "no duplicates",
			input: []types.OnCall{
				{Start: now, End: now.Add(8 * time.Hour), User: user1, Schedule: schedule1},
				{Start: now.Add(24 * time.Hour), End: now.Add(32 * time.Hour), User: user2, Schedule: schedule1},
			},
			expected: 2,
		},
		{
			name: "exact duplicates",
			input: []types.OnCall{
				{Start: now, End: now.Add(8 * time.Hour), User: user1, Schedule: schedule1},
				{Start: now, End: now.Add(8 * time.Hour), User: user1, Schedule: schedule1}, // duplicate
				{Start: now, End: now.Add(8 * time.Hour), User: user1, Schedule: schedule1}, // duplicate
			},
			expected: 1,
		},
		{
			name: "same user and time, different schedules",
			input: []types.OnCall{
				{Start: now, End: now.Add(8 * time.Hour), User: user1, Schedule: schedule1},
				{Start: now, End: now.Add(8 * time.Hour), User: user1, Schedule: schedule2}, // different schedule
			},
			expected: 2,
		},
		{
			name: "same schedule and time, different users",
			input: []types.OnCall{
				{Start: now, End: now.Add(8 * time.Hour), User: user1, Schedule: schedule1},
				{Start: now, End: now.Add(8 * time.Hour), User: user2, Schedule: schedule1}, // different user
			},
			expected: 2,
		},
		{
			name: "same user and schedule, different times",
			input: []types.OnCall{
				{Start: now, End: now.Add(8 * time.Hour), User: user1, Schedule: schedule1},
				{Start: now.Add(24 * time.Hour), End: now.Add(32 * time.Hour), User: user1, Schedule: schedule1}, // different time
			},
			expected: 2,
		},
		{
			name: "mixed duplicates and unique",
			input: []types.OnCall{
				{Start: now, End: now.Add(8 * time.Hour), User: user1, Schedule: schedule1},
				{Start: now, End: now.Add(8 * time.Hour), User: user1, Schedule: schedule1},                      // duplicate
				{Start: now.Add(24 * time.Hour), End: now.Add(32 * time.Hour), User: user2, Schedule: schedule1}, // unique
				{Start: now.Add(24 * time.Hour), End: now.Add(32 * time.Hour), User: user2, Schedule: schedule1}, // duplicate of above
			},
			expected: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DeduplicateOnCalls(tt.input)
			if len(result) != tt.expected {
				t.Errorf("DeduplicateOnCalls() returned %d items, expected %d", len(result), tt.expected)
			}

			// Verify no duplicates in result
			seen := make(map[string]bool)
			for _, shift := range result {
				key := fmt.Sprintf("%s-%s-%s-%s",
					shift.User.ID,
					shift.Schedule.ID,
					shift.Start.Format(time.RFC3339),
					shift.End.Format(time.RFC3339))
				if seen[key] {
					t.Errorf("Found duplicate in result: %s", key)
				}
				seen[key] = true
			}
		})
	}
}
