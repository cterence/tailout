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
	"github.com/tailscale/hujson"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize tailnet for xit",
	Long: `This command will initialize your tailnet for xit.
	
In details it will:
- add a tag:xit to your ACL
- update autoapprovers to allow exit nodes to be created
- add a ssh configuration allowing users to ssh into tagged xit machines`,
	Run: func(cmd *cobra.Command, args []string) {
		tsApiKey := viper.GetString("ts_api_key")
		tailnet := viper.GetString("ts_tailnet")
		dryRun := viper.GetBool("dry_run")

		// Get the ACL configuration
		acl, err := common.GetACL(tsApiKey, tailnet)
		if err != nil {
			fmt.Println("Failed to get ACL:", err)
			return
		}

		var config common.Configuration

		standardACL, err := hujson.Standardize(acl)
		if err != nil {
			fmt.Println("Failed to standardize ACL:", err)
			return
		}

		err = json.Unmarshal(standardACL, &config)
		if err != nil {
			fmt.Println("Failed to unmarshal ACL:", err)
			return
		}

		// Update the configuration
		config.TagOwners["tag:xit"] = []string{}
		config.AutoApprovers.ExitNode = []string{"tag:xit"}
		// TODO: find how to add a ssh configuration allowing users to ssh into tagged xit machines

		allowXitSSH := common.SSHConfiguration{
			Action: "check",
			Src:    []string{"autogroup:members"},
			Dst:    []string{"tag:xit"},
			Users:  []string{"autogroup:nonroot", "root"},
		}

		xitSSHConfigExists := false

		for _, sshConfig := range config.SSH {
			if sshConfig.Action == "check" && sshConfig.Src[0] == "autogroup:members" && sshConfig.Dst[0] == "tag:xit" && sshConfig.Users[0] == "autogroup:nonroot" && sshConfig.Users[1] == "root" {
				xitSSHConfigExists = true
			}
		}

		if !xitSSHConfigExists {
			config.SSH = append(config.SSH, allowXitSSH)
		}

		// Validate the updated ACL configuration
		err = common.ValidateACL(tsApiKey, tailnet, config)
		if err != nil {
			fmt.Println("Failed to validate ACL:", err)
			return
		}

		// Update the ACL configuration
		if !dryRun {
			err = common.UpdateACL(tsApiKey, tailnet, config)
			if err != nil {
				fmt.Println("Failed to update ACL:", err)
				return
			}
		}
	},
}

func init() {
	cobra.OnInitialize(InitConfig)

	rootCmd.AddCommand(initCmd)

	initCmd.PersistentFlags().StringP("ts-api-key", "", "", "TailScale API Key")
	initCmd.PersistentFlags().StringP("ts-tailnet", "", "", "TailScale Tailnet")

	viper.BindPFlag("ts_api_key", initCmd.PersistentFlags().Lookup("ts-api-key"))
	viper.BindPFlag("ts_tailnet", initCmd.PersistentFlags().Lookup("ts-tailnet"))

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// initCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
