package cmd

import (
	"fmt"

	"github.com/cterence/tailout/tailout"
	"github.com/spf13/cobra"
)

func buildStatusCommand(app *tailout.App) *cobra.Command {
	cmd := &cobra.Command{
		Args:  cobra.NoArgs,
		Use:   "status",
		Short: "Show tailout-related informations",
		Long: `Show tailout-related informations.

		This command will show the status of tailout nodes, including the node name and whether it is connected or not.

		Example : tailout status`,
		RunE: func(cmd *cobra.Command, args []string) error {
			err := app.Status(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to show status: %w", err)
			}
			return nil
		},
	}

	cmd.PersistentFlags().StringVar(&app.Config.Tailscale.APIKey, "tailscale-api-key", "", "Tailscale API key used to perform operations on your tailnet")
	cmd.PersistentFlags().StringVar(&app.Config.Tailscale.Tailnet, "tailscale-tailnet", "", "Tailscale Tailnet to use for operations")
	cmd.PersistentFlags().StringVar(&app.Config.Tailscale.BaseURL, "tailscale-base-url", "https://api.tailscale.com", "Tailscale base API URL, change this if you are using Headscale")

	return cmd
}
