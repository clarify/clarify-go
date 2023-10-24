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

package fields_test

import (
	"fmt"
	"testing"

	"github.com/clarify/clarify-go/fields"
)

func TestFilter(t *testing.T) {
	testStringer := func(f fields.ResourceFilter, expect string) func(t *testing.T) {
		return func(t *testing.T) {
			t.Helper()
			if result := fmt.Sprint(f); result != expect {
				t.Errorf("unexpected fields.String() value:\n got: %s\nwant: %s",
					result,
					expect,
				)
			}
		}
	}

	t.Run(`queryAll()`, testStringer(
		fields.FilterAll(),
		`{}`,
	))
	t.Run(`fields.And(fields.All(),fields.Field("id",fields.Equal("a")))`, testStringer(
		fields.And(fields.FilterAll(), fields.CompareField("id", fields.Equal("a"))),
		`{"id":{"$in":["a"]}}`, // Optimized to skip All fields.
	))
	t.Run(`fields.And(fields.All(),fields.Field("id",fields.In("a","b")))`, testStringer(
		fields.And(fields.FilterAll(), fields.CompareField("id", fields.In("a", "b"))),
		`{"id":{"$in":["a","b"]}}`, // Optimized to skip All fields.
	))
	t.Run(`fields.Or(fields.All(),fields.Field("id",fields.Equal("a")))`, testStringer(
		fields.Or(fields.FilterAll(), fields.CompareField("id", fields.Equal("a"))),
		`{}`, // Optimized to empty query (match all).
	))
}
