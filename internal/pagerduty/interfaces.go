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

package pagerduty

import (
	"net/url"

	"github.com/jdcasey/myshift-go/internal/types"
)

// PagerDutyClient defines the interface for PagerDuty API operations.
// This interface enables easy mocking for tests and provides a clean
// abstraction over the PagerDuty REST API v2.
//
// All methods handle authentication automatically using the API token
// provided during client creation. Errors are wrapped with contextual
// information to aid in debugging API issues.
type PagerDutyClient interface {
	// FindUserByEmail searches for a user by their email address.
	// Returns the first user matching the provided email address.
	FindUserByEmail(email string) (*types.User, error)

	// GetUser retrieves a user by their PagerDuty user ID.
	// Returns detailed user information including name, email, and type.
	GetUser(userID string) (*types.User, error)

	// GetOnCalls retrieves on-call shifts based on the provided parameters.
	// Supports filtering by time range, users, schedules, and other criteria.
	// Automatically handles pagination to return all matching results.
	GetOnCalls(params url.Values) ([]types.OnCall, error)

	// CreateOverrides creates one or more schedule overrides for the specified schedule.
	// Each override temporarily assigns a different user to handle on-call duties
	// during the specified time period.
	CreateOverrides(scheduleID string, overrides []types.Override) error
}

// Ensure Client implements PagerDutyClient interface
var _ PagerDutyClient = (*Client)(nil)
