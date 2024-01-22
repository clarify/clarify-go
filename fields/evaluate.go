// Copyright 2023-2024 Searis AS
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
	TimeAggregationDefault TimeAggregationMethod = iota
	TimeAggregationCount
	TimeAggregationMin
	TimeAggregationMax
	TimeAggregationSum
	TimeAggregationAvg
	TimeAggregationSeconds
	TimeAggregationPercent
	TimeAggregationRate
)

type TimeAggregationMethod uint8

const (
	GroupAggregationDefault GroupAggregation = iota
	GroupAggregationCount
	GroupAggregationMin
	GroupAggregationMax
	GroupAggregationSum
	GroupAggregationAvg
)

type GroupAggregation uint8

var (
	_ encoding.TextMarshaler   = TimeAggregationMethod(0)
	_ encoding.TextUnmarshaler = (*TimeAggregationMethod)(nil)
)

func (m TimeAggregationMethod) String() string {
	b, err := m.MarshalText()
	if err != nil {
		return "%(INVALID)!"
	}
	return string(b)
}

func (m TimeAggregationMethod) MarshalText() ([]byte, error) {
	switch m {
	case TimeAggregationDefault:
		return nil, nil
	case TimeAggregationCount:
		return []byte("count"), nil
	case TimeAggregationMin:
		return []byte("min"), nil
	case TimeAggregationMax:
		return []byte("max"), nil
	case TimeAggregationSum:
		return []byte("sum"), nil
	case TimeAggregationAvg:
		return []byte("avg"), nil
	case TimeAggregationSeconds:
		return []byte("state-seconds"), nil
	case TimeAggregationPercent:
		return []byte("state-percent"), nil
	case TimeAggregationRate:
		return []byte("state-rate"), nil
	}
	return nil, fmt.Errorf("bad aggregation method")
}

func (m *TimeAggregationMethod) UnmarshalText(data []byte) error {
	switch string(data) {
	case "":
		*m = TimeAggregationDefault
	case "count":
		*m = TimeAggregationCount
	case "min":
		*m = TimeAggregationMin
	case "max":
		*m = TimeAggregationMax
	case "sum":
		*m = TimeAggregationSum
	case "avg":
		*m = TimeAggregationAvg
	case "state-seconds", "state-histogram-seconds":
		*m = TimeAggregationSeconds
	case "state-percent", "state-histogram-percent":
		*m = TimeAggregationPercent
	case "state-rate", "state-histogram-rate":
		*m = TimeAggregationRate
	default:
		return fmt.Errorf("bad aggregation method")
	}
	return nil
}

func (m GroupAggregation) MarshalText() ([]byte, error) {
	switch m {
	case GroupAggregationDefault:
		return nil, nil
	case GroupAggregationCount:
		return []byte("count"), nil
	case GroupAggregationMin:
		return []byte("min"), nil
	case GroupAggregationMax:
		return []byte("max"), nil
	case GroupAggregationSum:
		return []byte("sum"), nil
	case GroupAggregationAvg:
		return []byte("avg"), nil
	}
	return nil, fmt.Errorf("bad aggregation method")
}

func (m *GroupAggregation) UnmarshalText(data []byte) error {
	switch string(data) {
	case "":
		*m = GroupAggregationDefault
	case "count":
		*m = GroupAggregationCount
	case "min":
		*m = GroupAggregationMin
	case "max":
		*m = GroupAggregationMax
	case "sum":
		*m = GroupAggregationSum
	case "avg":
		*m = GroupAggregationAvg
	default:
		return fmt.Errorf("bad aggregation method")
	}
	return nil
}

type EvaluateItem struct {
	Alias           string                `json:"alias,omitempty"`
	ID              string                `json:"id,omitempty"`
	TimeAggregation TimeAggregationMethod `json:"timeAggregation,omitempty"`
	State           int                   `json:"state"`
	Lead            int                   `json:"lead,omitempty"`
	Lag             int                   `json:"lag,omitempty"`
}

type EvaluateGroup struct {
	Alias            string                `json:"alias,omitempty"`
	ID               string                `json:"id,omitempty"`
	TimeAggregation  TimeAggregationMethod `json:"timeAggregation,omitempty"`
	GroupAggregation GroupAggregation      `json:"groupAggregation,omitempty"`
	State            int                   `json:"state"`
	Lead             int                   `json:"lead,omitempty"`
	Lag              int                   `json:"lag,omitempty"`
}

var _ json.Marshaler = EvaluateItem{}

func (ia EvaluateItem) MarshalJSON() ([]byte, error) {
	var v any
	switch ia.TimeAggregation {
	case TimeAggregationSeconds, TimeAggregationPercent, TimeAggregationRate:
		type encType EvaluateItem
		v = encType(ia)
	default:
		type encType struct {
			Alias           string                `json:"alias,omitempty"`
			ID              string                `json:"id,omitempty"`
			TimeAggregation TimeAggregationMethod `json:"aggregation,omitempty"`
			State           int                   `json:"-"`
			Lead            int                   `json:"lead,omitempty"`
			Lag             int                   `json:"lag,omitempty"`
		}
		v = encType(ia)
	}
	return json.Marshal(v)
}

func (ga EvaluateGroup) MarshalJSON() ([]byte, error) {
	var v any

	type encType struct {
		Alias            string                `json:"alias,omitempty"`
		ID               string                `json:"id,omitempty"`
		TimeAggregation  TimeAggregationMethod `json:"aggregation,omitempty"`
		GroupAggregation GroupAggregation      `json:"groupAggregation,omitempty"`
		State            int                   `json:"-"`
		Lead             int                   `json:"lead,omitempty"`
		Lag              int                   `json:"lag,omitempty"`
	}

	v = encType(ga)

	return json.Marshal(v)
}

type Calculation struct {
	Alias   string `json:"alias"`
	Formula string `json:"formula"`
}
