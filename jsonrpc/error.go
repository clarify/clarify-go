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
	"encoding/json"
	"fmt"
	"net/http"
)

// Client errors.
const (
	ErrBadRequest  strError = "bad request"
	ErrBadResponse strError = "bad response"
)

type strError string

func (err strError) Error() string { return string(err) }

type joinError struct {
	err, next error
	sep       string
}

func joinErrors(err, next error, sep string) error {
	switch {
	case err == nil:
		return next
	case next == nil:
		return err
	default:
		return joinError{
			err:  err,
			next: next,
			sep:  sep,
		}
	}
}

func appendOnError(target *error, f func() error, sep string) {
	*target = joinErrors(*target, f(), sep)
}

func (errs joinError) Error() string {
	return fmt.Sprintf("%s%s%s", errs.err, errs.sep, errs.next)
}

func (errs joinError) Is(other error) bool {
	return other == errs.err
}

func (errs joinError) Unwrap() error {
	return errs.next
}

// HTTPError described a transport-layer error that is returned when the RPC
// server can not be reached for some reason.
type HTTPError struct {
	StatusCode int
	Body       string
	Headers    http.Header
}

func (err HTTPError) Error() string {
	return fmt.Sprintf("%s (status: %d, headers: %+v)", err.Body, err.StatusCode, err.Headers)
}

// ServerError describes the error format returned by the RPC server.
type ServerError struct {
	Code    int       `json:"code"`
	Message string    `json:"message"`
	Data    ErrorData `json:"data"`
}

func (err ServerError) Error() string {
	jd, _ := json.Marshal(err.Data)
	return fmt.Sprintf("%s (code: %d, data: %s)", err.Message, err.Code, jd)
}

// ErrorData describes possible error data fields for the Clarify RPC server.
type ErrorData struct {
	Trace            string              `json:"trace"`
	Params           map[string][]string `json:"params,omitempty"`
	InvalidResources []InvalidResource   `json:"invalidResources,omitempty"`
	PartialResult    json.RawMessage     `json:"partialResult,omitempty"`
}

// InvalidResource describes an invalid resource.
type InvalidResource struct {
	ID            string              `json:"id,omitempty"`
	Type          string              `json:"type"`
	Message       string              `json:"message"`
	InvalidFields map[string][]string `json:"invalidFields,omitempty"`
}
