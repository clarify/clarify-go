package main

import (
	"context"
	"encoding/json"
	"os"
	"time"

	clarify "github.com/clarify/clarify-go"
	"github.com/clarify/clarify-go/fields"
)

func main() {
	// NOTE: To select item data or meta-data, you must grant the integration
	// access to the "clarify" namespace in the Clarify admin panel.
	creds, err := clarify.CredentialsFromFile("clarify-credentials.json")
	if err != nil {
		panic(err)
	}

	ctx := context.Background()
	client := creds.Client(ctx)

	t1 := time.Now().Add(-24 * time.Hour)
	t2 := time.Now().Add(24 * time.Hour)

	query := fields.Query().
		Where(fields.FilterAll()).
		Limit(10)

	data := fields.Data().
		Where(fields.TimeRange(t1, t2)).
		RollupDuration(time.Hour, time.Monday)

	// Fetch an item for use in evaluate.
	selection, err := client.Clarify().SelectItems(query).Do(ctx)
	if err != nil {
		panic(err)
	}
	if len(selection.Data) == 0 {
		panic("this example require your organization to expose at least one item")
	}
	items := []fields.EvaluateItem{
		{Alias: "i1", ID: selection.Data[0].ID, TimeAggregation: fields.TimeAggregationAvg},
	}
	calculations := []fields.Calculation{
		{Alias: "c1", Formula: "sin(g1)"},
		{Alias: "c2", Formula: "sin(2*PI*time_seconds/3600)"},
		{Alias: "c3", Formula: "max(c1,c2)"},
	}
	groups := []fields.EvaluateGroup{
		{Alias: "g1", Query: query, TimeAggregation: fields.TimeAggregationAvg, GroupAggregation: fields.GroupAggregationAvg},
	}

	result, err := client.Clarify().Evaluate(data).
		Items(items...).
		Groups(groups...).
		Calculations(calculations...).
		Do(ctx)
	if err != nil {
		panic(err)
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	enc.Encode(result)
}
