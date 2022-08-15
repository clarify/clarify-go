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

package clarify

import (
	"context"
	"time"

	"github.com/clarify/clarify-go/fields"
	"github.com/clarify/clarify-go/jsonrpc/resource"
	"github.com/clarify/clarify-go/query"
	"github.com/clarify/clarify-go/views"
)

// DataFrameRequest describes a data frame request.
type DataFrameRequest struct {
	parent resource.SelectRequest[DataFrameResult]
	data   query.Data
}

type DataFrameResult = struct {
	Meta     resource.SelectionMeta `json:"meta"`
	Data     views.DataFrame        `json:"data"`
	Included DataFrameInclude       `json:"included"`
}

// DataFrameInclude describe the included properties for the dataFrame select
// view.
type DataFrameInclude struct {
	Items []views.Item `json:"items"`
}

// Filter returns a new request that includes Items matching the provided
// filter.
func (req DataFrameRequest) Filter(filter query.FilterType) DataFrameRequest {
	req.parent = req.parent.Filter(filter)
	return req
}

// Limit returns a new request that limits the number of matches. Setting n < 0
// will use the max limit.
func (req DataFrameRequest) Limit(n int) DataFrameRequest {
	req.parent = req.parent.Limit(n)
	return req
}

// Skip returns a new request that skips the first n matches.
func (req DataFrameRequest) Skip(n int) DataFrameRequest {
	req.parent = req.parent.Skip(n)
	return req
}

// Sort returns a new request that sorts according to the specified fields. A
// minus (-) prefix can be used on the filed name to indicate inverse ordering.
func (req DataFrameRequest) Sort(fields ...string) DataFrameRequest {
	req.parent = req.parent.Sort(fields...)
	return req
}

// Include returns a new request set to include the specified relationships.
func (req DataFrameRequest) Include(relationships ...string) DataFrameRequest {
	req.parent = req.parent.Include(relationships...)
	return req
}

// TimeRange returns a new request that include data for matching items in the
// specified time range. Note that including data will reduce the maximum number
// of items that can be returned by the response.
func (req DataFrameRequest) TimeRange(gte, lt time.Time) DataFrameRequest {
	req.data.Filter.Times = query.DataTimesRange(gte, lt)
	return req
}

// RollupWindow sets the query to rollup all data into a single timestamp.
func (req DataFrameRequest) RollupWindow() DataFrameRequest {
	req.data.Rollup = "window"
	return req
}

// RollupBucket sets the query to rollup all data into fixed size bucket
// when d > 0. Otherwise, clear the rollup information from the query.
func (req DataFrameRequest) RollupBucket(d time.Duration) DataFrameRequest {
	if d > 0 {
		req.data.Rollup = fields.AsFixedDuration(d).String()
	} else {
		req.data.Rollup = ""
	}
	return req
}

// Last sets the query to only include the n last matching values per item.
func (req DataFrameRequest) Last(n int) DataFrameRequest {
	req.data.Last = n
	return req
}

// Do performs the request against the server and returns the result.
func (req DataFrameRequest) Do(ctx context.Context) (*DataFrameResult, error) {
	return req.parent.Do(ctx, paramData.Value(req.data))
}
