// Copyright 2023-2024 Searis AS
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

package main

import (
	"context"
	"fmt"
	"log/slog"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/clarify/clarify-go/automation"
	"github.com/clarify/clarify-go/fields"
	"github.com/clarify/clarify-go/views"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

const (
	exampleName = "automation_cli"

	// For more advanced applications, defining your annotations as constants,
	// is less error prone. Annotation keys should be prefixed to avoid
	// collision.
	keyTransformVersion     = "clarify/clarify-go/example/transform"
	keySignalAttributesHash = "clarify/clarify-go/example/source-signal/attributes-hash"
	keySignalID             = "clarify/clarify-go/example/source-signal/id"

	// In this example we filter which signals to expose using the following
	// annotation keys and values.
	keyExampleName    = "clarify/clarify-go/example/name"
	keyExamplePublish = "clarify/clarify-go/example/publish"
	annotationTrue    = "true"

	// transformVersion must be incremented when updating transform, in order to
	// force updates of already exposed items when there are no changes to the
	// underlying signal attributes.
	transformVersion = "v2"
)

var routines = automation.Routines{
	"devdata": automation.Routines{
		"save-signals":  automation.RoutineFunc(saveSignals),
		"insert-random": automation.RoutineFunc(insertRandom),
	},
	"publish": publishSignals,
	"evaluate": automation.Routines{
		"detect-fire": detectFire,
	},
}

// insertRandom is an example of a custom routine that inserts a random value
// for the "banana-stand/status" signal.
func insertRandom(ctx context.Context, cfg *automation.Config) error {
	logger := cfg.Logger()
	client := cfg.Client()

	series := make(views.DataSeries)
	until := fields.AsTimestamp(time.Now()).Truncate(time.Minute)
	start := until.Add(-time.Hour)
	for t := start; t < until; t = t.Add(time.Minute) {
		var value float64
		if rand.Intn(100) > 1 {
			// 1% chance of state fire.
			value = 1
		}
		series[t] = value
	}
	df := views.DataFrame{
		"banana-stand/status": series,
	}

	logger.Debug("Insert status signal", automation.AttrDataFrame(df))
	if !cfg.DryRun() {
		result, err := client.Insert(df).Do(ctx)
		if err != nil {
			return fmt.Errorf("insert: %w", err)
		}
		logger.Debug("Insert signal result", slog.Any("result", result))
	}
	return nil
}

// saveSignals is an example of a custom automation routine that updates signal
// meta-data for the "banana-stand/status" signal.
func saveSignals(ctx context.Context, cfg *automation.Config) error {
	logger := cfg.Logger()
	client := cfg.Client()

	signalsByInput := map[string]views.SignalSave{
		"banana-stand/status": {
			MetaSave: views.MetaSave{
				Annotations: fields.Annotations{
					keyExampleName:                       "save_signals",
					"clarify/clarify-go/example/publish": "true",
				},
			},
			SignalSaveAttributes: views.SignalSaveAttributes{
				Name:        "Building status",
				Description: "Overall building status, aggregated from environmental sensors.",
				Labels: fields.Labels{
					"data-source": {"clarify-go/examples"},
					"location":    {"banana stand", "pier"},
				},
				SourceType: views.Aggregation,
				ValueType:  views.Enum,
				EnumValues: fields.EnumValues{
					0: "not on fire",
					1: "on fire",
				},
				SampleInterval: fields.AsFixedDurationNullZero(15 * time.Minute),
				GapDetection:   fields.AsFixedDurationNullZero(2 * time.Hour),
			},
		},
	}
	logger.Debug("Save status signal", slog.Any("signalsByInput", signalsByInput))
	if !cfg.DryRun() {
		result, err := client.SaveSignals(signalsByInput).Do(ctx)
		if err != nil {
			return fmt.Errorf("save signals: %w", err)
		}
		logger.Debug("Save signals result", slog.Any("result", result))
	}

	return nil
}

// publishSignals contain an automation that publish items from signals.
var publishSignals = automation.PublishSignals{
	// NOTE: CLARIFY_EXAMPLE_PUBLISH_INTEGRATION_ID is read from env to allow
	// the example to be runnable without code change. For production, you are
	// recommended to hard-code the integration IDs to publish from.
	Integrations: []string{os.Getenv("CLARIFY_EXAMPLE_PUBLISH_INTEGRATION_ID")},
	SignalsFilter: fields.Comparisons{
		"annotations." + keyExampleName:    fields.Equal(exampleName),
		"annotations." + keyExamplePublish: fields.Equal(annotationTrue),
	},
	TransformVersion: transformVersion,
	Transforms: []func(item *views.ItemSave){
		transformEnumValuesToFireEmoji,
		transformLabelValuesToTitle,
	},
}

// transformEnumValuesToFireEmoji is an example transform that replaces the enum
// values "on fire" and "not on fire" with appropriate emoji.
func transformEnumValuesToFireEmoji(item *views.ItemSave) {
	for i, v := range item.EnumValues {
		switch {
		case strings.EqualFold(v, "on fire"):
			item.EnumValues[i] = "🔥"
		case strings.EqualFold(v, "not on fire"):
			item.EnumValues[i] = "✅"
		}
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

var detectFire = automation.EvaluateActions{
	Evaluation: automation.Evaluation{
		Items: []fields.EvaluateItem{
			{Alias: "fire_rate", ID: os.Getenv("CLARIFY_EXAMPLE_STATUS_ITEM_ID"), TimeAggregation: fields.TimeAggregationRate, State: 1},
		},
		Calculations: []fields.Calculation{
			{Alias: "has_fire", Formula: "fire_rate > 0"}, // return 1.0 when true.
		},
		SeriesIn: []string{"has_fire"},
	},
	Actions: []automation.ActionFunc{
		automation.ActionSeriesContains("has_fire", 1),
		automation.ActionRoutine(automation.LogInfo("FIRE! FIRE! FIRE!")),
	},
}
