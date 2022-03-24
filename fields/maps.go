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
	"encoding/json"

	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
)

// EnumValues maps integer Items values to strings.
type EnumValues map[int]string

var _ json.Marshaler = EnumValues{}

// Clone returns a deep clone of the enums structure.
func (e EnumValues) Clone() EnumValues {
	return maps.Clone(e)
}

func (e EnumValues) MarshalJSON() ([]byte, error) {
	if len(e) == 0 {
		return []byte(`{}`), nil
	}
	return json.Marshal(map[int]string(e))
}

type Labels map[string][]string

var _ json.Marshaler = Labels{}

// Clone returns a deep clone of the labels structure.
func (l Labels) Clone() Labels {
	if l == nil {
		return nil
	}
	n := make(Labels, len(l))
	for k, v := range l {
		n[k] = slices.Clone(v)
	}
	return n
}

func (l Labels) MarshalJSON() ([]byte, error) {
	if len(l) == 0 {
		return []byte(`{}`), nil
	}
	return json.Marshal(map[string][]string(l))
}

// Set relplace the set of values at the given location. Any provided duplicates
// are automatically removed. If there is no values, the key is deleted.
func (l *Labels) Set(key string, values []string) {
	switch {
	case len(values) == 0 && (*l) == nil:
	case len(values) == 0:
		delete(*l, key)
	case (*l) == nil:
		(*l) = make(Labels)
		fallthrough
	default:
		ll := slices.Clone(values)
		slices.Sort(ll)
		(*l)[key] = slices.Compact(ll)
	}
}

// Add adds the specified value to the relevant key if it's not already present.
// The resulting values is a sorted array.
func (l *Labels) Add(key string, value string) {
	if *l == nil {
		*l = make(Labels)
	}
	ll := (*l)[key]
	if len(ll) == 0 {
		(*l)[key] = []string{value}
		return
	}

	slices.Sort(ll)
	if i, found := slices.BinarySearch(ll, value); !found {
		slices.Insert(ll, i, value)
	}

	(*l)[key] = ll
}

// Remove remvoes the specified value from the relevant key. If there are no
// values left for the key, the key is deleted.
func (l *Labels) Remove(key string, value string) {
	if *l == nil {
		return
	}
	ll := (*l)[key]
	if len(ll) == 0 {
		delete((*l), key)
		return
	}
	slices.Sort((*l)[key])
	if i, found := slices.BinarySearch(ll, value); !found {
		slices.Delete(ll, i, i+1)
	}
	if len(ll) == 0 {
		delete((*l), key)
		return
	}
	(*l)[key] = ll
}
