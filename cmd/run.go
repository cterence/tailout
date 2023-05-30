/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"encoding/base64"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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

		machineName := fmt.Sprintf("xit-%s-%s", region, *createdInstance.InstanceId)

		fmt.Printf(`You will be able to use this insance as an exit node within a few minutes with the following command:

sudo tailscale up --exit-node=%s

You will may have to add other parameters to your command depending on your configuration.
If you use a mobile client, you will be able to use it in a few minutes.
`, machineName)
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
