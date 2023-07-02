package tailscale

import "github.com/cterence/xit/xit/config"

type Client struct {
	config *config.TailscaleConfig
}

func NewClient(c *config.TailscaleConfig) *Client {
	return &Client{
		config: c,
	}
}
