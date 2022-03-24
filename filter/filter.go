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

package filter

import (
	"encoding/json"
	"fmt"
	"strings"
)

// Filter describes a search filter for matching resources.
type Filter struct {
	and   []Filter
	or    []Filter
	paths Comparisons
}

var (
	_ interface {
		json.Marshaler
		fmt.Stringer
	} = Filter{}
)

// Fields returns a new filter comparing multiple field paths.
func Fields(paths Comparisons) Filter {
	return Filter{paths: paths}
}

// Field returns a new filter comparing a single field path.
func Field(path string, cmp Comparison) Filter {
	return Filter{paths: Comparisons{path: cmp}}
}

// And returns an new filter that merges the passed in filters with logical AND.
func And(filters ...Filter) Filter {
	newF := Filter{
		and: make([]Filter, 0, len(filters)),
	}
	for _, f := range filters {
		switch {
		case len(f.or) == 0 && len(f.paths) == 0:
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

// Or returns an new filter that merges the passed in filters with logical OR.
func Or(filters ...Filter) Filter {
	newF := Filter{
		or: make([]Filter, 0, len(filters)),
	}
	for _, f := range filters {
		switch {
		case len(f.and) == 0 && len(f.paths) == 0:
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

func (f Filter) String() string {
	b, _ := f.MarshalJSON()
	return string(b)
}

func (f Filter) MarshalJSON() ([]byte, error) {
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
