/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os/exec"
	"regexp"
	"strings"

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
	Run: func(cmd *cobra.Command, args []string) {
		// Using the CLI on the host, run tailscale up and choose the exit node with the machine name provided
		tsApiKey := viper.GetString("ts_api_key")
		tailnet := viper.GetString("ts_tailnet")

		var machineConnect string

		if len(args) != 0 {
			machineConnect = args[0]
		} else {
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
		}

		fmt.Printf("Will run the command:\nsudo tailscale up --exit-node=%s\n", machineConnect)

		// Create a confirmation prompt

		// Use promptui for the confirmation prompt
		prompt := promptui.Select{
			Label: "Are you sure you want to connect to this machine?",
			Items: []string{"yes", "no"},
		}

		_, result, err := prompt.Run()
		if err != nil {
			fmt.Println("Failed to read input:", err)
			return
		}

		if result != "yes" {
			fmt.Println("Aborting...")
			return
		}

		// Run the command and parse the output

		out, err := exec.Command("sudo", "tailscale", "up", "--exit-node="+machineConnect).CombinedOutput()
		// If the command was unsuccessful, extract tailscale up command from error message with a regex and run it
		if err != nil {
			goto rerun
		}

		fmt.Println("Connected.")

		return

	rerun:
		// extract latest "tailscale up" command from output with a regex and run it
		regexp := regexp.MustCompile(`tailscale up .*`)
		tailscaleUpCommand := regexp.FindString(string(out))

		fmt.Printf("\nExisting configuration found, will run updated tailscale up command:\nsudo %s\n\n", tailscaleUpCommand)

		// Use promptui for the confirmation prompt
		prompt = promptui.Select{
			Label: "Are you sure you want to connect to this machine?",
			Items: []string{"yes", "no"},
		}

		_, result, err = prompt.Run()
		if err != nil {
			fmt.Println("Failed to read input:", err)
			return
		}

		if result == "yes" {
			_, err = exec.Command("sudo", strings.Split(tailscaleUpCommand, " ")...).CombinedOutput()
			if err != nil {
				fmt.Println("Failed to run command:", err)
			}

			fmt.Println("Connected.")
		}
	},
}

func init() {
	rootCmd.AddCommand(connectCmd)

	connectCmd.PersistentFlags().StringP("ts-api-key", "", "", "TailScale API Key")
	connectCmd.PersistentFlags().StringP("ts-tailnet", "", "", "TailScale Tailnet")

	viper.BindPFlag("ts_api_key", connectCmd.PersistentFlags().Lookup("ts-api-key"))
	viper.BindPFlag("ts_tailnet", connectCmd.PersistentFlags().Lookup("ts-tailnet"))
}
