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

// Package commands provides CLI command implementations for myshift-go.
package commands

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/jdcasey/myshift-go/internal/types"
)

// PlanFormatter defines the interface for formatting plan output.
type PlanFormatter interface {
	Format(writer io.Writer, shifts []types.OnCall, userMap map[string]string, start, end time.Time) error
}

// TextFormatter formats plan output as human-readable text.
type TextFormatter struct{}

// NewTextFormatter creates a new text formatter.
func NewTextFormatter() *TextFormatter {
	return &TextFormatter{}
}

// Format outputs the shifts in a human-readable text format.
func (f *TextFormatter) Format(writer io.Writer, shifts []types.OnCall, userMap map[string]string, start, end time.Time) error {
	if len(shifts) == 0 {
		_, err := fmt.Fprintf(writer, "No shifts found from %s to %s\n",
			start.Format("2006-01-02"), end.Format("2006-01-02"))
		return err
	}

	days := int(end.Sub(start).Hours() / 24)
	_, err := fmt.Fprintf(writer, "Shifts for the next %d days:\n", days)
	if err != nil {
		return err
	}

	for _, shift := range shifts {
		userName := userMap[shift.User.ID]
		_, err := fmt.Fprintf(writer, "%s to %s: %s\n",
			shift.Start.Format("2006-01-02 15:04 MST"),
			shift.End.Format("2006-01-02 15:04 MST"),
			userName,
		)
		if err != nil {
			return err
		}
	}

	return nil
}

// ICalFormatter formats plan output as iCalendar (.ics) format.
type ICalFormatter struct{}

// NewICalFormatter creates a new iCal formatter.
func NewICalFormatter() *ICalFormatter {
	return &ICalFormatter{}
}

// Format outputs the shifts in iCalendar format.
func (f *ICalFormatter) Format(writer io.Writer, shifts []types.OnCall, userMap map[string]string, start, end time.Time) error {
	// Write iCal header
	_, err := fmt.Fprintf(writer, "BEGIN:VCALENDAR\r\n")
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(writer, "VERSION:2.0\r\n")
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(writer, "PRODID:-//myshift-go//ON-CALL SCHEDULE//EN\r\n")
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(writer, "CALSCALE:GREGORIAN\r\n")
	if err != nil {
		return err
	}

	// Write events for each shift
	for i, shift := range shifts {
		userName := userMap[shift.User.ID]

		// Generate a unique UID for each event
		uid := fmt.Sprintf("oncall-%d-%s@myshift-go", i, shift.User.ID)

		_, err = fmt.Fprintf(writer, "BEGIN:VEVENT\r\n")
		if err != nil {
			return err
		}

		_, err = fmt.Fprintf(writer, "UID:%s\r\n", uid)
		if err != nil {
			return err
		}

		_, err = fmt.Fprintf(writer, "DTSTART:%s\r\n", shift.Start.UTC().Format("20060102T150405Z"))
		if err != nil {
			return err
		}

		_, err = fmt.Fprintf(writer, "DTEND:%s\r\n", shift.End.UTC().Format("20060102T150405Z"))
		if err != nil {
			return err
		}

		_, err = fmt.Fprintf(writer, "SUMMARY:On-Call: %s\r\n", f.escapeICalText(userName))
		if err != nil {
			return err
		}

		_, err = fmt.Fprintf(writer, "DESCRIPTION:On-call shift for %s\\nSchedule: %s\r\n",
			f.escapeICalText(userName), f.escapeICalText(shift.Schedule.Name))
		if err != nil {
			return err
		}

		_, err = fmt.Fprintf(writer, "CATEGORIES:ON-CALL\r\n")
		if err != nil {
			return err
		}

		_, err = fmt.Fprintf(writer, "STATUS:CONFIRMED\r\n")
		if err != nil {
			return err
		}

		_, err = fmt.Fprintf(writer, "TRANSP:OPAQUE\r\n")
		if err != nil {
			return err
		}

		// Add timestamp
		_, err = fmt.Fprintf(writer, "DTSTAMP:%s\r\n", time.Now().UTC().Format("20060102T150405Z"))
		if err != nil {
			return err
		}

		_, err = fmt.Fprintf(writer, "END:VEVENT\r\n")
		if err != nil {
			return err
		}
	}

	// Write iCal footer
	_, err = fmt.Fprintf(writer, "END:VCALENDAR\r\n")
	return err
}

// escapeICalText escapes special characters in iCalendar text values.
func (f *ICalFormatter) escapeICalText(text string) string {
	// Escape special characters according to RFC 5545
	text = strings.ReplaceAll(text, "\\", "\\\\")
	text = strings.ReplaceAll(text, ",", "\\,")
	text = strings.ReplaceAll(text, ";", "\\;")
	text = strings.ReplaceAll(text, "\n", "\\n")
	text = strings.ReplaceAll(text, "\r", "\\r")
	return text
}

// GetFormatter returns the appropriate formatter based on the format string.
func GetFormatter(format string) (PlanFormatter, error) {
	switch strings.ToLower(format) {
	case "text", "txt":
		return NewTextFormatter(), nil
	case "ical", "ics":
		return NewICalFormatter(), nil
	default:
		return nil, fmt.Errorf("unsupported format: %s (supported: text, ical)", format)
	}
}

// DeduplicateOnCalls removes duplicate OnCall entries that have the same
// user, schedule, start time, and end time. This is needed because PagerDuty
// returns duplicate entries for schedules with multiple escalation paths.
func DeduplicateOnCalls(onCalls []types.OnCall) []types.OnCall {
	if len(onCalls) == 0 {
		return onCalls
	}

	seenShifts := make(map[string]bool)
	var uniqueShifts []types.OnCall

	for _, shift := range onCalls {
		// Create a unique key based on user, schedule, start, and end times
		key := fmt.Sprintf("%s-%s-%s-%s",
			shift.User.ID,
			shift.Schedule.ID,
			shift.Start.Format(time.RFC3339),
			shift.End.Format(time.RFC3339))

		if !seenShifts[key] {
			seenShifts[key] = true
			uniqueShifts = append(uniqueShifts, shift)
		}
	}

	return uniqueShifts
}
