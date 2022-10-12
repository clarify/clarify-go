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

package fields_test

import (
	"errors"
	"testing"
	"time"

	"github.com/clarify/clarify-go/fields"
)

func TestParseFixedDuration(t *testing.T) {
	testCases := []struct {
		s   string
		d   time.Duration
		err error
	}{
		// Invalid
		{s: "10s", d: 0, err: fields.ErrBadFixedDuration},
		{s: "P1Y", d: 0, err: fields.ErrBadFixedDuration},
		{s: "P1M", d: 0, err: fields.ErrBadFixedDuration},
		{s: "P-3H", d: 0, err: fields.ErrBadFixedDuration},
		// Valid
		{s: "PT0.001S", d: time.Millisecond},
		{s: "-PT0.001S", d: -time.Millisecond},
		{s: "PT2M", d: 2 * time.Minute},
		{s: "PT3H", d: 3 * time.Hour},
		{s: "P4D", d: 4 * 24 * time.Hour},
		{s: "-P1W1DT3H2M0.001S", d: -8*24*time.Hour - 3*time.Hour - 2*time.Minute - time.Millisecond},
		{s: "P1W", d: time.Hour * 24 * 7},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.s, func(t *testing.T) {
			d, err := fields.ParseFixedDuration(tc.s)
			if d.Duration != tc.d {
				t.Errorf("got duration %v, want %v", d, tc.d)
			}
			if !errors.Is(err, tc.err) {
				t.Errorf("unexpected error:\n got: %v\nwant %v", err, tc.err)
			}
		})
	}
}
