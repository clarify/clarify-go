package clarify

import (
	"context"

	"github.com/clarify/clarify-go/jsonrpc"
)

// SaveSummary describe the effect of a save operation.
type SaveSummary struct {
	ID      string `json:"id"`
	Created bool   `json:"created"`
	Updated bool   `json:"updated"`
}

// KeyedSaveRequest allows creating or updating properties based on a keyed
// relation.
type KeyedSaveRequest[E, R any] struct {
	method     string
	entryParam string

	contextParams map[string]any
	entries       map[string]E
	createOnly    bool

	h jsonrpc.Handler
}

// CreateOnly returns a request with the createOnly property set to true. When
// set to true, existing resources are not updated.
func (req KeyedSaveRequest[E, R]) CreateOnly() KeyedSaveRequest[E, R] {
	req.createOnly = true
	return req
}

// Do performs the request against the server and returns the result.
func (req KeyedSaveRequest[E, R]) Do(ctx context.Context) (*R, error) {
	var res R
	params := make(map[string]any, len(req.contextParams)+3)
	for k, v := range req.contextParams {
		params[k] = v
	}
	params[req.entryParam] = req.entries
	params["createOnly"] = req.createOnly
	rpcReq := jsonrpc.NewRequest(
		req.method,
		params,
	)
	if err := req.h.Do(ctx, rpcReq, &res); err != nil {
		return nil, err
	}
	return &res, nil
}
