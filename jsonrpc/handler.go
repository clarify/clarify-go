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

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const headerAPIVersion = "X-API-Version"
const defaultAPIVersion = "1.0"

// Handler describe the interface for handling arbitrary RPC requests.
type Handler interface {
	Do(ctx context.Context, req Request, result any) error
}

// HTTPHandler performs RPC requests via HTTP POST against the specified URL.
type HTTPHandler struct {
	Client http.Client
	URL    string
}

// Do sends the passed in request to the server, and decodes the result or error
// from the response. Result must be a pointer.
func (c *HTTPHandler) Do(ctx context.Context, req Request, result any) error {
	body, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrBadRequest, err)
	}
	httpReq, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		c.URL,
		bytes.NewReader(body),
	)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrBadRequest, err)
	}
	defer appendOnError(&err, httpReq.Body.Close, "; ")

	httpReq.Header.Set(headerAPIVersion, req.APIVersion)
	httpReq.Header.Set("Content-Type", "application/json")
	httpResp, err := c.Client.Do(httpReq)
	if err != nil {
		return err
	}
	defer appendOnError(&err, httpResp.Body.Close, "; ")

	if httpResp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(httpResp.Body)
		return HTTPError{
			StatusCode: httpResp.StatusCode,
			Headers:    httpResp.Header,
			Body:       string(b),
		}
	}
	resp := rpcResponse{
		Result:     result,
		APIVersion: httpResp.Header.Get(headerAPIVersion),
	}
	dec := json.NewDecoder(httpResp.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&resp); err != nil {
		return fmt.Errorf("%w: %v", ErrBadResponse, err)
	}
	if resp.JSONRPC != "2.0" {
		return fmt.Errorf(`%w: jsonrpc must be "2.0"`, ErrBadResponse)
	}
	if resp.ID != req.ID {
		return fmt.Errorf(`%w: id must match request`, ErrBadResponse)
	}
	if err := resp.Error; err != nil {
		return err
	}
	return nil
}

type rpcResponse struct {
	JSONRPC string       `json:"jsonrpc"`
	Error   *ServerError `json:"error"`
	ID      int          `json:"id"`
	Result  any          `json:"result"`

	// Transport layer parameters.
	APIVersion string `json:"-"`
}
