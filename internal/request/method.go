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

package request

import (
	"context"

	"github.com/clarify/clarify-go/jsonrpc"
)

// Method is a constructor for an RPC request for a specific RPC method and API
// version.
type Method[R any] struct {
	APIVersion string
	Method     string
}

func (cfg Method[R]) NewRequest(h jsonrpc.Handler, params ...jsonrpc.Param) Request[R] {
	return Request[R]{
		apiVersion: cfg.APIVersion,
		method:     cfg.Method,

		baseParams: params,
		h:          h,
	}
}

// Request describe an initialized RPC request with access to a request handler.
type Request[R any] struct {
	apiVersion string
	method     string

	baseParams []jsonrpc.Param

	h jsonrpc.Handler
}

// Do performs the request against the server and returns the result.
func (req Request[R]) Do(ctx context.Context) (*R, error) {
	return req.do(ctx)
}

func (req Request[R]) do(ctx context.Context, params ...jsonrpc.Param) (*R, error) {
	allParams := make([]jsonrpc.Param, 0, len(req.baseParams)+len(params))
	allParams = append(allParams, req.baseParams...)
	allParams = append(allParams, params...)

	rpcReq := jsonrpc.NewRequest(req.method, allParams...)
	if req.apiVersion != "" {
		rpcReq.APIVersion = req.apiVersion
	}

	var res R
	if err := req.h.Do(ctx, rpcReq, &res); err != nil {
		return nil, err
	}
	return &res, nil
}
