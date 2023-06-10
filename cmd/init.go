/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/cterence/xit/common"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize tailnet for xit",
	Long: `This command will initialize your tailnet for xit.
	
In details it will:
- add a tag:xit to your policy
- update autoapprovers to allow exit nodes to be created
- add a ssh configuration allowing users to ssh into tagged xit machines`,
	Run: func(cmd *cobra.Command, args []string) {
		tsApiKey := viper.GetString("ts_api_key")
		tailnet := viper.GetString("ts_tailnet")
		dryRun := viper.GetBool("dry_run")

		// Get the policy configuration
		policy, err := common.GetPolicy(tsApiKey, tailnet)
		if err != nil {
			fmt.Println("Failed to get policy:", err)
			return
		}

		// Update the configuration
		policy.TagOwners["tag:xit"] = []string{}
		policy.AutoApprovers.ExitNode = []string{"tag:xit"}

		allowXitSSH := common.SSHConfiguration{
			Action: "check",
			Src:    []string{"autogroup:members"},
			Dst:    []string{"tag:xit"},
			Users:  []string{"autogroup:nonroot", "root"},
		}

		xitSSHConfigExists := false

		for _, sshConfig := range policy.SSH {
			if sshConfig.Action == "check" && sshConfig.Src[0] == "autogroup:members" && sshConfig.Dst[0] == "tag:xit" && sshConfig.Users[0] == "autogroup:nonroot" && sshConfig.Users[1] == "root" {
				xitSSHConfigExists = true
			}
		}

		if !xitSSHConfigExists {
			policy.SSH = append(policy.SSH, allowXitSSH)
		}

		// Validate the updated policy configuration
		err = common.ValidatePolicy(tsApiKey, tailnet, policy)
		if err != nil {
			fmt.Println("Failed to validate policy:", err)
			return
		}

		// Update the policy configuration
		if !dryRun {
			policyJSON, err := json.MarshalIndent(policy, "", "  ")
			if err != nil {
				fmt.Println("Failed to marshal policy:", err)
				return
			}

			// Make a prompt to show the update that will be done
			fmt.Printf(`The following update to the policy will be done:")
Add tag:xit to tagOwners
Update autoapprovers to allow exit nodes to be approved when tagged with tag:xit
Add a ssh configuration allowing users to ssh into tagged xit machines

Your new policy document will look like this:
%s

Do you want to continue? [y/N]
`, policyJSON)

			var answer string
			fmt.Scanln(&answer)
			if answer != "y" {
				fmt.Println("Aborting")
				return
			}

			err = common.UpdatePolicy(tsApiKey, tailnet, policy)
			if err != nil {
				fmt.Println("Failed to update policy:", err)
				return
			}

			fmt.Println("Policy updated")
		}
	},
}

func init() {
	rootCmd.AddCommand(initCmd)

	initCmd.PersistentFlags().StringP("ts-api-key", "", "", "TailScale API Key")
	initCmd.PersistentFlags().StringP("ts-tailnet", "", "", "TailScale Tailnet")

	viper.BindPFlag("ts_api_key", initCmd.PersistentFlags().Lookup("ts-api-key"))
	viper.BindPFlag("ts_tailnet", initCmd.PersistentFlags().Lookup("ts-tailnet"))
}
