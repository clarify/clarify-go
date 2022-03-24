package main

import (
	"context"
	"log"
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

	result, err := client.DataFrame().Limit(5).TimeRange(t1, t2).RollupBucket(time.Hour).FilterField("id", query.Equal("c37rniaj2jujsk6jn2rg")).Do(ctx)
	if err != nil {
		panic(err)
	}
	log.Printf("result: %#v\n", result)
}
