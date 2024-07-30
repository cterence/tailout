package internal

import (
	"context"
	"fmt"
	"time"

	"github.com/manifoldco/promptui"
	"github.com/tailscale/tailscale-client-go/tailscale"
)

const (
	baseURL = "https://api.tailscale.com"
)

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

func GetActiveNodes(c *tailscale.Client) ([]tailscale.Device, error) {
	devices, err := c.Devices(context.TODO())
	if err != nil {
		return nil, err
	}

	tailoutDevices := make([]tailscale.Device, 0)
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
