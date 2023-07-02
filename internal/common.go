package internal

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"regexp"
	"sort"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/manifoldco/promptui"
)

const (
	baseURL = "https://api.tailscale.com"
)

// Function that uses promptui to return an AWS region fetched from the aws sdk
func SelectRegion() (string, error) {
	sess, err := session.NewSession(&aws.Config{})
	if err != nil {
		fmt.Println("Failed to create session:", err)
		return "", err
	}

	svc := ec2.New(sess, aws.NewConfig().WithRegion("us-east-1"))
	regions, err := svc.DescribeRegions(&ec2.DescribeRegionsInput{})
	if err != nil {
		fmt.Println("Failed to describe regions:", err)
		return "", err
	}

	regionNames := []string{}
	for _, region := range regions.Regions {
		regionNames = append(regionNames, *region.RegionName)
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

// Function that takes every common code in the above function and makes it a function
func sendRequest(tsApiKey, tailnet, method, url string, body []byte) ([]byte, error) {
	req, err := http.NewRequest(method, url, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+tsApiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send HTTP request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get OK status code: %s", resp.Status)
	}

	body, err = io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read HTTP response: %w", err)
	}

	return body, nil
}

// Function that uses promptui to return a boolean value
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

func RunTailscaleUpCommand(command string, nonInteractive bool) error {
	tailscaleCommand := strings.Split(command, " ")

	if nonInteractive {
		tailscaleCommand = append([]string{"-n"}, tailscaleCommand...)
	}

	fmt.Println("Running command:\nsudo", strings.Join(tailscaleCommand, " "))

	if !nonInteractive {
		result, err := PromptYesNo("Are you sure you want to run this command?")
		if err != nil {
			return fmt.Errorf("failed to prompt for confirmation: %w", err)
		}

		if !result {
			fmt.Println("Aborting...")
			return nil
		}
	}

	out, err := exec.Command("sudo", tailscaleCommand...).CombinedOutput()
	// If the command was unsuccessful, extract tailscale up command from error message with a regex and run it
	if err != nil {
		// extract latest "tailscale up" command from output with a regex and run it
		regexp := regexp.MustCompile(`tailscale up .*`)
		loggedTailscaleCommand := regexp.FindString(string(out))

		if loggedTailscaleCommand == "" {
			return fmt.Errorf("failed to find tailscale up command in output: %s", string(out))
		}

		fmt.Printf("Existing Tailscale configuration found, will run updated tailscale up command:\nsudo %s\n", loggedTailscaleCommand)

		// Use promptui for the confirmation prompt
		if !nonInteractive {
			result, err := PromptYesNo("Are you sure you want to run this command?")
			if err != nil {
				return fmt.Errorf("failed to prompt for confirmation: %w", err)
			}

			if !result {
				fmt.Println("Aborting...")
				return nil
			}
		}

		tailscaleCommand = strings.Split(loggedTailscaleCommand, " ")

		if nonInteractive {
			tailscaleCommand = append([]string{"-n"}, tailscaleCommand...)
		}

		_, err = exec.Command("sudo", tailscaleCommand...).CombinedOutput()
		if err != nil {
			return fmt.Errorf("failed to run command: %w", err)
		}
	}
	return nil
}
