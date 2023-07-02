package tailscale

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/cterence/tailout/tailout/config"
	"github.com/tailscale/hujson"
)

const (
	baseURL = "https://api.tailscale.com"
)

func sendTailscaleAPIRequest(c *Client, method, url string, body []byte) ([]byte, error) {
	req, err := http.NewRequest(method, url, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.config.APIKey)
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

func (c *Client) GetNode(id string) (config.Node, error) {
	url := fmt.Sprintf("%s/api/v2/device/%s", baseURL, id)

	body, err := sendTailscaleAPIRequest(c, "GET", url, nil)
	if err != nil {
		return config.Node{}, fmt.Errorf("failed to get node: %w", err)
	}

	var node config.Node
	err = json.Unmarshal(body, &node)
	if err != nil {
		return config.Node{}, fmt.Errorf("failed to unmarshal node: %w", err)
	}

	return node, nil
}

func (c *Client) GetNodes() ([]config.Node, error) {
	url := fmt.Sprintf("%s/api/v2/tailnet/%s/devices", baseURL, c.config.Tailnet)

	body, err := sendTailscaleAPIRequest(c, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get nodes: %w", err)
	}

	var userNodes config.UserNodes
	err = json.Unmarshal(body, &userNodes)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal nodes: %w", err)
	}

	return userNodes.Nodes, nil
}

func (c *Client) DeleteNode(id string) error {
	url := fmt.Sprintf("%s/api/v2/device/%s", baseURL, id)

	_, err := sendTailscaleAPIRequest(c, "DELETE", url, nil)
	if err != nil {
		return fmt.Errorf("failed to delete node: %w", err)
	}

	return nil
}

func (c *Client) GetActiveXitNodes() ([]config.Node, error) {
	nodes, err := c.GetNodes()
	if err != nil {
		return nil, fmt.Errorf("failed to get active tailout nodes: %w", err)
	}
	var foundNodes []config.Node
	for _, node := range nodes {
		lastSeen, err := time.Parse(time.RFC3339, node.LastSeen)
		if err != nil {
			return nil, err
		}

		hasTag := false
		for _, t := range node.Tags {
			if t == "tag:tailout" {
				hasTag = true
			}
		}

		if hasTag && time.Since(lastSeen) < 10*time.Minute {
			foundNodes = append(foundNodes, node)
		}
	}

	return foundNodes, nil
}

func (c *Client) GetPolicy() (config.Policy, error) {
	url := fmt.Sprintf("%s/api/v2/tailnet/%s/acl", baseURL, c.config.Tailnet)

	body, err := sendTailscaleAPIRequest(c, "GET", url, nil)
	if err != nil {
		return config.Policy{}, fmt.Errorf("failed to get policy: %w", err)
	}

	standardBody, err := hujson.Standardize(body)
	if err != nil {
		return config.Policy{}, fmt.Errorf("failed to standardize body: %w", err)
	}

	var policy config.Policy
	err = json.Unmarshal(standardBody, &policy)
	if err != nil {
		return config.Policy{}, fmt.Errorf("failed to unmarshal policy: %w", err)
	}

	return policy, nil
}

func (c *Client) ValidatePolicy(policy config.Policy) error {
	policyString, err := json.Marshal(policy)
	if err != nil {
		return fmt.Errorf("failed to marshal policy: %w", err)
	}

	url := fmt.Sprintf("%s/api/v2/tailnet/%s/acl/validate", baseURL, c.config.Tailnet)

	_, err = sendTailscaleAPIRequest(c, "POST", url, policyString)
	if err != nil {
		return fmt.Errorf("failed to validate policy: %w", err)
	}

	return nil
}

func (c *Client) UpdatePolicy(policy config.Policy) error {
	policyString, err := json.Marshal(policy)
	if err != nil {
		return fmt.Errorf("failed to marshal policy: %w", err)
	}

	url := fmt.Sprintf("%s/api/v2/tailnet/%s/acl", baseURL, c.config.Tailnet)

	_, err = sendTailscaleAPIRequest(c, "POST", url, policyString)
	if err != nil {
		return fmt.Errorf("failed to update policy: %w", err)
	}

	return nil
}
