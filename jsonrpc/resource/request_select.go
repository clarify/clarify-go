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

package resource

import (
	"context"

	"github.com/clarify/clarify-go/jsonrpc"
	"github.com/clarify/clarify-go/query"
)

const (
	paramQuery               jsonrpc.ParamName = "query"
	paramInclude             jsonrpc.ParamName = "include"
	paramGroupIncludedByType jsonrpc.ParamName = "groupIncludedByType"
)

type SelectMethod[R any] struct {
	APIVersion string
	Method     string
}

func (cfg SelectMethod[R]) NewRequest(h jsonrpc.Handler, params ...jsonrpc.Param) SelectRequest[R] {
	return SelectRequest[R]{
		apiVersion: cfg.APIVersion,
		method:     cfg.Method,
		query:      query.New(),
		baseParams: params,

		h: h,
	}
}

type SelectRequest[R any] struct {
	apiVersion string
	method     string

	baseParams []jsonrpc.Param
	query      query.Query
	includes   []string

	h jsonrpc.Handler
}

// Filter returns a new request with the specified filter added to existing
// filters with logical AND.
func (req SelectRequest[R]) Filter(f query.FilterType) SelectRequest[R] {
	req.query.Filter = query.And(req.query.Filter, f)
	return req
}

// Limit returns a new request that limits the number of matches. Setting n < 0
// will use the default limit.
func (req SelectRequest[R]) Limit(n int) SelectRequest[R] {
	req.query.Limit = n
	return req
}

// Skip returns a new request that skips the first n matches.
func (req SelectRequest[R]) Skip(n int) SelectRequest[R] {
	req.query.Skip = n
	return req
}

// Sort returns a new request that sorts according to the specified fields. A
// minus (-) prefix can be used on the filed name to indicate inverse ordering.
func (req SelectRequest[R]) Sort(fields ...string) SelectRequest[R] {
	req.query.Sort = fields
	return req
}

// Total returns a new request that includes a total count of matches in the
// result.
func (req SelectRequest[R]) Total() SelectRequest[R] {
	req.query.Total = true
	return req
}

// Include returns a new request that includes the specified related resources.
// the provided list is appended to any existing include properties.
func (req SelectRequest[R]) Include(relationships ...string) SelectRequest[R] {
	a := make([]string, 0, len(req.includes)+len(relationships))
	a = append(a, req.includes...)
	a = append(a, relationships...)
	req.includes = a
	return req
}

// Do performs the request against the server and returns the result.
func (req SelectRequest[R]) Do(ctx context.Context, extraParams ...jsonrpc.Param) (*R, error) {
	params := make([]jsonrpc.Param, 0, len(req.baseParams)+3+len(extraParams))
	params = append(params, req.baseParams...)
	params = append(params,
		paramQuery.Value(req.query),
		paramInclude.Value(req.includes),
		paramGroupIncludedByType.Value(true),
	)
	params = append(params, extraParams...)

	rpcReq := jsonrpc.NewRequest(req.method, params...)
	if req.apiVersion != "" {
		rpcReq.APIVersion = req.apiVersion
	}

	var res R
	if err := req.h.Do(ctx, rpcReq, &res); err != nil {
		return nil, err
	}
	return &res, nil
}
