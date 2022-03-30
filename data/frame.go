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
	"encoding/json"
	"math"
	"sort"
)

var (
	_ json.Unmarshaler = (*Frame)(nil)

	_ json.Marshaler = Frame{}

	_ sort.Interface = rawDataFrame{}
)

// Series contain a map of timestamps in micro seconds since the epoch to
// a floating point value.
type Series map[Timestamp]float64

// Frame provides JSON encoding and decoding for a map of series identified
// by an arbitrary key.
type Frame map[string]Series

// ordered returns a valid and ordered RawDataFrame with duplicated entries
// removed.
func (df Frame) ordered() rawDataFrame {
	times := map[Timestamp]struct{}{}
	for _, series := range df {
		for ts := range series {
			times[ts] = struct{}{}
		}
	}

	ordered := make([]Timestamp, 0, len(times))
	for ts := range times {
		ordered = append(ordered, ts)
	}
	sort.Slice(ordered, func(i, j int) bool {
		return ordered[i] < ordered[j]
	})

	out := rawDataFrame{
		Times:  ordered,
		Series: make(map[string][]Number, len(df)),
	}
	for sid, series := range df {
		values := make([]Number, 0, len(series))
		for _, ts := range out.Times {
			f, ok := series[ts]
			switch ok {
			case false:
				values = append(values, Number(math.NaN()))
			default:
				values = append(values, Number(f))
			}
		}
		out.Series[sid] = values
	}
	return out
}

func (df Frame) MarshalJSON() ([]byte, error) {
	return json.Marshal(df.ordered())
}

func (df *Frame) UnmarshalJSON(b []byte) error {
	in := rawDataFrame{
		Series: make(map[string][]Number),
	}

	if err := json.Unmarshal(b, &in); err != nil {
		return err
	}

	*df = in.DataFrame()
	return nil
}

// rawDataFrame describes a data frame that isn't necessarily valid or ordered.
// Series can have different length, and there can be multiple instances of the
// same time.
type rawDataFrame struct {
	Times  []Timestamp         `json:"times"`
	Series map[string][]Number `json:"series"`
}

func (raw rawDataFrame) Len() int {
	return len(raw.Times)
}

func (raw rawDataFrame) Less(i, j int) bool {
	return raw.Times[i] < raw.Times[j]
}

func (raw rawDataFrame) Swap(i, j int) {
	raw.Times[i], raw.Times[j] = raw.Times[j], raw.Times[i]
	for _, series := range raw.Series {
		series[i], series[j] = series[j], series[i]
	}
}

// DataFrame converts the raw data into a map of series. This method does not
// validate the raw data frame, but converts it at best effort. Specifically, if
// a series contain more values than we have timestamps, the additional values
// are dropped. Likewise, if a series contain fewer samples than timestamps,
// that's fine as far as the conversion is concerned.
func (raw rawDataFrame) DataFrame() Frame {
	out := make(Frame, len(raw.Series))
	for sid, values := range raw.Series {
		l := len(values)
		if len(values) > len(raw.Times) {
			l = len(raw.Times)
		}
		series := make(Series, l)
		for i := 0; i < l; i++ {
			f := float64(values[i])
			if !math.IsNaN(f) {
				series[raw.Times[i]] = f
			}
		}
		out[sid] = series
	}

	return out
}
