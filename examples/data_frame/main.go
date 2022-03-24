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
	t2 := time.Now()

	result, err := client.DataFrame().Limit(5).TimeRange(t1, t2).RollupBucket(15 * time.Minute).Last(3).Filter(
		query.Field("annotations.clarify/clarify-go/example/name", query.Equal("publish_signals")),
	).Do(ctx)
	if err != nil {
		panic(err)
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	enc.Encode(result)
}
