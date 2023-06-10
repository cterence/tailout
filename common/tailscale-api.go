package common

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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

type Device struct {
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
	MachineKey                string   `json:"machineKey,omitempty"`
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

type UserDevices struct {
	User    string   `json:"user"`
	Devices []Device `json:"devices"`
}

const (
	baseURL = "https://api.tailscale.com"
)

func GetDevices(tsApiKey, tailnet string) ([]byte, error) {
	url := fmt.Sprintf("%s/api/v2/tailnet/%s/devices", baseURL, tailnet)

	body, err := sendRequest(tsApiKey, tailnet, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get devices: %w", err)
	}

	return body, nil
}

func GetACL(tsApiKey, tailnet string) ([]byte, error) {
	url := fmt.Sprintf("%s/api/v2/tailnet/%s/acl", baseURL, tailnet)

	body, err := sendRequest(tsApiKey, tailnet, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get ACL: %w", err)
	}

	return body, nil
}

func ValidateACL(tsApiKey, tailnet string, config Configuration) error {
	configString, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal ACL: %w", err)
	}

	url := fmt.Sprintf("%s/api/v2/tailnet/%s/acl/validate", baseURL, tailnet)

	_, err = sendRequest(tsApiKey, tailnet, "POST", url, configString)
	if err != nil {
		return fmt.Errorf("failed to validate ACL: %w", err)
	}

	return nil
}

func UpdateACL(tsApiKey, tailnet string, config Configuration) error {
	configString, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal ACL: %w", err)
	}

	url := fmt.Sprintf("%s/api/v2/tailnet/%s/acl", baseURL, tailnet)

	_, err = sendRequest(tsApiKey, tailnet, "PUT", url, configString)
	if err != nil {
		return fmt.Errorf("failed to update ACL: %w", err)
	}

	return nil
}

func GetDevice(device, tsApiKey string) ([]byte, error) {
	url := fmt.Sprintf("%s/api/v2/device/%s", baseURL, device)

	body, err := sendRequest(tsApiKey, "", "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get device: %w", err)
	}

	return body, nil
}

// Function that takes every common code in the above function and makes it a function
func sendRequest(tsApiKey, tailnet, method, url string, body []byte) ([]byte, error) {
	req, err := http.NewRequest(method, url, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+tsApiKey)
	req.Header.Set("Content-Type", "application/json")

	client := http.DefaultClient
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send HTTP request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to send HTTP request: %s", resp.Status)
	}

	body, err = io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read HTTP response: %w", err)
	}

	return body, nil
}
