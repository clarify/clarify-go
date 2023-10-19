package main

import (
	"context"
	"encoding/json"
	"os"

	clarify "github.com/clarify/clarify-go"
	"github.com/clarify/clarify-go/params"
)

func main() {
	// To select or publish signals, you must grant the integration access to
	// the "admin" namespace in the Clarify admin panel.
	creds, err := clarify.CredentialsFromFile("clarify-credentials.json")
	if err != nil {
		panic(err)
	}

	ctx := context.Background()
	client := creds.Client(ctx)

	// For this example, the signals we want to select are created by the same
	// integration that we are using to select them. Note that this isn't a
	// requirement; for production cases, you may want this integration ID to be
	// configured to be something else.
	integrationID := creds.Integration

	query := params.Query().
		Where(params.Comparisons{"annotations.clarify/clarify-go/example/name": params.Equal("save_signals")}).
		Limit(10)

	result, err := client.Admin().SelectSignals(integrationID, query).Do(ctx)
	if err != nil {
		panic(err)
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	enc.Encode(result)
}
