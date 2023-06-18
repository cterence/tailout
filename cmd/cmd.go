package cmd

import (
	"github.com/cterence/xit/xit"
	"github.com/spf13/cobra"
)

func New(app *xit.App) *cobra.Command {
	return buildXitCommand(app)
}

func buildXitCommand(app *xit.App) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "xit",
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

	return cmd
}
