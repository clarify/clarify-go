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

package resource

import (
	"bytes"
	"crypto/sha1"
	"encoding/json"
	"io"
	"time"

	"github.com/clarify/clarify-go/fields"
)

// Normalizer describes a type that should be normalized before encoding.
type Normalizer interface {
	Normalize()
}

// Resource describes a generic resource entry select view.
type Resource[A, R any] struct {
	Identifier
	Meta          ResourceMeta `json:"meta"`
	Attributes    A            `json:"attributes"`
	Relationships R            `json:"relationships"`
}

var _ json.Marshaler = Resource[struct{}, struct{}]{}

func (e Resource[A, R]) MarshalJSON() ([]byte, error) {
	var target = struct {
		Identifier
		Meta          ResourceMeta    `json:"meta"`
		Attributes    json.RawMessage `json:"attributes"`
		Relationships R               `json:"relationships"`
	}{
		Identifier:    e.Identifier,
		Meta:          e.Meta,
		Relationships: e.Relationships,
	}

	hash := sha1.New()
	if n, ok := any(&e.Attributes).(Normalizer); ok {
		n.Normalize()
	}
	var buf bytes.Buffer
	enc := json.NewEncoder(io.MultiWriter(hash, &buf))
	if err := enc.Encode(e.Attributes); err != nil {
		return nil, err
	}
	target.Attributes = buf.Bytes()
	target.Meta.AttributesHash = fields.Binary(hash.Sum(nil))
	return json.Marshal(target)
}

// ToOne describes a to one relationship entry.
type ToOne struct {
	Meta map[string]json.RawMessage `json:"meta,omitempty"`
	Data NullIdentifier             `json:"data"`
}

// ToMany describes a to many relationship entry.
type ToMany struct {
	Meta map[string]json.RawMessage `json:"meta,omitempty"`
	Data []Identifier               `json:"data"`
}

// Identifier uniquely identifies a resource entry.
type Identifier struct {
	Type string `json:"type"`
	ID   string `json:"id"`
}

// NullIdentifier is a version of Identifier where the zero-value is encode as
// null in JSON.
type NullIdentifier Identifier

var (
	_ json.Marshaler   = NullIdentifier{}
	_ json.Unmarshaler = (*NullIdentifier)(nil)
)

func (id NullIdentifier) MarshalJSON() ([]byte, error) {
	if id.ID == "" && id.Type == "" {
		return []byte(`null`), nil
	}
	return json.Marshal(Identifier(id))
}

func (id *NullIdentifier) UnmarshalJSON(data []byte) error {
	data = bytes.TrimSpace(data)
	if bytes.Equal(data, []byte(`null`)) {
		*id = NullIdentifier{}
		return nil
	}
	return json.Unmarshal(data, (*Identifier)(id))
}

// ResourceMeta holds the meta data fields for a resource entry select view.
type ResourceMeta struct {
	Annotations    map[string]string `json:"annotations,omitempty"`
	AttributesHash fields.Binary     `json:"attributesHash,omitempty"`
	CreatedAt      time.Time         `json:"createdAt"`
	UpdatedAt      time.Time         `json:"updatedAt"`
}

// ResourceMetaSave holds the mutable meta data fields for a resource entry.
type ResourceMetaSave struct {
	Annotations map[string]string `json:"annotations,omitempty"`
}
