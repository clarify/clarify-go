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
)

const (
	paramCreateOnly jsonrpc.ParamName = "createOnly"
)

// SaveMethod is a constructor for Requests against a given RPC method.
type SaveMethod[D, R any] struct {
	APIVersion string
	Method     string
	DataParam  string
}

func (cfg SaveMethod[D, R]) NewRequest(h jsonrpc.Handler, data D, params ...jsonrpc.Param) SaveRequest[D, R] {
	return SaveRequest[D, R]{
		apiVersion: cfg.APIVersion,
		method:     cfg.Method,
		dataParam:  jsonrpc.ParamName(cfg.DataParam),
		data:       data,
		baseParams: params,
		h:          h,
	}
}

// SaveRequest allows creating or updating properties based on a keyed
// relation.
type SaveRequest[D, R any] struct {
	apiVersion string
	method     string
	dataParam  jsonrpc.ParamName

	baseParams []jsonrpc.Param
	data       D
	createOnly bool

	h jsonrpc.Handler
}

// CreateOnly returns a request with the createOnly property set to true. When
// set to true, existing resources are not updated.
func (req SaveRequest[D, R]) CreateOnly() SaveRequest[D, R] {
	req.createOnly = true
	return req
}

// Do performs the request against the server and returns the result.
func (req SaveRequest[D, R]) Do(ctx context.Context, extraParams ...jsonrpc.Param) (*R, error) {
	params := make([]jsonrpc.Param, 0, len(req.baseParams)+2+len(extraParams))
	params = append(params, req.baseParams...)
	params = append(params,
		req.dataParam.Value(req.data),
		paramCreateOnly.Value(req.createOnly),
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
