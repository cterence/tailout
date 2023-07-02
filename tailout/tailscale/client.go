package tailscale

import "github.com/cterence/tailout/tailout/config"

type Client struct {
	config *config.TailscaleConfig
}

func NewClient(c *config.TailscaleConfig) *Client {
	return &Client{
		config: c,
	}
}
