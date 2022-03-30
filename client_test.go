package clarify_test

import (
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
	if err := json.Unmarshal(resp.rawResult, result); err != nil {
		return fmt.Errorf("%w: %v", jsonrpc.ErrBadResponse, err)
	}
	return nil
}

type mockRPCResponse struct {
	err       *jsonrpc.Error
	rawResult json.RawMessage
}
