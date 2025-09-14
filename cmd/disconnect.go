package cmd

import (
	"fmt"

	"github.com/cterence/tailout/tailout"
	"github.com/spf13/cobra"
)

// disconnectCmd represents the disconnect command.
func buildDisconnectCommand(app *tailout.App) *cobra.Command {
	cmd := &cobra.Command{
		Args:  cobra.NoArgs,
		Use:   "disconnect",
		Short: "Disconnect from an exit node in your tailnet",
		RunE: func(cmd *cobra.Command, args []string) error {
			err := app.Disconnect(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to disconnect: %w", err)
			}
			return nil
		},
	}

	cmd.PersistentFlags().BoolVarP(&app.Config.NonInteractive, "non-interactive", "n", false, "Disable interactive prompts")

	return cmd
}
