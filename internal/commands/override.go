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

	"github.com/jdcasey/myshift-go/internal/types"
)

// OverrideCommand handles the "override" command functionality.
type OverrideCommand struct {
	*BaseCommand
}

// NewOverrideCommand creates a new OverrideCommand instance.
func NewOverrideCommand(ctx *CommandContext) *OverrideCommand {
	return &OverrideCommand{
		BaseCommand: NewBaseCommand(ctx.Client, ctx.Config, ctx.Writer),
	}
}

// Execute runs the override command to create schedule overrides.
func (o *OverrideCommand) Execute(args []string) error {
	flags, err := ParseOverrideFlags(args)
	if err != nil {
		return err
	}

	// If flags is nil, help was displayed - exit gracefully
	if flags == nil {
		return nil
	}

	// Get schedule ID
	scheduleID, err := o.GetScheduleID()
	if err != nil {
		return err
	}

	// Parse time range
	start, end, err := ParseTimeRange(flags.Start, flags.End)
	if err != nil {
		return err
	}

	// Find user who will take over the shift
	user, err := o.client.FindUserByEmail(flags.User)
	if err != nil {
		return fmt.Errorf("error finding user %s: %w", flags.User, err)
	}

	// Find target user whose shifts will be overridden
	targetUser, err := o.client.FindUserByEmail(flags.Target)
	if err != nil {
		return fmt.Errorf("error finding target user %s: %w", flags.Target, err)
	}

	// Get existing shifts for the target user in the time range
	onCalls, err := o.GetOnCallsForUser(scheduleID, targetUser.ID, start, end)
	if err != nil {
		return fmt.Errorf("error fetching target shifts: %w", err)
	}

	if len(onCalls) == 0 {
		return fmt.Errorf("no shifts found for target user %s in the specified time range", flags.Target)
	}

	// Create overrides for each shift
	var overrides []types.Override
	for _, shift := range onCalls {
		override := types.Override{
			Start: shift.Start,
			End:   shift.End,
			User: types.UserReference{
				ID:   user.ID,
				Type: "user_reference",
			},
			TimeZone: "UTC",
		}
		overrides = append(overrides, override)
	}

	// Create the overrides
	if err := o.client.CreateOverrides(scheduleID, overrides); err != nil {
		return fmt.Errorf("error creating overrides: %w", err)
	}

	fmt.Fprintf(o.writer, "Successfully created %d override(s) for %s\n", len(overrides), user.Name)
	for i, override := range overrides {
		fmt.Fprintf(o.writer, "Override %d:\n", i+1)
		fmt.Fprintf(o.writer, "  Start: %s\n", override.Start.Format("2006-01-02 15:04 MST"))
		fmt.Fprintf(o.writer, "  End: %s\n", override.End.Format("2006-01-02 15:04 MST"))
	}

	return nil
}

// Usage returns the usage information for the override command
func (o *OverrideCommand) Usage() string {
	return `Usage: myshift override [options]

Options:
  --user string     User email to override with (required)
  --target string   Target user email to override (required)  
  --start string    Start time (YYYY-MM-DD HH:MM) (required)
  --end string      End time (YYYY-MM-DD HH:MM) (required)

`
}
