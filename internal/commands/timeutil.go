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
	"time"
)

// CalculateTimeRange calculates start and end times from string inputs and days
func CalculateTimeRange(startDate, endDate string, days int) (time.Time, time.Time, error) {
	var start, end time.Time
	var err error

	if startDate != "" {
		start, err = time.Parse("2006-01-02", startDate)
		if err != nil {
			return time.Time{}, time.Time{}, fmt.Errorf("invalid start date: %w", err)
		}
	} else {
		start = time.Now()
	}

	if endDate != "" {
		end, err = time.Parse("2006-01-02", endDate)
		if err != nil {
			return time.Time{}, time.Time{}, fmt.Errorf("invalid end date: %w", err)
		}
	} else {
		end = start.AddDate(0, 0, days)
	}

	return start, end, nil
}

// ParseTimeRange parses start and end times with specific format
func ParseTimeRange(startTime, endTime string) (time.Time, time.Time, error) {
	start, err := time.Parse("2006-01-02 15:04", startTime)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid start time: %w", err)
	}

	end, err := time.Parse("2006-01-02 15:04", endTime)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid end time: %w", err)
	}

	return start, end, nil
}

// FormatTimeRange formats a time range for display
func FormatTimeRange(start, end time.Time) string {
	return fmt.Sprintf("%s to %s",
		start.Format("2006-01-02 15:04 MST"),
		end.Format("2006-01-02 15:04 MST"))
}
