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

package clarify

import (
	"github.com/clarify/clarify-go/jsonrpc"
	"github.com/clarify/clarify-go/jsonrpc/resource"
	"github.com/clarify/clarify-go/views"
)

const (
	paramIntegration jsonrpc.ParamName = "integration"
	paramData        jsonrpc.ParamName = "data"
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
func (c *Client) Insert(data views.DataFrame) resource.Request[InsertResult] {
	return methodInsert.NewRequest(c.h, paramIntegration.Value(c.integration), paramData.Value(data))
}

// InsertResult holds the result of an Insert operation.
type InsertResult struct {
	SignalsByInput map[string]resource.CreateSummary `json:"signalsByInput"`
}

var methodInsert = resource.Method[InsertResult]{
	APIVersion: "1.1rc1",
	Method:     "integration.insert",
}

// SaveSignals returns a new save signals request that can be modified though a
// chainable API before it's executed. Keys in inputs are scoped to the current
// integration. Requires access to the integration namespace.
func (c *Client) SaveSignals(inputs map[string]views.SignalSave) resource.SaveRequest[map[string]views.SignalSave, SaveSignalsResult] {
	return methodSaveSignals.NewRequest(c.h, inputs, paramIntegration.Value(c.integration))
}

// SaveSignalsResults holds the result of a SaveSignals operation.
type SaveSignalsResult struct {
	SignalsByInput map[string]resource.SaveSummary `json:"signalsByInput"`
}

var methodSaveSignals = resource.SaveMethod[map[string]views.SignalSave, SaveSignalsResult]{
	APIVersion: "1.1rc1",
	Method:     "integration.saveSignals",
	DataParam:  "inputs",
}

// PublishSignals returns a new request for publishing signals as Items.
// Requires access to the admin namespace.
//
// Disclaimer: this method is based on a pre-release of the Clarify API, and
// might be unstable or stop working.
func (c *Client) PublishSignals(integration string, itemsBySignal map[string]views.ItemSave) resource.SaveRequest[map[string]views.ItemSave, PublishSignalsResult] {
	return methodPublishSignals.NewRequest(c.h, itemsBySignal, paramIntegration.Value(integration))
}

// PublishSignalsResult holds the result of a PublishSignals operation.
type PublishSignalsResult struct {
	ItemsBySignals map[string]resource.SaveSummary `json:"itemsBySignal"`
}

var methodPublishSignals = resource.SaveMethod[map[string]views.ItemSave, PublishSignalsResult]{
	APIVersion: "1.1rc1",
	Method:     "admin.publishSignals",
	DataParam:  "itemsBySignal",
}

// SelectSignals returns a new request for querying signals and related
// resources.
//
// Disclaimer: this method is based on a pre-release of the Clarify API, and
// might be unstable or stop working.
func (c *Client) SelectSignals(integration string) resource.SelectRequest[SelectSignalsResult] {
	return methodSelectSignals.NewRequest(c.h, paramIntegration.Value(integration))
}

// SelectSignalsResult holds the result of a SelectSignals operation.
type SelectSignalsResult = resource.Selection[views.Signal, views.SignalInclude]

var methodSelectSignals = resource.SelectMethod[SelectSignalsResult]{
	APIVersion: "1.1rc1",
	Method:     "admin.selectSignals",
}

// SelectItems returns a new request for querying signals and related
// resources.
//
// Disclaimer: this method is based on a pre-release of the Clarify API, and
// might be unstable or stop working.
func (c *Client) SelectItems() resource.SelectRequest[SelectItemsResult] {
	return methodSelectItems.NewRequest(c.h)
}

// SelectItemsResult holds the result of a SelectItems operation.
type SelectItemsResult = resource.Selection[views.Item, views.ItemInclude]

var methodSelectItems = resource.SelectMethod[SelectItemsResult]{
	APIVersion: "1.1rc1",
	Method:     "clarify.selectItems",
}

// DataFrame returns a new request from retrieving data from clarify. The
// request can be furthered modified by a chainable API before it's executed. By
// default, the data section is set to be included in the response.
//
// Disclaimer: this method is based on a pre-release of the Clarify API, and
// might be unstable or stop working.
func (c *Client) DataFrame() DataFrameRequest {
	return DataFrameRequest{
		parent: methodDataFrame.NewRequest(c.h),
	}
}

var methodDataFrame = resource.SelectMethod[DataFrameResult]{
	APIVersion: "1.1rc1",
	Method:     "clarify.dataFrame",
}
