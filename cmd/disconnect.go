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

	"github.com/cterence/xit/common"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

// disconnectCmd represents the disconnect command
var disconnectCmd = &cobra.Command{
	Use:   "disconnect",
	Short: "Disconnect from the current exit node",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		var status common.TailscaleStatus

		out, err := exec.Command("tailscale", "debug", "prefs").CombinedOutput()
		if err != nil {
			fmt.Println("Failed to get tailscale preferences:", err)
			return
		}

		json.Unmarshal(out, &status)

		if status.ExitNodeID == "" {
			fmt.Println("Not connected to an exit node.")
			return
		}

		command := "tailscale up --exit-node="

		// Use promptui for the confirmation prompt
		prompt := promptui.Select{
			Label: "Are you sure you want to disconnect from this machine?",
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

		out, err = exec.Command("sudo", strings.Split(command, " ")...).CombinedOutput()
		// If the command was unsuccessful, extract tailscale up command from error message with a regex and run it
		if err != nil {
			// extract latest "tailscale up" command from output with a regex and run it
			regexp := regexp.MustCompile(`tailscale up .*`)
			command = regexp.FindString(string(out))

			fmt.Printf("\nExisting configuration found, will run updated tailscale up command:\nsudo %s\n", command)

			// Use promptui for the confirmation prompt
			prompt = promptui.Select{
				Label: "Are you sure you want to disconnect from this machine?",
				Items: []string{"yes", "no"},
			}

			_, result, err = prompt.Run()
			if err != nil {
				fmt.Println("Failed to read input:", err)
				return
			}

			if result != "yes" {
				fmt.Println("Aborting...")
				return
			}

			_, err = exec.Command("sudo", strings.Split(command, " ")...).CombinedOutput()
			if err != nil {
				fmt.Println("Failed to run command:", err)
			}

			fmt.Println("Disconnected.")
		}

		fmt.Println("Disconnected.")
	},
}

func init() {
	rootCmd.AddCommand(disconnectCmd)
}
