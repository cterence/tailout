package cmd

import (
	"github.com/cterence/tailout/tailout"
	"github.com/spf13/cobra"
)

// uiCmd represents the UI command
func buildUiCommand(app *tailout.App) *cobra.Command {
	cmd := &cobra.Command{
		Args:  cobra.ArbitraryArgs,
		Use:   "ui",
		Short: "Start the Tailout UI",
		RunE: func(cmd *cobra.Command, args []string) error {
			err := app.Ui(args)
			if err != nil {
				return err
			}
			return nil
		},
	}

	cmd.PersistentFlags().BoolVarP(&app.Config.NonInteractive, "non-interactive", "n", false, "Disable interactive prompts")
	cmd.PersistentFlags().StringVarP(&app.Config.Ui.Address, "address", "a", "127.0.0.1", "Address to bind the UI to")
	cmd.PersistentFlags().StringVarP(&app.Config.Ui.Port, "port", "p", "3000", "Port to bind the UI to")

	return cmd
}
