package gcp

import (
	"context"
	"fmt"

	compute "cloud.google.com/go/compute/apiv1"
	"cloud.google.com/go/compute/apiv1/computepb"
	"github.com/cterence/tailout/pkg/types"
	"google.golang.org/protobuf/proto"
)

type Provider struct {
	Client *compute.InstancesClient
}

func (p *Provider) CreateNode(ctx *context.Context, nc types.NodeConfig, dryRun bool) (types.Node, error) {
	defer p.Client.Close()

	n := Node{
		AuthKey:  nc.AuthKey,
		Shutdown: nc.Shutdown,
		Region:   nc.Region,
	}

	instance := &computepb.Instance{
		Name:        proto.String("tailout"),
		MachineType: proto.String(fmt.Sprintf("zones/%s/machineTypes/%s", nc.Region, nc.InstanceType)), // Specify the machine type
		Disks: []*computepb.AttachedDisk{
			{
				Boot: proto.Bool(true),
				InitializeParams: &computepb.AttachedDiskInitializeParams{
					DiskSizeGb:  proto.Int64(10),
					SourceImage: proto.String("projects/debian-cloud/global/images/family/debian-10"), // Specify the image
				},
			},
		},
		NetworkInterfaces: []*computepb.NetworkInterface{
			{
				Name: proto.String("default"), // Use the default network
			},
		},
	}

	// Create the instance
	op, err := p.Client.Insert(*ctx, &computepb.InsertInstanceRequest{
		InstanceResource: instance,
	})
	if err != nil {
		return n, fmt.Errorf("failed to create instance: %v", err)
	}

	// Wait for the operation to complete
	fmt.Println("Waiting for the instance to be created...")
	err = op.Wait(*ctx)
	if err != nil {
		return n, fmt.Errorf("failed to wait for instance creation: %v", err)
	}

	fmt.Println("Instance created successfully")

	return n, nil
}

func (p *Provider) GetNode(nodeId string) (types.Node, error) {
	var node types.Node
	return node, nil
}

func (p *Provider) ListNodes() ([]types.Node, error) {
	return nil, nil
}

func (p *Provider) DeleteNode(nodeID string) error {
	return nil
}
