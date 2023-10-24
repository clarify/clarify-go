package main

import (
	"context"
	"log"
	"strings"

	clarify "github.com/clarify/clarify-go"
	"github.com/clarify/clarify-go/automation"
	"github.com/clarify/clarify-go/fields"
	"github.com/clarify/clarify-go/views"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

const (
	// For more advanced applications, defining your annotations as constants,
	// is less error prone. Annotation keys should be prefixed to avoid
	// collision.
	keyTransformVersion     = "clarify/clarify-go/example/transform"
	keySignalAttributesHash = "clarify/clarify-go/example/source-signal/attributes-hash"
	keySignalID             = "clarify/clarify-go/example/source-signal/id"

	// In this example we filter which signals to expose using the following
	// annotation keys and values.
	keyExampleName    = "clarify/clarify-go/example/name"
	keyExamplePublish = "clarify/clarify-go/example/publish"
	exampleName       = "save_signals"
	annotationTrue    = "true"

	// transformVersion must be incremented when updating transform, in order to
	// force updates of already exposed items when there are no changes to the
	// underlying signal attributes.
	transformVersion = "v2"
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

	routine := automation.PublishSignals{
		// For this example, the signals we want to publish are created by the same
		// integration that we are using to select them. Note that this isn't a
		// requirement; for production cases, you may want this integration ID to be
		// configured to be something else.
		Integrations: []string{creds.Integration},
		SignalsFilter: fields.Comparisons{
			"annotations." + keyExampleName:    fields.Equal(exampleName),
			"annotations." + keyExamplePublish: fields.Equal(annotationTrue),
		},
		TransformVersion: transformVersion,
		Transforms: []func(item *views.ItemSave){
			transformEnumValuesToFireEmoji,
			transformLabelValuesToTitle,
		},
	}
	if err := routine.Do(ctx, automation.NewConfig(client)); err != nil {
		log.Fatalf("fatal: %v", err)
	}
}

// transformEnumValuesToFireEmoji is an example transform that replaces the enum
// values "on fire" and "not on fire" with appropriate emoji.
func transformEnumValuesToFireEmoji(item *views.ItemSave) {
	for i, v := range item.EnumValues {
		switch {
		case strings.EqualFold(v, "on fire"):
			item.EnumValues[i] = "🔥"
		case strings.EqualFold(v, "not on fire"):
			item.EnumValues[i] = "✅"
		}
	}
}

// transformLabelValuesToTitle transforms label values from format "multiple
// words" to "Multiple Words".
func transformLabelValuesToTitle(item *views.ItemSave) {
	for k, labels := range item.Labels {
		for i, label := range labels {
			item.Labels[k][i] = cases.Title(language.AmericanEnglish).String(label)
		}
	}
}
