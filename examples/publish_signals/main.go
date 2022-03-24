package main

import (
	"context"
	"encoding/json"
	"log"
	"os"

	clarify "github.com/clarify/clarify-go"
	"github.com/clarify/clarify-go/query"
	"github.com/clarify/clarify-go/views"
)

const (
	// Annotation keys should be prefixed to avoid collision. We recommend:
	// <company domain or GitHub org>/<software name>/<variable name>
	keyExampleName          = "clarify/clarify-go/example/name"
	keyTransform            = "clarify/clarify-go/example/transform"
	keySignalAttributesHash = "clarify/clarify-go/example/source-signal/attributes-hash"
	keySignalID             = "clarify/clarify-go/example/source-signal/id"

	// The value to describe our transformVersion; incrementing or changing this
	// number forces items to be republished.
	transformVersion = "v1"
)

func main() {
	creds, err := clarify.CredentialsFromFile("clarify-credentials.json")
	if err != nil {
		panic(err)
	}

	ctx := context.Background()
	client := creds.Client(ctx)

	// To select or publish signals, you must enable access to the "admin"
	// namespace in clarify; this allows selecting and publishing signals from all integrations
	// in the organization, but for the example, we will use the integrationID
	// from the credentials file.
	integrationID := creds.Integration

	selectResult, err := client.SelectSignals(integrationID).Filter(
		query.Field("annotations."+keyExampleName, query.Equal("save_signals")),
	).Include("item").Do(ctx)
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
		ok = ok && prevItem.Meta.Annotations.Get(keyTransform) == transformVersion
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

		newItems[signal.ID] = views.PublishedItem(signal, func(item *views.ItemSave) {
			item.Annotations.Set(keyExampleName, "publish_signals")
			item.Annotations.Set(keyTransform, transformVersion)
			item.Annotations.Set(keySignalAttributesHash, signal.Meta.AttributesHash.String())
			item.Annotations.Set(keySignalID, signal.ID)
			item.Visible = visible
		})
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
