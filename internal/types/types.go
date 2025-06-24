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

// Package types provides shared types and utilities for the myshift-go CLI tool.
package types

import "time"

// User represents a PagerDuty user object.
type User struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Type  string `json:"type"`
}

// UserReference represents a reference to a PagerDuty user.
type UserReference struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}

// Schedule represents a PagerDuty schedule.
type Schedule struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	TimeZone    string `json:"time_zone"`
}

// OnCall represents an on-call shift from the PagerDuty API.
type OnCall struct {
	Start    time.Time `json:"start"`
	End      time.Time `json:"end"`
	User     User      `json:"user"`
	Schedule Schedule  `json:"schedule"`
}

// Override represents a schedule override.
type Override struct {
	Start    time.Time     `json:"start"`
	End      time.Time     `json:"end"`
	User     UserReference `json:"user"`
	TimeZone string        `json:"time_zone,omitempty"`
}

// Shift represents a simplified shift with local timezone conversion.
type Shift struct {
	Start  time.Time
	End    time.Time
	UserID string
}

// Config represents the application configuration.
type Config struct {
	PagerDutyToken string `yaml:"pagerduty_token"`
	ScheduleID     string `yaml:"schedule_id,omitempty"`
	MyUser         string `yaml:"my_user,omitempty"`
}

// Version represents the application version.
const Version = "0.1.0"
