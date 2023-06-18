package xit

import (
	"fmt"

	"github.com/cterence/xit/internal"
	"github.com/manifoldco/promptui"
)

func (app *App) Connect(args []string) error {
	var nodeConnect string

	nonInteractive := app.Config.NonInteractive
	tsApiKey := app.Config.Tailscale.APIKey
	tailnet := app.Config.Tailscale.Tailnet

	if len(args) != 0 {
		nodeConnect = args[0]
	} else if !nonInteractive {
		xitNodes, err := internal.FindActiveXitNodes(tsApiKey, tailnet)
		if err != nil {
			return fmt.Errorf("failed to find active xit nodes: %w", err)
		}

		if len(xitNodes) == 0 {
			return fmt.Errorf("no xit node found in your tailnet")
		}

		// Use promptui to select a node
		prompt := promptui.Select{
			Label: "Select a node",
			Items: xitNodes,
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

		nodeConnect = xitNodes[idx].Hostname
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
