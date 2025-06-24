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
	"net/url"
	"time"

	"github.com/jdcasey/myshift-go/internal/pagerduty"
	"github.com/jdcasey/myshift-go/internal/types"
)

// OverrideCommand handles the "override" command functionality.
type OverrideCommand struct {
	client pagerduty.PagerDutyClient
}

// NewOverrideCommand creates a new OverrideCommand instance.
func NewOverrideCommand(client pagerduty.PagerDutyClient) *OverrideCommand {
	return &OverrideCommand{client: client}
}

// Execute runs the override command to create schedule overrides.
func (o *OverrideCommand) Execute(scheduleID, userEmail, targetEmail string, start, end time.Time) error {
	if userEmail == "" {
		return fmt.Errorf("user email is required")
	}
	if targetEmail == "" {
		return fmt.Errorf("target user email is required")
	}

	// Find user who will take over the shift
	user, err := o.client.FindUserByEmail(userEmail)
	if err != nil {
		return fmt.Errorf("error finding user %s: %w", userEmail, err)
	}

	// Find target user whose shifts will be overridden
	targetUser, err := o.client.FindUserByEmail(targetEmail)
	if err != nil {
		return fmt.Errorf("error finding target user %s: %w", targetEmail, err)
	}

	// Get existing shifts for the target user in the time range
	params := url.Values{
		"since":          []string{start.Format(time.RFC3339)},
		"until":          []string{end.Format(time.RFC3339)},
		"user_ids[]":     []string{targetUser.ID},
		"schedule_ids[]": []string{scheduleID},
		"overflow":       []string{"true"},
	}

	onCalls, err := o.client.GetOnCalls(params)
	if err != nil {
		return fmt.Errorf("error fetching target shifts: %w", err)
	}

	if len(onCalls) == 0 {
		return fmt.Errorf("no shifts found for target user %s in the specified time range", targetEmail)
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

	fmt.Printf("Successfully created %d override(s) for %s\n", len(overrides), user.Name)
	for i, override := range overrides {
		fmt.Printf("Override %d:\n", i+1)
		fmt.Printf("  Start: %s\n", override.Start.Format("2006-01-02 15:04 MST"))
		fmt.Printf("  End: %s\n", override.End.Format("2006-01-02 15:04 MST"))
	}

	return nil
}
