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

package resource

// SaveSummary holds the identity of a resource, and describes weather it was
// updated or created due to your operation.
type SaveSummary struct {
	ID      string `json:"id"`
	Created bool   `json:"created"`
	Updated bool   `json:"updated"`
}

// CreateSummary holds the identity of a resource, and describes weather it was
// created due to your operation.
type CreateSummary struct {
	ID      string `json:"id"`
	Created bool   `json:"created"`
}

// Selection holds resource selection results.
type Selection[E, I any] struct {
	Meta     SelectionMeta `json:"meta"`
	Data     []E           `json:"data"`
	Included I             `json:"included"`
}

// SelectionMeta contains top-level meta information about a resource
// selection.
type SelectionMeta struct {
	Total               int  `json:"total"`
	GroupIncludedByType bool `json:"groupIncludedByType"`
}
