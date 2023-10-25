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

package fields

// Parsing errors.
const (
	ErrBadCalendarDuration   strError = "must be RFC 3339 duration in range year to fraction"
	ErrMixedCalendarDuration strError = "can not combine month or year components with day or time components"
	ErrBadFixedDuration      strError = "must be RFC 3339 duration in range week to fraction"
)

type strError string

func (err strError) Error() string { return string(err) }
