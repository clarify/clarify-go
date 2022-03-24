// Copyright 2022 Searis AS
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

package data

import (
	"encoding"
	"time"
)

// Timestamp provides a hashable and comparable alternative to time.Time, stored
// as microseconds since the epoch.
type Timestamp int64

var (
	_ encoding.TextMarshaler   = Timestamp(0)
	_ encoding.TextUnmarshaler = (*Timestamp)(nil)
)

const (
	// OriginTime defines midnight of the first Monday of year 2000 in the
	// UTC time-zone (2000-01-03T00:00:00Z) as microseconds since the epoch.
	OriginTime Timestamp = 946857600000000
)

// Truncate returns the result of rounding ts down to a multiple of d (since
// OriginTime). Note that this is not fully equivalent to using Truncate on the
// time.Time type, as we are deliberately using a different origin.
func (ts Timestamp) Truncate(d time.Duration) Timestamp {
	if d == 0 {
		return ts
	}
	r := (ts - OriginTime) % Timestamp(d)
	return ts - r
}

// Add adds the fixed duration to the time-stamp.
func (ts Timestamp) Add(d time.Duration) Timestamp {
	td := Timestamp(d) / 1e3
	return ts + td
}

// Sub returns the differences between ts and ts2 as a fixed duration.
func (ts Timestamp) Sub(ts2 Timestamp) time.Duration {
	return time.Duration(ts-ts2) * 1e3
}

// AsTimestamp converts a time.Time to Timestamp.
func AsTimestamp(t time.Time) Timestamp {
	return Timestamp(t.UnixMicro())
}

// Time returns the Timestamp as time.Time.
func (ts Timestamp) Time() time.Time {
	return time.UnixMicro(int64(ts)).UTC()
}

func (ts Timestamp) MarshalText() ([]byte, error) {
	return ts.Time().MarshalText()
}

func (ts *Timestamp) UnmarshalText(data []byte) error {
	var t time.Time
	err := t.UnmarshalText(data)
	if err != nil {
		return err
	}
	tmp := Timestamp(t.UnixMicro())
	*ts = tmp
	return nil
}
