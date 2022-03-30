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

package clarify

import (
	"github.com/clarify/clarify-go/data"
	"github.com/clarify/clarify-go/resource"
)

// SignalSave describe the writable attributes.
type SignalSave struct {
	SignalAttributes
	resource.MetaSave
}

// SignalAttributes contains writable signal attributes.
type SignalAttributes struct {
	Name           string             `json:"name"`
	Description    string             `json:"description"`
	Type           ValueType          `json:"type"`
	SourceType     SourceType         `json:"sourceType"`
	EngUnit        string             `json:"engUnit"`
	SampleInterval data.FixedDuration `json:"sampleInterval"`
	GapDetection   data.FixedDuration `json:"gapDetection"`
	Labels         resource.Labels    `json:"labels"`
	EnumValues     map[int]string     `json:"enumValues"`
}

type SignalRelationships struct {
	Integration resource.ToOne `json:"integration"`
	Item        resource.ToOne `json:"item"`
}

func (attr *SignalAttributes) Normalize() {
	if attr.Labels == nil {
		attr.Labels = resource.Labels{}
	}
	if attr.EnumValues == nil {
		attr.EnumValues = map[int]string{}
	}
}

// ValueType determine how Items data values should be interpreted.
type ValueType string

// Allowed timeseires types.
const (
	// Numeric indicates that Items values should be treated as numeric
	// (floating point) values.
	Numeric ValueType = "numeric"

	// Enum indicates that Items values should be treated as integer
	// indices to the Items EnumValues map.
	Enum ValueType = "enum"
)

// SourceType describe the the source of a Items instance.
type SourceType string

// Allowed Items source types.
const (
	// Measurement indicates that Items values are retrieved "directly"
	// from a sensor.
	Measurement SourceType = "measurement"

	// Aggregation indicates that Items values are aggregated from one or
	// more sources.
	Aggregation SourceType = "aggregation"

	// Prediction indicates that Items values are predictions or forecasts.
	Prediction SourceType = "prediction"
)
