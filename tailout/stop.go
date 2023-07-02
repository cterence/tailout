package tailout

import (
	"fmt"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/cterence/tailout/internal"
	"github.com/cterence/tailout/tailout/config"
	"github.com/cterence/tailout/tailout/tailscale"
	"github.com/ktr0731/go-fuzzyfinder"
)

func (app *App) Stop(args []string) error {
	nonInteractive := app.Config.NonInteractive
	dryRun := app.Config.DryRun
	stopAll := app.Config.Stop.All

	c := tailscale.NewClient(&app.Config.Tailscale)

	var nodesToStop []config.Node

	tailoutNodes, err := c.GetActiveNodes()
	if err != nil {
		return err
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

		nodesToStop = []config.Node{}
		for _, i := range idx {
			nodesToStop = append(nodesToStop, tailoutNodes[i])
		}
	} else {
		if !stopAll {
			nodesToStop = []config.Node{}
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
	// TODO: warning when stopping a deice to which you are connected, propose to disconnect before
	for _, Node := range nodesToStop {
		fmt.Println("Stopping", Node.Hostname)

		// Create a session to share configuration, and load external configuration.
		sess, err := session.NewSession(&aws.Config{})
		if err != nil {
			return fmt.Errorf("failed to create session: %w", err)
		}

		// Extract the region from the Node name with a regex
		region := regexp.MustCompile(`(?i)(eu|us|ap|sa|ca|cn|me|af|us-gov|us-iso)-[a-z]{2,}-[0-9]`).FindString(Node.Hostname)

		// Create EC2 service client
		svc := ec2.New(sess, aws.NewConfig().WithRegion(region))

		// Extract the instance ID from the Node name with a regex

		instanceID := regexp.MustCompile(`i\-[a-z0-9]{17}$`).FindString(Node.Hostname)

		_, err = svc.TerminateInstances(&ec2.TerminateInstancesInput{
			DryRun:      aws.Bool(dryRun),
			InstanceIds: []*string{aws.String(instanceID)},
		})

		if err != nil {
			return fmt.Errorf("failed to terminate instance: %w", err)
		}

		fmt.Println("Successfully terminated instance", Node.Hostname)

		err = c.DeleteNode(Node.ID)
		if err != nil {
			return fmt.Errorf("failed to delete node from tailnet: %w", err)
		}

		fmt.Println("Successfully deleted node", Node.Hostname)
	}
	return nil
}
