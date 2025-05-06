package main

import (
	"context"
	"encoding/json"
	"os"

	clarify "github.com/clarify/clarify-go"

	"github.com/clarify/clarify-go/views"
	clarifyx "github.com/clarify/clarify-go/x"
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
	xclient := clarifyx.Upgrade(client)

	// For this example, the signals we want to select are created by the same
	// integration that we are using to select them. Note that this isn't a
	// requirement; for production cases, you may want this integration ID to be
	// configured to be something else.
	integrationID := creds.Integration
	signalID := "cl1d4kobi2aq0ttkc4b0"

	data := []views.SignalAnnotate{{
		ID: signalID,
		Annotations: map[string]string{
			"obviously-this-blue-part-here": "is-the-land",
			"i-ve-made-a-huge-mistake":      "", // Empty values deletes the annotation
		},
	}}

	result, err := xclient.Admin().Signals().Annotate(integrationID, data).Do(ctx)
	if err != nil {
		panic(err)
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(result); err != nil {
		panic(err)
	}
}
