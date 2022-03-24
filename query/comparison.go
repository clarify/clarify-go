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
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
)

// Comparisons maps field paths joined by dot to a comparison.
type Comparisons map[string]Comparison

// Comparison allows comparing a particular value with one or more operators.
// The zero-value is treated equivalent to Equal(null).
type Comparison struct {
	value *opComparison
}

var (
	_ interface {
		json.Marshaler
		fmt.Stringer
	} = Comparison{}
	_ json.Unmarshaler = (*Comparison)(nil)
)

func (cmp Comparison) String() string {
	b, _ := json.Marshal(cmp)
	return string(b)
}

// MultiOperator merges multiple comparisons with different operators together
// to a single comparison entry.
//
// When conflicting operators ar encountered, the right most value is selected
// for the result. Operators are resolved based on operator keys, where some
// initializers have an overlap:
//
//    - Equal and In both users $in.
//    - NotEqual and NotIn both users $nin.
//    - Range and GreaterThanOrEqual both uses $gte.
//    - Range and LessThan both uses $lt.
//
// Example valid usage:
//
//    MultiOperator(Equal(nil), LessThan(49))        // {"$in":[nil],"$lt":49}
//    MultiOperator(GreaterOrEqual(0), LessThan(49)) // {"$gte":0,"$lt":49}
//
// Example of conflicting operators:
//
//    MultiOperator(Equal(0), In(1, 2))       // {"$in":[1,2]}
//    MultiOperator(In(1, 2), Equal(nil))     // null
//    MultiOperator(NotIn(1, 2), NotEqual(0)) // {"$nin":[0]}
//    MultiOperator(NotEqual(0), NotIn(1, 2)) // {"$nin":[1,2]}
func MultiOperator(cmps ...Comparison) Comparison {
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
			if v.GreaterThan != nil {
				target.GreaterThan = v.GreaterThan
			}
			if v.GreaterThanOrEqual != nil {
				target.GreaterThanOrEqual = v.GreaterThanOrEqual
			}
			if v.LessThan != nil {
				target.LessThan = v.LessThan
			}
			if v.LessThanOrEqual != nil {
				target.LessThanOrEqual = v.LessThanOrEqual
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
	In                 []json.RawMessage `json:"$in,omitempty"`
	NotIn              []json.RawMessage `json:"$nin,omitempty"`
	GreaterThan        json.RawMessage   `json:"$gt,omitempty"`
	GreaterThanOrEqual json.RawMessage   `json:"$gte,omitempty"`
	LessThan           json.RawMessage   `json:"$lt,omitempty"`
	LessThanOrEqual    json.RawMessage   `json:"$lte,omitempty"`
	Regex              string            `json:"$regex,omitempty"`
}

func (cmp *opComparison) normalize() *opComparison {
	// Normalizes the following:
	//
	//   - opComparison{In:{"null"}} -> nil
	//   - opComparison{} -> nil
	//
	isEmptyExceptIn := (cmp.NotIn == nil &&
		cmp.GreaterThan == nil &&
		cmp.GreaterThanOrEqual == nil &&
		cmp.LessThan == nil &&
		cmp.LessThanOrEqual == nil &&
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
// JSON marshalable into a simple JSON type (string, number, bool or null).
func Equal(v any) Comparison {
	return Comparison{
		value: (&opComparison{In: []json.RawMessage{simpleJSON(v)}}).normalize(),
	}
}

// NotEqual returns a comparison that match values not equal to v. Panics if v
// is not JSON marshalable into a simple JSON type (string, number, bool or
// null).
func NotEqual(v any) Comparison {
	return Comparison{
		value: &opComparison{NotIn: []json.RawMessage{simpleJSON(v)}},
	}
}

// In returns a comparison that match values in elements. Panics if any element
// is not JSON marshalable into an a simple JSON type (string, number, bool or
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
// element is not JSON marshalable into an a simple JSON type (string, number,
// bool or null).
func NotIn[E any](elements ...E) Comparison {
	nin := make([]json.RawMessage, 0, len(elements))
	for _, elem := range elements {
		nin = append(nin, simpleJSON(elem))
	}
	return Comparison{
		value: &opComparison{In: nin},
	}
}

// GreaterThan returns a comparison that matches values > gte. Panics if gt is
// not JSON marshalable into an a sortable JSON type (string or number).
func GreaterThan(gt any) Comparison {
	return Comparison{
		value: &opComparison{GreaterThan: orderedJSONType(gt)},
	}
}

// GreaterThanOrEqual returns a comparison that matches values >= gte. Panics if
// gte is not JSON marshalable into an a sortable JSON type (string or number).
func GreaterThanOrEqual(gte any) Comparison {
	return Comparison{
		value: &opComparison{GreaterThanOrEqual: orderedJSONType(gte)},
	}
}

// LessThan returns a comparison that matches values < lt. Panics if lt is not
// JSON marshalable into an a sortable JSON type (string or number).
func LessThan(lt any) Comparison {
	return Comparison{
		value: &opComparison{LessThan: orderedJSONType(lt)},
	}
}

// LessThanOrEqual returns a comparison that matches values <= lte. Panics if
// lte is not JSON marshalable into an a sortable JSON type (string or number).
func LessThanOrEqual(lte any) Comparison {
	return Comparison{
		value: &opComparison{LessThanOrEqual: orderedJSONType(lte)},
	}
}

// Range is a short-hand for:
//
//     MergeComparisons(GreaterThanOrEqual(gte), LessThan(lt))
func Range(gte, lt any) Comparison {
	return Comparison{
		value: &opComparison{
			GreaterThanOrEqual: orderedJSONType(gte),
			LessThan:           orderedJSONType(lt),
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
		// got object, treat as operator comparison.
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
