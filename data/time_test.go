package data_test

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/clarify/clarify-go/data"
)

func TestOriginTime(t *testing.T) {
	expectTime := time.Date(2000, 01, 03, 0, 0, 0, 0, time.UTC)
	if result := data.OriginTime.Time(); !result.Equal(expectTime) {
		t.Errorf("expected OriginZeroTime.Time() equal %v, got %v", expectTime, result)
	}
	if result := int64(data.OriginTime); result != expectTime.UnixMicro() {
		t.Errorf("expected OriginZeroTime equal %v, got %v", expectTime.UnixMicro(), result)
	}
}

func TestZeroTime(t *testing.T) {
	expectTime := time.Date(1970, 01, 01, 0, 0, 0, 0, time.UTC)
	if result := data.Timestamp(0).Time(); !result.Equal(expectTime) {
		t.Errorf("expected Timestamp(0).Time() equal %v, got %v", expectTime, result)
	}
	if result := int64(data.Timestamp(0)); result != expectTime.UnixMicro() {
		t.Errorf("expected Timestamp(0) equal %v, got %v", expectTime.UnixMicro(), result)
	}
}

func TestParseFixedDuration(t *testing.T) {
	tcs := []struct {
		s   string
		d   data.FixedDuration
		err error
	}{
		// Invalid
		{s: "10s", d: 0, err: data.ErrBadFixedDuration},
		{s: "P1Y", d: 0, err: data.ErrBadFixedDuration},
		{s: "P1M", d: 0, err: data.ErrBadFixedDuration},
		{s: "P-3H", d: 0, err: data.ErrBadFixedDuration},
		// Valid
		{s: "PT0.001S", d: data.Millisecond},
		{s: "-PT0.001S", d: -data.Millisecond},
		{s: "PT2M", d: 2 * data.Minute},
		{s: "PT3H", d: 3 * data.Hour},
		{s: "P4D", d: 4 * 24 * data.Hour},
		{s: "-P1W1DT3H2M0.001S", d: -8*24*data.Hour - 3*data.Hour - 2*data.Minute - data.Millisecond},
		{s: "P1W", d: data.Hour * 24 * 7},
	}

	for _, tc := range tcs {
		tc := tc
		t.Run(tc.s, func(t *testing.T) {
			d, err := data.ParseFixedDuration(tc.s)
			if d != tc.d {
				t.Errorf("got duration %v, want %v", d, tc.d)
			}
			if !errors.Is(err, tc.err) {
				t.Errorf("got error %v, want %v", fmtErr(err), fmtErr(tc.err))
			}
		})
	}
}

func fmtErr(err error) string {
	if err == nil {
		return `<nil>`
	}
	return fmt.Sprintf("%q", err.Error())
}
