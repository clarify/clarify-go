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

	inputs := map[string]views.SignalSave{
		"banana-stand/amount": {
			MetaSave: views.MetaSave{
				Annotations: fields.Annotations{
					// Annotation keys should be prefixed to avoid collision.
					"clarify/clarify-go/example/name": "save_signals",
				},
			},
			SignalSaveAttributes: views.SignalSaveAttributes{
				Name:        "Amount",
				Description: "Amount at location, counted manually.",
				Labels: fields.Labels{
					// Label keys generally does not need to be prefixed.
					"data-source": {"manual"},
					"location":    {"banana stand", "pier"},
				},
				EngUnit: "USD",
			},
		},
		"banana-stand/status": {
			MetaSave: views.MetaSave{
				Annotations: fields.Annotations{
					// Annotation keys should be prefixed to avoid collision.
					"clarify/clarify-go/example/name":    "save_signals",
					"clarify/clarify-go/example/publish": "true",
				},
			},
			SignalSaveAttributes: views.SignalSaveAttributes{
				Name:        "Building status",
				Description: "Overall building status, aggregated from environmental sensors.",
				Labels: fields.Labels{
					// Label keys generally does not need to be prefixed.
					"data-source": {"environmental sensors"},
					"location":    {"banana stand", "pier"},
				},
				SourceType: views.Aggregation,
				ValueType:  views.Enum,
				EnumValues: fields.EnumValues{
					0: "not on fire",
					1: "on fire",
				},
				SampleInterval: fields.AsFixedDuration(15 * time.Minute),
				GapDetection:   fields.AsFixedDuration(2 * time.Hour),
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
