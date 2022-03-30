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
