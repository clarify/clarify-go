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

package views

import (
	"encoding/json"
	"math"
	"slices"
	"sort"

	"github.com/clarify/clarify-go/fields"
)

type DataFrameInclude struct {
	Items []Item
}

// DataSeries contain a map of timestamps in micro seconds since the epoch to
// a floating point value.
type DataSeries map[fields.Timestamp]float64

// Timestamps returns an ordered set of all timestamps in the data-series where
// there is at least one NaN value.
func (s DataSeries) Timestamps() []fields.Timestamp {
	ordered := make([]fields.Timestamp, 0, len(s))
	for t, v := range s {
		if math.IsNaN(v) {
			continue
		}
		ordered = append(ordered, t)
	}
	slices.Sort(ordered)
	return ordered
}

// DataFrame provides JSON encoding and decoding for a map of series identified
// by a series key.
type DataFrame map[string]DataSeries

var (
	_ json.Marshaler   = DataFrame{}
	_ json.Unmarshaler = (*DataFrame)(nil)
)

// Timestamps returns an ordered set of all timestamps in the data-frame where
// there is at least one non-empty (not NaN) value.
func (df DataFrame) Timestamps() []fields.Timestamp {
	m := make(map[fields.Timestamp]struct{})
	for _, s := range df {
		for t, v := range s {
			if math.IsNaN(v) {
				continue
			}
			m[t] = struct{}{}
		}
	}
	ordered := make([]fields.Timestamp, 0, len(m))
	for t := range m {
		ordered = append(ordered, t)
	}
	slices.Sort(ordered)
	return ordered
}

// ordered returns a valid and ordered RawDataFrame with duplicated entries
// removed.
func (df DataFrame) ordered() rawDataFrame {
	out := rawDataFrame{
		Times:  df.Timestamps(),
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
	return json.Marshal(df.ordered())
}

func (df *DataFrame) UnmarshalJSON(b []byte) error {
	in := rawDataFrame{
		Series: make(map[string][]fields.Number),
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
	Times  []fields.Timestamp         `json:"times"`
	Series map[string][]fields.Number `json:"series"`
}

var _ sort.Interface = rawDataFrame{}

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
func (raw rawDataFrame) DataFrame() DataFrame {
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
