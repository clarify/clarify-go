package main

import (
	"context"
	"time"

	clarify "github.com/clarify/clarify-go"
	"github.com/clarify/clarify-go/fields"
	"github.com/clarify/clarify-go/views"
)

func main() {
	creds, err := clarify.CredentialsFromFile("clarify-credentials.json")
	if err != nil {
		panic(err)
	}

	ctx := context.Background()
	client := creds.Client(ctx)

	t1 := fields.AsTimestamp(time.Now())
	t2 := t1.Add(time.Hour)
	df := views.DataFrame{
		"a": {t1: 1.0, t2: 1.2},
		"b": {t1: 2.0},
	}

	if _, err := client.Insert(df).Do(ctx); err != nil {
		panic(err)
	}
}
