package types

import (
	"context"
)

type ProviderConfig struct {
	Region  string
	Account string
}

type Provider interface {
	Init(pc ProviderConfig) Provider
	CreateNode(ctx context.Context, config NodeConfig, dryRun bool) (Node, error)
	GetNode(nodeId string) (Node, error)
	ListNodes() ([]Node, error)
	DeleteNode(nodeID string) error
}
