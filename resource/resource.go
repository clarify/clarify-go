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
)

// Normalizer describes a type that should be normalized before encoding.
type Normalizer interface {
	Normalize()
}

// SelectEntry describes the generic selection view of a resource entry.
type SelectEntry[A, R any] struct {
	Identifier
	Meta          MetaSelect `json:"meta"`
	Attributes    A          `json:"attributes"`
	Relationships R          `json:"relationships"`
}

var _ json.Marshaler = SelectEntry[struct{}, struct{}]{}

func (e SelectEntry[A, R]) MarshalJSON() ([]byte, error) {
	var target = struct {
		Identifier
		Meta          MetaSelect      `json:"meta"`
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
	target.Meta.AttributesHash = Binary(hash.Sum(nil))
	return json.Marshal(target)
}

// ToOne describes a to one relationship entry.
type ToOne struct {
	Meta map[string]json.RawMessage `json:"meta,omitempty"`
	Data *Identifier                `json:"data"`
}

// ToMany describes a to many relationship entry.
type ToMany struct {
	Meta map[string]json.RawMessage `json:"meta,omitempty"`
	Data []Identifier               `json:"data"`
}

// Identifier identifies a resource.
type Identifier struct {
	Type string `json:"type"`
	ID   string `json:"id"`
}

// MetaSelect holds the standardized meta data fields for a resource entry's
// selection view.
type MetaSelect struct {
	Annotations    map[string]string `json:"annotations,omitempty"`
	AttributesHash Binary            `json:"attributesHash,omitempty"`
	CreatedAt      time.Time         `json:"createdAt"`
	UpdatedAt      time.Time         `json:"updatedAt"`
}

// MetaSave holds the mutable meta data fields for a resource entry.
type MetaSave struct {
	Annotations map[string]string `json:"annotations,omitempty"`
}
