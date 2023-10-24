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
	"encoding/json"
)

const defaultQueryLimit = 50

type resourceQuery struct {
	Filter ResourceFilter `json:"filter,omitempty"`
	Sort   []string       `json:"sort,omitempty"`
	Limit  int            `json:"limit"`
	Skip   int            `json:"skip"`
	Total  bool           `json:"total"`
}

// ResourceQuery holds a resource fields. Although it does not expose any
// fields, the type can be decoded from and encoded to JSON.
type ResourceQuery struct {
	limitSet bool
	query    resourceQuery
}

var (
	_ json.Marshaler   = ResourceQuery{}
	_ json.Unmarshaler = (*ResourceQuery)(nil)
)

// Query returns a new resource query which joins the passed in filters with
// logical AND.
func Query() ResourceQuery {
	return ResourceQuery{
		query: resourceQuery{},
	}
}

func (q *ResourceQuery) UnmarshalJSON(data []byte) error {
	q.limitSet = true
	q.query = resourceQuery{Limit: defaultQueryLimit}
	return json.Unmarshal(data, &q.query)
}

func (q ResourceQuery) MarshalJSON() ([]byte, error) {
	if !q.limitSet {
		q.query.Limit = defaultQueryLimit
	}
	return json.Marshal(q.query)
}

// Where returns a new query with the given where condition added to existing
// query non-empty filters with logical AND.
//
// See the API reference documentation to determine which fields are filterable
// for each resource: https://docs.clarify.io/api/1.1/types/resources.
func (q ResourceQuery) Where(filter ResourceFilterType) ResourceQuery {
	q.query.Filter = And(q.query.Filter, filter)
	return q
}

// Sort returns a new query that sorts results using the provided fields. To get
// descending sort, prefix the field with a minus (-).
//
// See the API reference documentation to determine which fields are sortable
// for each resource: https://docs.clarify.io/api/1.1/types/resources.
func (q ResourceQuery) Sort(fields ...string) ResourceQuery {
	sort := make([]string, 0, len(q.query.Sort)+len(fields))
	sort = append(sort, q.query.Sort...)
	sort = append(sort, fields...)
	q.query.Sort = sort
	return q
}

// Skip returns a query that skips the first n entries matching the fields.
func (q ResourceQuery) Skip(n int) ResourceQuery {
	q.query.Skip = n
	return q
}

// GetSkip returns the query skip value.
func (q ResourceQuery) GetSkip() int {
	return q.query.Skip
}

// Limit returns a new query that limits the number of results to n. Set limit
// to -1 to use the maximum allowed value.
func (q ResourceQuery) Limit(n int) ResourceQuery {
	q.limitSet = true
	q.query.Limit = n
	return q
}

// GetLimit returns the query limit value.
func (q ResourceQuery) GetLimit() int {
	if !q.limitSet {
		return defaultQueryLimit
	}
	return q.query.Limit
}

// NextPage returns a new query where the skip value is incremented by the query
// limit value.
func (q ResourceQuery) NextPage() ResourceQuery {
	q.query.Skip += q.GetLimit()
	return q
}

// Total returns a query that forces the inclusion of a total count in the
// response when force is true, or includes it only if it can be calculated for
// free if force is false.
func (q ResourceQuery) Total(force bool) ResourceQuery {
	q.query.Total = force
	return q
}
