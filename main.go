package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/cterence/tailout/pkg/aws"
	"github.com/cterence/tailout/pkg/gcp"
	"github.com/cterence/tailout/pkg/types"
	"github.com/urfave/cli/v2"
	"github.com/urfave/cli/v2/altsrc"
)

func main() {
	var (
		authKey          string
		gcpProjectIDFlag string
		regionFlag       string
		instanceTypeFlag string
		shutdownFlag     int
		dryRunFlag       bool
	)

	const flagPrefix = "TAILOUT_"

	flags := map[string]cli.Flag{
		"gcp-project-id": altsrc.NewStringFlag(&cli.StringFlag{
			Name:        "gcp-project-id",
			Usage:       "GCP project ID of the exit node",
			Destination: &gcpProjectIDFlag,
			EnvVars:     []string{flagPrefix + "GCP_PROJECT_ID"},
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
			Aliases:     []string{"k"},
			Usage:       "Tailscale auth key for the node",
			Destination: &authKey,
			EnvVars:     []string{flagPrefix + "AUTH_KEY"},
		}),
		"shutdown": altsrc.NewIntFlag(&cli.IntFlag{
			Name:        "shutdown",
			Value:       15,
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

	addAWSFlags := []cli.Flag{
		flags["region"],
		flags["instance-type"],
		flags["auth-key"],
		flags["shutdown"],
		flags["dry-run"],
	}

	addGCPFlags := []cli.Flag{
		flags["account"],
		flags["region"],
		flags["instance-type"],
		flags["auth-key"],
		flags["shutdown"],
		flags["dry-run"],
	}

	// TODO: good config file handling like viper

	app := &cli.App{
		Name:  "tailout",
		Usage: "Create a cheap VPN using Tailscale and cloud providers (AWS, GCP...)",
		Commands: []*cli.Command{
			{
				Name:    "add",
				Aliases: []string{"a"},
				Usage:   "Add an exit node to your tailnet",
				Subcommands: []*cli.Command{
					{
						Name: "aws",
						Before: altsrc.InitInputSourceWithContext(addAWSFlags, func(cCtx *cli.Context) (altsrc.InputSourceContext, error) {
							return altsrc.NewYamlSourceFromFile(".tailout.yaml")
						}),
						Action: func(ctx *cli.Context) error {
							pc := types.ProviderConfig{
								Region: regionFlag,
							}
							p := aws.Provider{}
							if err := p.Init(ctx.Context, pc); err != nil {
								return fmt.Errorf("failed to init AWS provider: %v", err)
							}
							nc := types.NodeConfig{
								AuthKey:      authKey,
								Shutdown:     strconv.Itoa(shutdownFlag),
								InstanceType: instanceTypeFlag,
							}
							_, err := p.CreateNode(ctx.Context, nc, dryRunFlag)
							if err != nil {
								return fmt.Errorf("failed to add exit node: %v", err)
							}
							return nil
						},
					},
					{
						Name: "gcp",
						Before: altsrc.InitInputSourceWithContext(addGCPFlags, func(cCtx *cli.Context) (altsrc.InputSourceContext, error) {
							return altsrc.NewYamlSourceFromFile(".tailout.yaml")
						}),

						Action: func(ctx *cli.Context) error {
							shutdown := strconv.Itoa(shutdownFlag)
							pc := types.ProviderConfig{
								Account: gcpProjectIDFlag,
								Region:  regionFlag,
							}
							p := gcp.Provider{}
							if err := p.Init(ctx.Context, pc); err != nil {
								return fmt.Errorf("failed to init GCP provider: %v", err)
							}

							nc := types.NodeConfig{
								AuthKey:      authKey,
								Shutdown:     shutdown,
								InstanceType: instanceTypeFlag,
							}
							_, err := p.CreateNode(ctx.Context, nc, dryRunFlag)
							if err != nil {
								return fmt.Errorf("failed to add exit node: %v", err)
							}
							return nil
						},
					},
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
