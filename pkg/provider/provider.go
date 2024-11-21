package provider

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/cterence/tailout/pkg/node"
)

type CloudProvider interface {
	CreateNode(ctx *context.Context, config node.Config) (node.Node, error)
	GetNode(nodeId string) (node.Node, error)
	ListNodes() ([]node.Node, error)
	DeleteNode(nodeID string) error
}

func GetCloudProvider(providerName string, region string) (CloudProvider, error) {
	switch providerName {
	case "aws":
		cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(region))
		if err != nil {
			return nil, fmt.Errorf("unable to load SDK config, %v", err)
		}
		return &AWSProvider{Config: cfg}, nil
	// case "gcp":
	// 		return &GCPProvider{client: gcp.NewClient()}, nil
	// case "azure":
	// 		return &AzureProvider{client: azure.NewClient()}, nil
	default:
		return nil, fmt.Errorf("unsupported cloud provider: %s", providerName)
	}
}
