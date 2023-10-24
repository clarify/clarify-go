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
		Where(fields.CompareField("annotations.clarify/clarify-go/example/name", fields.Equal("publish_signals"))).
		Limit(10)

	data := fields.Data().
		Where(fields.TimeRange(t1, t2)).
		RollupDuration(time.Hour, time.Monday)

	result, err := client.Clarify().DataFrame(query, data).Do(ctx)
	if err != nil {
		panic(err)
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	enc.Encode(result)
}
