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

package fields

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

const (
	patternYearToFraction = `^(?P<sign>-)?P((?P<years>\d+)Y)?((?P<months>\d+)M)?((?P<weeks>\d+)W)?((?P<days>\d+)D)?(T((?P<hours>\d+)H)?((?P<minutes>\d+)M)?((?P<fractions>\d+(\.\d+)?)S)?)?$`
	patternWeekToFraction = `^(?P<sign>-)?P((?P<weeks>\d+)W)?((?P<days>\d+)D)?(T((?P<hours>\d+)H)?((?P<minutes>\d+)M)?((?P<fractions>\d+(\.\d+)?)S)?)?$`
)

var (
	reYearToFraction = regexp.MustCompile(patternYearToFraction)
	reWeekToFraction = regexp.MustCompile(patternWeekToFraction)
)

// CalendarDurationNullZero is a variant of CalendarDuration that JSON encodes
// the zero-value to null.
type CalendarDurationNullZero CalendarDuration

var (
	_ fmt.Stringer     = CalendarDurationNullZero{}
	_ json.Marshaler   = CalendarDurationNullZero{}
	_ json.Unmarshaler = (*CalendarDurationNullZero)(nil)
)

func (cd CalendarDurationNullZero) IsZero() bool {
	return CalendarDuration(cd).IsZero()
}

func (cd CalendarDurationNullZero) String() string {
	if cd.IsZero() {
		return ""
	}
	return CalendarDuration(cd).String()
}

func (cd *CalendarDurationNullZero) UnmarshalJSON(data []byte) error {
	if bytes.Equal(data, []byte(`null`)) {
		cd.Duration = 0
		cd.Months = 0
		return nil
	}
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	_cd, ok := parseYearToFraction(s)
	if !ok {
		return ErrBadCalendarDuration
	}

	*cd = CalendarDurationNullZero(_cd)
	return nil
}

func (cd CalendarDurationNullZero) MarshalJSON() ([]byte, error) {
	if cd.Duration == 0 && cd.Months == 0 {
		return []byte(`null`), nil
	}
	s, err := formatCalendarDuration(CalendarDuration(cd))
	if err != nil {
		return nil, err
	}
	return json.Marshal(s)
}

// CalendarDuration allows encoding either a fixed duration or a monthly
// duration. Setting both a month and a duration value is regarded as an error, and
// will fail to encoded.
type CalendarDuration struct {
	Months   int
	Duration time.Duration
}

var (
	_ fmt.Stringer             = CalendarDuration{}
	_ encoding.TextMarshaler   = CalendarDuration{}
	_ encoding.TextUnmarshaler = (*CalendarDuration)(nil)
)

// MonthDuration returns a duration that spans a given number of months.
func MonthDuration(m int) CalendarDuration {
	return CalendarDuration{Months: m}
}

func (cd CalendarDuration) IsZero() bool {
	return cd.Duration == 0 && cd.Months == 0
}

func (cd CalendarDuration) String() string {
	res, _ := formatCalendarDuration(cd)
	return res
}

func (dd CalendarDuration) MarshalText() ([]byte, error) {
	s, err := formatCalendarDuration(dd)
	if err != nil {
		return nil, err
	}
	return []byte(s), nil
}

func (cd *CalendarDuration) UnmarshalText(b []byte) error {
	_cd, ok := parseYearToFraction(string(b))
	if !ok {
		return fmt.Errorf("json: %w", ErrBadCalendarDuration)
	}
	cd.Months = _cd.Months
	cd.Duration = _cd.Duration
	return nil
}

// ParseCalendarDuration converts text-encoded RFC 3339 duration to its
// CalendarDuration representation.
func ParseCalendarDuration(s string) (CalendarDuration, error) {
	dd, ok := parseYearToFraction(s)
	if !ok {
		return CalendarDuration{}, ErrBadCalendarDuration
	}
	return dd, nil
}

func formatCalendarDuration(dd CalendarDuration) (string, error) {
	var s string
	switch {
	case dd.Months != 0 && dd.Duration != 0:
		return "", fmt.Errorf("can't specify both months and duration")
	case dd.Months != 0:
		m := dd.Months
		if m < 0 {
			s = "-P"
			m = -m
		} else {
			s = "P"
		}
		if y := m / 12; y > 0 {
			s += fmt.Sprintf("%dY", y)
			m %= 12
		}
		if m > 0 {
			s += fmt.Sprintf("%dM", m)
		}
	case dd.Duration != 0:
		s = formatFixedDuration(dd.Duration)
	default:
		s = "PT0S"
	}

	return s, nil
}

func parseYearToFraction(s string) (CalendarDuration, bool) {
	var err error
	var di int64
	var df float64
	var dd CalendarDuration
	sign := 1

	matches := reYearToFraction.FindStringSubmatch(strings.ToUpper(s))
	if matches == nil {
		return dd, false
	}
	for i, name := range reYearToFraction.SubexpNames() {
		if matches[i] == "" || name == "" {
			continue
		}
		switch name {
		case "sign":
			sign = -1
		case "years":
			di, err = strconv.ParseInt(matches[i], 10, 64)
			dd.Months += 12 * int(di)
		case "months":
			di, err = strconv.ParseInt(matches[i], 10, 64)
			dd.Months += int(di)
		case "weeks":
			di, err = strconv.ParseInt(matches[i], 10, 64)
			dd.Duration += time.Duration(di) * 7 * 24 * time.Hour
		case "days":
			di, err = strconv.ParseInt(matches[i], 10, 64)
			dd.Duration += time.Duration(di) * 24 * time.Hour
		case "hours":
			di, err = strconv.ParseInt(matches[i], 10, 64)
			dd.Duration += time.Duration(di) * time.Hour
		case "minutes":
			di, err = strconv.ParseInt(matches[i], 10, 64)
			dd.Duration += time.Duration(di) * time.Minute
		case "fractions":
			df, err = strconv.ParseFloat(matches[i], 64)
			dd.Duration += time.Duration(df * float64(time.Second))
		}
		if err != nil {
			// If this happens, it's a programming error that must be corrected;
			// regex should validate the format for matches.
			panic(fmt.Errorf("%s: %s", name, err))
		}
	}
	if dd.IsZero() {
		return dd, false
	}

	dd.Duration *= time.Duration(sign)
	dd.Months *= sign
	return dd, true
}

// FixedDurationNullZero is a variant of FixedDuration that JSON encodes the
// zero-value as null.
type FixedDurationNullZero FixedDuration

var (
	_ json.Marshaler   = FixedDurationNullZero{}
	_ json.Unmarshaler = (*FixedDurationNullZero)(nil)
)

// AsFixedDurationNullZero converts d to a FixedDurationNullZero instance.
func AsFixedDurationNullZero(d time.Duration) FixedDurationNullZero {
	return FixedDurationNullZero{Duration: d}
}

func (d FixedDurationNullZero) MarshalJSON() ([]byte, error) {
	if d.Duration == 0 {
		return []byte(`null`), nil
	}
	return json.Marshal(formatFixedDuration(d.Duration))
}

func (d *FixedDurationNullZero) UnmarshalJSON(b []byte) error {
	if bytes.Equal(b, []byte(`null`)) {
		d.Duration = 0
		return nil
	}

	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	_d, ok := parseWeekToFraction(s)
	if !ok {
		return ErrBadFixedDuration
	}

	d.Duration = _d
	return nil
}

// FixedDuration wraps a time.Duration so that it's JSON encoded as an RFC 3339
// duration string.
type FixedDuration struct {
	time.Duration
}

var (
	_ fmt.Stringer             = FixedDuration{}
	_ encoding.TextMarshaler   = FixedDuration{}
	_ encoding.TextUnmarshaler = (*FixedDuration)(nil)
)

// AsFixedDuration converts d to a FixedDuration instance.
func AsFixedDuration(d time.Duration) FixedDuration {
	return FixedDuration{Duration: d}
}

func (d FixedDuration) String() string {
	return formatFixedDuration(d.Duration)
}

func (d FixedDuration) MarshalText() ([]byte, error) {
	return []byte(formatFixedDuration(d.Duration)), nil
}

// ParseFixedDuration parses a RFC 3339 string accepting weeks, days, hours,
// minute, seconds and fractions.
func ParseFixedDuration(s string) (FixedDurationNullZero, error) {
	d, ok := parseWeekToFraction(s)
	if !ok {
		return FixedDurationNullZero{}, ErrBadFixedDuration
	}
	return FixedDurationNullZero{Duration: d}, nil
}

func (d *FixedDuration) UnmarshalText(b []byte) error {
	_d, ok := parseWeekToFraction(string(b))
	if !ok {
		return ErrBadFixedDuration
	}

	d.Duration = _d
	return nil
}

func (d FixedDurationNullZero) String() string {
	return formatFixedDuration(d.Duration)
}

func formatFixedDuration(d time.Duration) string {
	s := "PT"

	if d < 0 {
		d = -d
		s = "-PT"
	}

	if hour := d / time.Hour; hour > 0 {
		s += fmt.Sprintf("%dH", hour)
		d %= time.Hour
	}
	if min := d / time.Minute; min > 0 {
		s += fmt.Sprintf("%dM", min)
		d %= time.Minute
	}
	if fraction := float64(d) / float64(time.Second); fraction > 0 {
		// FormatFloat with -1 as precision to return a string without trainling
		// zeros after the decimal point. E.g. 0.5S, not 0.500000S
		s += strconv.FormatFloat(fraction, 'f', -1, 64) + "S"
	}

	return s
}

func parseWeekToFraction(s string) (time.Duration, bool) {
	var err error
	var di int64
	var df float64
	var d time.Duration
	sign := time.Duration(1)

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
			d += time.Duration(di) * 7 * 24 * time.Hour
		case "days":
			di, err = strconv.ParseInt(matches[i], 10, 64)
			d += time.Duration(di) * 24 * time.Hour
		case "hours":
			di, err = strconv.ParseInt(matches[i], 10, 64)
			d += time.Duration(di) * time.Hour
		case "minutes":
			di, err = strconv.ParseInt(matches[i], 10, 64)
			d += time.Duration(di) * time.Minute
		case "fractions":
			df, err = strconv.ParseFloat(matches[i], 64)
			d += time.Duration(df * float64(time.Second))
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
