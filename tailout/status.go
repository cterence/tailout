package tailout

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/netip"
	"slices"

	"github.com/cterence/tailout/internal"
	tsapi "github.com/tailscale/tailscale-client-go/tailscale"
	"tailscale.com/client/tailscale"
)

func (app *App) Status() error {
	client, err := tsapi.NewClient(app.Config.Tailscale.APIKey, app.Config.Tailscale.Tailnet, tsapi.WithBaseURL(app.Config.Tailscale.BaseURL))
	if err != nil {
		return fmt.Errorf("failed to create tailscale client: %w", err)
	}

	nodes, err := internal.GetActiveNodes(client)
	if err != nil {
		return fmt.Errorf("failed to get active nodes: %w", err)
	}

	var localClient tailscale.LocalClient
	status, err := localClient.Status(context.TODO())
	if err != nil {
		return fmt.Errorf("failed to get tailscale preferences: %w", err)
	}

	var currentNode tsapi.Device

	if status.ExitNodeStatus != nil {
		i := slices.IndexFunc(nodes, func(e tsapi.Device) bool {
			return netip.MustParsePrefix(e.Addresses[0]+"/32") == status.ExitNodeStatus.TailscaleIPs[0]
		})
		currentNode = nodes[i]
	}

	if len(nodes) == 0 {
		fmt.Println("No active node created by tailout found.")
	} else {
		fmt.Println("Active nodes created by tailout:")
		for _, node := range nodes {
			if currentNode.Hostname == node.Hostname {
				fmt.Println("-", node.Hostname, "[Connected]")
			} else {
				fmt.Println("-", node.Hostname)
			}
		}
	}

	// Query for the public IP address of this Node
	httpClient := &http.Client{}
	req, err := http.NewRequestWithContext(context.TODO(), http.MethodGet, "https://ifconfig.me/ip", nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to get public IP: %w", err)
	}
	defer resp.Body.Close()

	ipAddr, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to get public IP: %w", err)
	}

	fmt.Println("Public IP: " + string(ipAddr))
	return nil
}
