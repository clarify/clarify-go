// Copyright 2022 Searis AS, 2017-2022 Timescale, Inc.
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

package fields_test

// timestampTimeBucket is a direct Go translation of C code from Timescale. We
// use it as a reference check for our own implementation.
// https://github.com/timescale/timescaledb/blob/a6b5f9002cf4f3894aa8cbced7f862a73784cada/src/time_bucket.c#L18
func timescaleTimeBucket(period, timestamp, offset, min, max int64) int64 {
	if period <= 0 {
		panic("period must be greater than 0")
	}
	if offset != 0 {
		// We need to ensure that the timestamp is in range _after_ the
		// offset is applied: when the offset is positive we need to make
		// sure the resultant time is at least min, and when negative that
		// it is less than the max.
		offset = offset % period
		if (offset > 0 && timestamp < min+offset) || (offset < 0 && timestamp > max+offset) {
			panic("timestamp out of range")
		}
		timestamp -= offset
	}

	result := (timestamp / period) * period
	if timestamp < 0 && timestamp%period != 0 {
		if result < (min)+(period) {
			panic("timestamp out of range")
		}
		result -= period
	}
	result += offset

	return result
}
