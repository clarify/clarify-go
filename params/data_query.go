// Copyright 2022-2023 Searis AS
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

package params

import (
	"encoding/json"
	"time"

	"github.com/clarify/clarify-go/fields"
)

type dataQuery struct {
	Filter         DataFilter `json:"filter"`
	Rollup         string     `json:"rollup,omitempty"`
	Last           int        `json:"last,omitempty"`
	Origin         string     `json:"origin,omitempty"`
	FirstDayOfWeek int        `json:"firstDayOfWeek,omitempty"`
	TimeZone       string     `json:"timeZone,omitempty"`
}

// DataQuery holds a data params. Although it does not expose any fields, the
// type can be decoded from and encoded to JSON.
type DataQuery struct {
	query dataQuery
}

// Data returns a new DataQuery that joins passed in filters with logical AND.
func Data() DataQuery {
	return DataQuery{}
}

func (q *DataQuery) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &q.query)
}

func (q DataQuery) MarshalJSON() ([]byte, error) {
	return json.Marshal(q.query)
}

// Origin returns a new data query with a custom rollup bucket origin. The
// origin is used by DurationRollup and MonthRollup. This setting takes
// precedence over the firstDayOfWeek setting passed to DurationRollup.
func (dq DataQuery) Origin(o time.Time) DataQuery {
	dq.query.Origin = o.Format(time.RFC3339Nano)
	return dq
}

// RollupWindow returns a new data query with a window based rollup.
func (dq DataQuery) RollupWindow() DataQuery {
	dq.query.Rollup = "window"
	return dq
}

// RollupDuration returns a new data query with a fixed duration bucket rollup.
//
// The default bucket origin is set to time 00:00:00 according to the query
// time-zone for the first date in 2000 where the weekday matches the
// firstDayOfWeek parameter.
func (dq DataQuery) RollupDuration(d time.Duration, firstDayOfWeek time.Weekday) DataQuery {
	dq.query.Rollup = fields.AsFixedDurationNullZero(d).String()
	isoDay := int(firstDayOfWeek) % 7
	if isoDay == 0 {
		isoDay = 7
	}
	dq.query.FirstDayOfWeek = isoDay
	return dq
}

// RollupMonths returns a new data query with a calendar month bucket rollup.
//
// The default bucket origin is set to time 00:00:00 according to the query
// time-zone for January 1 year 2000.
func (dq DataQuery) RollupMonths(months int) DataQuery {
	dq.query.Rollup = fields.CalendarDurationNullZero{Months: months}.String()
	return dq
}

// TimeZoneLocation returns a new data query with the time-zone set to TZ
// database name of the passed in loc.
//
// The method is equivalent to dq.TimeZone(loc.String()).
func (dq DataQuery) TimeZoneLocation(loc *time.Location) DataQuery {
	dq.query.TimeZone = loc.String() // nil values return "UTC".
	return dq
}

// TimeZone returns a new data query with TimeZone set to name. The name should
// contain a valid TZ Database reference, such as "UTC", "Europe/Berlin" or
// "America/New_York". The default value is "UTC".
//
// See https://en.wikipedia.org/wiki/List_of_tz_database_time_zones for
// available values.
//
// The time zone of a data query affects how the rollup bucket origin is aligned
// when there is no custom origin provided. If the time-zone location includes
// daylight saving time adjustments (DST), then resulting bucket times are
// adjusted according to local clock times if the bucket width is above the DST
// adjustment offset (normally 1 hour).
func (dq DataQuery) TimeZone(name string) DataQuery {
	dq.query.TimeZone = name
	return dq
}

// Where returns a new data query which joins the passed in filter conditions
// with existing filer conditions using logical and.
func (dq DataQuery) Where(filter DataFilter) DataQuery {
	dq.query.Filter = DataAnd(dq.query.Filter, filter)
	return dq
}

// Last returns a new data query where only the last n non-empty data-points per
// series that match the query is included. If n is <= 0, no limit is applied.
func (dq DataQuery) Last(n int) DataQuery {
	dq.query.Last = n
	return dq
}
