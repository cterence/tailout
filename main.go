package main

import (
	"fmt"
	"os"

	"github.com/cterence/tailout/cmd"
	"github.com/cterence/tailout/tailout"
)

func main() {
	app, err := tailout.New()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		os.Exit(1)
	}

	if err := cmd.New(app).Execute(); err != nil {
		os.Exit(1)
	}
}
