package types

import (
	"context"
)

type CloudProvider interface {
	CreateNode(ctx *context.Context, config NodeConfig, dryRun bool) (Node, error)
	GetNode(nodeId string) (Node, error)
	ListNodes() ([]Node, error)
	DeleteNode(nodeID string) error
}
