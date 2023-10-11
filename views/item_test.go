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

package views_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/clarify/clarify-go/views"
)

func TestItemSelectMarshalJSON(t *testing.T) {
	itemID := "123"
	now := time.Now()
	expectSHA1 := "cc65e86b937c9d743a2b004b4f5de9ccde46bfc8"
	item := views.Item{
		Identifier: views.Identifier{
			Type: "items",
			ID:   itemID,
		},
		Meta: views.Meta{
			CreatedAt: now,
			UpdatedAt: now,
		},
		Attributes: views.ItemAttributes{
			ItemSaveAttributes: views.ItemSaveAttributes{
				Name: "test",
			},
		},
	}
	data, err := item.MarshalJSON()
	if err != nil {
		t.Errorf("json.Marshal returns an error: %v", err)
	}
	var check struct {
		Meta views.Meta
	}
	if err := json.Unmarshal(data, &check); err != nil {
		t.Errorf("json.Unmarshal returns an error: %v", err)
	}
	if sha1 := check.Meta.AttributesHash.String(); sha1 != expectSHA1 {
		t.Errorf("Unexpected value for meta.attributesHash:\n got: %q\nwant: %q", sha1, expectSHA1)
	}
	if sha1 := item.Meta.AttributesHash.String(); sha1 != "" {
		t.Errorf("Expectted item.Meta.AttributesHash is unchanged:\n got: %q\nwant: \"\"", sha1)
	}
}
