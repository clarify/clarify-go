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

package views

import (
	"encoding/json"
	"math"
	"sort"

	"github.com/clarify/clarify-go/fields"
	"golang.org/x/exp/slices"
)

var (
	_ json.Unmarshaler = (*DataFrame)(nil)

	_ json.Marshaler = DataFrame{}

	_ sort.Interface = RawDataFrame{}
)

// DataSeries contain a map of timestamps in micro seconds since the epoch to
// a floating point value.
type DataSeries map[fields.Timestamp]float64

// DataFrame provides JSON encoding and decoding for a map of series identified
// by an arbitrary key.
type DataFrame map[string]DataSeries

// Ordered returns a valid and Ordered RawDataFrame with duplicated entries
// removed.
func (df DataFrame) Ordered() RawDataFrame {
	times := map[fields.Timestamp]struct{}{}
	for _, series := range df {
		for ts := range series {
			times[ts] = struct{}{}
		}
	}

	ordered := make([]fields.Timestamp, 0, len(times))
	for ts := range times {
		ordered = append(ordered, ts)
	}
	slices.Sort(ordered)

	out := RawDataFrame{
		Times:  ordered,
		Series: make(map[string][]fields.Number, len(df)),
	}
	for sid, series := range df {
		values := make([]fields.Number, 0, len(series))
		for _, ts := range out.Times {
			f, ok := series[ts]
			switch ok {
			case false:
				values = append(values, fields.Number(math.NaN()))
			default:
				values = append(values, fields.Number(f))
			}
		}
		out.Series[sid] = values
	}
	return out
}

func (df DataFrame) MarshalJSON() ([]byte, error) {
	return json.Marshal(df.Ordered())
}

func (df *DataFrame) UnmarshalJSON(b []byte) error {
	in := RawDataFrame{
		Series: make(map[string][]fields.Number),
	}

	if err := json.Unmarshal(b, &in); err != nil {
		return err
	}

	*df = in.DataFrame()
	return nil
}

// RawDataFrame describes a data frame that isn't necessarily valid or ordered.
// Series can have different length, and there can be multiple instances of the
// same time.
type RawDataFrame struct {
	Times  []fields.Timestamp         `json:"times"`
	Series map[string][]fields.Number `json:"series"`
}

func (raw RawDataFrame) Len() int {
	return len(raw.Times)
}

func (raw RawDataFrame) Less(i, j int) bool {
	return raw.Times[i] < raw.Times[j]
}

func (raw RawDataFrame) Swap(i, j int) {
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
func (raw RawDataFrame) DataFrame() DataFrame {
	out := make(DataFrame, len(raw.Series))
	for sid, values := range raw.Series {
		l := len(values)
		if len(values) > len(raw.Times) {
			l = len(raw.Times)
		}
		series := make(DataSeries, l)
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
