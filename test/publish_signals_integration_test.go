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

	clarify "github.com/clarify/clarify-go"
	"github.com/clarify/clarify-go/automation"
	"github.com/clarify/clarify-go/fields"
	"github.com/clarify/clarify-go/views"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func TestPublishSignals(t *testing.T) {
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

	applyTestArgs(a, onlyError(insertDefault), onlyError(saveSignalsDefault))

	type testCase struct {
		testArgs       TestArgs
		config         *automation.Config
		expectedFields func(*clarify.PublishSignalsResult) bool
	}

	test := func(tc testCase) func(t *testing.T) {
		return func(t *testing.T) {
			t.Helper()

			result, err := publishSignals(tc.testArgs.ctx, tc.testArgs.integration, tc.testArgs.prefix, tc.config)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			} else if !tc.expectedFields(result) {
				t.Errorf("unexpected field found!")
			}

			// FIXME: before enabling `jsonEncode(t, result)` i need to move away from using automation to using the publish signals function directly.
		}
	}

	t.Run("basic publish signals test", test(testCase{
		testArgs: a,
		config:   defaultCfg(a.client),
		expectedFields: func(*clarify.PublishSignalsResult) bool {
			return true
		},
	}))
}

func publishSignals(ctx context.Context, integration string, prefix string, cfg *automation.Config) (*clarify.PublishSignalsResult, error) {
	keyExamplePublish := "clarify/clarify-go/example/publish"
	annotationTrue := "true"
	transformVersion := "v2"
	routine := automation.PublishSignals{
		Integrations: []string{integration},
		SignalsFilter: fields.Comparisons{
			"annotations." + prefix + AnnotationKey:     fields.Equal(AnnotationValue),
			"annotations." + prefix + keyExamplePublish: fields.Equal(annotationTrue),
		},
		TransformVersion: transformVersion,
		Transforms: []func(item *views.ItemSave){
			transformEnumValuesToEmoji,
			transformLabelValuesToTitle,
			createTransform(prefix),
		},
	}
	err := routine.Do(ctx, cfg)

	return nil, err
}

func transformEnumValuesToEmoji(item *views.ItemSave) {
	for i := range item.EnumValues {
		item.EnumValues[i] = getRandom(emojiList)
	}
}

// transformLabelValuesToTitle transforms label values from format "multiple
// words" to "Multiple Words".
func transformLabelValuesToTitle(item *views.ItemSave) {
	for k, labels := range item.Labels {
		for i, label := range labels {
			item.Labels[k][i] = cases.Title(language.AmericanEnglish).String(label)
		}
	}
}

func createTransform(prefix string) func(item *views.ItemSave) {
	return func(item *views.ItemSave) {
		item.Annotations.Set(prefix+AnnotationKey, AnnotationValue)
	}
}

func getRandom[T any](as []T) T {
	return as[rand.Intn(len(as))]
}

var emojiList = []string{
	"🔴", "🟠", "🟡", "🟢", "🔵", "🟣",
}

func defaultCfg(client *clarify.Client) *automation.Config {
	cfg := automation.NewConfig(client)
	cfg = cfg.WithEarlyOut(true)
	cfg = cfg.WithAppName("rotrotrot")

	return cfg
}

func publishSignalsDefault(a TestArgs) (*clarify.PublishSignalsResult, error) {
	return publishSignals(a.ctx, a.integration, a.prefix, defaultCfg(a.client))
}
