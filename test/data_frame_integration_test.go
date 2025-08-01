// Copyright 2024 Searis AS
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

package test

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"testing"
	"time"

	clarify "github.com/clarify/clarify-go"

	"github.com/clarify/clarify-go/fields"
	"github.com/clarify/clarify-go/jsonrpc"
)

func TestDataFrame(t *testing.T) {
	ctx := context.Background()
	creds := credentialsFromEnv(t)
	h, err := creds.HTTPHandler(ctx)
	if err != nil {
		t.Fatal(err)
	}
	h.RequestLogger = func(request jsonrpc.Request, trace string, latency time.Duration, err error) {
		var b bytes.Buffer
		enc := json.NewEncoder(&b)
		_ = enc.Encode(request)
		slog.Debug("Performing JSON RPC request", "trace", trace, "latency", latency, "err", err, "body", json.RawMessage(b.Bytes()))
	}
	client := clarify.NewClient(creds.Integration, h)
	prefix := prefixFromTestName()
	a := TestArgs{
		ctx:         ctx,
		integration: creds.Integration,
		client:      client,
		prefix:      prefix,
	}

	mustApplyTestArgs(a, onlyError(insertDefault), onlyError(saveSignalsDefault), onlyError(publishSignalsDefault))

	t0, t1 := timeRange()

	type testCase struct {
		testArgs       TestArgs
		items          fields.ResourceQuery
		data           fields.DataQuery
		expectedFields func(*clarify.DataFrameResult) bool
	}

	test := func(tc testCase) func(t *testing.T) {
		return func(t *testing.T) {
			t.Helper()

			result, err := dataFrame(tc.testArgs.ctx, tc.testArgs.client, tc.items, tc.data)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			} else if !tc.expectedFields(result) {
				t.Errorf("unexpected field found!")
			}

			mustPrintJSON(t, result)
		}
	}

	t.Run("basic data frame test", test(testCase{
		testArgs: a,
		items:    annotationQuery(a.prefix),
		data:     fields.Data().Where(fields.TimeRange(t0, t1)).RollupDuration(time.Hour, time.Monday),
		expectedFields: func(dfr *clarify.DataFrameResult) bool {
			return true
		},
	}))

	t.Run("less basic data frame test", test(testCase{
		testArgs: a,
		items:    annotationQuery(a.prefix),
		data:     fields.Data().Where(fields.TimeRange(t1, t0)).RollupDuration(time.Hour, time.Monday),
		expectedFields: func(dfr *clarify.DataFrameResult) bool {
			return true
		},
	}))
}

func dataFrame(ctx context.Context, client *clarify.Client, items fields.ResourceQuery, data fields.DataQuery) (*clarify.DataFrameResult, error) {
	result, err := client.Clarify().DataFrame(items, data).Do(ctx)

	return result, err
}
