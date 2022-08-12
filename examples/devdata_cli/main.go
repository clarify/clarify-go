package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	root := rootCommand()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	defer stop()

	if err := root.ParseAndRun(ctx, os.Args[1:]); err != nil {
		stop()
		log.Fatal(err)
	}
}
