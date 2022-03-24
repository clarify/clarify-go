package main

import (
	"context"

	clarify "github.com/clarify/clarify-go"
	"github.com/clarify/clarify-go/fields"
)

func main() {
	creds, err := clarify.CredentialsFromFile("clarify-credentials.json")
	if err != nil {
		panic(err)
	}

	ctx := context.Background()
	client := creds.Client(ctx)

	inputs := map[string]clarify.SignalSave{
		"a": {
			SignalWriteAttributes: clarify.SignalWriteAttributes{
				Name: "Signal A",
				Labels: fields.Labels{
					"data-source": {"<your data-source name>"},
					"location":    {"<your location name>"},
				},
			},
		},
		"b": {
			SignalWriteAttributes: clarify.SignalWriteAttributes{
				Name: "Signal B",
				Labels: fields.Labels{
					"data-source": {"<your data-source name>"},
					"location":    {"<your location name>"},
				},
			},
		},
	}
	if _, err := client.SaveSignals(inputs).Do(ctx); err != nil {
		panic(err)
	}
}
