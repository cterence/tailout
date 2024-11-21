package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/cterence/tailout/pkg/factory"
	"github.com/cterence/tailout/pkg/types"
	"github.com/urfave/cli/v2"
	"github.com/urfave/cli/v2/altsrc"
)

func main() {
	var (
		authKey          string
		providerFlag     string
		regionFlag       string
		instanceTypeFlag string
		shutdownFlag     int
		dryRunFlag       bool
	)

	const flagPrefix = "TAILOUT_"

	flags := map[string]cli.Flag{
		"load": &cli.StringFlag{Name: "load"},
		"provider": altsrc.NewStringFlag(&cli.StringFlag{
			Name:        "provider",
			Aliases:     []string{"p"},
			Usage:       "Provider for the exit node",
			Destination: &providerFlag,
			EnvVars:     []string{flagPrefix + "PROVIDER"},
			Required:    true,
		}),
		"region": altsrc.NewStringFlag(&cli.StringFlag{
			Name:        "region",
			Aliases:     []string{"r"},
			Usage:       "Region for the exit node",
			Destination: &regionFlag,
			EnvVars:     []string{flagPrefix + "REGION"},
		}),
		"instance-type": altsrc.NewStringFlag(&cli.StringFlag{
			Name:        "instance-type",
			Aliases:     []string{"t"},
			Usage:       "Type of the instance",
			Destination: &instanceTypeFlag,
			EnvVars:     []string{flagPrefix + "INSTANCE_TYPE"},
		}),
		"auth-key": altsrc.NewStringFlag(&cli.StringFlag{
			Name:        "auth-key",
			Aliases:     []string{"a"},
			Usage:       "Auth key for the node",
			Destination: &authKey,
			EnvVars:     []string{flagPrefix + "AUTH_KEY"},
		}),
		"shutdown": altsrc.NewIntFlag(&cli.IntFlag{
			Name:        "shutdown",
			Aliases:     []string{"s"},
			Usage:       "Delay after which the node will shutdown, in minutes",
			Destination: &shutdownFlag,
			EnvVars:     []string{flagPrefix + "SHUTDOWN"},
		}),
		"dry-run": altsrc.NewBoolFlag(&cli.BoolFlag{
			Name:        "dry-run",
			Aliases:     []string{"d"},
			Usage:       "Dry-run all provider commands",
			Destination: &dryRunFlag,
			EnvVars:     []string{flagPrefix + "DRY_RUN"},
		}),
	}

	addFlags := []cli.Flag{
		flags["load"],
		flags["provider"],
		flags["region"],
		flags["instance-type"],
		flags["auth-key"],
		flags["shutdown"],
		flags["dry-run"],
	}

	app := &cli.App{
		Name:  "tailout",
		Usage: "Create a cheap VPN using Tailscale and cloud providers (AWS, GCP...)",
		Commands: []*cli.Command{
			{
				Name:    "add",
				Aliases: []string{"a"},
				Usage:   "Add an exit node to your tailnet",
				Flags:   addFlags,
				Before:  altsrc.InitInputSourceWithContext(addFlags, altsrc.NewYamlSourceFromFlagFunc("load")),
				Action: func(ctx *cli.Context) error {
					shutdown := strconv.Itoa(shutdownFlag)
					nc := types.NodeConfig{
						AuthKey:      authKey,
						Shutdown:     shutdown,
						InstanceType: instanceTypeFlag,
					}
					p, err := factory.GetCloudProvider(providerFlag, regionFlag)
					if err != nil {
						log.Fatalf("failed to get cloud provider, %v", err)
					}
					_, err = p.CreateNode(&ctx.Context, nc, dryRunFlag)
					if err != nil {
						return fmt.Errorf("failed to add exit node: %v", err)
					}
					return nil
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
