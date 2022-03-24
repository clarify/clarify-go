package main

import (
	"context"
	"os"
	"time"

	clarify "github.com/clarify/clarify-go"
	"github.com/clarify/clarify-go/data"
)

func main() {
	creds := clarify.BasicAuthCredentials(
		os.Getenv("CLARIFY_INTEGRATION_ID"),
		os.Getenv("CLARIFY_PASSWORD"),
	)

	ctx := context.Background()
	client := creds.Client(ctx)

	t1 := data.AsTimestamp(time.Now())
	t2 := t1.Add(time.Hour)
	df := data.Frame{
		"a": {t1: 1.0, t2: 1.2},
		"b": {t1: 2.0},
	}

	if _, err := client.Insert(df).Do(ctx); err != nil {
		panic(err)
	}
}
