// Copyright 2023 Searis AS
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

package automation

import (
	"context"
	"log/slog"
	"time"

	"github.com/clarify/clarify-go/fields"
	"github.com/clarify/clarify-go/views"
)

// Evaluation describe the instruction for performing an evaluation request as
// well as annotations for passing on to actions.
type Evaluation struct {
	// Items lists item aggregations for items with known IDs.
	Items []fields.ItemAggregation

	// Calculations lists calculations to perform. A calculation can reference
	// items or previous calculations.
	Calculations []fields.Calculation

	// SeriesIn filters which series keys (aliases) to return in the server
	// result. If the value is nil, all series are returned.
	SeriesIn []string
}

// EvaluateActions allows running a single evaluation and pass the result onto
// a chained list of actions.
type EvaluateActions struct {
	// Evaluation contains the evaluations to perform.
	Evaluation Evaluation

	// TimeFunc is provided the current time, and should return a time range to
	// evaluate. If not specified, a default window containing the last hour
	// will be evaluated.
	TimeFunc func(time.Time) (gte, lt time.Time)

	// RollupBucket describe the rollup bucket to use for the evaluation. If not
	// specified, a window rollup is used.
	RollupBucket fields.CalendarDuration

	// Actions describe a chain of functions that are run in order for each
	// evaluation result. If an action modifies the evaluation result, this
	// change is visible to the next action in the chain. If an action returns
	// false, the chain is broken and no further actions are run for the given
	// evaluation result.
	//
	// Actions are responsible for checking opts.DryRun, and for logging their
	// own errors.
	Actions []ActionFunc
}

func (e EvaluateActions) Do(ctx context.Context, cfg *Config) error {
	logger := cfg.Logger()
	client := cfg.Client()

	var gte, lt time.Time
	dataQuery := fields.Data().Where(fields.TimeRange(gte, lt))
	if e.Evaluation.SeriesIn != nil {
		dataQuery = dataQuery.Where(fields.SeriesIn(e.Evaluation.SeriesIn...))
	}

	selection, err := client.Clarify().
		Evaluate(e.Evaluation.Items, e.Evaluation.Calculations, dataQuery).
		Do(ctx)
	if err != nil {
		return err
	}

	result := EvaluateResult{
		Data: selection.Data,
	}
	logger.LogAttrs(
		ctx, slog.LevelDebug, "Evaluation result",
		slog.Any("annotations", result.Annotations),
		slog.Any("data_frame", result.Data),
	)
	for _, action := range e.Actions {
		if !action(ctx, cfg, &result) {
			break
		}
	}
	return nil
}

// ActionFunc describes a function that is run in response to an evaluation. The
// return value indicates whether the next action in a chain of action should run
// or not.
//
// Implementations are responsible for handling their own errors and for acting
// according to the cfg.DryRun() and cfg.EarlyOut() configuration values.
type ActionFunc func(ctx context.Context, cfg *Config, result *EvaluateResult) bool

// ActionAnnotate returns an action that adds an annotation with the passed-in
// key and value to the result.
func ActionAnnotate(key, value string) ActionFunc {
	return func(ctx context.Context, cfg *Config, result *EvaluateResult) bool {
		result.Annotations.Set(key, value)
		return true
	}
}

// ActionSeriesContains returns an action that return true if the named series
// contain at least one instance of value.
func ActionSeriesContains(series string, value float64) ActionFunc {
	return func(ctx context.Context, cfg *Config, result *EvaluateResult) bool {
		for _, v := range result.Data[series] {
			if v == value {
				return true
			}
		}
		return false
	}
}

// ActionRoutine returns a new action function that runs the passed-in routine.
// If the routine succeed, the action returns true. If it fails, the error is
// logged and the action returns false.
func ActionRoutine(r Routine) ActionFunc {
	return func(ctx context.Context, cfg *Config, _ *EvaluateResult) bool {
		if err := r.Do(ctx, cfg); err != nil {
			cfg.Logger().Error(err.Error())
			return false
		}
		return true
	}
}

// EvaluateResult describe the result of an evaluation.
type EvaluateResult struct {
	Annotations fields.Annotations
	Data        views.DataFrame
}
