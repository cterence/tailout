/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"encoding/json"
	"fmt"
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/cterence/xit/common"
	"github.com/ktr0731/go-fuzzyfinder"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// stopCmd represents the stop command
var stopCmd = &cobra.Command{
	Use:   "stop [machine names...]",
	Args:  cobra.ArbitraryArgs,
	Short: "Terminates instances created by xit",
	Long: `By default, terminates all instances created by xit. 

If one or more machine names are specified, only those instances will be terminated.

Example : xit stop xit-eu-west-3-i-048afd4880f66c596`,
	Run: func(cmd *cobra.Command, args []string) {
		// TODO: fetch all devices, make a multi fuzzy finder and stop all the selected devices while changing the region if needed
		dryRun := viper.GetBool("dry_run")
		tsApiKey := viper.GetString("ts_api_key")
		tailnet := viper.GetString("ts_tailnet")

		machineStop := args

		if len(machineStop) == 0 {
			devicesResponse, err := common.GetDevices(tsApiKey, tailnet)
			if err != nil {
				fmt.Println("Failed to get devices:", err)
				return
			}

			var userDevices common.UserDevices

			json.Unmarshal(devicesResponse, &userDevices)

			xitDevices := []string{}
			// Try to find a device with the tag : tag:xit
			for _, device := range userDevices.Devices {
				for _, tag := range device.Tags {
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

			// Create a fuzzy finder selector with the xit devices
			idx, err := fuzzyfinder.FindMulti(xitDevices, func(i int) string {
				return xitDevices[i]
			})
			if err != nil {
				fmt.Println("Failed to find device:", err)
				return
			}

			machineStop = []string{}
			for _, i := range idx {
				machineStop = append(machineStop, xitDevices[i])
			}
		}

		for _, machine := range machineStop {
			fmt.Println("Stopping", machine)

			// Create a session to share configuration, and load external configuration.
			sess, err := session.NewSession(&aws.Config{})
			if err != nil {
				fmt.Println("Failed to create session:", err)
				return
			}

			// Extract the region from the machine name with a regex
			region := regexp.MustCompile(`(?i)(eu|us|ap|sa|ca|cn|me|af|us-gov|us-iso)-[a-z]{2,}-[0-9]`).FindString(machine)

			// Create EC2 service client
			svc := ec2.New(sess, aws.NewConfig().WithRegion(region))

			// Extract the instance ID from the machine name with a regex

			instanceID := regexp.MustCompile(`i\-[a-z0-9]{17}$`).FindString(machine)

			_, err = svc.TerminateInstances(&ec2.TerminateInstancesInput{
				DryRun:      aws.Bool(dryRun),
				InstanceIds: []*string{aws.String(instanceID)},
			})

			if err != nil {
				fmt.Println("Failed to terminate instance:", err)
				return
			}

			fmt.Println("Successfully terminated instance", machine)
		}
	},
}

func init() {
	rootCmd.AddCommand(stopCmd)
	stopCmd.PersistentFlags().StringP("ts-api-key", "", "", "TailScale API Key")
	stopCmd.PersistentFlags().StringP("ts-tailnet", "", "", "TailScale Tailnet")
	viper.BindPFlag("ts_api_key", stopCmd.PersistentFlags().Lookup("ts-api-key"))
	viper.BindPFlag("ts_tailnet", stopCmd.PersistentFlags().Lookup("ts-tailnet"))
}
