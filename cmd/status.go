/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"time"

	"github.com/cterence/xit/common"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// statusCmd represents the status command
var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show the status of xit devices",
	Long: `Show the status of xit devices.
	
	This command will show the status of xit devices, including the device name and whether it is connected or not.
	
	Example : xit status`,
	Run: func(cmd *cobra.Command, args []string) {
		tsApiKey := viper.GetString("ts_api_key")
		tailnet := viper.GetString("ts_tailnet")

		devices, err := common.GetDevices(tsApiKey, tailnet)
		if err != nil {
			fmt.Println("Failed to get devices:", err)
			return
		}

		found := []string{}

		// Try to find a device with the tag : tag:xit
		for _, device := range devices {
			for _, tag := range device.Tags {
				// Check if lastSeen is within the last 5 minutes, time looks like 2023-06-10T13:13:38Z
				lastSeen, err := time.Parse(time.RFC3339, device.LastSeen)
				if err != nil {
					fmt.Println("Failed to parse lastSeen:", err)
					return
				}

				if tag == "tag:xit" && time.Since(lastSeen) < 5*time.Minute {
					found = append(found, device.Hostname)
				}
			}
		}

		out, err := exec.Command("tailscale", "debug", "prefs").CombinedOutput()

		if err != nil {
			fmt.Println("Failed to get tailscale preferences:", err)
			return
		}

		var status common.TailscaleStatus
		var currentDevice common.Device

		json.Unmarshal(out, &status)

		if status.ExitNodeID != "" {
			currentDevice, err = common.GetDevice(tsApiKey, status.ExitNodeID)
			if err != nil {
				fmt.Println("Failed to get device:", err)
				return
			}
		}

		if len(found) == 0 {
			fmt.Println("No active device created by xit found.")
		} else {
			fmt.Println("Active devices created by xit:")
			for _, device := range found {
				if currentDevice.Hostname == device {
					fmt.Println("-", device, "[Connected]")
				} else {
					fmt.Println("-", device)
				}
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)

	statusCmd.PersistentFlags().StringP("ts-api-key", "", "", "TailScale API Key")
	statusCmd.PersistentFlags().StringP("ts-tailnet", "", "", "TailScale Tailnet")

	viper.BindPFlag("ts_api_key", statusCmd.PersistentFlags().Lookup("ts-api-key"))
	viper.BindPFlag("ts_tailnet", statusCmd.PersistentFlags().Lookup("ts-tailnet"))
}
