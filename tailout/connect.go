package tailout

import (
	"context"
	"errors"
	"fmt"
	"net/netip"
	"net/url"
	"slices"

	"github.com/cterence/tailout/internal"
	"github.com/manifoldco/promptui"
	"tailscale.com/client/tailscale"
	tsapi "tailscale.com/client/tailscale/v2"
	"tailscale.com/ipn"
	"tailscale.com/tailcfg"
)

func (app *App) Connect(ctx context.Context, args []string) error {
	var nodeConnect string

	nonInteractive := app.Config.NonInteractive

	baseURL, err := url.Parse(app.Config.Tailscale.BaseURL)
	if err != nil {
		return fmt.Errorf("failed to parse base URL: %w", err)
	}

	apiClient := &tsapi.Client{
		APIKey:  app.Config.Tailscale.APIKey,
		Tailnet: app.Config.Tailscale.Tailnet,
		BaseURL: baseURL,
	}

	var deviceToConnectTo tsapi.Device

	tailoutDevices, err := internal.GetActiveNodes(ctx, apiClient)
	if err != nil {
		return fmt.Errorf("failed to get active nodes: %w", err)
	}

	if len(args) != 0 {
		nodeConnect = args[0]
		i := slices.IndexFunc(tailoutDevices, func(e tsapi.Device) bool {
			return e.Hostname == nodeConnect
		})
		deviceToConnectTo = tailoutDevices[i]
	} else if !nonInteractive {
		if len(tailoutDevices) == 0 {
			return errors.New("no tailout node found in your tailnet")
		}

		// Use promptui to select a node
		prompt := promptui.Select{
			Label: "Select a node",
			Items: tailoutDevices,
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

		deviceToConnectTo = tailoutDevices[idx]
		nodeConnect = deviceToConnectTo.ID
	} else {
		return errors.New("no node name provided")
	}

	var localClient tailscale.LocalClient

	prefs := ipn.NewPrefs()

	prefs.ExitNodeID = tailcfg.StableNodeID(nodeConnect)
	prefs.ExitNodeIP = netip.MustParseAddr(deviceToConnectTo.Addresses[0])

	_, err = localClient.EditPrefs(ctx, &ipn.MaskedPrefs{
		Prefs:         *prefs,
		ExitNodeIDSet: true,
		ExitNodeIPSet: true,
	})
	if err != nil {
		return fmt.Errorf("failed to run tailscale up command: %w", err)
	}

	fmt.Println("Connected.")
	return nil
}
