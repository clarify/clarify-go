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

package fields

import (
	"encoding"
	"encoding/json"
	"fmt"
)

const (
	AggregateDefault AggregateMethod = iota
	AggregateCount
	AggregateMin
	AggregateMax
	AggregateSum
	AggregateAvg
	AggregateStateHistSeconds
	AggregateStateHistPercent
	AggregateStateHistRate
)

type AggregateMethod uint8

var (
	_ encoding.TextMarshaler   = AggregateMethod(0)
	_ encoding.TextUnmarshaler = (*AggregateMethod)(nil)
)

func (m AggregateMethod) String() string {
	b, err := m.MarshalText()
	if err != nil {
		return "%(INVALID)!"
	}
	return string(b)
}

func (m AggregateMethod) MarshalText() ([]byte, error) {
	switch m {
	case AggregateDefault:
		return nil, nil
	case AggregateCount:
		return []byte("count"), nil
	case AggregateMin:
		return []byte("min"), nil
	case AggregateMax:
		return []byte("max"), nil
	case AggregateSum:
		return []byte("sum"), nil
	case AggregateAvg:
		return []byte("avg"), nil
	case AggregateStateHistSeconds:
		return []byte("state-histogram-seconds"), nil
	case AggregateStateHistPercent:
		return []byte("state-histogram-percent"), nil
	case AggregateStateHistRate:
		return []byte("state-histogram-rate"), nil
	}
	return nil, fmt.Errorf("unknown aggregation method")
}

func (m *AggregateMethod) UnmarshalText(data []byte) error {
	switch string(data) {
	case "":
		*m = AggregateDefault
	case "count":
		*m = AggregateCount
	case "min":
		*m = AggregateMin
	case "max":
		*m = AggregateMax
	case "sum":
		*m = AggregateSum
	case "avg":
		*m = AggregateAvg
	case "state-histogram-seconds":
		*m = AggregateStateHistSeconds
	case "state-histogram-percent":
		*m = AggregateStateHistPercent
	case "state-histogram-rate":
		*m = AggregateStateHistRate
	default:
		return fmt.Errorf("bad aggregation method")
	}
	return nil
}

type ItemAggregation struct {
	Alias       string          `json:"alias"`
	ID          string          `json:"id"`
	Aggregation AggregateMethod `json:"aggregation"`
	State       int             `json:"state"`
}

var _ json.Marshaler = ItemAggregation{}

func (ia ItemAggregation) MarshalJSON() ([]byte, error) {
	var v any
	switch ia.Aggregation {
	case AggregateStateHistSeconds, AggregateStateHistPercent, AggregateStateHistRate:
		type encType ItemAggregation
		v = encType(ia)
	default:
		type encType struct {
			Alias       string          `json:"alias"`
			ID          string          `json:"id"`
			Aggregation AggregateMethod `json:"aggregation"`
			State       int             `json:"-"`
		}
		v = encType(ia)
	}
	return json.Marshal(v)
}

type Calculation struct {
	Alias   string `json:"alias"`
	Formula string `json:"formula"`
}
