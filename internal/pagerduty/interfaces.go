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

	"github.com/jdcasey/myshift-go/pkg/myshift"
)

// PagerDutyClient defines the interface for PagerDuty API operations.
// This enables easy mocking for tests.
type PagerDutyClient interface {
	FindUserByEmail(email string) (*myshift.User, error)
	GetUser(userID string) (*myshift.User, error)
	GetOnCalls(params url.Values) ([]myshift.OnCall, error)
	CreateOverrides(scheduleID string, overrides []myshift.Override) error
}

// Ensure Client implements PagerDutyClient interface
var _ PagerDutyClient = (*Client)(nil)
