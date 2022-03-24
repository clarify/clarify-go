package main

import (
	"context"
	"flag"
	"log"
	"time"

	clarify "github.com/clarify/clarify-go"
	"github.com/clarify/clarify-go/query"
)

func main() {
	flag.Parse()
	integration := flag.Arg(0)

	creds, err := clarify.CredentialsFromFile("clarify-credentials.json")
	if err != nil {
		panic(err)
	}

	ctx := context.Background()
	client := creds.Client(ctx)

	t1 := time.Now().Add(-24 * time.Hour)
	t2 := time.Now()
	rf := query.Or(
		query.Field("createdAt", query.Range(t1, t2)),
		query.Fields(query.Comparisons{
			"name": query.MultiOperator(
				query.NotEqual("hellow world"),
				query.Regex("^hello.*$"),
			),
			"createdAt": query.Range(t1, t2),
		}),
	)
	result, err := client.SelectSignals(integration).Filter(rf).Do(ctx)
	if err != nil {
		panic(err)
	}
	log.Printf("result: %#v\n", result)
}
