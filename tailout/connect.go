package tailout

import (
	"fmt"

	"github.com/cterence/tailout/internal"
	"github.com/cterence/tailout/tailout/tailscale"
	"github.com/manifoldco/promptui"
)

func (app *App) Connect(args []string) error {
	var nodeConnect string

	nonInteractive := app.Config.NonInteractive
	c := tailscale.NewClient(&app.Config.Tailscale)

	if len(args) != 0 {
		nodeConnect = args[0]
	} else if !nonInteractive {
		tailoutNodes, err := c.GetActiveNodes()
		if err != nil {
			return err
		}

		if len(tailoutNodes) == 0 {
			return fmt.Errorf("no tailout node found in your tailnet")
		}

		// Use promptui to select a node
		prompt := promptui.Select{
			Label: "Select a node",
			Items: tailoutNodes,
			Templates: &promptui.SelectTemplates{
				Label:    "{{ . }}",
				Active:   "{{ .Hostname | cyan }}",
				Inactive: "{{ .Hostname }}",
				Selected: "{{ .Hostname | yellow }}",
			},
		}

		idx, _, err := prompt.Run()
		if err != nil {
			return fmt.Errorf("failed to select node: %w", err)
		}

		nodeConnect = tailoutNodes[idx].Hostname
	} else {
		return fmt.Errorf("no node name provided")
	}

	err := internal.RunTailscaleUpCommand("tailscale up --exit-node="+nodeConnect, nonInteractive)
	if err != nil {
		return fmt.Errorf("failed to run tailscale up command: %w", err)
	}

	fmt.Println("Connected.")
	return nil
}
