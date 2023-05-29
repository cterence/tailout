/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// stopCmd represents the stop command
var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		region := viper.GetString("region")
		dryRun := viper.GetBool("dry_run")

		sess, err := session.NewSession(&aws.Config{
			Region: aws.String(region),
		})
		if err != nil {
			fmt.Println("Failed to create session:", err)
			return
		}

		// Create EC2 service client
		svc := ec2.New(sess)

		// Define the tag key and value
		tagKey := "App"
		tagValue := "xit"

		// Filter to describe instances with the specified tag
		tagFilter := &ec2.Filter{
			Name:   aws.String("tag:" + tagKey),
			Values: []*string{aws.String(tagValue)},
		}

		statusFilter := &ec2.Filter{
			Name:   aws.String("instance-state-name"),
			Values: []*string{aws.String("running")},
		}

		// DescribeInstances to get the instances with the specified tag
		instancesOutput, err := svc.DescribeInstances(&ec2.DescribeInstancesInput{
			Filters: []*ec2.Filter{tagFilter, statusFilter},
		})
		if err != nil {
			fmt.Println("Failed to describe instances:", err)
			return
		}

		// Extract the instance IDs
		var instanceIDs []*string
		for _, reservation := range instancesOutput.Reservations {
			for _, instance := range reservation.Instances {
				instanceIDs = append(instanceIDs, instance.InstanceId)
			}
		}

		if instanceIDs == nil {
			fmt.Println("No running instances with tag App=xit found.")
			return
		}

		var instanceIDList []string
		for _, instanceID := range instanceIDs {
			instanceIDList = append(instanceIDList, *instanceID)
		}

		// TerminateInstances to terminate the instances
		_, err = svc.TerminateInstances(&ec2.TerminateInstancesInput{
			InstanceIds: instanceIDs,
			DryRun:      aws.Bool(dryRun),
		})
		if err != nil {
			fmt.Println("Failed to terminate instances:", err)
			return
		}

		fmt.Printf("Instances with tag App=xit terminated successfully: %+v", instanceIDList)
	},
}

func init() {
	rootCmd.AddCommand(stopCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// stopCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// stopCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
