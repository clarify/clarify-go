// Copyright 2022-2025 Searis AS
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

package clarifyx

import (
	"github.com/clarify/clarify-go"
	"github.com/clarify/clarify-go/fields"
	"github.com/clarify/clarify-go/internal/request"
	"github.com/clarify/clarify-go/jsonrpc"
	"github.com/clarify/clarify-go/views"
)

const (
	apiVersion             = "1.1"
	apiVersionExperimental = "1.2alpha1"

	paramFormat      jsonrpc.ParamName = "format"
	paramIntegration jsonrpc.ParamName = "integration"
	paramItem        jsonrpc.ParamName = "item"
	paramQuery       jsonrpc.ParamName = "query"
)

type Client struct {
	clarify.Client
}

func Upgrade(c *clarify.Client) Client {
	return Client{Client: *c}
}

func (c Client) Admin() AdminNamespace {
	return AdminNamespace{AdminNamespace: c.Client.Admin()}
}

type AdminNamespace struct {
	clarify.AdminNamespace
}

// ConnectSignals returns a new request for publishing signals as items.
func (ns AdminNamespace) ConnectSignals(integration, item string, q fields.ResourceQuery) ConnectSignalsRequest {
	return methodConnectSignals.NewRequest(ns.Handler(),
		paramIntegration.Value(integration),
		paramItem.Value(item),
		paramQuery.Value(q),
		paramFormat.Value(views.SelectionFormat{
			DataAsArray:         true,
			GroupIncludedByType: true,
		}),
	)
}

type (
	// ConnectSignalsRequest describe an initialized admin.publishSignal RPC
	// request with access to a request handler.
	ConnectSignalsRequest = request.Request[ConnectSignalsResult]

	// ConnectSignalsResult describe the result format for a
	// ConnectSignalsRequest.
	ConnectSignalsResult = views.Selection[[]views.Signal, views.SignalInclude]
)

var methodConnectSignals = request.Method[ConnectSignalsResult]{
	APIVersion: apiVersionExperimental,
	Method:     "admin.connectSignals",
}

// DisconnectSignals returns a new request for publishing signals as items.
func (ns AdminNamespace) DisconnectSignals(integration string, q fields.ResourceQuery) DisconnectSignalsRequest {
	return methodDisconnectSignals.NewRequest(ns.Handler(),
		paramIntegration.Value(integration),
		paramQuery.Value(q),
		paramFormat.Value(views.SelectionFormat{
			DataAsArray:         true,
			GroupIncludedByType: true,
		}),
	)
}

type (
	// DisconnectSignalsRequest describe an initialized admin.publishSignal RPC
	// request with access to a request handler.
	DisconnectSignalsRequest = request.Request[DisconnectSignalsResult]

	// DisconnectSignalsResult describe the result format for a
	// DisconnectSignalsRequest.
	DisconnectSignalsResult = views.Selection[[]views.Signal, views.SignalInclude]
)

var methodDisconnectSignals = request.Method[DisconnectSignalsResult]{
	APIVersion: apiVersionExperimental,
	Method:     "admin.disconnectSignals",
}
