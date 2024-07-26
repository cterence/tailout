package tailout

import (
	"context"
	"fmt"
	"net/netip"

	"tailscale.com/client/tailscale"
	"tailscale.com/ipn"
)

func (app *App) Disconnect() error {
	var localClient tailscale.LocalClient
	prefs, err := localClient.GetPrefs(context.TODO())
	if err != nil {
		return fmt.Errorf("failed to get prefs: %w", err)
	}

	if prefs.ExitNodeID == "" {
		return fmt.Errorf("not connected to an exit node")
	}

	disconnectPrefs := ipn.NewPrefs()

	disconnectPrefs.ExitNodeID = ""
	disconnectPrefs.ExitNodeIP = netip.Addr{}

	_, err = localClient.EditPrefs(context.TODO(), &ipn.MaskedPrefs{
		Prefs:         *disconnectPrefs,
		ExitNodeIDSet: true,
		ExitNodeIPSet: true,
	})

	if err != nil {
		return fmt.Errorf("failed to run tailscale up command: %w", err)
	}

	fmt.Println("Disconnected.")
	return nil
}
