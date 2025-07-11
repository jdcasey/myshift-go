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

// NextCommand handles the "next" command functionality.
type NextCommand struct {
	*BaseCommand
}

// NewNextCommand creates a new NextCommand instance.
func NewNextCommand(ctx *CommandContext) *NextCommand {
	return &NextCommand{
		BaseCommand: NewBaseCommand(ctx.Client, ctx.Config, ctx.Writer),
	}
}

// Execute runs the next command to show the next on-call shift for a user.
func (n *NextCommand) Execute(args []string) error {
	parser := NewFlagParser("next").
		AddUserFlag("", "User email address (uses my_user from config if not provided)").
		AddDaysFlag(90, "Number of days to look ahead").
		SetUsage(func() {
			fmt.Print(`Usage: myshift next [options]

Options:
  --user string   User email address (uses my_user from config if not provided)
  --days int      Number of days to look ahead (default: 90)

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
	scheduleID, err := n.GetScheduleID()
	if err != nil {
		return err
	}

	// Find user by email
	user, err := n.ResolveUser(flags.User)
	if err != nil {
		return err
	}

	// Calculate time range
	now := time.Now()
	until := now.AddDate(0, 0, flags.Days)

	// Get on-call shifts
	onCalls, err := n.GetOnCallsForUser(scheduleID, user.ID, now, until)
	if err != nil {
		return err
	}

	if len(onCalls) == 0 {
		fmt.Fprintln(n.writer, "No upcoming shifts found")
		return nil
	}

	// Find the next shift
	nextShift := onCalls[0]
	for _, shift := range onCalls {
		if shift.Start.Before(nextShift.Start) {
			nextShift = shift
		}
	}

	// Check if currently on call
	if nextShift.Start.Before(now) && nextShift.End.After(now) {
		fmt.Fprintln(n.writer, "Currently on call")
		fmt.Fprintf(n.writer, "Shift ends: %s\n", nextShift.End.Format("2006-01-02 15:04 MST"))
	} else if nextShift.Start.After(now) {
		fmt.Fprintln(n.writer, "Next shift:")
		fmt.Fprintf(n.writer, "Starts: %s\n", nextShift.Start.Format("2006-01-02 15:04 MST"))
		fmt.Fprintf(n.writer, "Ends: %s\n", nextShift.End.Format("2006-01-02 15:04 MST"))
	} else {
		fmt.Fprintln(n.writer, "No upcoming shifts found")
	}

	return nil
}

// Usage returns the usage information for the next command
func (n *NextCommand) Usage() string {
	return `Usage: myshift next [options]

Options:
  --user string   User email address (uses my_user from config if not provided)
  --days int      Number of days to look ahead (default: 90)

`
}
