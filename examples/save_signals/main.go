package main

import (
	"context"

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

	inputs := map[string]views.SignalSave{
		"a": {
			SignalSaveAttributes: views.SignalSaveAttributes{
				Name: "Signal A",
				Labels: fields.Labels{
					"data-source": {"<your data-source name>"},
					"location":    {"<your location name>"},
				},
			},
		},
		"b": {
			SignalSaveAttributes: views.SignalSaveAttributes{
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
