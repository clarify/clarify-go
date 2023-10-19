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

package jsonrpc

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"runtime/debug"
	"time"

	"golang.org/x/oauth2"
)

const (
	headerAPIVersion  = "X-API-Version"
	defaultAPIVersion = "1.0"
)

var userAgent = "clarify-go/unknown"

func init() {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return
	}
	for _, dep := range info.Deps {
		if dep.Path == "github.com/clarify/clarify-go" {
			userAgent = "clarify-go/" + dep.Version
		}
	}
}

// Handler describe the interface for handling arbitrary RPC requests.
type Handler interface {
	Do(ctx context.Context, req Request, result any) error
}

// HTTPHandler performs RPC requests via HTTP POST against the specified URL.
type HTTPHandler struct {
	Client        http.Client
	URL           string
	RequestLogger func(request Request, trace string, latency time.Duration, err error)
}

// Do sends the passed in request to the server, and decodes the result or error
// from the response. Result must be a pointer.
func (c *HTTPHandler) Do(ctx context.Context, req Request, result any) (retErr error) {
	var trace string
	var err error
	if c.RequestLogger != nil {
		start := time.Now()
		defer func() {
			c.RequestLogger(req, trace, time.Since(start), err)
		}()
	}

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
	defer appendOnError(&retErr, httpReq.Body.Close, "; ")

	httpReq.Header.Set(headerAPIVersion, req.APIVersion)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("User-Agent", userAgent)
	httpResp, err := c.Client.Do(httpReq)

	var authErr *oauth2.RetrieveError
	switch {
	case errors.As(err, &authErr):
		trace = authErr.Response.Header.Get("traceparent")
		return HTTPError{
			StatusCode: authErr.Response.StatusCode,
			Headers:    authErr.Response.Header,
			Body:       string(authErr.Body),
		}
	case err != nil:
		return err
	}

	trace = httpResp.Header.Get("traceparent")
	defer appendOnError(&retErr, httpResp.Body.Close, "; ")

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

	var buf bytes.Buffer
	dec := json.NewDecoder(io.TeeReader(httpResp.Body, &buf))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&resp); err != nil {
		data := buf.Bytes()
		return fmt.Errorf("%w: %v (traceparent: %s, body: %s)", ErrBadResponse, err, trace, data)
	}
	if resp.JSONRPC != "2.0" {
		data := buf.Bytes()
		return fmt.Errorf(`%w: jsonrpc must be "2.0" (traceparent: %s, body: %s)`, ErrBadResponse, trace, data)
	}
	if resp.ID != req.ID {
		data := buf.Bytes()
		return fmt.Errorf(`%w: id must match request (traceparent: %s, body: %s)`, ErrBadResponse, trace, data)
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
