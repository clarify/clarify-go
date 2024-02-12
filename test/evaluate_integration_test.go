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

func TestEvaluate(t *testing.T) {
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

	t0, t1 := getDefaultTimeRange()
	ir, err := selectItemsDefault(a)
	if err != nil {
		t.Errorf("%v", err)
	}

	itemIDs := getListFromResponse(ir)

	type testCase struct {
		testArgs         TestArgs
		itemIDs          []string
		query            fields.ResourceQuery
		data             fields.DataQuery
		timeAggregation  fields.TimeAggregation
		groupAggregation fields.GroupAggregation
		expectedFields   func(*clarify.EvaluateResult) bool
	}

	test := func(tc testCase) func(t *testing.T) {
		return func(t *testing.T) {
			t.Helper()

			result, err := evaluate(tc.testArgs.ctx, tc.testArgs.client, tc.itemIDs, tc.query, tc.data, tc.timeAggregation, tc.groupAggregation)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			} else if !tc.expectedFields(result) {
				t.Errorf("unexpected field found!")
			}

			jsonEncode(t, result)
		}
	}

	ef := func(er *clarify.EvaluateResult) bool {
		// this could be done better!
		// like some kind of order of items, groups, and calculations as test input,
		// looped through here to verify they exist in the output.
		keys := []string{"i0", "i1", "c1", "c2", "c3", "g1"}
		ok := true
		for _, key := range keys {
			_, kok := er.Data[key]
			ok = ok && kok
		}

		return ok
	}

	t.Run("basic evaluate test", test(testCase{
		testArgs:         a,
		itemIDs:          itemIDs,
		query:            createAnnotationQuery(a.prefix),
		data:             fields.Data().Where(fields.TimeRange(t0, t1)).RollupDuration(time.Hour, time.Monday),
		timeAggregation:  fields.TimeAggregationAvg,
		groupAggregation: fields.GroupAggregationAvg,
		expectedFields:   ef,
	}))

	taggs := []fields.TimeAggregation{
		fields.TimeAggregationCount,
		fields.TimeAggregationMin,
		fields.TimeAggregationMax,
		fields.TimeAggregationSum,
		fields.TimeAggregationAvg,
		fields.TimeAggregationSeconds,
		fields.TimeAggregationPercent,
		fields.TimeAggregationRate,
	}
	for _, taggs := range taggs {
		t.Run("time aggregation test type "+fmt.Sprint(taggs), test(testCase{
			testArgs:         a,
			itemIDs:          itemIDs,
			query:            createAnnotationQuery(a.prefix),
			data:             fields.Data().Where(fields.TimeRange(t0, t1)).RollupDuration(time.Hour, time.Monday),
			timeAggregation:  taggs,
			groupAggregation: fields.GroupAggregationAvg,
			expectedFields:   ef,
		}))
	}

	gaggs := []fields.GroupAggregation{
		fields.GroupAggregationCount,
		fields.GroupAggregationMin,
		fields.GroupAggregationMax,
		fields.GroupAggregationSum,
		fields.GroupAggregationAvg,
	}
	for _, gagg := range gaggs {
		t.Run("group aggregation test type "+fmt.Sprint(gagg), test(testCase{
			testArgs:         a,
			itemIDs:          itemIDs,
			query:            createAnnotationQuery(a.prefix),
			data:             fields.Data().Where(fields.TimeRange(t0, t1)).RollupDuration(time.Hour, time.Monday),
			timeAggregation:  fields.TimeAggregationAvg,
			groupAggregation: gagg,
			expectedFields:   ef,
		}))
	}
}

func evaluate(ctx context.Context, client *clarify.Client, itemIDs []string, query fields.ResourceQuery, data fields.DataQuery, timeAggregation fields.TimeAggregation, groupAggregation fields.GroupAggregation) (*clarify.EvaluateResult, error) {
	f := func(i int, itemID string) fields.EvaluateItem {
		return fields.EvaluateItem{
			Alias:           fmt.Sprintf("i%d", i),
			ID:              itemID,
			TimeAggregation: timeAggregation,
			State:           10,
		}
	}
	items := MapIndex(f, itemIDs)
	groups := []fields.EvaluateGroup{
		{
			Alias:            "g1",
			Query:            query,
			TimeAggregation:  timeAggregation,
			GroupAggregation: groupAggregation,
			State:            10,
		},
	}
	calculations := []fields.Calculation{
		{Alias: "c1", Formula: "sin(g1)"},
		{Alias: "c2", Formula: "sin(2*PI*time_seconds/3600)"},
		{Alias: "c3", Formula: "max(c1,c2)"},
	}
	format := views.SelectionFormat{
		GroupIncludedByType: true,
	}
	result, err := client.Clarify().Evaluate(data).
		Items(items...).
		Groups(groups...).
		Calculations(calculations...).
		Include([]string{"items"}...).
		Format(format).
		Do(ctx)

	return result, err
}

func getListFromResponse(sir *clarify.SelectItemsResult) []string {
	itemIDs := make([]string, 0, len(sir.Data))
	for _, d := range sir.Data {
		itemIDs = append(itemIDs, d.ID)
	}

	return itemIDs
}
