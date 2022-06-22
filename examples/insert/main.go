package main

import (
	"context"
	"encoding/json"
	"os"
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

	t0 := fields.AsTimestamp(time.Now()).Truncate(time.Hour)
	t1 := t0.Add(15 * time.Minute)
	t2 := t1.Add(15 * time.Minute)
	t3 := t2.Add(15 * time.Minute)
	t4 := t0.Add(24 * time.Hour)
	df := views.DataFrame{
		"banana-stand/amount": {t0: 250000, t4: 0},
		"banana-stand/status": {t0: 0, t1: 0, t2: 1, t3: 1, t4: 0},
	}

	result, err := client.Insert(df).Do(ctx)
	if err != nil {
		panic(err)
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	enc.Encode(result)
}
