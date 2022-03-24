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

package clarify_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/clarify/clarify-go/jsonrpc"
)

const exampleTrace = "123"

type mockRPCHandler map[string]mockRPCResponse

func (m mockRPCHandler) Do(ctx context.Context, req jsonrpc.Request, result any) error {
	resp, ok := m[strings.ToLower(req.Method)]
	if !ok {
		return &jsonrpc.Error{
			Code:    jsonrpc.CodeMethodNotFound,
			Message: "Method not found",
			Data: jsonrpc.ErrorData{
				Trace: exampleTrace,
			},
		}
	}
	if resp.err != nil {
		return resp.err
	}
	dec := json.NewDecoder(bytes.NewReader(resp.rawResult))
	// DisallowUnknownFields is useful for discovering issues in testdata or
	// models; production clients should not use it.
	dec.DisallowUnknownFields()
	if err := dec.Decode(result); err != nil {
		return fmt.Errorf("%w: %v", jsonrpc.ErrBadResponse, err)
	}
	return nil
}

type mockRPCResponse struct {
	err       *jsonrpc.Error
	rawResult json.RawMessage
}
