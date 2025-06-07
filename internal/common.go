package internal

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/manifoldco/promptui"
	tsapi "tailscale.com/client/tailscale/v2"
)

func GetRegions(ctx context.Context) ([]string, error) {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion("us-east-1"))
	if err != nil {
		return nil, fmt.Errorf("failed to load default config: %w", err)
	}
	ec2Svc := ec2.NewFromConfig(cfg)

	regions, err := ec2Svc.DescribeRegions(ctx, &ec2.DescribeRegionsInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to describe regions: %w", err)
	}

	regionNames := []string{}
	for _, region := range regions.Regions {
		regionNames = append(regionNames, *region.RegionName)
	}

	return regionNames, nil
}

// Function that uses promptui to return an AWS region fetched from the aws sdk.
func SelectRegion(ctx context.Context) (string, error) {
	regionNames, err := GetRegions(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to retrieve regions: %w", err)
	}

	sort.Slice(regionNames, func(i, j int) bool {
		return regionNames[i] < regionNames[j]
	})

	// Prompt for region with promptui displaying 15 regions at a time, sorted alphabetically and searchable
	prompt := promptui.Select{
		Label: "Select AWS region",
		Items: regionNames,
	}

	_, region, err := prompt.Run()
	if err != nil {
		return "", fmt.Errorf("failed to select region: %w", err)
	}

	return region, nil
}

// Function that uses promptui to return a boolean value.
func PromptYesNo(question string) (bool, error) {
	prompt := promptui.Select{
		Label: question,
		Items: []string{"Yes", "No"},
	}

	_, result, err := prompt.Run()
	if err != nil {
		return false, fmt.Errorf("failed to prompt for yes/no: %w", err)
	}

	if result == "Yes" {
		return true, nil
	}

	return false, nil
}

func GetActiveNodes(ctx context.Context, c *tsapi.Client) ([]tsapi.Device, error) {
	devices, err := c.Devices().List(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get devices: %w", err)
	}

	tailoutDevices := make([]tsapi.Device, 0)
	for _, device := range devices {
		for _, tag := range device.Tags {
			if tag == "tag:tailout" {
				if time.Duration(device.LastSeen.Minute()) < 10*time.Minute {
					tailoutDevices = append(tailoutDevices, device)
				}
			}
		}
	}

	return tailoutDevices, nil
}
