package cmd

import (
	"github.com/cterence/tailout/tailout"
	"github.com/spf13/cobra"
)

func New(app *tailout.App) *cobra.Command {
	return buildTailoutCommand(app)
}

func buildTailoutCommand(app *tailout.App) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "tailout",
		Short:        "Quickly create a cloud-based exit node in your tailnet",
		SilenceUsage: true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return app.Config.Load(cmd.Flags(), cmd.Name())
		},
	}

	cmd.AddCommand(buildCreateCommand(app))
	cmd.AddCommand(buildDisconnectCommand(app))
	cmd.AddCommand(buildConnectCommand(app))
	cmd.AddCommand(buildInitCommand(app))
	cmd.AddCommand(buildStatusCommand(app))
	cmd.AddCommand(buildStopCommand(app))
	cmd.AddCommand(buildUICommand(app))
	cmd.AddCommand(buildVersionCommand())

	return cmd
}
