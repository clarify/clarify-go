// Copyright 2023 Searis AS
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

package request

import (
	"context"

	"github.com/clarify/clarify-go/jsonrpc"
)

const (
	includeParam jsonrpc.ParamName = "include"
)

// RelationalMethod is a constructor for an RPC request for a specific RPC
// method and API version where named relationships can be included.
type RelationalMethod[R any] struct {
	APIVersion string
	Method     string
}

func (cfg RelationalMethod[R]) NewRequest(h jsonrpc.Handler, params ...jsonrpc.Param) Relational[R] {
	return Relational[R]{
		parent: Request[R]{
			apiVersion: cfg.APIVersion,
			method:     cfg.Method,

			baseParams: params,
			h:          h,
		},
	}
}

// Relational describe an initialized RPC request with access to a request
// handler and the option to include related resources.
type Relational[R any] struct {
	parent  Request[R]
	include []string
}

// Include returns a request that appends the named relationships to the
// request include list.
func (req Relational[R]) Include(relationships ...string) Relational[R] {
	include := make([]string, 0, len(req.include)+len(relationships))
	include = append(include, req.include...)
	include = append(include, relationships...)
	req.include = include
	return req
}

// Do performs the request against the server and returns the result.
func (req Relational[R]) Do(ctx context.Context) (*R, error) {
	return req.parent.do(ctx, includeParam.Value(req.include))
}
