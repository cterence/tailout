package tailout

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"net/url"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/adhocore/chin"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/cterence/tailout/internal"
	tsapi "tailscale.com/client/tailscale/v2"
)

func (app *App) Create() error {
	nonInteractive := app.Config.NonInteractive
	region := app.Config.Region
	dryRun := app.Config.DryRun
	connect := app.Config.Create.Connect
	shutdown := app.Config.Create.Shutdown

	if app.Config.Tailscale.AuthKey == "" {
		return errors.New("no tailscale auth key found")
	}

	// TODO: add option for no shutdown
	duration, err := time.ParseDuration(shutdown)
	if err != nil {
		return fmt.Errorf("failed to parse duration: %w", err)
	}

	durationMinutes := int(duration.Minutes())
	if durationMinutes < 1 {
		return errors.New("duration must be at least 1 minute")
	}

	// Create EC2 service client

	if region == "" && !nonInteractive {
		region, err = internal.SelectRegion()
		if err != nil {
			return fmt.Errorf("failed to select region: %w", err)
		}
	} else if region == "" && nonInteractive {
		return errors.New("selected non-interactive mode but no region was explicitly specified")
	}

	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(region))
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}

	ec2Svc := ec2.NewFromConfig(cfg)

	// DescribeImages to get the latest Amazon Linux AMI
	amazonLinuxImages, err := ec2Svc.DescribeImages(context.TODO(), &ec2.DescribeImagesInput{
		Filters: []types.Filter{
			{
				Name:   aws.String("name"),
				Values: []string{"al2023-ami-*"},
			},
			{
				Name:   aws.String("state"),
				Values: []string{"available"},
			},
			{
				Name:   aws.String("is-public"),
				Values: []string{"true"},
			},
			{
				Name:   aws.String("architecture"),
				Values: []string{"x86_64"},
			},
		},
		Owners: []string{"amazon"},
	})
	if err != nil {
		return fmt.Errorf("failed to describe Amazon Linux images: %w", err)
	}

	if len(amazonLinuxImages.Images) == 0 {
		return errors.New("no Amazon Linux images found")
	}

	sort.Slice(amazonLinuxImages.Images, func(i, j int) bool {
		return *amazonLinuxImages.Images[i].CreationDate > *amazonLinuxImages.Images[j].CreationDate
	})

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

TOKEN=$(curl -sSL -X PUT "http://169.254.169.254/latest/api/token" -H "X-aws-ec2-metadata-token-ttl-seconds: 30")
INSTANCE_ID=$(curl -sSL -H "X-aws-ec2-metadata-token: ${TOKEN}" http://169.254.169.254/latest/meta-data/instance-id)

curl -fsSL https://tailscale.com/install.sh | sh
sudo tailscale up --auth-key=` + app.Config.Tailscale.AuthKey + ` --hostname=tailout-` + region + `-${INSTANCE_ID} --advertise-exit-node --ssh
sudo echo "sudo shutdown" | at now + ` + strconv.Itoa(durationMinutes) + ` minutes`

	// Encode the string in base64
	userDataScriptBase64 := base64.StdEncoding.EncodeToString([]byte(userDataScript))

	// Create instance input parameters
	runInput := &ec2.RunInstancesInput{
		ImageId:      aws.String(imageID),
		InstanceType: types.InstanceTypeT3aMicro,
		MinCount:     aws.Int32(1),
		MaxCount:     aws.Int32(1),
		UserData:     aws.String(userDataScriptBase64),
		DryRun:       aws.Bool(dryRun),
		InstanceMarketOptions: &types.InstanceMarketOptionsRequest{
			MarketType: types.MarketTypeSpot,
			SpotOptions: &types.SpotMarketOptions{
				InstanceInterruptionBehavior: types.InstanceInterruptionBehaviorTerminate,
			},
		},
	}

	stsSvc := sts.NewFromConfig(cfg)

	identity, err := stsSvc.GetCallerIdentity(context.TODO(), &sts.GetCallerIdentityInput{})
	if err != nil {
		return fmt.Errorf("failed to get account ID: %w", err)
	}

	fmt.Printf(`Creating tailout node in AWS with the following parameters:
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
	runResult, err := ec2Svc.RunInstances(context.TODO(), runInput)
	if err != nil {
		return fmt.Errorf("failed to create EC2 instance: %w", err)
	}

	if len(runResult.Instances) == 0 {
		fmt.Println("No instances created.")
		return nil
	}

	createdInstance := runResult.Instances[0]
	fmt.Println("EC2 instance created successfully:", *createdInstance.InstanceId)
	nodeName := fmt.Sprintf("tailout-%s-%s", region, *createdInstance.InstanceId)
	fmt.Println("Instance will be named", nodeName)
	// Create tags for the instance
	tags := []types.Tag{
		{
			Key:   aws.String("App"),
			Value: aws.String("tailout"),
		},
		{
			Key:   aws.String("Name"),
			Value: aws.String(nodeName),
		},
	}

	// Add the tags to the instance
	_, err = ec2Svc.CreateTags(context.TODO(), &ec2.CreateTagsInput{
		Resources: []string{*createdInstance.InstanceId},
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
	err = ec2.NewInstanceExistsWaiter(ec2Svc).Wait(context.TODO(), &ec2.DescribeInstancesInput{
		InstanceIds: []string{*createdInstance.InstanceId},
	}, time.Minute*2)
	if err != nil {
		return fmt.Errorf("failed to wait for instance to be created: %w", err)
	}

	fmt.Println("OK.")
	fmt.Println("Waiting for instance to join tailnet...")

	// Call internal.GetNodes periodically and search for the instance
	// If the instance is found, print the command to use it as an exit node

	timeout := time.Now().Add(3 * time.Minute)

	baseURL, err := url.Parse(app.Config.Tailscale.BaseURL)
	if err != nil {
		return fmt.Errorf("failed to parse base URL: %w", err)
	}

	client := &tsapi.Client{
		APIKey:  app.Config.Tailscale.APIKey,
		Tailnet: app.Config.Tailscale.Tailnet,
		BaseURL: baseURL,
	}

	for {
		nodes, err := client.Devices().List(context.TODO())
		if err != nil {
			return fmt.Errorf("failed to get devices: %w", err)
		}

		for _, node := range nodes {
			if node.Hostname == nodeName {
				goto found
			}
		}

		// Timeouts after 2 minutes
		if time.Now().After(timeout) {
			return errors.New("timeout waiting for instance to join tailnet")
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
		InstanceIds: []string{*createdInstance.InstanceId},
	}

	describeResult, err := ec2Svc.DescribeInstances(context.TODO(), describeInput)
	if err != nil {
		return fmt.Errorf("failed to describe EC2 instance: %w", err)
	}

	if len(describeResult.Reservations) == 0 {
		return errors.New("no reservations found")
	}

	reservation := describeResult.Reservations[0]
	if len(reservation.Instances) == 0 {
		return errors.New("no instances found")
	}

	instance := reservation.Instances[0]
	if instance.PublicIpAddress == nil {
		return errors.New("no public IP address found")
	}

	fmt.Printf("Node %s joined tailnet.\n", nodeName)
	fmt.Println("Public IP address:", *instance.PublicIpAddress)
	fmt.Println("Planned termination time:", time.Now().Add(duration).Format(time.RFC3339))

	if connect {
		fmt.Println()
		args := []string{nodeName}
		err = app.Connect(args)
		if err != nil {
			return fmt.Errorf("failed to connect to node: %w", err)
		}
	}
	return nil
}
