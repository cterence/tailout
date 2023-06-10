/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"regexp"

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
		dryRun := viper.GetBool("dry_run")
		tsApiKey := viper.GetString("ts_api_key")
		tailnet := viper.GetString("ts_tailnet")

		var devicesToStop []common.Device

		xitDevices, err := common.FindActiveXitDevices(tsApiKey, tailnet)
		if err != nil {
			fmt.Println("Failed to find active xit devices:", err)
			return
		}

		if len(xitDevices) == 0 {
			fmt.Println("No xit devices found")
			return
		}

		if len(args) == 0 {
			// Create a fuzzy finder selector with the xit devices
			idx, err := fuzzyfinder.FindMulti(xitDevices, func(i int) string {
				return xitDevices[i].Hostname
			})
			if err != nil {
				fmt.Println("Failed to find device:", err)
				return
			}

			devicesToStop = []common.Device{}
			for _, i := range idx {
				devicesToStop = append(devicesToStop, xitDevices[i])
			}
		} else {
			for _, device := range xitDevices {
				for _, arg := range args {
					if device.Hostname == arg {
						devicesToStop = append(devicesToStop, device)
					}
				}
			}
		}

		for _, machine := range devicesToStop {
			fmt.Println("Stopping", machine.Hostname)

			// Create a session to share configuration, and load external configuration.
			sess, err := session.NewSession(&aws.Config{})
			if err != nil {
				fmt.Println("Failed to create session:", err)
				return
			}

			// Extract the region from the machine name with a regex
			region := regexp.MustCompile(`(?i)(eu|us|ap|sa|ca|cn|me|af|us-gov|us-iso)-[a-z]{2,}-[0-9]`).FindString(machine.Hostname)

			// Create EC2 service client
			svc := ec2.New(sess, aws.NewConfig().WithRegion(region))

			// Extract the instance ID from the machine name with a regex

			instanceID := regexp.MustCompile(`i\-[a-z0-9]{17}$`).FindString(machine.Hostname)

			_, err = svc.TerminateInstances(&ec2.TerminateInstancesInput{
				DryRun:      aws.Bool(dryRun),
				InstanceIds: []*string{aws.String(instanceID)},
			})

			if err != nil {
				fmt.Println("Failed to terminate instance:", err)
				return
			}

			fmt.Println("Successfully terminated instance", machine.Hostname)

			err = common.DeleteDevice(tsApiKey, machine.ID)
			if err != nil {
				fmt.Println("Failed to delete device from tailnet:", err)
				return
			}

			fmt.Println("Successfully terminated device", machine.Hostname)
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
