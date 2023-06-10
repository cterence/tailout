/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/cterence/xit/common"
	"github.com/ktr0731/go-fuzzyfinder"
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

		devicesResponse, err := common.GetDevices(tsApiKey, tailnet)
		if err != nil {
			fmt.Println("Failed to get devices:", err)
			return
		}

		var userDevices common.UserDevices

		json.Unmarshal(devicesResponse, &userDevices)

		var machineConnect string

		if len(args) != 0 {
			machineConnect = args[0]
		} else {
			xitDevices := []string{}
			// Try to find a device with the tag : tag:xit
			for _, device := range userDevices.Devices {
				for _, tag := range device.Tags {
					// Check if lastSeen is within the last 5 minutes, time looks like 2023-06-10T13:13:38Z
					lastSeen, err := time.Parse(time.RFC3339, device.LastSeen)
					if err != nil {
						fmt.Println("Failed to parse lastSeen:", err)
						return
					}

					if tag == "tag:xit" && time.Since(lastSeen) < 5*time.Minute {
						xitDevices = append(xitDevices, device.Hostname)
					}
				}
			}

			if len(xitDevices) == 0 {
				fmt.Println("No xit devices found")
				return
			}

			idx, err := fuzzyfinder.Find(xitDevices, func(i int) string {
				return xitDevices[i]
			})
			if err != nil {
				fmt.Println("Failed to find device:", err)
				return
			}

			machineConnect = xitDevices[idx]
		}

		fmt.Printf("Will run the command:\nsudo tailscale up --exit-node=%s\n", machineConnect)

		// Create a confirmation prompt

		var confirm string

		fmt.Println("\nAre you sure you want to connect to this machine? (y/n)")
		_, err = fmt.Scanln(&confirm)
		if err != nil {
			fmt.Println("Failed to read input:", err)
			return
		}

		if confirm != "y" {
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

		fmt.Printf("\nExisting configuration found, will run updated tailscale up command:\nsudo %s\n", tailscaleUpCommand)

		fmt.Println("\nAre you sure you want to run this command? (y/n)")
		_, err = fmt.Scanln(&confirm)
		if err != nil {
			fmt.Println("Failed to read input:", err)
			return
		}

		if confirm != "y" {
			fmt.Println("Aborting...")
			return
		}

		_, err = exec.Command("sudo", strings.Split(tailscaleUpCommand, " ")...).CombinedOutput()
		if err != nil {
			fmt.Println("Failed to run command:", err)
		}

		fmt.Println("Connected.")
	},
}

func init() {
	rootCmd.AddCommand(connectCmd)

	connectCmd.PersistentFlags().StringP("ts-api-key", "", "", "TailScale API Key")
	connectCmd.PersistentFlags().StringP("ts-tailnet", "", "", "TailScale Tailnet")
	// TODO: Add a --yes flag to skip confirmation prompt
	// connectCmd.Flags().Bool("yes", false, "Skip confirmation prompt")

	viper.BindPFlag("ts_api_key", connectCmd.PersistentFlags().Lookup("ts-api-key"))
	viper.BindPFlag("ts_tailnet", connectCmd.PersistentFlags().Lookup("ts-tailnet"))

	// Create a machine argument for this command

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// connectCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// connectCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
