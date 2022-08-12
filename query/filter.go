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

package query

import (
	"encoding/json"
	"fmt"
	"strings"
)

// FilterType describe any type that can generate a filter.
type FilterType interface {
	Filter() Filter
}

// Filter describes a search filter for matching resources.
type Filter struct {
	And   []Filter
	Or    []Filter
	Paths Comparisons
}

func (f Filter) Filter() Filter { return f }

var (
	_ interface {
		json.Marshaler
		fmt.Stringer
		FilterType
	} = Filter{}
	_ json.Unmarshaler = (*Filter)(nil)
)

// Field returns a new filter comparing a single field path.
func Field(path string, cmp Comparison) Filter {
	return Filter{Paths: Comparisons{path: cmp}}
}

// And returns an new filter that merges the passed in filters with logical AND.
func And(filters ...FilterType) Filter {
	newF := Filter{
		And: make([]Filter, 0, len(filters)),
	}
	for _, ft := range filters {
		f := ft.Filter()
		switch {
		case len(f.Or) == 0 && len(f.Paths) == 0:
			newF.And = append(newF.And, f.And...)
		default:
			newF.And = append(newF.And, f)
		}
	}
	if len(newF.And) == 1 {
		return newF.And[0]
	}
	return newF
}

// Or returns an new filter that merges the passed in filters with logical OR.
func Or(filters ...FilterType) Filter {
	newF := Filter{
		Or: make([]Filter, 0, len(filters)),
	}
	for _, ft := range filters {
		f := ft.Filter()
		switch {
		case len(f.And) == 0 && len(f.Paths) == 0:
			newF.Or = append(newF.Or, f.Or...)
		default:
			newF.Or = append(newF.Or, f)
		}
	}
	if len(newF.Or) == 1 {
		return newF.Or[0]
	}
	return newF
}

func (f Filter) String() string {
	b, _ := f.MarshalJSON()
	return string(b)
}

func (f Filter) MarshalJSON() ([]byte, error) {
	m := make(map[string]json.RawMessage, 2+len(f.Paths))
	for k, v := range f.Paths {
		if strings.HasPrefix(k, "$") {
			return nil, fmt.Errorf("path %q: operator prefix ($) not allowed in path filters", k)
		}
		j, err := json.Marshal(v)
		if err != nil {
			return nil, fmt.Errorf("path %s: %v", k, err)
		}
		m[k] = j
	}
	if len(f.And) > 0 {
		j, err := json.Marshal(f.And)
		if err != nil {
			return nil, fmt.Errorf("$and: %v", err)
		}
		m["$and"] = j
	}
	if len(f.Or) > 0 {
		j, err := json.Marshal(f.And)
		if err != nil {
			return nil, fmt.Errorf("$or: %v", err)
		}
		m["$or"] = j
	}
	return json.Marshal(m)
}

func (f *Filter) UnmarshalJSON(data []byte) error {
	var m map[string]json.RawMessage
	if err := json.Unmarshal(data, &m); err != nil {
		return err
	}

	if v, ok := m["$and"]; ok {
		if err := json.Unmarshal(v, &f.And); err != nil {
			return err
		}
		delete(m, "$and")
	}
	if v, ok := m["$or"]; ok {
		if err := json.Unmarshal(v, &f.Or); err != nil {
			return err
		}
		delete(m, "$or")
	}
	f.Paths = make(Comparisons, len(m))
	for k, v := range m {
		var cmp Comparison
		if len(k) > 0 && k[0] == '$' {
			return fmt.Errorf("bad conjunction %q", k)
		}
		if err := json.Unmarshal(v, &cmp); err != nil {
			return err
		}
		f.Paths[k] = cmp
	}

	// Minor optimization: simplify and/or clauses with only one entry.
	switch {
	case len(f.Paths) == 0 && len(f.Or) == 0 && len(f.And) == 1:
		f.Paths = f.And[0].Paths
		f.Or = f.And[0].Or
		f.And = nil
	case len(f.Paths) == 0 && len(f.Or) == 1 && len(f.And) == 0:
		f.Paths = f.Or[0].Paths
		f.And = f.Or[0].And
		f.Or = nil
	}
	return nil
}
