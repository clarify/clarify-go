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

package fields

import (
	"bytes"
	"encoding/json"
	"math"
)

// Number is a float64 value that JSON encodes the IEEE 754 "not-a-number" value
// to 'null'.
type Number float64

var (
	_ json.Unmarshaler = (*Number)(nil)
	_ json.Marshaler   = Number(0)
)

func (f Number) IsNaN() bool {
	return math.IsNaN(float64(f))
}

func (f Number) Float64() float64 {
	return float64(f)
}

func (f Number) MarshalJSON() ([]byte, error) {
	if math.IsNaN(float64(f)) {
		return []byte(`null`), nil
	}
	return json.Marshal(float64(f))
}

func (f *Number) UnmarshalJSON(data []byte) error {
	data = bytes.TrimSpace(data)
	if bytes.Equal(data, []byte(`null`)) {
		*f = Number(math.NaN())
		return nil
	}
	return json.Unmarshal(data, (*float64)(f))
}
