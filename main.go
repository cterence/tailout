package main

import (
	"fmt"
	"log"
	"os"

	"github.com/cterence/tailout/pkg/node"
	"github.com/cterence/tailout/pkg/provider"
	"github.com/davecgh/go-spew/spew"

	"github.com/urfave/cli/v2"
)

func main() {
	var (
		providerFlag string
		regionFlag   string
	)
	const flagPrefix = "TAILOUT_"

	app := &cli.App{
		Name:  "tailout",
		Usage: "Create a cheap VPN using Tailscale and cloud providers (AWS, GCP...)",
		Commands: []*cli.Command{
			{
				Name:    "add",
				Aliases: []string{"a"},
				Usage:   "Add an exit node to your tailnet",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:        "provider",
						Aliases:     []string{"p"},
						Usage:       "Provider for the exit node",
						Destination: &providerFlag,
						EnvVars:     []string{flagPrefix + "PROVIDER"},
					},
					&cli.StringFlag{
						Name:        "region",
						Aliases:     []string{"r"},
						Usage:       "Region for the exit node",
						Destination: &regionFlag,
						EnvVars:     []string{flagPrefix + "REGION"},
					},
				},
				Action: func(ctx *cli.Context) error {
					nc := node.Config{}
					p, err := provider.GetCloudProvider(providerFlag, regionFlag)
					if err != nil {
						log.Fatalf("failed to get cloud provider, %v", err)
					}
					node, err := p.CreateNode(&ctx.Context, nc)
					if err != nil {
						return fmt.Errorf("failed to create node: %v", err)
					}
					spew.Dump(node)
					return nil
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
