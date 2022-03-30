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

	"github.com/clarify/clarify-go/data"
	"github.com/clarify/clarify-go/jsonrpc"
)

// InsertResult describe the effect of a save operation.
type InsertResult struct {
	SignalsByInput map[string]InsertSummary `json:"signalsByInput"`
}

// InsertSummary reference the signal where data was inserted and
// weather or not the signal was created.
type InsertSummary struct {
	ID      string `json:"id"`
	Created bool   `json:"created"`
}

type InsertRequest struct {
	integration string
	data        data.Frame

	h jsonrpc.Handler
}

// Do performs the request against the server and returns the result.
func (req InsertRequest) Do(ctx context.Context) (*InsertResult, error) {
	var res InsertResult
	rpcReq := jsonrpc.NewRequest(
		methodInsert,
		[]any{req.integration, req.data},
	)
	if err := req.h.Do(ctx, rpcReq, &res); err != nil {
		return nil, err
	}
	return &res, nil
}
