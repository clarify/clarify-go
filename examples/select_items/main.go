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
	// namespace in clarify; this allows selecting signals from all integrations
	// in the organization, but for the example, we will use the integration
	// from the credentials file.
	result, err := client.SelectItems().Filter(query.Comparisons{
		"annotations.clarify/clarify-go/example/name": query.Equal("publish_signals"),
	}).Do(ctx)
	if err != nil {
		panic(err)
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	enc.Encode(result)
}
