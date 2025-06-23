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

// UpcomingCommand handles the "upcoming" command functionality.
type UpcomingCommand struct {
	client pagerduty.PagerDutyClient
}

// NewUpcomingCommand creates a new UpcomingCommand instance.
func NewUpcomingCommand(client pagerduty.PagerDutyClient) *UpcomingCommand {
	return &UpcomingCommand{client: client}
}

// Execute runs the upcoming command to show all upcoming shifts for a user.
func (u *UpcomingCommand) Execute(scheduleID, userEmail string, days int, formatter PlanFormatter, writer io.Writer) error {
	if userEmail == "" {
		return fmt.Errorf("user email is required")
	}

	// Find user by email
	user, err := u.client.FindUserByEmail(userEmail)
	if err != nil {
		return fmt.Errorf("error finding user: %w", err)
	}

	// Calculate time range
	now := time.Now()
	until := now.AddDate(0, 0, days)

	// Get on-call shifts
	params := url.Values{
		"since":          []string{now.Format(time.RFC3339)},
		"until":          []string{until.Format(time.RFC3339)},
		"user_ids[]":     []string{user.ID},
		"schedule_ids[]": []string{scheduleID},
		"overflow":       []string{"true"},
	}

	onCalls, err := u.client.GetOnCalls(params)
	if err != nil {
		return fmt.Errorf("error fetching shifts: %w", err)
	}

	// Create user map with the user we already fetched
	userMap := map[string]string{
		user.ID: user.Name,
	}

	// If no writer is provided, default to stdout
	if writer == nil {
		writer = os.Stdout
	}

	// Use the formatter to output the data
	return formatter.Format(writer, onCalls, userMap, now, until)
}
