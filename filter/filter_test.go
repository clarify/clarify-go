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

package filter_test

import (
	"testing"

	"github.com/clarify/clarify-go/filter"
)

func TestFilter(t *testing.T) {
	testStringer := func(f filter.Filter, expect string) func(t *testing.T) {
		return func(t *testing.T) {
			t.Helper()
			if result := f.String(); result != expect {
				t.Errorf("unexpected filter.String() value:\n got: %s\nwant: %s",
					result,
					expect,
				)
			}
		}
	}

	t.Run(`filter.Filter{}`, testStringer(
		filter.Filter{},
		`{}`,
	))
	t.Run(`filter.And(filter.Filter{},filter.Field("id",filter.Equal("a")))`, testStringer(
		filter.And(filter.Filter{}, filter.Field("id", filter.Equal("a"))),
		`{"id":{"$in":["a"]}}`,
	))
	t.Run(`filter.And(filter.Filter{},filter.Field("id",filter.In("a")))`, testStringer(
		filter.And(filter.Filter{}, filter.Field("id", filter.In("a"))),
		`{"id":{"$in":["a"]}}`,
	))
	t.Run(`filter.And(filter.Filter{},filter.Field("id",filter.In("a","b")))`, testStringer(
		filter.And(filter.Filter{}, filter.Field("id", filter.In("a", "b"))),
		`{"id":{"$in":["a","b"]}}`,
	))
}
