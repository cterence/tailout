package tailout

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/url"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/cterence/tailout/internal"
	"github.com/ktr0731/go-fuzzyfinder"
	tsapi "tailscale.com/client/tailscale/v2"
)

func (app *App) Stop(ctx context.Context, args []string) error {
	nonInteractive := app.Config.NonInteractive
	dryRun := app.Config.DryRun
	stopAll := app.Config.Stop.All

	nodesToStop := []tsapi.Device{}

	baseURL, err := url.Parse(app.Config.Tailscale.BaseURL)
	if err != nil {
		return fmt.Errorf("failed to parse base URL: %w", err)
	}

	client := &tsapi.Client{
		APIKey:  app.Config.Tailscale.APIKey,
		Tailnet: app.Config.Tailscale.Tailnet,
		BaseURL: baseURL,
	}

	tailoutNodes, err := internal.GetActiveNodes(ctx, client)
	if err != nil {
		return fmt.Errorf("failed to get active nodes: %w", err)
	}

	if len(tailoutNodes) == 0 {
		fmt.Println("No tailout node found in your tailnet")
		return nil
	}

	if len(args) == 0 && !nonInteractive && !stopAll {
		// Create a fuzzy finder selector with the tailout nodes
		idx, err := fuzzyfinder.FindMulti(tailoutNodes, func(i int) string {
			return tailoutNodes[i].Hostname
		})
		if err != nil {
			return fmt.Errorf("failed to find node: %w", err)
		}

		nodesToStop = []tsapi.Device{}
		for _, i := range idx {
			nodesToStop = append(nodesToStop, tailoutNodes[i])
		}
	} else {
		if !stopAll {
			for _, node := range tailoutNodes {
				for _, arg := range args {
					if node.Hostname == arg {
						nodesToStop = append(nodesToStop, node)
					}
				}
			}
		} else {
			nodesToStop = tailoutNodes
		}
	}

	if !nonInteractive {
		fmt.Println("The following nodes will be stopped:")
		for _, node := range nodesToStop {
			fmt.Println("-", node.Hostname)
		}

		result, err := internal.PromptYesNo("Are you sure you want to stop these Nodes?")
		if err != nil {
			return fmt.Errorf("failed to prompt for confirmation: %w", err)
		}

		if !result {
			fmt.Println("Aborting...")
			return nil
		}
	}

	// TODO: cleanup tailout instances that were not last seen recently
	// TODO: warning when stopping a device to which you are connected, propose to disconnect before
	for _, node := range nodesToStop {
		fmt.Println("Stopping", node.Hostname)

		regionNames, err := internal.GetRegions(ctx)
		if err != nil {
			return fmt.Errorf("failed to retrieve regions: %w", err)
		}
		var region string
		for _, regionName := range regionNames {
			if strings.Contains(node.Hostname, regionName) {
				region = regionName
			}
		}

		if region == "" {
			return errors.New("failed to extract region from node name")
		}

		// Create a session to share configuration, and load external configuration.
		cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
		if err != nil {
			log.Fatalf("unable to load SDK config, %v", err)
		}

		ec2Svc := ec2.NewFromConfig(cfg)

		// Extract the instance ID from the Node name with a regex

		instanceID := regexp.MustCompile(`i\-[a-z0-9]{17}$`).FindString(node.Hostname)

		_, err = ec2Svc.TerminateInstances(ctx, &ec2.TerminateInstancesInput{
			DryRun:      aws.Bool(dryRun),
			InstanceIds: []string{instanceID},
		})
		if err != nil {
			return fmt.Errorf("failed to terminate instance: %w", err)
		}

		fmt.Println("Successfully terminated instance", node.Hostname)

		err = client.Devices().Delete(ctx, node.ID)
		if err != nil {
			return fmt.Errorf("failed to delete node from tailnet: %w", err)
		}

		fmt.Println("Successfully deleted node", node.Hostname)
	}
	return nil
}
