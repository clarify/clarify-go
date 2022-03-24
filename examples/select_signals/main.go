package main

import (
	"context"
	"flag"
	"log"
	"time"

	clarify "github.com/clarify/clarify-go"
	"github.com/clarify/clarify-go/filter"
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
	rf := filter.Or(
		filter.Field("createdAt", filter.Range(t1, t2)),
		filter.Fields(filter.Comparisons{
			"name": filter.MultiOperator(
				filter.NotEqual("hellow world"),
				filter.Regex("^hello.*$"),
			),
			"createdAt": filter.Range(t1, t2),
		}),
	)
	result, err := client.SelectSignals(integration).Filter(rf).Do(ctx)
	if err != nil {
		panic(err)
	}
	log.Printf("result: %#v\n", result)
}
