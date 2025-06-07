package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/cterence/tailout/cmd"
	"github.com/cterence/tailout/tailout"
)

func main() {
	app, err := tailout.New()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(context.Background())

	// Handle signals
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigCh
		cancel()
	}()

	if err := cmd.New(app).ExecuteContext(ctx); err != nil {
		os.Exit(1)
	}
}
