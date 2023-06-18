package main

import (
	"fmt"
	"os"

	"github.com/cterence/xit/cmd"
	"github.com/cterence/xit/xit"
)

func main() {
	app, err := xit.New()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		os.Exit(1)
	}

	if err := cmd.New(app).Execute(); err != nil {
		os.Exit(1)
	}
}
