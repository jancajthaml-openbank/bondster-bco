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

package utils

import (
  "time"
)

func GetMonthsWithin(startDate time.Time, endDate time.Time) []time.Time {
  dates := make([]time.Time, 0)
  for ; startDate.Before(endDate); startDate = startDate.AddDate(0, 1, 0) {
    date := startDate.AddDate(0, 1, 0).Add(time.Nanosecond*-1)
    date = time.Date(date.Year(), date.Month(), 1, 0, 0, 0, 0, time.UTC)
    dates = append(dates, date)
  }
  return dates
}
