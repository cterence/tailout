package internal

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/manifoldco/promptui"
	"github.com/tailscale/hujson"
)

type Policy struct {
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

type Node struct {
	Addresses                 []string `json:"addresses"`
	Authorized                bool     `json:"authorized"`
	BlocksIncomingConnections bool     `json:"blocksIncomingConnections"`
	ClientVersion             string   `json:"clientVersion"`
	Created                   string   `json:"created"`
	Expires                   string   `json:"expires"`
	Hostname                  string   `json:"hostname"`
	ID                        string   `json:"id"`
	IsExternal                bool     `json:"isExternal"`
	KeyExpiryDisabled         bool     `json:"keyExpiryDisabled"`
	LastSeen                  string   `json:"lastSeen"`
	MachineKey                string   `json:"NodeKey,omitempty"`
	Name                      string   `json:"name,omitempty"`
	NodeID                    string   `json:"nodeId"`
	NodeKey                   string   `json:"nodeKey"`
	OS                        string   `json:"os"`
	TailnetLockError          string   `json:"tailnetLockError,omitempty"`
	TailnetLockKey            string   `json:"tailnetLockKey,omitempty"`
	UpdateAvailable           bool     `json:"updateAvailable"`
	User                      string   `json:"user,omitempty"`
	Tags                      []string `json:"tags,omitempty"`
}

type TailscaleStatus struct {
	ControlURL             string `json:"ControlURL"`
	RouteAll               bool   `json:"RouteAll"`
	AllowSingleHosts       bool   `json:"AllowSingleHosts"`
	ExitNodeID             string `json:"ExitNodeID"`
	ExitNodeIP             string `json:"ExitNodeIP"`
	ExitNodeAllowLANAccess bool   `json:"ExitNodeAllowLANAccess"`
	CorpDNS                bool   `json:"CorpDNS"`
	RunSSH                 bool   `json:"RunSSH"`
	WantRunning            bool   `json:"WantRunning"`
	LoggedOut              bool   `json:"LoggedOut"`
	ShieldsUp              bool   `json:"ShieldsUp"`
	AdvertiseTags          string `json:"AdvertiseTags"`
	Hostname               string `json:"Hostname"`
	NotepadURLs            bool   `json:"NotepadURLs"`
	AdvertiseRoutes        string `json:"AdvertiseRoutes"`
	NoSNAT                 bool   `json:"NoSNAT"`
	NetfilterMode          int    `json:"NetfilterMode"`
	Config                 struct {
		PrivateMachineKey string `json:"PrivateMachineKey"`
		PrivateNodeKey    string `json:"PrivateNodeKey"`
		OldPrivateNodeKey string `json:"OldPrivateNodeKey"`
		Provider          string `json:"Provider"`
		LoginName         string `json:"LoginName"`
		UserProfile       struct {
			ID            int64    `json:"ID"`
			LoginName     string   `json:"LoginName"`
			DisplayName   string   `json:"DisplayName"`
			ProfilePicURL string   `json:"ProfilePicURL"`
			Roles         []string `json:"Roles"`
		} `json:"UserProfile"`
		NetworkLockKey string `json:"NetworkLockKey"`
		NodeID         string `json:"NodeID"`
	} `json:"Config"`
}

type UserNodes struct {
	User  string `json:"user"`
	Nodes []Node `json:"devices"`
}

const (
	baseURL = "https://api.tailscale.com"
)

// Create a method HasTag for Node
func (d Node) HasTag(tag string) bool {
	for _, t := range d.Tags {
		if t == tag {
			return true
		}
	}
	return false
}

func GetNodes(tsApiKey, tailnet string) ([]Node, error) {
	url := fmt.Sprintf("%s/api/v2/tailnet/%s/devices", baseURL, tailnet)

	body, err := sendRequest(tsApiKey, tailnet, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get nodes: %w", err)
	}

	var userNodes UserNodes
	err = json.Unmarshal(body, &userNodes)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal nodes: %w", err)
	}

	return userNodes.Nodes, nil
}

func GetPolicy(tsApiKey, tailnet string) (Policy, error) {
	url := fmt.Sprintf("%s/api/v2/tailnet/%s/acl", baseURL, tailnet)

	body, err := sendRequest(tsApiKey, tailnet, "GET", url, nil)
	if err != nil {
		return Policy{}, fmt.Errorf("failed to get ACL: %w", err)
	}

	standardBody, err := hujson.Standardize(body)
	if err != nil {
		return Policy{}, fmt.Errorf("failed to standardize body: %w", err)
	}

	var policy Policy
	err = json.Unmarshal(standardBody, &policy)
	if err != nil {
		return Policy{}, fmt.Errorf("failed to unmarshal policy: %w", err)
	}

	return policy, nil
}

func ValidatePolicy(tsApiKey, tailnet string, config Policy) error {
	configString, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal policy: %w", err)
	}

	url := fmt.Sprintf("%s/api/v2/tailnet/%s/acl/validate", baseURL, tailnet)

	_, err = sendRequest(tsApiKey, tailnet, "POST", url, configString)
	if err != nil {
		return fmt.Errorf("failed to validate policy: %w", err)
	}

	return nil
}

func UpdatePolicy(tsApiKey, tailnet string, config Policy) error {
	configString, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal policy: %w", err)
	}

	url := fmt.Sprintf("%s/api/v2/tailnet/%s/acl", baseURL, tailnet)

	_, err = sendRequest(tsApiKey, tailnet, "POST", url, configString)
	if err != nil {
		return fmt.Errorf("failed to update policy: %w", err)
	}

	return nil
}

func GetNode(tsApiKey, id string) (Node, error) {
	url := fmt.Sprintf("%s/api/v2/device/%s", baseURL, id)

	body, err := sendRequest(tsApiKey, "", "GET", url, nil)
	if err != nil {
		return Node{}, fmt.Errorf("failed to get node: %w", err)
	}

	var node Node
	err = json.Unmarshal(body, &node)
	if err != nil {
		return Node{}, fmt.Errorf("failed to unmarshal node: %w", err)
	}

	return node, nil
}

func DeleteNode(tsApiKey, id string) error {
	url := fmt.Sprintf("%s/api/v2/device/%s", baseURL, id)

	_, err := sendRequest(tsApiKey, "", "DELETE", url, nil)
	if err != nil {
		return fmt.Errorf("failed to delete node: %w", err)
	}

	return nil
}

func FindNodeByHostname(tsApiKey, hostname, tailnet string) (Node, error) {
	nodes, err := GetNodes(tsApiKey, tailnet)
	if err != nil {
		return Node{}, fmt.Errorf("failed to get nodes: %w", err)
	}

	if err != nil {
		return Node{}, fmt.Errorf("failed to unmarshal nodes: %w", err)
	}

	for _, node := range nodes {
		if node.Hostname == hostname {
			return node, nil
		}
	}

	return Node{}, fmt.Errorf("node with hostname %s not found", hostname)
}

func FindNodesByHostname(tsApiKey, tailnet string, hostnames []string) ([]Node, error) {
	nodes, err := GetNodes(tsApiKey, tailnet)
	if err != nil {
		return nil, fmt.Errorf("failed to get nodes: %w", err)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal nodes: %w", err)
	}

	var foundNodes []Node
	for _, node := range nodes {
		for _, hostname := range hostnames {
			if node.Hostname == hostname {
				foundNodes = append(foundNodes, node)
			}
		}
	}

	return foundNodes, nil
}

func FindActiveXitNodes(tsApiKey, tailnet string) ([]Node, error) {
	nodes, err := GetNodes(tsApiKey, tailnet)
	if err != nil {
		return nil, fmt.Errorf("failed to get nodes: %w", err)
	}
	var foundNodes []Node
	for _, node := range nodes {
		lastSeen, err := time.Parse(time.RFC3339, node.LastSeen)
		if err != nil {
			fmt.Println("Failed to parse lastSeen:", err)
			return nil, err
		}
		if node.HasTag("tag:xit") && time.Since(lastSeen) < 10*time.Minute {
			foundNodes = append(foundNodes, node)
		}
	}

	return foundNodes, nil
}

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
