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
	"github.com/clarify/clarify-go/resource"
)

const (
	keyFromAttributesHash = "clarify/clarify-go/from/signal/attributes-hash"
	keyFromSignal         = "clarify/clarify-go/from/signal/id"
)

// Item describe the select view for an item.
type Item = resource.Resource[ItemAttributes, ItemRelationships]

type ItemInclude struct{}

// ItemSave describe the save view for an item.
type ItemSave struct {
	ItemSaveAttributes
	resource.ResourceMetaSave
}

// PublishedItem constructs a view for an item based on the passed in signal,
// including a set of base annotations referring the item back to the signal it
// was generated from. The passed in transforms, if any, are run in order.
func PublishedItem(signal Signal, transforms ...ItemTransformFunc) ItemSave {
	item := ItemSave{
		ResourceMetaSave: resource.ResourceMetaSave{
			Annotations: map[string]string{
				keyFromAttributesHash: signal.Meta.AttributesHash.String(),
				keyFromSignal:         signal.ID,
			},
		},
		ItemSaveAttributes: ItemSaveAttributes{
			Name:           signal.Attributes.Name,
			Description:    signal.Attributes.Description,
			ValueType:      signal.Attributes.ValueType,
			SourceType:     signal.Attributes.SourceType,
			EngUnit:        signal.Attributes.EngUnit,
			SampleInterval: signal.Attributes.SampleInterval,
			GapDetection:   signal.Attributes.GapDetection,
			Labels:         signal.Attributes.Labels.Clone(),
			EnumValues:     signal.Attributes.EnumValues.Clone(),
			Visible:        false,
		},
	}

	for _, f := range transforms {
		f(&item)
	}
	return item
}

// ItemTransformFunc is a function that inspects and transforms an item.
type ItemTransformFunc func(dest *ItemSave)

// ItemAttributes contains attributes for the item select view.
type ItemAttributes struct {
	ItemSaveAttributes
}

// ItemSaveAttributes contains attributes that are part of the item save view.
type ItemSaveAttributes struct {
	Name           string               `json:"name"`
	Description    string               `json:"description"`
	ValueType      ValueType            `json:"valueType"`
	SourceType     SourceType           `json:"sourceType"`
	EngUnit        string               `json:"engUnit"`
	SampleInterval fields.FixedDuration `json:"sampleInterval"`
	GapDetection   fields.FixedDuration `json:"gapDetection"`
	Labels         fields.Labels        `json:"labels"`
	EnumValues     fields.EnumValues    `json:"enumValues"`
	Visible        bool                 `json:"visible"`
}

// ItemRelationships describe the item relationships that's exposed by the API.
type ItemRelationships struct{}
