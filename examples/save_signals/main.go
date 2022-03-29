package main

import (
	"context"

	clarify "github.com/clarify/clarify-go"
	"github.com/clarify/clarify-go/resource"
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
			SignalAttributes: clarify.SignalAttributes{
				Name: "Signal A",
				Labels: resource.Labels{
					"data-source": {"<your data-source name>"},
					"location":    {"<your location name>"},
				},
			},
		},
		"b": {
			SignalAttributes: clarify.SignalAttributes{
				Name: "Signal B",
				Labels: resource.Labels{
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
