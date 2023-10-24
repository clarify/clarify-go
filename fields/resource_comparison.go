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

package fields

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
)

type Comparisons map[string]Comparison

var _ ResourceFilterType = Comparisons{}

// CompareField returns a new filter comparing a single field path.
func CompareField(path string, cmp Comparison) Comparisons {
	return Comparisons{path: cmp}
}

func (c Comparisons) filter() ResourceFilter {
	return ResourceFilter{
		paths: c,
	}
}

// Comparison allows comparing a particular value with one or more operators.
// The zero-value is treated equivalent to Equal(null).
type Comparison struct {
	value *opComparison
}

var (
	_ fmt.Stringer     = Comparison{}
	_ json.Marshaler   = Comparison{}
	_ json.Unmarshaler = (*Comparison)(nil)
)

// MergeOperators merges multiple comparisons with different operators together
// to a single comparison entry.
//
// When conflicting operators ar encountered, the right most value is selected
// for the result. MergeOperators are resolved based on operator keys, where
// some initializers have an overlap:
//
//   - Equal and In both uses $in.
//   - NotEqual and NotIn both uses $nin.
//   - Range and GreaterThanOrEqual both uses $gte.
//   - Range and LessThan both uses $lt.
//
// Example valid usage:
//
//	MergeOperators(Equal(nil), LessThan(49))        // {"$in":[nil],"$lt":49}
//	MergeOperators(GreaterOrEqual(0), LessThan(49)) // {"$gte":0,"$lt":49}
//
// Example of conflicting operators:
//
//	MergeOperators(Equal(0), In(1, 2))       // {"$in":[1,2]}
//	MergeOperators(In(1, 2), Equal(nil))     // null
//	MergeOperators(NotIn(1, 2), NotEqual(0)) // {"$nin":[0]}
//	MergeOperators(NotEqual(0), NotIn(1, 2)) // {"$nin":[1,2]}
func MergeOperators(cmps ...Comparison) Comparison {
	var target opComparison
	for _, cmp := range cmps {
		v := cmp.value
		switch v {
		case nil:
			target.In = []json.RawMessage{json.RawMessage(`null`)}
		default:
			if len(v.In) > 0 {
				target.In = v.In
			}
			if len(v.NotIn) > 0 {
				target.NotIn = v.NotIn
			}
			if v.Greater != nil {
				target.Greater = v.Greater
			}
			if v.GreaterOrEqual != nil {
				target.GreaterOrEqual = v.GreaterOrEqual
			}
			if v.Less != nil {
				target.Less = v.Less
			}
			if v.LessOrEqual != nil {
				target.LessOrEqual = v.LessOrEqual
			}
			if v.Regex != "" {
				target.Regex = v.Regex
			}
		}
	}
	return Comparison{
		value: target.normalize(),
	}
}

type opComparison struct {
	In             []json.RawMessage `json:"$in,omitempty"`
	NotIn          []json.RawMessage `json:"$nin,omitempty"`
	Greater        json.RawMessage   `json:"$gt,omitempty"`
	GreaterOrEqual json.RawMessage   `json:"$gte,omitempty"`
	Less           json.RawMessage   `json:"$lt,omitempty"`
	LessOrEqual    json.RawMessage   `json:"$lte,omitempty"`
	Regex          string            `json:"$regex,omitempty"`
}

func (cmp *opComparison) normalize() *opComparison {
	// Normalizes the following:
	//
	//   - opComparison{In:{"null"}} -> nil
	//   - opComparison{} -> nil
	//
	isEmptyExceptIn := (cmp.NotIn == nil &&
		cmp.Greater == nil &&
		cmp.GreaterOrEqual == nil &&
		cmp.Less == nil &&
		cmp.LessOrEqual == nil &&
		cmp.Regex == "")
	switch {
	case isEmptyExceptIn && cmp.In == nil:
		// Convert to equal null comparison.
		return nil
	case isEmptyExceptIn && len(cmp.In) == 1 && bytes.Equal(cmp.In[0], []byte(`null`)):
		// Convert to equal null comparison.
		return nil
	default:
		return cmp
	}
}

// Equal returns a comparison that match values equal to v. Panics if v is not
// JSON marshalled into a simple JSON type (string, number, bool or null).
func Equal(v any) Comparison {
	return Comparison{
		value: (&opComparison{In: []json.RawMessage{simpleJSON(v)}}).normalize(),
	}
}

// NotEqual returns a comparison that match values not equal to v. Panics if v
// is not JSON marshalled into a simple JSON type (string, number, bool or
// null).
func NotEqual(v any) Comparison {
	return Comparison{
		value: &opComparison{NotIn: []json.RawMessage{simpleJSON(v)}},
	}
}

// In returns a comparison that match values in elements. Panics if any element
// is not JSON marshalled into an a simple JSON type (string, number, bool or
// null).
func In[E any](elements ...E) Comparison {
	in := make([]json.RawMessage, 0, len(elements))
	for _, elem := range elements {
		in = append(in, simpleJSON(elem))
	}
	return Comparison{
		value: (&opComparison{In: in}).normalize(),
	}
}

// NotIn returns a comparison that match values not in elements. Panics if any
// element is not JSON marshalled into an a simple JSON type (string, number,
// bool or null).
func NotIn[E any](elements ...E) Comparison {
	nin := make([]json.RawMessage, 0, len(elements))
	for _, elem := range elements {
		nin = append(nin, simpleJSON(elem))
	}
	return Comparison{
		value: &opComparison{NotIn: nin},
	}
}

// Greater returns a comparison that matches values > gte. Panics if gt is not
// JSON marshalled into an a sortable JSON type (string or number).
func Greater(gt any) Comparison {
	return Comparison{
		value: &opComparison{Greater: orderedJSONType(gt)},
	}
}

// GreaterOrEqual returns a comparison that matches values >= gte. Panics if gte
// is not JSON marshalled into an a sortable JSON type (string or number).
func GreaterOrEqual(gte any) Comparison {
	return Comparison{
		value: &opComparison{GreaterOrEqual: orderedJSONType(gte)},
	}
}

// Less returns a comparison that matches values < lt. Panics if lt is not JSON
// marshalled into an a sortable JSON type (string or number).
func Less(lt any) Comparison {
	return Comparison{
		value: &opComparison{Less: orderedJSONType(lt)},
	}
}

// LessOrEqual returns a comparison that matches values <= lte. Panics if lte is
// not JSON marshalled into an a sortable JSON type (string or number).
func LessOrEqual(lte any) Comparison {
	return Comparison{
		value: &opComparison{LessOrEqual: orderedJSONType(lte)},
	}
}

// Range is a short-hand for:
//
//	MergeComparisons(GreaterThanOrEqual(gte), LessThan(lt))
func Range(gte, lt any) Comparison {
	return Comparison{
		value: &opComparison{
			GreaterOrEqual: orderedJSONType(gte),
			Less:           orderedJSONType(lt),
		},
	}
}

// Regex returns a comparison that match values that matches the provided regexp
// pattern.
func Regex(pattern string) Comparison {
	return Comparison{
		value: &opComparison{Regex: pattern},
	}
}

func (cmp Comparison) String() string {
	b, _ := json.Marshal(cmp)
	return string(b)
}

func (c Comparison) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value.normalize())
}

func (c *Comparison) UnmarshalJSON(data []byte) error {
	data = bytes.TrimSpace(data)
	if len(data) == 0 {
		return fmt.Errorf("cannot unmarshal empty bytes into target of type Comparison")
	}

	var target struct {
		opComparison
		NotEqual json.RawMessage `json:"$ne,omitempty"`
	}

	switch data[0] {
	case '{':
		// Got object, treat as operator comparison.
		if err := json.Unmarshal(data, &target); err != nil {
			return err
		}

		if target.NotEqual != nil {
			target.NotIn = append(target.NotIn, target.NotEqual)
		}
	default:
		// Got non-object, treat as equality comparison.
		var eq json.RawMessage
		if err := json.Unmarshal(data, &eq); err != nil {
			return err
		}
		target.In = append(target.In, eq)
	}

	c.value = target.normalize()
	return nil
}

func simpleJSON(v any) json.RawMessage {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	b = bytes.TrimSpace(b)
	if len(b) == 0 || !strings.ContainsRune(`"0123456789.tfn`, rune(b[0])) {
		panic("does not marshal to simple JSON type (string, number, bool or null)")
	}
	return b
}

func orderedJSONType(v any) json.RawMessage {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	b = bytes.TrimSpace(b)
	if len(b) == 0 || !strings.ContainsRune(`"0123456789.`, rune(b[0])) {
		panic("does not marshal to sortable JSON type (string or number)")
	}
	return b
}
