package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"strings"

	clarify "github.com/clarify/clarify-go"
	"github.com/clarify/clarify-go/query"
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

	// For this example, the signals we want to publish are created by the same
	// integration that we are using to select them. Note that this isn't a
	// requirement; for production cases, you may want this integration ID to be
	// configured to be something else.
	integrationID := creds.Integration

	selectResult, err := client.SelectSignals(integrationID).Filter(query.Comparisons{
		"annotations." + keyExampleName:    query.Equal(exampleName),
		"annotations." + keyExamplePublish: query.Equal(annotationTrue),
	}).Include("item").Limit(100).Do(ctx)
	if err != nil {
		panic(err)
	}

	// This logic only updates items if the signals has changed since last time
	// the item was published, or if our transform version has changed.
	prevItemsBySignal := make(map[string]views.Item, len(selectResult.Included.Items))
	for _, item := range selectResult.Included.Items {
		prevItemsBySignal[item.Meta.Annotations.Get(keySignalID)] = item
	}

	newItems := make(map[string]views.ItemSave, len(selectResult.Data))

	for _, signal := range selectResult.Data {
		// Only update item if the source signal has changed.
		prevItem, ok := prevItemsBySignal[signal.ID]
		ok = ok && prevItem.Meta.Annotations.Get(keyTransformVersion) == transformVersion
		ok = ok && prevItem.Meta.Annotations.Get(keySignalAttributesHash) == signal.Meta.AttributesHash.String()
		if ok {
			log.Printf("Skip signal %s: item is up-to-date", signal.ID)
			continue
		}

		// Use visible status from previous item.
		visible := prevItem.Attributes.Visible
		if prevItem.ID == "" {
			visible = true
		}

		newItems[signal.ID] = views.PublishedItem(signal,
			transformEnumValuesToFireEmoji,
			transformLabelValuesToTitle,
			func(item *views.ItemSave) {
				item.Annotations.Set(keyExampleName, "publish_signals")
				item.Annotations.Set(keyTransformVersion, transformVersion)
				item.Annotations.Set(keySignalAttributesHash, signal.Meta.AttributesHash.String())
				item.Annotations.Set(keySignalID, signal.ID)
				item.Visible = visible
			},
		)
	}

	if len(newItems) == 0 {
		log.Printf("All items up-to-date")
		return
	}
	log.Printf("Updating %d/%d items", len(newItems), len(selectResult.Data))
	result, err := client.PublishSignals(integrationID, newItems).Do(ctx)
	if err != nil {
		panic(err)
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	enc.Encode(result)
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

// transformLabelValuesToTitle transforms label values from format "multiple words"
// to "Multiple Words".
func transformLabelValuesToTitle(item *views.ItemSave) {
	for k, labels := range item.Labels {
		for i, label := range labels {
			item.Labels[k][i] = cases.Title(language.AmericanEnglish).String(label)
		}
	}
}
