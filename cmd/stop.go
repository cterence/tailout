package cmd

import (
	"github.com/cterence/tailout/tailout"
	"github.com/spf13/cobra"
)

func buildStopCommand(app *tailout.App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stop [node names...]",
		Args:  cobra.ArbitraryArgs,
		Short: "Terminates instances created by tailout",
		Long: `By default, terminates all instances created by tailout.

	If one or more Node names are specified, only those instances will be terminated.

	Example : tailout stop tailout-eu-west-3-i-048afd4880f66c596`,
		RunE: func(cmd *cobra.Command, args []string) error {
			err := app.Stop(args)
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
	cmd.PersistentFlags().BoolVarP(&app.Config.Stop.All, "all", "a", false, "Terminate all instances created by tailout")

	return cmd
}
