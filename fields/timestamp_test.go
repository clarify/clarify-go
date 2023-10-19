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

package fields_test

import (
	"fmt"
	"math"
	"testing"
	"time"

	"github.com/clarify/clarify-go/fields"
)

func TestOriginTime(t *testing.T) {
	expectTime := time.Date(2000, 01, 03, 0, 0, 0, 0, time.UTC)
	if result := fields.OriginTime.Time(); !result.Equal(expectTime) {
		t.Errorf("expected OriginZeroTime.Time() equal %v, got %v", expectTime, result)
	}
	if result := int64(fields.OriginTime); result != expectTime.UnixMicro() {
		t.Errorf("expected OriginZeroTime equal %v, got %v", expectTime.UnixMicro(), result)
	}
}

func TestZeroTime(t *testing.T) {
	expectTime := time.Date(1970, 01, 01, 0, 0, 0, 0, time.UTC)
	if result := fields.Timestamp(0).Time(); !result.Equal(expectTime) {
		t.Errorf("expected Timestamp(0).Time() equal %v, got %v", expectTime, result)
	}
	if result := int64(fields.Timestamp(0)); result != expectTime.UnixMicro() {
		t.Errorf("expected Timestamp(0) equal %v, got %v", expectTime.UnixMicro(), result)
	}
}

func TestTimestampTruncate(t *testing.T) {
	// test1 is valid only when the different origin between Timestamp and
	// Time is insignificant.

	test := func(ts fields.Timestamp, d time.Duration) func(t *testing.T) {
		return func(t *testing.T) {
			t.Helper()

			// Compare result against multiple alternate implementations; the
			// first case is allowed to fail.
			t.Run("ts.Truncate(d).Time()==ts.Time().Truncate(d)", softCompareTimestampTruncateTime(ts, d))
			t.Run("ts.Truncate(d)==timestampTimeBucket(ts,d,origin,min,max)", compareTimestampTruncateTimescale(ts, d))
			t.Run("ts.Truncate(d)==custom", compareTimestampTruncateCustom(ts, d))
		}
	}

	// Many cases are found/adapted from FuzzTimestampTruncate.
	now := time.Now().UTC()
	t.Run("now.Truncate(24h)", test(fields.AsTimestamp(now), 24*time.Hour))
	t.Run("now.Truncate(7*24h)", test(fields.AsTimestamp(now), 7*24*time.Hour))
	t.Run("now.Truncate(1h)", test(fields.AsTimestamp(now), time.Hour))
	t.Run("now.Truncate(-1h)", test(fields.AsTimestamp(now), -time.Hour))
	t.Run("now.Truncate(1min)", test(fields.AsTimestamp(now), time.Minute))
	t.Run("now.Truncate(-1min)", test(fields.AsTimestamp(now), -time.Minute))
	t.Run("now.Truncate(1µs)", test(fields.AsTimestamp(now), time.Microsecond))
	t.Run("now.Truncate(3µs)", test(fields.AsTimestamp(now), time.Microsecond))
	t.Run("now.Truncate(-1µs)", test(fields.AsTimestamp(now), -time.Microsecond))
	t.Run("now.Truncate(35µs)", test(fields.AsTimestamp(now), 35*time.Microsecond))
	t.Run("now.Truncate(-35µs)", test(fields.AsTimestamp(now), -35*time.Microsecond))
	t.Run("Timestamp(1).Truncate(1µs)", test(fields.Timestamp(1), 1*time.Microsecond))
	t.Run("Timestamp(origin-1h30min).Truncate(1h)", test(fields.OriginTime.Add(-1*time.Hour-30*time.Minute), time.Hour))
	t.Run("Timestamp(origin-1).Truncate(2µs)", test(fields.OriginTime-1, 2*time.Microsecond))
	t.Run("Timestamp(origin-1).Truncate(3µs)", test(fields.OriginTime-1, 3*time.Microsecond))
	t.Run("Timestamp(origin+1).Truncate(2µs)", test(fields.OriginTime+1, 2*time.Microsecond))
	t.Run("Timestamp(origin+1).Truncate(3µs)", test(fields.OriginTime+1, 3*time.Microsecond))
	t.Run("Timestamp(-1).Truncate(2µs)", test(-1, 2*time.Microsecond))
	t.Run("Timestamp(1).Truncate(2µs)", test(1, 2*time.Microsecond))
	t.Run("Timestamp(1).Truncate(3µs)", test(1, 3*time.Microsecond))
	t.Run("Timestamp(3).Truncate(3µs)", test(3, 3*time.Microsecond))
	t.Run("Timestamp(-3).Truncate(3µs)", test(3, 3*time.Microsecond))
	t.Run("Timestamp(-1).Truncate(3µs)", test(-1, 3*time.Microsecond))
	t.Run("Timestamp(1).Truncate(11µs)", test(1, 11*time.Microsecond))
	t.Run("Timestamp(1).Truncate(13s)", test(1, 13*time.Microsecond))
}

func FuzzTimestampTruncate(f *testing.F) {
	//origTime := time.Date(2000, 01, 03, 0, 0, 0, 0, time.UTC)
	f.Fuzz(func(t *testing.T, msec, dMsec int64) {
		ts := fields.Timestamp(msec)
		d := time.Duration(dMsec) * time.Microsecond

		testName := fmt.Sprintf("Timestamp(%v).Truncate(%v)",
			ts.Time().Format(time.RFC3339Nano+" (Mon)"),
			d,
		)

		// Compare result against multiple alternate implementations; the
		// first case is allowed to fail.
		t.Run(testName+".Time()==ts.Time().Truncate(d)", softCompareTimestampTruncateTime(ts, d))
		t.Run(testName+"==timestampTimeBucket(ts,d,origin,min,max)", compareTimestampTruncateTimescale(ts, d))
		t.Run(testName+"==custom", compareTimestampTruncateCustom(ts, d))
	})
}

func softCompareTimestampTruncateTime(ts fields.Timestamp, d time.Duration) func(t *testing.T) {
	return func(t *testing.T) {
		t.Helper()

		expect := ts.Time().Truncate(d)
		result := ts.Truncate(d).Time()
		if !expect.Equal(result) {
			// This test is not expected tp pass for all cases; only log the
			// error.
			t.Logf("Test does not pass:\nGot:  %v\nWant: %v", result, expect)
		}
	}
}

func compareTimestampTruncateTimescale(ts fields.Timestamp, d time.Duration) func(t *testing.T) {
	return func(t *testing.T) {
		t.Helper()

		var expect int64
		if d > 0 {
			expect = timescaleTimeBucket(d.Microseconds(), int64(ts), int64(fields.OriginTime), math.MinInt64, math.MaxInt64)
		} else {
			expect = int64(ts)
		}

		result := int64(ts.Truncate(d))
		if expect != result {
			t.Errorf("Unexpected result:\nGot:  %v\nWant: %v", result, expect)
		}
	}
}

func compareTimestampTruncateCustom(ts fields.Timestamp, d time.Duration) func(t *testing.T) {
	return func(t *testing.T) {
		t.Helper()

		td := fields.Timestamp(d.Microseconds())
		result := ts.Truncate(d)

		var expect fields.Timestamp
		if td > 0 {
			tsRelOrigin := ts - fields.OriginTime
			m := tsRelOrigin / td
			if tsRelOrigin < 0 && (tsRelOrigin)%td != 0 {
				m--
			}
			expect = fields.OriginTime + td*m
		} else {
			expect = ts
		}

		if expect != result {
			t.Errorf("Unexpected result:\nGot:  %v\nWant: %v",
				result, expect,
			)
		}
	}
}
