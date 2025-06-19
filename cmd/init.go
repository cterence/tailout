package cmd

import (
	"fmt"

	"github.com/cterence/tailout/tailout"
	"github.com/spf13/cobra"
)

func buildInitCommand(app *tailout.App) *cobra.Command {
	cmd := &cobra.Command{
		Args:  cobra.NoArgs,
		Use:   "init",
		Short: "Initialize tailnet policy for tailout",
		Long: `Initialize tailnet policy for tailout.

	 This command will update your tailnet policy by:
	 - adding a new tag 'tag:tailout',
	 - adding exit nodes tagged with 'tag:tailout to auto approvers',
	 - allowing your tailnet devices to SSH into tailout nodes.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			err := app.Init(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to initialize tailnet policy: %w", err)
			}
			return nil
		},
	}

	cmd.PersistentFlags().StringVar(&app.Config.Tailscale.APIKey, "tailscale-api-key", "", "Tailscale API key used to perform operations on your tailnet")
	cmd.PersistentFlags().StringVar(&app.Config.Tailscale.Tailnet, "tailscale-tailnet", "", "Tailscale Tailnet to use for operations")
	cmd.PersistentFlags().StringVar(&app.Config.Tailscale.BaseURL, "tailscale-base-url", "https://api.tailscale.com", "Tailscale base API URL, change this if you are using Headscale")
	cmd.PersistentFlags().BoolVarP(&app.Config.NonInteractive, "non-interactive", "n", false, "Disable interactive prompts")
	cmd.PersistentFlags().BoolVarP(&app.Config.DryRun, "dry-run", "d", false, "Dry run mode (no changes will be made)")

	return cmd
}
