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
	"time"
)

// UpcomingCommand handles the "upcoming" command functionality.
type UpcomingCommand struct {
	*BaseCommand
}

// NewUpcomingCommand creates a new UpcomingCommand instance.
func NewUpcomingCommand(ctx *CommandContext) *UpcomingCommand {
	return &UpcomingCommand{
		BaseCommand: NewBaseCommand(ctx.Client, ctx.Config, ctx.Writer),
	}
}

// Execute runs the upcoming command to show all upcoming shifts for a user.
func (u *UpcomingCommand) Execute(args []string) error {
	parser := NewFlagParser("upcoming").
		AddUserFlag("", "User email address (uses my_user from config if not provided)").
		AddDaysFlag(28, "Number of days to look ahead").
		AddFormatFlag("text", "Output format (text, ical)").
		SetUsage(func() {
			fmt.Print(`Usage: myshift upcoming [options]

Options:
  --user string        User email address (uses my_user from config if not provided)
  --days int           Number of days to look ahead (default: 28)
  --format, -o string  Output format: text, ical (default: text)

`)
		})

	flags, err := parser.Parse(args)
	if err != nil {
		return err
	}

	// If flags is nil, help was displayed - exit gracefully
	if flags == nil {
		return nil
	}

	// Get schedule ID
	scheduleID, err := u.GetScheduleID()
	if err != nil {
		return err
	}

	// Find user by email
	user, err := u.ResolveUser(flags.User)
	if err != nil {
		return err
	}

	// Calculate time range
	now := time.Now()
	until := now.AddDate(0, 0, flags.Days)

	// Get on-call shifts
	onCalls, err := u.GetOnCallsForUser(scheduleID, user.ID, now, until)
	if err != nil {
		return err
	}

	// Create user map with the user we already fetched
	userMap := map[string]string{
		user.ID: user.Name,
	}

	// Get the appropriate formatter
	formatter, err := GetFormatter(flags.Format)
	if err != nil {
		return err
	}

	// Use the formatter to output the data
	return formatter.Format(u.writer, onCalls, userMap, now, until)
}

// Usage returns the usage information for the upcoming command
func (u *UpcomingCommand) Usage() string {
	return `Usage: myshift upcoming [options]

Options:
  --user string        User email address (uses my_user from config if not provided)
  --days int           Number of days to look ahead (default: 28)
  --format, -o string  Output format: text, ical (default: text)

`
}
