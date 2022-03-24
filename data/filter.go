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

import "time"

// Filters allows filtering which data to include in a data frame.
type Filter struct {
	Times TimeComparison `json:"times"`
}

// TimeComparison allows filtering times. The zero-value indicate API defaults.
type TimeComparison struct {
	GreaterThanOrEqual time.Time `json:"$gte,omitempty"`
	LessThan           time.Time `json:"$lt,omitempty"`
}

// TimeRange matches times within the specified range.
func TimeRange(gte, lt time.Time) TimeComparison {
	return TimeComparison{
		GreaterThanOrEqual: gte,
		LessThan:           lt,
	}
}
