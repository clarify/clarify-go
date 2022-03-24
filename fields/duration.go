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

package fields

import (
	"bytes"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// FixedDuration wraps a time.Duration so that it's JSON encoded as an RFC 3339
// duration string. The zero-value is encoded as null.
type FixedDuration struct {
	time.Duration
}

var (
	_ interface {
		json.Marshaler
		fmt.Stringer
	} = FixedDuration{}
	_ json.Unmarshaler = (*FixedDuration)(nil)
)

const (
	patternWeekToFraction = `^(?P<sign>-)?P((?P<weeks>\d+)W)?((?P<days>\d+)D)?(T((?P<hours>\d+)H)?((?P<minutes>\d+)M)?((?P<fractions>\d+(\.\d+)?)S)?)?$`
)

var reWeekToFraction = regexp.MustCompile(patternWeekToFraction)

// AsFixedDuration converts d to a FixedDuration.
func AsFixedDuration(d time.Duration) FixedDuration {
	return FixedDuration{Duration: d}
}

// ParseFixedDuration parses a RFC 3339 string accepting weeks, days, hours,
// minute, seconds and fractions.
func ParseFixedDuration(s string) (FixedDuration, error) {
	d, ok := parseWeekToFraction(s)
	if !ok {
		return FixedDuration{}, ErrBadFixedDuration
	}
	return FixedDuration{Duration: d}, nil
}

// UnmarshalJSON implements json.Unmarshaler.
func (d *FixedDuration) UnmarshalJSON(b []byte) error {
	if bytes.Equal(b, []byte(`null`)) {
		d.Duration = 0
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

	d.Duration = _d
	return nil
}

func (d FixedDuration) String() string {
	return formatFixedDuration(d.Duration)
}

// MarshalJSON implements json.Marshaler.
func (d FixedDuration) MarshalJSON() ([]byte, error) {
	if d.Duration == 0 {
		return []byte(`null`), nil
	}
	return json.Marshal(formatFixedDuration(d.Duration))
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
