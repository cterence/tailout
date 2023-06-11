package cmd

import (
	"fmt"

	"github.com/cterence/xit/common"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// connectCmd represents the connect command
var connectCmd = &cobra.Command{
	Use:   "connect [machine name]",
	Args:  cobra.MaximumNArgs(1),
	Short: "Connect to a machine",
	Long: `Connect to a machine.

	This command will run tailscale up and choose the exit node with the machine name provided.
	
	Example : xit connect xit-eu-west-3-i-048afd4880f66c596`,
	PreRun: func(cmd *cobra.Command, args []string) {
		viper.BindPFlag("ts_api_key", cmd.PersistentFlags().Lookup("ts-api-key"))
		viper.BindPFlag("ts_tailnet", cmd.PersistentFlags().Lookup("ts-tailnet"))
		viper.BindPFlag("non_interactive", cmd.PersistentFlags().Lookup("non-interactive"))
	},
	Run: func(cmd *cobra.Command, args []string) {
		// Using the CLI on the host, run tailscale up and choose the exit node with the machine name provided
		tsApiKey := viper.GetString("ts_api_key")
		tailnet := viper.GetString("ts_tailnet")
		nonInteractive := viper.GetBool("non_interactive")

		var machineConnect string

		if len(args) != 0 {
			machineConnect = args[0]
		} else if !nonInteractive {
			xitDevices, err := common.FindActiveXitDevices(tsApiKey, tailnet)
			if err != nil {
				fmt.Println("Failed to find active xit devices:", err)
				return
			}

			if len(xitDevices) == 0 {
				fmt.Println("No xit devices found")
				return
			}

			// Use promptui to select a device

			prompt := promptui.Select{
				Label: "Select a device",
				Items: xitDevices,
				Templates: &promptui.SelectTemplates{
					Label:    "{{ . }}",
					Active:   "{{ .Hostname | cyan }}",
					Inactive: "{{ .Hostname }}",
					Selected: "{{ .Hostname | yellow }}",
				},
			}

			idx, _, err := prompt.Run()
			if err != nil {
				fmt.Println("Failed to select device:", err)
				return
			}

			machineConnect = xitDevices[idx].Hostname
		} else {
			fmt.Println("No machine name provided")
			return
		}

		err := common.RunTailscaleUpCommand("tailscale up --exit-node="+machineConnect, nonInteractive)
		if err != nil {
			fmt.Println("Failed to run tailscale up command:", err)
			return
		}

		fmt.Println("Connected.")
	},
}

func init() {
	rootCmd.AddCommand(connectCmd)

	connectCmd.PersistentFlags().StringP("ts-api-key", "", "", "TailScale API Key")
	connectCmd.PersistentFlags().StringP("ts-tailnet", "", "", "TailScale Tailnet")
	connectCmd.PersistentFlags().BoolP("non-interactive", "", false, "Do not prompt for confirmation")
}
