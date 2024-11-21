package factory

import (
	"context"
	"fmt"

	compute "cloud.google.com/go/compute/apiv1"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/cterence/tailout/pkg/aws"
	"github.com/cterence/tailout/pkg/gcp"
	"github.com/cterence/tailout/pkg/types"
)

func GetCloudProvider(providerName string, region string) (types.CloudProvider, error) {
	switch providerName {
	case "aws":
		cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(region))
		if err != nil {
			return nil, fmt.Errorf("failed to load AWS SDK config, %v", err)
		}
		return &aws.Provider{Client: cfg}, nil
	case "gcp":
		instancesClient, err := compute.NewInstancesRESTClient(context.TODO())
		if err != nil {
			return nil, fmt.Errorf("failed to create GCP instances client: %v", err)
		}
		return &gcp.Provider{Client: instancesClient}, nil
	// case "azure":
	// 		return &AzureProvider{client: azure.NewClient()}, nil
	default:
		return nil, fmt.Errorf("unsupported cloud provider: %s", providerName)
	}
}
