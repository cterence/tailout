/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tailscale/hujson"
)

type Configuration struct {
	ACLs                []ACL               `json:"acls,omitempty"`
	Hosts               map[string]string   `json:"hosts,omitempty"`
	Groups              map[string][]string `json:"groups,omitempty"`
	Tests               []Test              `json:"tests,omitempty"`
	TagOwners           map[string][]string `json:"tagOwners,omitempty"`
	AutoApprovers       AutoApprovers       `json:"autoApprovers,omitempty"`
	SSH                 []SSHConfiguration  `json:"ssh,omitempty"`
	DerpMap             DerpMap             `json:"derpMap,omitempty"`
	DisableIPv4         bool                `json:"disableIPv4,omitempty"`
	RandomizeClientPort bool                `json:"randomizeClientPort,omitempty"`
}

type ACL struct {
	Action string   `json:"action,omitempty"`
	Src    []string `json:"src,omitempty"`
	Dst    []string `json:"dst,omitempty"`
	Proto  string   `json:"proto,omitempty"`
}

type Test struct {
	Src    string   `json:"src,omitempty"`
	Accept []string `json:"accept,omitempty"`
	Deny   []string `json:"deny,omitempty"`
}

type AutoApprovers struct {
	Routes   map[string][]string `json:"routes,omitempty"`
	ExitNode []string            `json:"exitNode,omitempty"`
}

type SSHConfiguration struct {
	Action string   `json:"action,omitempty"`
	Src    []string `json:"src,omitempty"`
	Dst    []string `json:"dst,omitempty"`
	Users  []string `json:"users,omitempty"`
}

type DerpMap struct {
	Regions map[string]DerpRegion `json:"regions,omitempty"`
}

type DerpRegion struct {
	RegionID int    `json:"regionID,omitempty"`
	HostName string `json:"hostName,omitempty"`
}

func getACL(tsApiKey, baseURL, tailnet string) ([]byte, error) {
	url := fmt.Sprintf("%s/api/v2/tailnet/%s/acl", baseURL, tailnet)

	// Create the HTTP request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Set the authentication header
	req.Header.Set("Authorization", "Bearer "+tsApiKey)
	req.Header.Set("Content-Type", "application/json")

	// Send the HTTP request
	client := http.DefaultClient
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send HTTP request: %w", err)
	}
	defer resp.Body.Close()

	// Check the response status code
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get ACL: %s", resp.Status)
	}

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read HTTP response: %w", err)
	}

	return body, nil
}

func validateACL(tsApiKey, baseURL, tailnet string, config Configuration) error {
	configString, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal ACL: %w", err)
	}

	url := fmt.Sprintf("%s/api/v2/tailnet/%s/acl/validate", baseURL, tailnet)

	// Create the HTTP request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(configString))
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Set the authentication header
	req.Header.Set("Authorization", "Bearer "+tsApiKey)
	req.Header.Set("Content-Type", "application/json")

	// Send the HTTP request
	client := http.DefaultClient
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send HTTP request: %w", err)
	}
	defer resp.Body.Close()

	// Check the response status code
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to validate ACL: %s", resp.Status)
	}

	return nil
}

func updateACL(tsApiKey, baseURL, tailnet string, config Configuration) error {
	configString, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal ACL: %w", err)
	}

	url := fmt.Sprintf("%s/api/v2/tailnet/%s/acl", baseURL, tailnet)

	// Create the HTTP request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(configString))
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Set the authentication header
	req.Header.Set("Authorization", "Bearer "+tsApiKey)
	req.Header.Set("Content-Type", "application/json")

	// Send the HTTP request
	client := http.DefaultClient
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send HTTP request: %w", err)
	}
	defer resp.Body.Close()

	// Check the response status code
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to update ACL: %s", resp.Status)
	}

	return nil
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		tsApiKey := viper.GetString("ts_api_key")
		tailnet := viper.GetString("ts_tailnet")

		fmt.Println("This command will add a tag:xit to your ACL and update autoapprovers to allow exit nodes to be created.")

		baseURL := "https://api.tailscale.com"

		// Get the ACL configuration
		acl, err := getACL(tsApiKey, baseURL, tailnet)
		if err != nil {
			fmt.Println("Failed to get ACL:", err)
			return
		}

		var config Configuration

		standardACL, err := hujson.Standardize(acl)
		if err != nil {
			fmt.Println("Failed to standardize ACL:", err)
			return
		}

		err = json.Unmarshal(standardACL, &config)
		if err != nil {
			fmt.Println("Failed to unmarshal ACL:", err)
			return
		}

		// Update the configuration
		config.TagOwners["tag:xit"] = []string{}
		config.AutoApprovers.ExitNode = []string{"tag:xit"}
		// TODO: find how to add a ssh configuration allowing users to ssh into tagged xit machines

		// Validate the updated ACL configuration
		err = validateACL(tsApiKey, baseURL, tailnet, config)
		if err != nil {
			fmt.Println("Failed to validate ACL:", err)
			return
		}

		// Update the ACL configuration
		err = updateACL(tsApiKey, baseURL, tailnet, config)
		if err != nil {
			fmt.Println("Failed to update ACL:", err)
			return
		}
	},
}

func init() {
	rootCmd.AddCommand(initCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// initCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
