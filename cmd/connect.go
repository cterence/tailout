package cmd

import (
	"fmt"

	"github.com/cterence/tailout/tailout"
	"github.com/spf13/cobra"
)

// connectCmd represents the connect command.
func buildConnectCommand(app *tailout.App) *cobra.Command {
	cmd := &cobra.Command{
		Args:  cobra.ArbitraryArgs,
		Use:   "connect",
		Short: "Connect to an exit node in your tailnet",
		RunE: func(cmd *cobra.Command, args []string) error {
			err := app.Connect(cmd.Context(), args)
			if err != nil {
				return fmt.Errorf("failed to connect: %w", err)
			}
			return nil
		},
	}

	cmd.PersistentFlags().BoolVarP(&app.Config.NonInteractive, "non-interactive", "n", false, "Disable interactive prompts")
	cmd.PersistentFlags().StringVar(&app.Config.Tailscale.APIKey, "tailscale-api-key", "", "Tailscale API key used to perform operations on your tailnet")
	cmd.PersistentFlags().StringVar(&app.Config.Tailscale.Tailnet, "tailscale-tailnet", "", "Tailscale Tailnet to use for operations")
	cmd.PersistentFlags().StringVar(&app.Config.Tailscale.BaseURL, "tailscale-base-url", "https://api.tailscale.com", "Tailscale base API URL, change this if you are using Headscale")

	return cmd
}
