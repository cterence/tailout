package cmd

import (
	"github.com/cterence/tailout/tailout"
	"github.com/spf13/cobra"
)

// createCmd represents the create command
func buildCreateCommand(app *tailout.App) *cobra.Command {
	cmd := &cobra.Command{
		Args:  cobra.NoArgs,
		Use:   "create",
		Short: "Create an exit node in your tailnet",
		Long: `Create an exit node in your tailnet.

 This command will create an EC2 instance in the targeted region with the following configuration:
 - Amazon Linux 2 AMI
 - t3a.micro instance type
 - Tailscale installed and configured to advertise as an exit node
 - SSH access enabled
 - Tagged with App=tailout
 - The instance will be created as a spot instance in the default VPC`,

		RunE: func(cmd *cobra.Command, args []string) error {
			err := app.Create()
			if err != nil {
				return err
			}
			return nil
		},
	}

	cmd.PersistentFlags().StringVar(&app.Config.Tailscale.APIKey, "tailscale-api-key", "", "Tailscale API key used to perform operations on your tailnet")
	cmd.PersistentFlags().StringVar(&app.Config.Tailscale.Tailnet, "tailscale-tailnet", "", "Tailscale Tailnet to use for operations")
	cmd.PersistentFlags().StringVar(&app.Config.Tailscale.AuthKey, "tailscale-auth-key", "", "Tailscale Auth Key to use for operations")
	cmd.PersistentFlags().BoolVarP(&app.Config.DryRun, "dry-run", "d", false, "Dry run mode (no changes will be made)")
	cmd.PersistentFlags().BoolVarP(&app.Config.NonInteractive, "non-interactive", "n", false, "Disable interactive prompts")
	cmd.PersistentFlags().StringVarP(&app.Config.Region, "region", "r", "", "Cloud-provider region to use")

	cmd.PersistentFlags().StringVarP(&app.Config.Create.Shutdown, "shutdown", "s", "2h", "Shutdown the instance after the specified duration. Valid time units are \"s\", \"m\", \"h\"")
	cmd.PersistentFlags().BoolVarP(&app.Config.Create.Connect, "connect", "c", false, "Connect to the instance after creation")

	return cmd
}
