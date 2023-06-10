/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/adhocore/chin"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/cterence/xit/common"
	"github.com/ktr0731/go-fuzzyfinder"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/exp/slices"
)

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Create an exit node in your tailnet",
	Long: `Create an exit node in your tailnet.

This command will create an EC2 instance in the targeted region with the following configuration:
- Amazon Linux 2 AMI
- t3a.micro instance type
- Tailscale installed and configured to advertise as an exit node
- SSH access enabled
- Tagged with App=xit
- The instance will be created as a spot instance in the default VPC`,
	Run: func(cmd *cobra.Command, args []string) {
		// Set up AWS session in the desired region
		tsAuthKey := viper.GetString("ts_auth_key")
		tsApiKey := viper.GetString("ts_api_key")
		tailnet := viper.GetString("ts_tailnet")
		region := viper.GetString("region")
		dryRun := viper.GetBool("dry_run")
		shutdown := viper.GetString("shutdown")

		duration, err := time.ParseDuration(shutdown)
		if err != nil {
			fmt.Println("Failed to parse duration:", err)
			return
		}

		durationMinutes := int(duration.Minutes())
		if durationMinutes < 1 {
			fmt.Println("Duration must be at least 1 minute")
			return
		}

		sess, err := session.NewSession(&aws.Config{})
		if err != nil {
			fmt.Println("Failed to create session:", err)
			return
		}

		// Create EC2 service client

		if region == "" {
			svc := ec2.New(sess, aws.NewConfig().WithRegion("us-east-1"))
			regions, err := svc.DescribeRegions(&ec2.DescribeRegionsInput{})
			if err != nil {
				fmt.Println("Failed to describe regions:", err)
				return
			}

			regionNames := []string{}
			for _, region := range regions.Regions {
				regionNames = append(regionNames, *region.RegionName)
			}

			slices.SortFunc(regionNames, func(a, b string) bool {
				return a > b
			})

			// Create a fuzzy finder selector with the region names
			idx, err := fuzzyfinder.Find(regionNames, func(i int) string {
				return regionNames[i]
			})
			if err != nil {
				fmt.Println("Failed to select region:", err)
				return
			}

			region = regionNames[idx]
		}

		svc := ec2.New(sess, aws.NewConfig().WithRegion(region))

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
sudo tailscale up --authkey=` + tsAuthKey + ` --hostname=xit-` + region + `-$INSTANCE_ID --advertise-exit-node --ssh
sudo echo "sudo shutdown" | at now + ` + fmt.Sprint(durationMinutes) + ` minutes`

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

		var wg sync.WaitGroup

		s := chin.New().WithWait(&wg)

		go s.Start()

		fmt.Println("Waiting for instance to be running...")

		// Add a handler for the instance state change event
		err = svc.WaitUntilInstanceRunning(&ec2.DescribeInstancesInput{
			InstanceIds: []*string{aws.String(*createdInstance.InstanceId)},
		})
		if err != nil {
			fmt.Println("Failed to wait for instance to be created:", err)
			return
		}

		fmt.Println("OK.")
		fmt.Println("Waiting for instance to join tailnet...")

		// Call common.GetDevices periodically and search for the instance
		// If the instance is found, print the command to use it as an exit node

		machineName := fmt.Sprintf("xit-%s-%s", region, *createdInstance.InstanceId)

		timeout := time.Now().Add(2 * time.Minute)

		for {
			devices, err := common.GetDevices(tsApiKey, tailnet)
			if err != nil {
				fmt.Println("Failed to get devices:", err)
				return
			}

			var userDevices common.UserDevices

			json.Unmarshal(devices, &userDevices)

			for _, device := range userDevices.Devices {
				if device.Hostname == machineName {
					goto found
				}
			}

			// Timeouts after 2 minutes
			if time.Now().After(timeout) {
				fmt.Println("Timeout waiting for instance to join tailnet.")
				return
			}

			time.Sleep(2 * time.Second)
		}

	found:
		s.Stop()
		wg.Wait()

		// Get public IP address of created instance
		describeInput := &ec2.DescribeInstancesInput{
			InstanceIds: []*string{aws.String(*createdInstance.InstanceId)},
		}

		describeResult, err := svc.DescribeInstances(describeInput)
		if err != nil {
			fmt.Println("Failed to describe EC2 instance:", err)
			return
		}

		if len(describeResult.Reservations) == 0 {
			fmt.Println("No reservations found.")
			return
		}

		reservation := describeResult.Reservations[0]
		if len(reservation.Instances) == 0 {
			fmt.Println("No instances found.")
			return
		}

		instance := reservation.Instances[0]
		if instance.PublicIpAddress == nil {
			fmt.Println("No public IP address found.")
			return
		}

		fmt.Printf("Instance %s joined tailnet.\n", machineName)
		fmt.Println("Public IP address:", *instance.PublicIpAddress)
		fmt.Println("Planned termination time:", time.Now().Add(duration).Format(time.RFC3339))
		fmt.Println()

		connectCmd.Run(cmd, []string{machineName})
	},
}

func init() {
	cobra.OnInitialize(InitConfig)

	rootCmd.AddCommand(runCmd)

	runCmd.PersistentFlags().StringP("ts-auth-key", "", "", "TailScale Auth Key")
	runCmd.PersistentFlags().StringP("region", "", "", "AWS Region to create the instance into")
	runCmd.PersistentFlags().StringP("shutdown", "s", "2h", "Terminate the instance after the specified duration (e.g. 2h, 1h30m, 30m)")
	viper.BindPFlag("ts_auth_key", runCmd.PersistentFlags().Lookup("ts-auth-key"))
	viper.BindPFlag("region", runCmd.PersistentFlags().Lookup("region"))
	viper.BindPFlag("shutdown", runCmd.PersistentFlags().Lookup("shutdown"))

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
}
