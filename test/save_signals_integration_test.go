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
	"fmt"
	"testing"
	"time"

	clarify "github.com/clarify/clarify-go"
	"github.com/clarify/clarify-go/fields"
	"github.com/clarify/clarify-go/views"
)

func TestSaveSignals(t *testing.T) {
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

	applyTestArgs(a, onlyError(insertDefault))

	type testCase struct {
		testArgs       TestArgs
		expectedFields func(*clarify.SaveSignalsResult) bool
	}

	test := func(tc testCase) func(t *testing.T) {
		return func(t *testing.T) {
			t.Helper()

			result, err := saveSignals(tc.testArgs.ctx, tc.testArgs.client, tc.testArgs.prefix)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			} else if !tc.expectedFields(result) {
				t.Errorf("unexpected field found!")
			}

			jsonEncode(t, result)
		}
	}

	t.Run("basic save signals test", test(testCase{
		testArgs: a,
		expectedFields: func(ssr *clarify.SaveSignalsResult) bool {
			return true
		},
	}))
}

func saveSignals(ctx context.Context, client *clarify.Client, prefix string) (*clarify.SaveSignalsResult, error) {
	inputs := map[string]views.SignalSave{
		prefix + "banana-stand/amount": {
			MetaSave: views.MetaSave{
				Annotations: fields.Annotations{
					prefix + AnnotationKey:                        AnnotationValue,
					prefix + "clarify/clarify-go/example/publish": "true",
				},
			},
			SignalSaveAttributes: views.SignalSaveAttributes{
				Name:        prefix + "Amount",
				Description: "Amount at location, counted manually.",
				Labels: fields.Labels{
					"data-source": {"manual"},
					"location":    {"banana stand", "pier"},
				},
				EngUnit: "USD",
			},
		},
		prefix + "banana-stand/status": {
			MetaSave: views.MetaSave{
				Annotations: fields.Annotations{
					prefix + AnnotationKey:                        AnnotationValue,
					prefix + "clarify/clarify-go/example/publish": "true",
				},
			},
			SignalSaveAttributes: views.SignalSaveAttributes{
				Name:        prefix + "Building status",
				Description: "Overall building status, aggregated from environmental sensors.",
				Labels: fields.Labels{
					"data-source": {"environmental sensors"},
					"location":    {"banana stand", "pier"},
				},
				SourceType:     views.Aggregation,
				ValueType:      views.Enum,
				EnumValues:     createEnumValues(),
				SampleInterval: fields.AsFixedDurationNullZero(15 * time.Minute),
				GapDetection:   fields.AsFixedDurationNullZero(2 * time.Hour),
			},
		},
	}
	result, err := client.SaveSignals(inputs).Do(ctx)

	return result, err
}

func createEnumValues() fields.EnumValues {
	ev := make(fields.EnumValues)

	for i := 0; i < 10; i++ {
		ev[i] = fmt.Sprintf("%d", i)
	}

	return ev
}

func saveSignalsDefault(a TestArgs) (*clarify.SaveSignalsResult, error) {
	return saveSignals(a.ctx, a.client, a.prefix)
}
