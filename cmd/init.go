package cmd

import (
	"github.com/cterence/xit/xit"
	"github.com/spf13/cobra"
)

func buildInitCommand(app *xit.App) *cobra.Command {
	cmd := &cobra.Command{
		Args:  cobra.NoArgs,
		Use:   "init",
		Short: "Initialize your tailnet's policy for xit",
		Long: `Initialize your tailnet's policy for xit.
		
	 This command will update your tailnet policy by:
	 - adding a new tag 'tag:xit',
	 - adding exit nodes tagged with 'tag:xit to auto approvers',
	 - allowing your tailnet devices to SSH into xit nodes.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			err := app.Init()
			if err != nil {
				return err
			}
			return nil
		},
	}

	cmd.PersistentFlags().StringVar(&app.Config.Tailscale.APIKey, "tailscale-api-key", "", "Tailscale API key used to perform operations on your tailnet")
	cmd.PersistentFlags().StringVar(&app.Config.Tailscale.Tailnet, "tailscale-tailnet", "", "Tailscale Tailnet to use for operations")
	cmd.PersistentFlags().BoolVarP(&app.Config.NonInteractive, "non-interactive", "n", false, "Disable interactive prompts")
	cmd.PersistentFlags().BoolVarP(&app.Config.DryRun, "dry-run", "d", false, "Dry run mode (no changes will be made)")

	return cmd
}
