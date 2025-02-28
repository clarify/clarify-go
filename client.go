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

package clarify

import (
	"context"

	"github.com/clarify/clarify-go/fields"
	"github.com/clarify/clarify-go/internal/request"
	"github.com/clarify/clarify-go/jsonrpc"
	"github.com/clarify/clarify-go/views"
)

const (
	apiVersion             = "1.1"
	apiVersionExperimental = "1.2alpha1"

	paramCalculations   jsonrpc.ParamName = "calculations"
	paramData           jsonrpc.ParamName = "data"
	paramFormat         jsonrpc.ParamName = "format"
	paramIntegration    jsonrpc.ParamName = "integration"
	paramItems          jsonrpc.ParamName = "items"
	paramGroups         jsonrpc.ParamName = "groups"
	paramItemsBySignal  jsonrpc.ParamName = "itemsBySignal"
	paramQuery          jsonrpc.ParamName = "query"
	paramSignalsByInput jsonrpc.ParamName = "signalsByInput"
)

// Client allows calling JSON RPC methods against Clarify.
type Client struct {
	ns IntegrationNamespace
}

// NewClient can be used to initialize an integration client from a
// jsonrpc.Handler implementation.
func NewClient(integration string, h jsonrpc.Handler) *Client {
	return &Client{ns: IntegrationNamespace{integration: integration, h: h}}
}

// Insert returns a new request for inserting data to clarify. When referencing
// input IDs that don't exist for the current integration, new signals are
// created automatically on demand.
//
// c.Insert(data) is a short-hand for c.Integration().Insert(data).
func (c Client) Insert(data views.DataFrame) InsertRequest {
	return c.ns.Insert(data)
}

// SaveSignals returns a new request for updating signal meta-data in Clarify.
// When referencing input IDs that don't exist for the current integration, new
// signals are created automatically on demand.
//
// c.SaveSignals(inputs) si a short-hand for:
//
//	c.Integration().SaveSignals(inputs)
func (c Client) SaveSignals(inputs map[string]views.SignalSave) SaveSignalRequest {
	return c.ns.SaveSignals(inputs)
}

// Integration return a handler for initializing methods that require access to
// the integration namespace.
//
// All integrations got access to the integration namespace.
func (c Client) Integration() IntegrationNamespace {
	return c.ns
}

// Admin return a handler for initializing methods that require access to the
// admin namespace.
//
// Access to the admin namespace must be explicitly granted per integration in
// the Clarify admin panel. Do not grant excessive permissions.
func (c Client) Admin() AdminNamespace {
	return AdminNamespace{h: c.ns.h}
}

// Clarify return a handler for initializing methods that require access to
// the clarify namespace.
//
// Access to the clarify namespace must be explicitly granted per integration in
// the Clarify admin panel.  Do not grant excessive permissions.
func (c Client) Clarify() ClarifyNamespace {
	return ClarifyNamespace{h: c.ns.h}
}

type IntegrationNamespace struct {
	integration string

	h jsonrpc.Handler
}

// Insert returns a new request for inserting data to clarify. When referencing
// input IDs that don't exist for the current integration, new signals are
// created automatically on demand.
func (ns IntegrationNamespace) Insert(data views.DataFrame) InsertRequest {
	return methodInsert.NewRequest(ns.h, paramIntegration.Value(ns.integration), paramData.Value(data))
}

type InsertRequest = request.Request[InsertResult]

// InsertResult holds the result of an Insert operation.
type InsertResult struct {
	SignalsByInput map[string]views.CreateSummary `json:"signalsByInput"`
}

var methodInsert = request.Method[InsertResult]{
	APIVersion: apiVersion,
	Method:     "integration.insert",
}

// SaveSignals returns a new request for updating signal meta-data in Clarify.
// When referencing input IDs that don't exist for the current integration, new
// signals are created automatically on demand.
func (ns IntegrationNamespace) SaveSignals(inputs map[string]views.SignalSave) SaveSignalRequest {
	return methodSaveSignals.NewRequest(ns.h,
		paramIntegration.Value(ns.integration),
		paramSignalsByInput.Value(inputs),
	)
}

type (
	// SaveSignalRequest describe an initialized integration.saveSignals RPC
	// request with access to a request handler.
	SaveSignalRequest = request.Request[SaveSignalsResult]

	// SaveSignalsResults describe the result format for a SaveSignalRequest.
	SaveSignalsResult struct {
		SignalsByInput map[string]views.SaveSummary `json:"signalsByInput"`
	}
)

var methodSaveSignals = request.Method[SaveSignalsResult]{
	APIVersion: apiVersion,
	Method:     "integration.saveSignals",
}

type AdminNamespace struct {
	h jsonrpc.Handler
}

func (ns AdminNamespace) Handler() jsonrpc.Handler {
	return ns.h
}

// SelectSignals returns a new request for querying signals and related
// resources.
func (ns AdminNamespace) SelectSignals(integration string, q fields.ResourceQuery) SelectSignalsRequest {
	return methodSelectSignals.NewRequest(ns.h,
		paramIntegration.Value(integration),
		paramQuery.Value(q),
		paramFormat.Value(views.SelectionFormat{
			DataAsArray:         true,
			GroupIncludedByType: true,
		}),
	)
}

type (
	// SelectSignalsRequest describe an initialized admin.selectSignals RPC
	// request with access to a request handler.
	SelectSignalsRequest = request.Relational[SelectSignalsResult]

	// SelectSignalsResult describe the result format for a
	// SelectSignalsRequest.
	SelectSignalsResult = views.Selection[[]views.Signal, views.SignalInclude]
)

var methodSelectSignals = request.RelationalMethod[SelectSignalsResult]{
	APIVersion: apiVersion,
	Method:     "admin.selectSignals",
}

// PublishSignals returns a new request for publishing signals as items.
func (ns AdminNamespace) PublishSignals(integration string, itemsBySignal map[string]views.ItemSave) PublishSignalsRequest {
	return methodPublishSignals.NewRequest(ns.h,
		paramIntegration.Value(integration),
		paramItemsBySignal.Value(itemsBySignal),
	)
}

type (
	// PublishSignalsRequest describe an initialized admin.publishSignal RPC
	// request with access to a request handler.
	PublishSignalsRequest = request.Request[PublishSignalsResult]

	// PublishSignalsResult describe the result format for a
	// PublishSignalsRequest.
	PublishSignalsResult struct {
		ItemsBySignals map[string]views.SaveSummary `json:"itemsBySignal"`
	}
)

var methodPublishSignals = request.Method[PublishSignalsResult]{
	APIVersion: apiVersion,
	Method:     "admin.publishSignals",
}

type ClarifyNamespace struct {
	h jsonrpc.Handler
}

func (ns ClarifyNamespace) Handler() jsonrpc.Handler {
	return ns.h
}

// SelectItems returns a request for querying items.
func (ns ClarifyNamespace) SelectItems(q fields.ResourceQuery) SelectItemsRequest {
	return methodSelectItems.NewRequest(ns.h,
		paramQuery.Value(q),
		paramFormat.Value(views.SelectionFormat{
			DataAsArray:         true,
			GroupIncludedByType: true,
		}),
	)
}

type (
	// SelectItemsRequest describe an initialized clarify.selectItems RPC
	// request with access to a request handler.
	SelectItemsRequest = request.Relational[SelectItemsResult]

	// SelectItemsResult describe the result format for a SelectItemsRequest.
	SelectItemsResult = views.Selection[[]views.Item, views.ItemInclude]
)

var methodSelectItems = request.RelationalMethod[SelectItemsResult]{
	APIVersion: apiVersion,
	Method:     "clarify.selectItems",
}

// DataFrame returns a new request from retrieving raw or aggregated data from
// Clarify. When a data query rollup is specified, data is aggregated using the
// default aggregation methods for each item is used. That is statistical
// aggregation values (count, min, max, sum, avg) for numeric items and a state
// histogram aggregation in seconds (duration spent in each state per bucket)
// for enum items.
func (ns ClarifyNamespace) DataFrame(items fields.ResourceQuery, data fields.DataQuery) DataFrameRequest {
	return methodDataFrame.NewRequest(ns.h,
		paramQuery.Value(items),
		paramData.Value(data),
		paramFormat.Value(views.SelectionFormat{
			GroupIncludedByType: true,
		}),
	)
}

type (
	DataFrameRequest = request.Relational[DataFrameResult]
	DataFrameResult  = views.Selection[views.DataFrame, views.DataFrameInclude]
)

var methodDataFrame = request.RelationalMethod[DataFrameResult]{
	APIVersion: apiVersion,
	Method:     "clarify.dataFrame",
}

// Evaluate returns a new request for retrieving aggregated data from Clarify
// and perform calculations.
func (ns ClarifyNamespace) Evaluate(data fields.DataQuery) EvaluateRequest {
	return EvaluateRequest{
		data: data,
		h:    ns.h,
	}
}

func (er EvaluateRequest) Items(items ...fields.EvaluateItem) EvaluateRequest {
	newItems := make([]fields.EvaluateItem, 0, len(er.items)+len(items))
	newItems = append(append(newItems, er.items...), items...)
	er.items = newItems

	return er
}

func (er EvaluateRequest) Groups(groups ...fields.EvaluateGroup) EvaluateRequest {
	newGroups := make([]fields.EvaluateGroup, 0, len(er.groups)+len(groups))
	newGroups = append(append(newGroups, er.groups...), groups...)
	er.groups = newGroups

	return er
}

func (er EvaluateRequest) Calculations(calculations ...fields.Calculation) EvaluateRequest {
	newCalculations := make([]fields.Calculation, 0, len(er.calculations)+len(calculations))
	newCalculations = append(append(newCalculations, er.calculations...), calculations...)
	er.calculations = newCalculations

	return er
}

func (er EvaluateRequest) Include(relationships ...string) EvaluateRequest {
	newRelationships := make([]string, 0, len(er.relationships)+len(relationships))
	newRelationships = append(append(newRelationships, er.relationships...), relationships...)
	er.relationships = newRelationships

	return er
}

func (er EvaluateRequest) Format(format views.SelectionFormat) EvaluateRequest {
	er.format = format

	return er
}

func (er EvaluateRequest) Do(ctx context.Context) (*EvaluateResult, error) {
	r := methodEvaluate.NewRequest(er.h,
		paramData.Value(er.data),
		paramItems.Value(er.items),
		paramGroups.Value(er.groups),
		paramCalculations.Value(er.calculations),
		paramFormat.Value(er.format)).
		Include(er.relationships...)

	return r.Do(ctx)
}

// EvaluateRequest describe an initialized clarify.evaluate RPC request with
// access to a request handler.
type EvaluateRequest struct {
	data          fields.DataQuery
	items         []fields.EvaluateItem
	groups        []fields.EvaluateGroup
	calculations  []fields.Calculation
	relationships []string
	format        views.SelectionFormat
	h             jsonrpc.Handler
}

// EvaluateResult describe the result format for a EvaluateRequest.
type EvaluateResult = views.Selection[views.DataFrame, views.DataFrameInclude]

var methodEvaluate = request.RelationalMethod[EvaluateResult]{
	APIVersion: apiVersionExperimental,
	Method:     "clarify.evaluate",
}
