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
	Short: "Initialize tailnet for xit",
	Long: `This command will initialize your tailnet for xit.
	
In details it will:
- add a tag:xit to your ACL
- update autoapprovers to allow exit nodes to be created
- add a ssh configuration allowing users to ssh into tagged xit machines`,
	Run: func(cmd *cobra.Command, args []string) {
		tsApiKey := viper.GetString("ts_api_key")
		tailnet := viper.GetString("ts_tailnet")
		dryRun := viper.GetBool("dry_run")

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

		allowXitSSH := SSHConfiguration{
			Action: "check",
			Src:    []string{"autogroup:members"},
			Dst:    []string{"tag:xit"},
			Users:  []string{"autogroup:nonroot", "root"},
		}

		xitSSHConfigExists := false

		for _, sshConfig := range config.SSH {
			if sshConfig.Action == "check" && sshConfig.Src[0] == "autogroup:members" && sshConfig.Dst[0] == "tag:xit" && sshConfig.Users[0] == "autogroup:nonroot" && sshConfig.Users[1] == "root" {
				xitSSHConfigExists = true
			}
		}

		if !xitSSHConfigExists {
			config.SSH = append(config.SSH, allowXitSSH)
		}

		// Validate the updated ACL configuration
		err = validateACL(tsApiKey, baseURL, tailnet, config)
		if err != nil {
			fmt.Println("Failed to validate ACL:", err)
			return
		}

		// Update the ACL configuration
		if !dryRun {
			err = updateACL(tsApiKey, baseURL, tailnet, config)
			if err != nil {
				fmt.Println("Failed to update ACL:", err)
				return
			}
		}
	},
}

func init() {
	cobra.OnInitialize(InitConfig)

	rootCmd.AddCommand(initCmd)

	initCmd.PersistentFlags().StringP("ts-api-key", "", "", "TailScale API Key")
	initCmd.PersistentFlags().StringP("ts-tailnet", "", "", "TailScale Tailnet")

	viper.BindPFlag("ts_api_key", initCmd.PersistentFlags().Lookup("ts-api-key"))
	viper.BindPFlag("ts_tailnet", initCmd.PersistentFlags().Lookup("ts-tailnet"))

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// initCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
