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
)

// NextCommand handles the "next" command functionality.
type NextCommand struct {
	client pagerduty.PagerDutyClient
}

// NewNextCommand creates a new NextCommand instance.
func NewNextCommand(client pagerduty.PagerDutyClient) *NextCommand {
	return &NextCommand{client: client}
}

// Execute runs the next command to show the next on-call shift for a user.
func (n *NextCommand) Execute(scheduleID, userEmail string, days int) error {
	if userEmail == "" {
		return fmt.Errorf("user email is required")
	}

	// Find user by email
	user, err := n.client.FindUserByEmail(userEmail)
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

	onCalls, err := n.client.GetOnCalls(params)
	if err != nil {
		return fmt.Errorf("error fetching shifts: %w", err)
	}

	if len(onCalls) == 0 {
		fmt.Println("No upcoming shifts found")
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
		fmt.Println("Currently on call")
		fmt.Printf("Shift ends: %s\n", nextShift.End.Format("2006-01-02 15:04 MST"))
	} else if nextShift.Start.After(now) {
		fmt.Println("Next shift:")
		fmt.Printf("Starts: %s\n", nextShift.Start.Format("2006-01-02 15:04 MST"))
		fmt.Printf("Ends: %s\n", nextShift.End.Format("2006-01-02 15:04 MST"))
	} else {
		fmt.Println("No upcoming shifts found")
	}

	return nil
}
