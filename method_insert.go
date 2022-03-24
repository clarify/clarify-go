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
