package main

import (
	"context"
	"encoding/json"
	"os"

	"github.com/clarify/clarify-go"
	"github.com/clarify/clarify-go/fields"
	clarifyx "github.com/clarify/clarify-go/x"
)

func main() {
	creds, err := clarify.CredentialsFromFile("clarify-credentials.json")
	if err != nil {
		panic(err)
	}

	ctx := context.Background()
	client := creds.Client(ctx)

	integration := "c2mo5rac8v1dtfnta5e0"
	item := "cuga77bpn4pb81evdgj0"
	signal := "cl1d4kobi2aq0ttkc4b0"

	q := fields.Query().Limit(1).Where(
		fields.CompareField("id", fields.In(signal)),
	).Total(true)

	xclient := clarifyx.Upgrade(client)

	result, err := xclient.Admin().ConnectSignals(integration, item, q).Do(ctx)
	if err != nil {
		panic(err)
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	enc.Encode(result)
}
