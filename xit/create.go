package xit

import (
	"encoding/base64"
	"fmt"
	"sync"
	"time"

	"github.com/adhocore/chin"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/cterence/xit/internal"
	"github.com/cterence/xit/xit/tailscale"
)

func (app *App) Create() error {
	nonInteractive := app.Config.NonInteractive
	region := app.Config.Region
	tsAuthKey := app.Config.Tailscale.AuthKey
	dryRun := app.Config.DryRun

	connect := app.Config.Create.Connect
	shutdown := app.Config.Create.Shutdown

	c := tailscale.NewClient(&app.Config.Tailscale)

	// TODO: add option for no shutdown
	duration, err := time.ParseDuration(shutdown)
	if err != nil {
		return fmt.Errorf("failed to parse duration: %w", err)
	}

	durationMinutes := int(duration.Minutes())
	if durationMinutes < 1 {
		return fmt.Errorf("duration must be at least 1 minute")
	}

	sess, err := session.NewSession(&aws.Config{})
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}

	// Create EC2 service client

	if region == "" && !nonInteractive {
		region, err = internal.SelectRegion()
		if err != nil {
			return fmt.Errorf("failed to select region: %w", err)
		}
	} else if region == "" && nonInteractive {
		return fmt.Errorf("selected non-interactive mode but no region was explicitly specified")
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
		return fmt.Errorf("failed to describe Amazon Linux images: %w", err)
	}

	if len(amazonLinuxImages.Images) == 0 {
		return fmt.Errorf("no Amazon Linux images found")
	}

	// Get the latest Amazon Linux AMI ID
	latestAMI := amazonLinuxImages.Images[0]
	imageID := *latestAMI.ImageId

	// Define the instance details
	// TODO: add option for instance type
	instanceType := "t3a.micro"
	userDataScript := `#!/bin/bash
# Allow ip forwarding
echo 'net.ipv4.ip_forward = 1' | sudo tee -a /etc/sysctl.conf
echo 'net.ipv6.conf.all.forwarding = 1' | sudo tee -a /etc/sysctl.conf
sudo sysctl -p /etc/sysctl.conf

export INSTANCE_ID=$(curl http://169.254.169.254/latest/meta-data/instance-id)

curl -fsSL https://tailscale.com/install.sh | sh
sudo tailscale up --auth-key=` + tsAuthKey + ` --hostname=xit-` + region + `-$INSTANCE_ID --advertise-exit-node --ssh
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

	identity, err := sts.New(sess).GetCallerIdentity(&sts.GetCallerIdentityInput{})
	if err != nil {
		return fmt.Errorf("failed to get account ID: %w", err)
	}

	fmt.Printf(`Creating xit node in AWS with the following parameters:
- AWS Account ID: %s
- AMI ID: %s
- Instance Type: %s
- Region: %s
- Auto shutdown after: %s
- Connect after instance up: %v
- Network: default VPC / Subnet / Security group of the region
`, *identity.Account, imageID, instanceType, region, shutdown, connect)

	if !nonInteractive {
		result, err := internal.PromptYesNo("Do you want to create this instance?")
		if err != nil {
			return fmt.Errorf("failed to prompt for confirmation: %w", err)
		}

		if !result {
			return nil
		}
	}

	// Run the EC2 instance
	runResult, err := svc.RunInstances(runInput)
	if err != nil {
		return fmt.Errorf("failed to create EC2 instance: %w", err)
	}

	if len(runResult.Instances) == 0 {
		fmt.Println("No instances created.")
		return nil
	}

	createdInstance := runResult.Instances[0]
	fmt.Println("EC2 instance created successfully:", *createdInstance.InstanceId)
	nodeName := fmt.Sprintf("xit-%s-%s", region, *createdInstance.InstanceId)
	fmt.Println("Instance will be named", nodeName)
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
		return fmt.Errorf("failed to add tags to the instance: %w", err)
	}

	// Initialize loading spinner
	var wg sync.WaitGroup
	var s *chin.Chin

	if !nonInteractive {
		s = chin.New().WithWait(&wg)
		go s.Start()
	}

	fmt.Println("Waiting for instance to be running...")

	// Add a handler for the instance state change event
	err = svc.WaitUntilInstanceRunning(&ec2.DescribeInstancesInput{
		InstanceIds: []*string{aws.String(*createdInstance.InstanceId)},
	})
	if err != nil {
		return fmt.Errorf("failed to wait for instance to be created: %w", err)
	}

	fmt.Println("OK.")
	fmt.Println("Waiting for instance to join tailnet...")

	// Call internal.GetNodes periodically and search for the instance
	// If the instance is found, print the command to use it as an exit node

	timeout := time.Now().Add(2 * time.Minute)

	for {
		nodes, err := c.GetNodes()
		if err != nil {
			return fmt.Errorf("failed to get nodes: %w", err)
		}

		for _, node := range nodes {
			if node.Hostname == nodeName {
				goto found
			}
		}

		// Timeouts after 2 minutes
		if time.Now().After(timeout) {
			return fmt.Errorf("timeout waiting for instance to join tailnet")
		}

		time.Sleep(2 * time.Second)
	}

found:
	// Stop the loading spinner
	if !nonInteractive {
		s.Stop()
		wg.Wait()
	}
	// Get public IP address of created instance
	describeInput := &ec2.DescribeInstancesInput{
		InstanceIds: []*string{aws.String(*createdInstance.InstanceId)},
	}

	describeResult, err := svc.DescribeInstances(describeInput)
	if err != nil {
		return fmt.Errorf("failed to describe EC2 instance: %w", err)
	}

	if len(describeResult.Reservations) == 0 {
		return fmt.Errorf("no reservations found")
	}

	reservation := describeResult.Reservations[0]
	if len(reservation.Instances) == 0 {
		return fmt.Errorf("no instances found")
	}

	instance := reservation.Instances[0]
	if instance.PublicIpAddress == nil {
		return fmt.Errorf("no public IP address found")
	}

	fmt.Printf("Instance %s joined tailnet.\n", nodeName)
	fmt.Println("Public IP address:", *instance.PublicIpAddress)
	fmt.Println("Planned termination time:", time.Now().Add(duration).Format(time.RFC3339))

	if connect {
		fmt.Println()
		args := []string{nodeName}
		if nonInteractive {
			args = append(args, "--non-interactive")
		}
		app.Connect(args)
	}
	return nil
}
