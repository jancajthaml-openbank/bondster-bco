// Copyright (c) 2016-2020, Jan Cajthaml <jan.cajthaml@gmail.com>
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

package timeshift

import "time"

// TimeRange represents segment of time
type TimeRange struct {
	StartTime time.Time
	EndTime   time.Time
}

func (value *TimeRange) String() string {
	if value == nil {
		return "<nil>"
	}
	return value.StartTime.Format(time.RFC3339) + " - " + value.EndTime.Format(time.RFC3339)
}

// SliceByMonths slices time range by months
func SliceByMonths(startDate time.Time, endDate time.Time) []TimeRange {
	dates := make([]TimeRange, 0)
	current := time.Date(startDate.Year(), startDate.Month(), 1, 0, 0, 0, 0, time.UTC)
	for current.Before(endDate) {
		date := current.AddDate(0, 1, 0).AddDate(0, 0, -1)
		dates = append(dates, TimeRange{
			StartTime: time.Date(current.Year(), current.Month(), 1, 0, 0, 0, 0, time.UTC),
			EndTime:   time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC),
		})
		current = date.AddDate(0, 0, 1)
	}
	return dates
}

// PartitionInterval partitions time range for optimal bondster search
func PartitionInterval(startDate time.Time, endDate time.Time) []TimeRange {
	timeline := SliceByMonths(startDate, endDate)
	if len(timeline) == 0 {
		return timeline
	}
	timeline[len(timeline)-1].EndTime = endDate
	timeline[0].StartTime = startDate
	if len(timeline) == 0 && timeline[0].StartTime == timeline[0].EndTime {
		return make([]TimeRange, 0)
	}
	return timeline
}
