package tailout

import (
	"context"
	"errors"
	"fmt"
	"net/netip"
	"slices"

	"github.com/cterence/tailout/internal"
	"github.com/manifoldco/promptui"
	tsapi "github.com/tailscale/tailscale-client-go/tailscale"
	"tailscale.com/client/tailscale"
	"tailscale.com/ipn"
	"tailscale.com/tailcfg"
)

func (app *App) Connect(args []string) error {
	var nodeConnect string

	nonInteractive := app.Config.NonInteractive

	apiClient, err := tsapi.NewClient(app.Config.Tailscale.APIKey, app.Config.Tailscale.Tailnet, tsapi.WithBaseURL(app.Config.Tailscale.BaseURL))
	if err != nil {
		return fmt.Errorf("failed to create tailscale client: %w", err)
	}

	var deviceToConnectTo tsapi.Device

	tailoutDevices, err := internal.GetActiveNodes(apiClient)
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

	_, err = localClient.EditPrefs(context.TODO(), &ipn.MaskedPrefs{
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
