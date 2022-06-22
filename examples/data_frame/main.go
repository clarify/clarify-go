package main

import (
	"context"
	"encoding/json"
	"os"
	"time"

	clarify "github.com/clarify/clarify-go"
	"github.com/clarify/clarify-go/query"
)

func main() {
	creds, err := clarify.CredentialsFromFile("clarify-credentials.json")
	if err != nil {
		panic(err)
	}

	ctx := context.Background()
	client := creds.Client(ctx)

	t1 := time.Now().Add(-24 * time.Hour)
	t2 := time.Now().Add(24 * time.Hour)

	// To select item data or meta-data, you must grant the integration access
	// to the "clarify" namespace in the Clarify admin panel.
	result, err := client.DataFrame().TimeRange(t1, t2).RollupBucket(time.Hour).Filter(
		query.Field("annotations.clarify/clarify-go/example/name", query.Equal("publish_signals")),
	).Limit(10).Do(ctx)
	if err != nil {
		panic(err)
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	enc.Encode(result)
}
