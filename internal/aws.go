package internal

import (
	"context"
	"fmt"
	"log"
	"sort"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/manifoldco/promptui"
)

func GetAWSRegions() ([]string, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion("us-east-1"))
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}
	ec2Svc := ec2.NewFromConfig(cfg)

	regions, err := ec2Svc.DescribeRegions(context.TODO(), &ec2.DescribeRegionsInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to describe regions: %w", err)
	}

	regionNames := []string{}
	for _, region := range regions.Regions {
		regionNames = append(regionNames, *region.RegionName)
	}

	return regionNames, nil
}

// Function that uses promptui to return an AWS region fetched from the aws sdk
func SelectAWSRegion() (string, error) {
	regionNames, err := GetAWSRegions()
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
