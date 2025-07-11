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
	"strconv"
	"time"
)

// ParamsBuilder provides a fluent interface for building PagerDuty API parameters
type ParamsBuilder struct {
	params url.Values
}

// NewParamsBuilder creates a new parameter builder
func NewParamsBuilder() *ParamsBuilder {
	return &ParamsBuilder{
		params: make(url.Values),
	}
}

// TimeRange sets the time range parameters
func (p *ParamsBuilder) TimeRange(start, end time.Time) *ParamsBuilder {
	p.params.Set("since", start.Format(time.RFC3339))
	p.params.Set("until", end.Format(time.RFC3339))
	return p
}

// Users adds user ID parameters
func (p *ParamsBuilder) Users(userIDs ...string) *ParamsBuilder {
	for _, id := range userIDs {
		p.params.Add("user_ids[]", id)
	}
	return p
}

// Schedules adds schedule ID parameters
func (p *ParamsBuilder) Schedules(scheduleIDs ...string) *ParamsBuilder {
	for _, id := range scheduleIDs {
		p.params.Add("schedule_ids[]", id)
	}
	return p
}

// Overflow sets the overflow parameter
func (p *ParamsBuilder) Overflow(overflow bool) *ParamsBuilder {
	p.params.Set("overflow", strconv.FormatBool(overflow))
	return p
}

// Query sets the query parameter
func (p *ParamsBuilder) Query(query string) *ParamsBuilder {
	p.params.Set("query", query)
	return p
}

// Limit sets the limit parameter
func (p *ParamsBuilder) Limit(limit int) *ParamsBuilder {
	p.params.Set("limit", strconv.Itoa(limit))
	return p
}

// Offset sets the offset parameter
func (p *ParamsBuilder) Offset(offset int) *ParamsBuilder {
	p.params.Set("offset", strconv.Itoa(offset))
	return p
}

// Build returns the constructed URL values
func (p *ParamsBuilder) Build() url.Values {
	return p.params
}
