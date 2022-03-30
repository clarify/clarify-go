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
	"encoding/json"
	"fmt"
	"net/http"
)

const (
	// Standard JSON RPC error codes.
	CodeInvalidJSON    = -32700
	CodeInvalidRequest = -32600
	CodeMethodNotFound = -32601
	CodeInvalidParams  = -32602
	CodeInternal       = -32603

	// Clarify error codes.
	CodeServerError            = -32000
	CodeProduceInvalidResource = -32001
	CodeFoundInvalidResource   = -32002
	CodeForbidden              = -32003
	CodeConflict               = -32009
	CodeTryAgain               = -32015
	CodePartialFailure         = -32021
)

// Client errors.
const (
	ErrBadCredentials strErr = "bad credentials"
	ErrBadResponse    strErr = "bad response"
	ErrBadRequest     strErr = "bad request"
)

type strErr string

func (err strErr) Error() string { return string(err) }

// PathErrors describes issues for fields in a data-structure. Nested field
// errors are reported with as `<parentField>.<subField>`. Field names are
// camelCased to match the preferred case used for JSON encoding.
type PathErrors map[string][]string

func (err PathErrors) Error() string { return fmt.Sprintf("%+v", map[string][]string(err)) }

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
	return fmt.Sprintf("%s (code: %d, data: %+v)", err.Message, err.Code, err.Data)
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
