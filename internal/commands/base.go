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
	"fmt"
	"io"
	"net/url"
	"os"
	"time"

	"github.com/jdcasey/myshift-go/internal/pagerduty"
	"github.com/jdcasey/myshift-go/internal/types"
)

// BaseCommand provides common functionality for all commands
type BaseCommand struct {
	client pagerduty.PagerDutyClient
	config *types.Config
	writer io.Writer
}

// NewBaseCommand creates a new base command with shared dependencies
func NewBaseCommand(client pagerduty.PagerDutyClient, config *types.Config, writer io.Writer) *BaseCommand {
	if writer == nil {
		writer = os.Stdout
	}
	return &BaseCommand{
		client: client,
		config: config,
		writer: writer,
	}
}

// ResolveUser resolves a user email, falling back to config if empty
func (b *BaseCommand) ResolveUser(email string) (*types.User, error) {
	if email == "" {
		email = b.config.MyUser
	}
	if email == "" {
		return nil, fmt.Errorf("user email is required (use --user flag or set my_user in config)")
	}
	return b.client.FindUserByEmail(email)
}

// GetScheduleID returns the schedule ID, ensuring it's configured
func (b *BaseCommand) GetScheduleID() (string, error) {
	if b.config.ScheduleID == "" {
		return "", fmt.Errorf("schedule_id must be configured")
	}
	return b.config.ScheduleID, nil
}

// BuildTimeRangeParams builds URL parameters for time range queries
func (b *BaseCommand) BuildTimeRangeParams(start, end time.Time) url.Values {
	return url.Values{
		"since":    []string{start.Format(time.RFC3339)},
		"until":    []string{end.Format(time.RFC3339)},
		"overflow": []string{"true"},
	}
}

// GetOnCallsForUser fetches on-call shifts for a specific user
func (b *BaseCommand) GetOnCallsForUser(scheduleID string, userID string, start, end time.Time) ([]types.OnCall, error) {
	params := b.BuildTimeRangeParams(start, end)
	params.Add("user_ids[]", userID)
	params.Add("schedule_ids[]", scheduleID)

	onCalls, err := b.client.GetOnCalls(params)
	if err != nil {
		return nil, fmt.Errorf("error fetching shifts: %w", err)
	}

	return DeduplicateOnCalls(onCalls), nil
}

// GetOnCallsForSchedule fetches all on-call shifts for a schedule
func (b *BaseCommand) GetOnCallsForSchedule(scheduleID string, start, end time.Time) ([]types.OnCall, error) {
	params := b.BuildTimeRangeParams(start, end)
	params.Add("schedule_ids[]", scheduleID)

	onCalls, err := b.client.GetOnCalls(params)
	if err != nil {
		return nil, fmt.Errorf("error fetching shifts: %w", err)
	}

	return DeduplicateOnCalls(onCalls), nil
}

// BuildUserMap creates a map of user IDs to names from on-call shifts
func (b *BaseCommand) BuildUserMap(onCalls []types.OnCall) map[string]string {
	userMap := make(map[string]string)
	for _, shift := range onCalls {
		if _, exists := userMap[shift.User.ID]; !exists {
			user, err := b.client.GetUser(shift.User.ID)
			if err != nil {
				// If we can't get user details, use what we have
				userMap[shift.User.ID] = shift.User.Name
			} else {
				userMap[shift.User.ID] = user.Name
			}
		}
	}
	return userMap
}
