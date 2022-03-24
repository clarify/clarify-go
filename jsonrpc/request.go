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

package jsonrpc

// ParamName can be used to define a parameter name.
type ParamName string

func (name ParamName) Value(v any) Param {
	return Param{
		Name:  string(name),
		Value: v,
	}
}

// Param described a named parameter.
type Param struct {
	Name  string
	Value any
}

// Request describe the structure of an RPC Request. The request should be
// initialized via NewRequest.
type Request struct {
	JSONRPC string `json:"jsonrpc"`
	Method  string `json:"method"`
	ID      int    `json:"id"`
	Params  any    `json:"params"`

	// Transport layer parameters.
	APIVersion string `json:"-"`
}

func NewRequest(method string, params ...Param) Request {
	m := make(map[string]any, len(params))
	for _, param := range params {
		m[param.Name] = param.Value
	}
	return Request{
		JSONRPC:    "2.0",
		Method:     method,
		ID:         1,
		Params:     m,
		APIVersion: defaultAPIVersion,
	}
}
