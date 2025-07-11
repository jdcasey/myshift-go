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
)

// PlanCommand handles the "plan" command functionality.
type PlanCommand struct {
	*BaseCommand
}

// NewPlanCommand creates a new PlanCommand instance.
func NewPlanCommand(ctx *CommandContext) *PlanCommand {
	return &PlanCommand{
		BaseCommand: NewBaseCommand(ctx.Client, ctx.Config, ctx.Writer),
	}
}

// Execute runs the plan command to show all shifts in a schedule.
func (p *PlanCommand) Execute(args []string) error {
	parser := NewFlagParser("plan").
		AddDaysFlag(28, "Number of days to show").
		AddStartFlag("", "Start date (YYYY-MM-DD)").
		AddEndFlag("", "End date (YYYY-MM-DD)").
		AddFormatFlag("text", "Output format (text, ical)").
		SetUsage(func() {
			fmt.Print(`Usage: myshift plan [options]

Options:
  --days int         Number of days to show (default: 28)
  --start string     Start date (YYYY-MM-DD) 
  --end string       End date (YYYY-MM-DD)
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
	scheduleID, err := p.GetScheduleID()
	if err != nil {
		return err
	}

	// Calculate time range
	start, end, err := CalculateTimeRange(flags.Start, flags.End, flags.Days)
	if err != nil {
		return err
	}

	// Get all on-call shifts for the schedule
	onCalls, err := p.GetOnCallsForSchedule(scheduleID, start, end)
	if err != nil {
		return err
	}

	// Build user map for display
	userMap := p.BuildUserMap(onCalls)

	// Get the appropriate formatter
	formatter, err := GetFormatter(flags.Format)
	if err != nil {
		return err
	}

	// Use the formatter to output the data
	return formatter.Format(p.writer, onCalls, userMap, start, end)
}

// Usage returns the usage information for the plan command
func (p *PlanCommand) Usage() string {
	return `Usage: myshift plan [options]

Options:
  --days int         Number of days to show (default: 28)
  --start string     Start date (YYYY-MM-DD) 
  --end string       End date (YYYY-MM-DD)
  --format, -o string  Output format: text, ical (default: text)

`
}
