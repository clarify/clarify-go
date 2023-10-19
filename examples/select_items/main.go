package main

import (
	"context"
	"encoding/json"
	"os"

	clarify "github.com/clarify/clarify-go"
	"github.com/clarify/clarify-go/params"
)

func main() {
	// To select item data or meta-data, you must enable access to the "clarify"
	// namespace in clarify.
	creds, err := clarify.CredentialsFromFile("clarify-credentials.json")
	if err != nil {
		panic(err)
	}

	ctx := context.Background()
	client := creds.Client(ctx)

	query := params.Query().
		Where(params.Comparisons{"annotations.clarify/clarify-go/example/name": params.Equal("publish_signals")}).
		Limit(10)

	result, err := client.Clarify().SelectItems(query).Do(ctx)
	if err != nil {
		panic(err)
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	enc.Encode(result)
}
