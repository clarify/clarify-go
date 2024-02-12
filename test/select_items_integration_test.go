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
	"testing"

	clarify "github.com/clarify/clarify-go"
	"github.com/clarify/clarify-go/fields"
)

func TestSelectItems(t *testing.T) {
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

	applyTestArgs(a, onlyError(insertDefault), onlyError(saveSignalsDefault), onlyError(publishSignalsDefault))

	type testCase struct {
		testArgs       TestArgs
		items          fields.ResourceQuery
		expectedFields func(*clarify.SelectItemsResult) bool
	}

	test := func(tc testCase) func(t *testing.T) {
		return func(t *testing.T) {
			t.Helper()

			result, err := selectItems(tc.testArgs.ctx, tc.testArgs.client, tc.items)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			} else if !tc.expectedFields(result) {
				t.Errorf("unexpected field found!")
			}

			jsonEncode(t, result)
		}
	}

	t.Run("basic select items test", test(testCase{
		testArgs: a,
		items:    createAnnotationQuery(a.prefix),
		expectedFields: func(sir *clarify.SelectItemsResult) bool {
			return true
		},
	}))
}

func selectItems(ctx context.Context, client *clarify.Client, items fields.ResourceQuery) (*clarify.SelectItemsResult, error) {
	result, err := client.Clarify().SelectItems(items).Do(ctx)

	return result, err
}

func selectItemsDefault(a TestArgs) (*clarify.SelectItemsResult, error) {
	items := createAnnotationQuery(a.prefix)

	return selectItems(a.ctx, a.client, items)
}
