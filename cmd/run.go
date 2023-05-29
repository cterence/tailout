/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"encoding/base64"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Create an exit node",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Set up AWS session in the desired region
		tsAuthKey := viper.GetString("ts_auth_key")
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

		// Filter to fetch the latest Ubuntu LTS AMI ID
		amazonLinuxFilter := []*ec2.Filter{
			{
				Name:   aws.String("name"),
				Values: []*string{aws.String("amzn2-ami-hvm-2.0.*-x86_64-gp2")},
			},
			{
				Name:   aws.String("architecture"),
				Values: []*string{aws.String("x86_64")},
			},
		}

		// DescribeImages to get the latest Amazon Linux AMI
		amazonLinuxImages, err := svc.DescribeImages(&ec2.DescribeImagesInput{
			Filters: amazonLinuxFilter,
			Owners:  []*string{aws.String("amazon")},
		})
		if err != nil {
			fmt.Println("Failed to describe Amazon Linux images:", err)
			return
		}

		if len(amazonLinuxImages.Images) == 0 {
			fmt.Println("No Amazon Linux images found.")
			return
		}

		// Get the latest Amazon Linux AMI ID
		latestAMI := amazonLinuxImages.Images[0]
		imageID := *latestAMI.ImageId

		// Define the instance details
		instanceType := "t3a.micro"
		userDataScript := `#!/bin/bash
# Allow ip forwarding
echo 'net.ipv4.ip_forward = 1' | sudo tee -a /etc/sysctl.conf
echo 'net.ipv6.conf.all.forwarding = 1' | sudo tee -a /etc/sysctl.conf
sudo sysctl -p /etc/sysctl.conf

export INSTANCE_ID=$(curl http://169.254.169.254/latest/meta-data/instance-id)

curl -fsSL https://tailscale.com/install.sh | sh
sudo tailscale up --authkey=` + tsAuthKey + ` --hostname=xit-` + region + `-$(INSTANCE_ID) --advertise-exit-node --ssh`

		// Encode the string in base64
		userDataScriptBase64 := base64.StdEncoding.EncodeToString([]byte(userDataScript))

		// Create instance input parameters
		runInput := &ec2.RunInstancesInput{
			ImageId:      aws.String(imageID),
			InstanceType: aws.String(instanceType),
			MinCount:     aws.Int64(1),
			MaxCount:     aws.Int64(1),
			UserData:     aws.String(userDataScriptBase64),
			DryRun:       aws.Bool(dryRun),
			InstanceMarketOptions: &ec2.InstanceMarketOptionsRequest{
				MarketType: aws.String(ec2.MarketTypeSpot),
				SpotOptions: &ec2.SpotMarketOptions{
					InstanceInterruptionBehavior: aws.String(ec2.InstanceInterruptionBehaviorTerminate),
				},
			},
		}

		fmt.Printf("Instance to be created in region %s: %v \n", region, runInput)

		// Run the EC2 instance
		runResult, err := svc.RunInstances(runInput)
		if err != nil {
			fmt.Println("Failed to create EC2 instance:", err)
			return
		}

		if len(runResult.Instances) == 0 {
			fmt.Println("No instances created.")
			return
		}

		createdInstance := runResult.Instances[0]
		fmt.Println("EC2 instance created successfully:", *createdInstance.InstanceId)

		// Create tags for the instance
		tags := []*ec2.Tag{
			{
				Key:   aws.String("App"),
				Value: aws.String("xit"),
			},
		}

		// Add the tags to the instance
		_, err = svc.CreateTags(&ec2.CreateTagsInput{
			Resources: []*string{aws.String(*createdInstance.InstanceId)},
			Tags:      tags,
		})
		if err != nil {
			fmt.Println("Failed to add tags to the instance:", err)
			return
		}
	},
}

func init() {
	rootCmd.AddCommand(runCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// runCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
