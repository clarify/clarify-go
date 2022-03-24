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

package views

import (
	"github.com/clarify/clarify-go/fields"
	"github.com/clarify/clarify-go/jsonrpc/resource"
)

// Signal describe the select view for a signal.
type Signal = resource.Resource[SignalAttributes, SignalRelationships]

type SignalInclude struct {
	Items []Item `json:"items"`
}

// SignalSave describe the save view for a signal.
type SignalSave struct {
	resource.MetaSave
	SignalSaveAttributes
}

// SignalAttributes contains attributes for the signal select view.
type SignalAttributes struct {
	SignalSaveAttributes
	SignalReadOnlyAttributes
}

// SignalReadOnlyAttributes contains read-only signal attributes.
type SignalReadOnlyAttributes struct {
	Input string `json:"input"`
}

// SignalSaveAttributes contains attributes that are part of the signal save
// view.
type SignalSaveAttributes struct {
	Name           string               `json:"name"`
	Description    string               `json:"description"`
	ValueType      ValueType            `json:"valueType"`
	SourceType     SourceType           `json:"sourceType"`
	EngUnit        string               `json:"engUnit"`
	SampleInterval fields.FixedDuration `json:"sampleInterval"`
	GapDetection   fields.FixedDuration `json:"gapDetection"`
	Labels         fields.Labels        `json:"labels"`
	EnumValues     fields.EnumValues    `json:"enumValues"`
}

// SignalRelationships declare the available relationships for the signal model.
type SignalRelationships struct {
	Integration resource.ToOne `json:"integration"`
	Item        resource.ToOne `json:"item"`
}

// ValueType determine how data values should be interpreted.
type ValueType string

// Allowed signal/items types.
const (
	// Numeric indicates that values should be treated as decimal numbers.
	Numeric ValueType = "numeric"

	// Enum indicates that values should be treated as integer indices to a map
	// of enumerated values.
	Enum ValueType = "enum"
)

// SourceType describe how data values where produced.
type SourceType string

// Allowed signal/item source types.
const (
	// Measurement indicates that values are retrieved "directly" from a sensor.
	Measurement SourceType = "measurement"

	// Aggregation indicates that values are aggregated from one or more
	// sources.
	Aggregation SourceType = "aggregation"

	// Prediction indicates that values are predictions or forecasts.
	Prediction SourceType = "prediction"
)
