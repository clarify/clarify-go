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

package clarify_test

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/clarify/clarify-go"
	"github.com/clarify/clarify-go/params"
	"github.com/clarify/clarify-go/testdata"
)

func ExampleAdminNamespace_SelectSignals() {
	const integrationID = "c8ktonqsahsmemfs7lv0"

	// In this example we use a mock client; real code should instead initialize
	// client using a clarify.Credentials instance.
	h := mockRPCHandler{
		"admin.selectsignals": {
			err:       nil,
			rawResult: json.RawMessage(testdata.ResultSelectSignals),
		},
	}
	c := clarify.NewClient(integrationID, h)

	ctx := context.Background()

	res, err := c.Admin().SelectSignals(integrationID,
		params.Query().Where(
			params.CompareField("id", params.In("c8keagasahsp3cpvma20", "c8l8bc2sahsgjg5cckcg")),
		).Limit(1),
	).Include("items").Do(ctx)

	if err != nil {
		fmt.Println("error:", err)
		return
	}

	fmt.Println("meta.total:", res.Meta.Total)
	fmt.Println("len(data):", len(res.Data))
	if len(res.Data) > 0 {
		fmt.Println("data.0.id:", res.Data[0].ID)
		fmt.Println("data.0.meta.attributesHash:", res.Data[0].Meta.AttributesHash)
	}
	fmt.Println("len(included.items):", len(res.Included.Items))
	if len(res.Included.Items) > 0 {
		fmt.Println("included.items.0.id:", res.Included.Items[0].ID)
		fmt.Println("included.items.0.meta.annotations:", res.Included.Items[0].Meta.Annotations)
	}
	// Output:
	// meta.total: 2
	// len(data): 1
	// data.0.id: c8keagasahsp3cpvma20
	// data.0.meta.attributesHash: 220596a7b7b4ea2ac5abb6a13e6198f161443226
	// len(included.items): 1
	// included.items.0.id: c8l95d2sahsh22imiabg
	// included.items.0.meta.annotations: map[clarify/clarify-go/from/signal/attributes-hash:220596a7b7b4ea2ac5abb6a13e6198f161443226 clarify/clarify-go/from/signal/id:c8keagasahsp3cpvma20]
}

func ExampleClarifyNamespace_DataFrame() {
	const integrationID = "c8ktonqsahsmemfs7lv0"
	const itemID = "c8l95d2sahsh22imiabg"

	// In this example we use a mock client; real code should instead initialize
	// client using a clarify.Credentials instance.
	h := mockRPCHandler{
		"clarify.dataframe": {
			err:       nil,
			rawResult: json.RawMessage(testdata.ResultDataFrameRollup),
		},
	}
	c := clarify.NewClient(integrationID, h)

	ctx := context.Background()

	res, err := c.Clarify().DataFrame(
		params.Query().Where(
			params.CompareField("id", params.In("c8keagasahsp3cpvma20")),
		).Limit(1),

		params.Data().Where(
			params.TimeRange(
				time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC),
				time.Date(2022, 1, 1, 4, 0, 0, 0, time.UTC),
			),
		).RollupDuration(time.Hour, time.Monday),
	).Include("items").Do(ctx)

	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println("meta.total:", res.Meta.Total)
	fmt.Println("len(data):", len(res.Data))
	if len(res.Data) > 0 {
		fmt.Println("data."+itemID+"_sum:", res.Data[itemID+"_sum"])
	}
	fmt.Println("len(included.items):", len(res.Included.Items))
	if len(res.Included.Items) > 0 {
		fmt.Println("included.items.0.id:", res.Included.Items[0].ID)
		fmt.Println("included.items.0.meta.annotations:", res.Included.Items[0].Meta.Annotations)
	}

	// Output:
	// meta.total: 2
	// len(data): 5
	// data.c8l95d2sahsh22imiabg_sum: map[1640995200000000:64 1640998800000000:32.1 1641002400000000:32.5]
	// len(included.items): 1
	// included.items.0.id: c8l95d2sahsh22imiabg
	// included.items.0.meta.annotations: map[clarify/clarify-go/from/signal/attributes-hash:220596a7b7b4ea2ac5abb6a13e6198f161443226 clarify/clarify-go/from/signal/id:c8keagasahsp3cpvma20]
}
