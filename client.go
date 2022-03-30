package clarify

import (
	"github.com/clarify/clarify-go/data"
	"github.com/clarify/clarify-go/jsonrpc"
)

const (
	// integration namespace methods.
	methodInsert      = "integration.insert"
	methodSaveSignals = "integration.saveSignals"
)

// Client allows calling JSON RPC methods against Clarify.
type Client struct {
	integration string

	h jsonrpc.Handler
}

// NewClient can be used to initialize an integration client from a
// jsonrpc.Handler implementation.
func NewClient(integration string, h jsonrpc.Handler) *Client {
	return &Client{integration: integration, h: h}
}

// Insert returns a new insert request that can be executed at will. Requires
// access to the integration namespace. Will insert the data to the integration
// set in c.
func (c *Client) Insert(data data.Frame) InsertRequest {
	return InsertRequest{
		integration: c.integration,
		data:        data,
		h:           c.h,
	}
}

// SaveSignals returns a new save signals request that can be modifed though a
// chainable API before it's executed. Keys in inputs are scoped to the current
// integration. Requires access to the integration namespace.
func (c *Client) SaveSignals(inputs map[string]SignalSave) SaveSignalsRequest {
	return SaveSignalsRequest{
		method:     methodSaveSignals,
		entryParam: "inputs",
		contextParams: map[string]any{
			"integration": c.integration,
		},
		entries: inputs,
		h:       c.h,
	}
}

type SaveSignalsRequest = KeyedSaveRequest[SignalSave, SaveSignalsResult]
type SaveSignalsResult struct {
	SignalsByInput map[string]SaveSummary `json:"signalsByInput"`
}
