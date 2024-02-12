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
	"context"
	"math/rand"
	"testing"
	"time"

	clarify "github.com/clarify/clarify-go"
	"github.com/clarify/clarify-go/fields"
	"github.com/clarify/clarify-go/views"
)

func TestInsert(t *testing.T) {
	ctx := context.Background()
	creds := getCredentials(t)
	client := creds.Client(ctx)
	prefix := createPrefix()
	a := TestArgs{
		ctx:         ctx,
		integration: creds.Integration,
		client:      client,
		prefix:      prefix,
	}

	type testCase struct {
		testArgs       TestArgs
		expectedFields func(*clarify.InsertResult) bool
	}

	test := func(tc testCase) func(t *testing.T) {
		return func(t *testing.T) {
			t.Helper()

			result, err := insert(tc.testArgs.ctx, tc.testArgs.client, tc.testArgs.prefix)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			} else if !tc.expectedFields(result) {
				t.Errorf("unexpected field found!")
			}

			jsonEncode(t, result)
		}
	}

	t.Run("basic insert test", test(testCase{
		testArgs: a,
		expectedFields: func(ir *clarify.InsertResult) bool {
			return true
		},
	}))
}

func insert(ctx context.Context, client *clarify.Client, prefix string) (*clarify.InsertResult, error) {
	tt0, tt1 := getDefaultTimeRange()
	t0 := fields.AsTimestamp(tt0)
	segmentSize := 15 * time.Minute
	segments := int(tt1.Sub(tt0)) / int(segmentSize)

	amount := make(views.DataSeries)
	for i := 0; i < segments; i++ {
		amount[t0.Add(time.Duration(i*int(segmentSize)))] = rand.Float64() * float64(i)
	}

	status := make(views.DataSeries)
	for i := 0; i < segments; i++ {
		status[t0.Add(time.Duration(i*int(segmentSize)))] = float64(rand.Intn(5))
	}

	df := views.DataFrame{
		prefix + "banana-stand/amount": amount,
		prefix + "banana-stand/status": status,
	}
	result, err := client.Insert(df).Do(ctx)

	return result, err
}

func insertDefault(a TestArgs) (*clarify.InsertResult, error) {
	return insert(a.ctx, a.client, a.prefix)
}
