package main

import (
	"context"
	"encoding/json"
	"os"

	clarify "github.com/clarify/clarify-go"
	"github.com/clarify/clarify-go/fields"
	"github.com/clarify/clarify-go/jsonrpc/resource"
	"github.com/clarify/clarify-go/views"
)

func main() {
	creds, err := clarify.CredentialsFromFile("clarify-credentials.json")
	if err != nil {
		panic(err)
	}

	ctx := context.Background()
	client := creds.Client(ctx)

	// Tools can add annotations to Clarify resources to track state or drive
	// logic. Annotations is the only field that is merged on save. Since the
	// live in the meta section on select, they do not affect attributesHash or
	// relationshipsHash in select views.
	annotations := fields.Annotations{
		// Annotations should include a prefix:
		// <company domain or github org>/<app name>/
		"clarify/clarify-go/example/name": "save_signals",
	}

	inputs := map[string]views.SignalSave{
		"a": {
			MetaSave: resource.MetaSave{
				Annotations: annotations,
			},
			SignalSaveAttributes: views.SignalSaveAttributes{
				Name: "Signal A",
				Labels: fields.Labels{
					"data-source": {"<your data-source name>"},
					"location":    {"<your location name>"},
				},
			},
		},
		"b": {
			MetaSave: resource.MetaSave{
				Annotations: annotations,
			},
			SignalSaveAttributes: views.SignalSaveAttributes{
				Name: "Signal B",
				Labels: fields.Labels{
					"data-source": {"<your data-source name>"},
					"location":    {"<your location name>"},
				},
			},
		},
	}
	result, err := client.SaveSignals(inputs).Do(ctx)
	if err != nil {
		panic(err)
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	enc.Encode(result)
}
