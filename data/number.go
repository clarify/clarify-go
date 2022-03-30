package data

import (
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
		return []byte(null), nil
	}
	return json.Marshal(float64(f))
}

func (f *Number) UnmarshalJSON(data []byte) error {
	if string(data) == null {
		*f = Number(math.NaN())
		return nil
	}
	return json.Unmarshal(data, (*float64)(f))
}
