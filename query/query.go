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

package query

import "time"

// Query describes the resource query structure.
type Query struct {
	Filter Filter   `json:"filter,omitempty"`
	Sort   []string `json:"sort,omitempty"`
	Limit  int      `json:"limit"`
	Skip   int      `json:"skip"`
}

// New returns a new resource query, using the API default limit.
func New() Query {
	return Query{Limit: -1}
}

// Data describes a data frame query structure.
type Data struct {
	Filter DataFilter `json:"filter"`
	Rollup string     `json:"rollup"`
	Last   int        `json:"last,omitempty"`
}

// DataFilter allows filtering which data to include in a data frame. The filter
// follows a similar structure to a resource Filter, but is more limited, and
// does not allow combinging filters with "$and" or "$or" conjunctions.
type DataFilter struct {
	Times DataTimesComparison `json:"times"`
}

// DataTimesComparison allows filtering times. The zero-value indicate API
// defaults.
type DataTimesComparison struct {
	GreaterThanOrEqual time.Time `json:"$gte,omitempty"`
	LessThan           time.Time `json:"$lt,omitempty"`
}

// DataTimesRange matches times within the specified range.
func DataTimesRange(gte, lt time.Time) DataTimesComparison {
	return DataTimesComparison{
		GreaterThanOrEqual: gte,
		LessThan:           lt,
	}
}
