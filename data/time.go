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
	"bytes"
	"encoding"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
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
func (ts Timestamp) Truncate(d FixedDuration) Timestamp {
	if d == 0 {
		return ts
	}
	r := (ts - OriginTime) % Timestamp(d)
	return ts - r
}

// Add adds the fixed duration to the time-stamp.
func (ts Timestamp) Add(d FixedDuration) Timestamp {
	return ts + Timestamp(d)
}

// Sub returns the differences between ts and ts2 as a fixed duration.
func (ts Timestamp) Sub(ts2 Timestamp) FixedDuration {
	return FixedDuration(ts - ts2)
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

// FixedDuration holds microseconds and encode them as a RFC 3339 compatible
// fixed duration string. The zero-value is encoded as JSON null.
type FixedDuration int64

var (
	_ json.Marshaler   = FixedDuration(0)
	_ json.Unmarshaler = (*FixedDuration)(nil)
)

const (
	null                  = `null`
	patternWeekToFraction = `^(?P<sign>-)?P((?P<weeks>\d+)W)?((?P<days>\d+)D)?(T((?P<hours>\d+)H)?((?P<minutes>\d+)M)?((?P<fractions>\d+(\.\d+)?)S)?)?$`
)

var reWeekToFraction = regexp.MustCompile(patternWeekToFraction)

// Constants for timestamp durations.
const (
	Microsecond FixedDuration = FixedDuration(time.Microsecond / 1e3)
	Millisecond FixedDuration = FixedDuration(time.Millisecond / 1e3)
	Second      FixedDuration = FixedDuration(time.Second / 1e3)
	Minute      FixedDuration = FixedDuration(time.Minute / 1e3)
	Hour        FixedDuration = FixedDuration(time.Hour / 1e3)
)

// AsFixedDuration converts d to a FixedDuration.
func AsFixedDuration(d time.Duration) FixedDuration {
	return FixedDuration(d.Microseconds())
}

// Duration returns d as a time.Duration.
func (d FixedDuration) Duration() time.Duration {
	return time.Duration(d) * time.Microsecond
}

// ParseFixedDuration parses a RFC 3339 string accepting weeks, days, hours,
// minute, seconds and fractions.
func ParseFixedDuration(s string) (FixedDuration, error) {
	d, ok := parseWeekToFraction(s)
	if !ok {
		return 0, ErrBadFixedDuration
	}
	return d, nil
}

// UnmarshalJSON implements json.Unmarshaler.
func (d *FixedDuration) UnmarshalJSON(b []byte) error {
	if bytes.Equal(b, []byte(`null`)) {
		*d = 0
		return nil
	}
	var s string
	err := json.Unmarshal(b, &s)
	if err != nil {
		return err
	}

	_d, ok := parseWeekToFraction(s)
	if !ok {
		return fmt.Errorf("json: %w", ErrBadFixedDuration)
	}

	*d = _d
	return nil
}

func (d FixedDuration) String() string {
	return formatFixedDuration(d)
}

// MarshalJSON implements json.Marshaler.
func (d FixedDuration) MarshalJSON() ([]byte, error) {
	if d == 0 {
		return []byte(null), nil
	}
	return json.Marshal(formatFixedDuration(d))
}

func formatFixedDuration(d FixedDuration) string {
	s := "PT"
	if d < 0 {
		d = -d
		s = "-PT"
	}

	if hour := d / Hour; hour > 0 {
		s += fmt.Sprintf("%dH", hour)
		d %= Hour
	}
	if min := d / Minute; min > 0 {
		s += fmt.Sprintf("%dM", min)
		d %= Minute
	}
	if fraction := float64(d) / float64(Second); fraction > 0 {
		// FormatFloat with -1 as precision to return a string without trainling
		// zeros after the decimal point. E.g. 0.5S, not 0.500000S
		s += strconv.FormatFloat(fraction, 'f', -1, 64) + "S"
	}

	return s
}

func parseWeekToFraction(s string) (FixedDuration, bool) {
	var err error
	var di int64
	var df float64
	var d FixedDuration
	sign := FixedDuration(1)

	matches := reWeekToFraction.FindStringSubmatch(strings.ToUpper(s))
	if matches == nil {
		return 0, false
	}
	for i, name := range reWeekToFraction.SubexpNames() {
		if matches[i] == "" || name == "" {
			continue
		}
		switch name {
		case "sign":
			sign = -1
		case "weeks":
			di, err = strconv.ParseInt(matches[i], 10, 64)
			d += FixedDuration(di) * 7 * 24 * Hour
		case "days":
			di, err = strconv.ParseInt(matches[i], 10, 64)
			d += FixedDuration(di) * 24 * Hour
		case "hours":
			di, err = strconv.ParseInt(matches[i], 10, 64)
			d += FixedDuration(di) * Hour
		case "minutes":
			di, err = strconv.ParseInt(matches[i], 10, 64)
			d += FixedDuration(di) * Minute
		case "fractions":
			df, err = strconv.ParseFloat(matches[i], 64)
			d += FixedDuration(df * float64(Second))
		}
		if err != nil {
			// If this happens, it's a programming error that must be corrected;
			// regex should validate the format for matches.
			panic(fmt.Errorf("%s: %s", name, err))
		}
	}
	d *= sign
	return d, true
}
