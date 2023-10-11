// Copyright 2022-2023 Searis AS
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

package params

import (
	"encoding/json"
	"fmt"
	"strings"
)

// ResourceFilterType is a sum type of all types that can be converted to a
// Filter instance. This is a sealed interface which means it cannot be
// implemented by end-users.
type ResourceFilterType interface{ filter() ResourceFilter }

// And returns a new resource filter that merges the passed-in filters with
// logical AND.
func And(filters ...ResourceFilterType) ResourceFilter {
	newF := ResourceFilter{
		and: make([]ResourceFilter, 0, len(filters)),
	}
	for _, ft := range filters {
		f := ft.filter()
		switch {
		case len(f.or) == 0 && len(f.paths) == 0:
			// Flatten AND values (and skip empty queries).
			newF.and = append(newF.and, f.and...)
		default:
			newF.and = append(newF.and, f)
		}
	}
	if len(newF.and) == 1 {
		return newF.and[0]
	}
	return newF
}

// Or returns a new resource filter that merges the passed-in filters with
// logical OR.
func Or(filters ...ResourceFilterType) ResourceFilter {
	newF := ResourceFilter{
		or: make([]ResourceFilter, 0, len(filters)),
	}

	for _, ft := range filters {
		f := ft.filter()
		switch {
		case f.matchAll():
			// Optimization:
			//   OR(matchAll,matchSome) == matchAll
			return ResourceFilter{}
		case len(f.and) == 0 && len(f.paths) == 0:
			// Flatten OR values if the filter contains only OR values.
			newF.or = append(newF.or, f.or...)
		default:
			newF.or = append(newF.or, f)
		}
	}
	if len(newF.or) == 1 {
		return newF.or[0]
	}
	return newF
}

// ResourceFilter describe a filter for matching clarify resources.
type ResourceFilter struct {
	and   []ResourceFilter
	or    []ResourceFilter
	paths Comparisons
}

func (f ResourceFilter) filter() ResourceFilter {
	return f
}

// FilterAll returns an empty filter, meaning it match all resources.
func FilterAll() ResourceFilter {
	return ResourceFilter{}
}

// matchAll return true if the filter matches all resources. A.k.a. the
// filter is empty.
func (f ResourceFilter) matchAll() bool {
	return len(f.and) == 0 && len(f.or) == 0 && len(f.paths) == 0
}

var (
	_ interface {
		json.Marshaler
		fmt.Stringer
	} = ResourceFilter{}
	_ json.Unmarshaler = (*ResourceFilter)(nil)
)

func (f ResourceFilter) String() string {
	b, _ := f.MarshalJSON()
	return string(b)
}

func (f ResourceFilter) MarshalJSON() ([]byte, error) {
	m := make(map[string]json.RawMessage, 2+len(f.paths))
	for k, v := range f.paths {
		if strings.HasPrefix(k, "$") {
			return nil, fmt.Errorf("path %q: operator prefix ($) not allowed in path filters", k)
		}
		j, err := json.Marshal(v)
		if err != nil {
			return nil, fmt.Errorf("path %s: %v", k, err)
		}
		m[k] = j
	}
	if len(f.and) > 0 {
		j, err := json.Marshal(f.and)
		if err != nil {
			return nil, fmt.Errorf("$and: %v", err)
		}
		m["$and"] = j
	}
	if len(f.or) > 0 {
		j, err := json.Marshal(f.and)
		if err != nil {
			return nil, fmt.Errorf("$or: %v", err)
		}
		m["$or"] = j
	}
	return json.Marshal(m)
}

func (f *ResourceFilter) UnmarshalJSON(data []byte) error {
	var m map[string]json.RawMessage
	if err := json.Unmarshal(data, &m); err != nil {
		return err
	}

	if v, ok := m["$and"]; ok {
		if err := json.Unmarshal(v, &f.and); err != nil {
			return err
		}
		delete(m, "$and")
	}
	if v, ok := m["$or"]; ok {
		if err := json.Unmarshal(v, &f.or); err != nil {
			return err
		}
		delete(m, "$or")
	}
	f.paths = make(Comparisons, len(m))
	for k, v := range m {
		var cmp Comparison
		if len(k) > 0 && k[0] == '$' {
			return fmt.Errorf("bad conjunction %q", k)
		}
		if err := json.Unmarshal(v, &cmp); err != nil {
			return err
		}
		f.paths[k] = cmp
	}

	// Minor optimization: simplify and/or clauses with only one entry.
	switch {
	case len(f.paths) == 0 && len(f.or) == 0 && len(f.and) == 1:
		f.paths = f.and[0].paths
		f.or = f.and[0].or
		f.and = nil
	case len(f.paths) == 0 && len(f.or) == 1 && len(f.and) == 0:
		f.paths = f.or[0].paths
		f.and = f.or[0].and
		f.or = nil
	}
	return nil
}
