package main

import (
	"context"
	"encoding/json"
	"os"

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

	// To select or publish signals, you must enable access to the "admin"
	// namespace in clarify; this allows selecting and publishing signals from
	// all integrations in the organization, but for the example, we will use
	// the integrationID from the credentials file.
	integrationID := creds.Integration

	result, err := client.SelectSignals(integrationID).Filter(query.Comparisons{
		"annotations.clarify/clarify-go/example/name": query.Equal("save_signals"),
		"input": query.In("a", "b"),
	}).Include("item").Do(ctx)
	if err != nil {
		panic(err)
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	enc.Encode(result)
}
