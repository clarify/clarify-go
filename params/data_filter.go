// Copyright 2023 Searis AS
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
	"slices"
	"time"
)

// DataFilter describe a type can return an internal data filter structure. Data
// filters helps reduce the amount of data that is returned by a method.
type DataFilter struct {
	filter dataFilter
}

var (
	_ json.Marshaler   = DataFilter{}
	_ json.Unmarshaler = (*DataFilter)(nil)
)

func (q DataFilter) MarshalJSON() ([]byte, error) {
	return json.Marshal(q.filter)
}

func (q *DataFilter) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &q.filter)
}

// DataAnd joins one or more data filters with logical and.
func DataAnd(filters ...DataFilter) DataFilter {
	var result DataFilter

	for _, f := range filters {
		// Use greatest non-zero times.$gte value.
		switch {
		case f.filter.Times.GreaterOrEqual.IsZero():
			//pass
		case result.filter.Times.GreaterOrEqual.IsZero(), f.filter.Times.GreaterOrEqual.After(result.filter.Times.GreaterOrEqual):
			result.filter.Times.GreaterOrEqual = f.filter.Times.GreaterOrEqual
		}

		// Use least non-zero times.$lt value.
		switch {
		case f.filter.Times.Less.IsZero():
			// pass
		case result.filter.Times.Less.IsZero(), f.filter.Times.Less.Before(result.filter.Times.Less):
			// Use least value.
			result.filter.Times.Less = f.filter.Times.Less
		}

		// Use the union of non-zero series.$in values.
		switch {
		case f.filter.Series.In == nil:
			// pass
		case result.filter.Series.In == nil:
			result.filter.Series.In = f.filter.Series.In
		default:
			sizeHint := len(result.filter.Series.In)
			if l := len(f.filter.Series.In); l < sizeHint {
				sizeHint = l
			}
			union := make([]string, 0, sizeHint)
			for _, k := range result.filter.Series.In {
				if slices.Contains(f.filter.Series.In, k) {
					union = append(union, k)
				}
			}
			result.filter.Series.In = union
		}
	}
	return result
}

type dataFilter struct {
	Times  timesFilter  `json:"times"`
	Series seriesFilter `json:"series"`
}

type timesFilter struct {
	GreaterOrEqual time.Time `json:"$gte,omitempty"`
	Less           time.Time `json:"$lt,omitempty"`
}

// TimeRange return a TimesFilter that matches times in range [gte,lt).
//
// Be aware of API limits according to how large time ranges you can query with
// different query resolutions. In order to query larger time windows in a
// single query, you can increase the width of your rollup duration.
//
// See the API documentation for the method you are calling for more details:
// https://docs.clarify.io/api/1.1/.
func TimeRange(gte, lt time.Time) DataFilter {
	return DataFilter{
		filter: dataFilter{
			Times: timesFilter{
				GreaterOrEqual: gte,
				Less:           lt,
			},
		},
	}
}

type seriesFilter struct {
	In []string `json:"$in,omitempty"`
}

// SeriesIn return a data filter that reduce the time-series to encode in the
// final result to the ones that are in the specified list of keys.
func SeriesIn(keys ...string) DataFilter {
	return DataFilter{
		filter: dataFilter{
			Series: seriesFilter{
				In: keys,
			},
		},
	}
}
