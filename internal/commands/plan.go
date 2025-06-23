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
	"net/url"
	"os"
	"time"

	"github.com/jdcasey/myshift-go/internal/pagerduty"
)

// PlanCommand handles the "plan" command functionality.
type PlanCommand struct {
	client pagerduty.PagerDutyClient
}

// NewPlanCommand creates a new PlanCommand instance.
func NewPlanCommand(client pagerduty.PagerDutyClient) *PlanCommand {
	return &PlanCommand{client: client}
}

// Execute runs the plan command to show all shifts in a schedule.
func (p *PlanCommand) Execute(scheduleID string, start, end time.Time, formatter PlanFormatter, writer io.Writer) error {
	// Get all on-call shifts for the schedule
	params := url.Values{
		"since":          []string{start.Format(time.RFC3339)},
		"until":          []string{end.Format(time.RFC3339)},
		"schedule_ids[]": []string{scheduleID},
		"overflow":       []string{"true"},
	}

	onCalls, err := p.client.GetOnCalls(params)
	if err != nil {
		return fmt.Errorf("error fetching shifts: %w", err)
	}

	// Build user map for display
	userMap := make(map[string]string)
	for _, shift := range onCalls {
		if _, exists := userMap[shift.User.ID]; !exists {
			user, err := p.client.GetUser(shift.User.ID)
			if err != nil {
				// If we can't get user details, use what we have
				userMap[shift.User.ID] = shift.User.Name
			} else {
				userMap[shift.User.ID] = user.Name
			}
		}
	}

	// If no writer is provided, default to stdout
	if writer == nil {
		writer = os.Stdout
	}

	// Use the formatter to output the data
	return formatter.Format(writer, onCalls, userMap, start, end)
}
