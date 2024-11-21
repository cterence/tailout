package provider

import (
	"bytes"
	"context"
	"fmt"
	"text/template"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/cterence/tailout/pkg/node"
	"github.com/davecgh/go-spew/spew"
)

type AWSProvider struct {
	Config aws.Config
	Region types.Region
}

func (p *AWSProvider) CreateNode(ctx *context.Context, nc node.Config) (node.Node, error) {
	var n node.Node
	ec2Svc := ec2.NewFromConfig(p.Config)

	// TODO: give more control on this ?
	// TODO: paginated request
	amazonLinuxImages, err := ec2Svc.DescribeImages(*ctx, &ec2.DescribeImagesInput{
		Filters: []types.Filter{
			{
				Name:   aws.String("name"),
				Values: []string{"al2023-ami-2023.6*"},
			},
			{
				Name:   aws.String("architecture"),
				Values: []string{"x86_64"},
			},
		},
		Owners: []string{"amazon"},
	})
	if err != nil {
		return n, fmt.Errorf("failed to get AMIs: %v", err)
	}
	if len(amazonLinuxImages.Images) == 0 {
		return n, fmt.Errorf("failed to get AMIs: no AMI retrieved from filter")
	}

	bestImage := amazonLinuxImages.Images[0]

	if len(amazonLinuxImages.Images) > 0 {
		for _, i := range amazonLinuxImages.Images[1:] {
			bestImageTime, err := time.Parse(time.RFC3339, *bestImage.DeprecationTime)
			if err != nil {
				return n, fmt.Errorf("failed to parse AMI deprecation time: %v", err)
			}
			currentImageTime, err := time.Parse(time.RFC3339, *bestImage.DeprecationTime)
			if err != nil {
				return n, fmt.Errorf("failed to parse AMI deprecation time: %v", err)
			}
			if currentImageTime.After(bestImageTime) {
				bestImage = i
			}
		}
	}

	spew.Dump(bestImage)

	// TODO: give control over this
	// instanceTypes, err := ec2Svc.DescribeInstanceTypes(*ctx, &ec2.DescribeInstanceTypesInput{
	// 	Filters: []types.Filter{
	// 		{
	// 			Name:   aws.String("instance-type"),
	// 			Values: []string{"t3a*"},
	// 		},
	// 	},
	// })
	// if err != nil {
	// 	return node, fmt.Errorf("failed to get instance types: %v", err)
	// }

	// TODO: generic method for logging in to tailscale or headscale
	userDataTmpl, err := template.New("userData").Parse(node.UserDataTmpl)
	if err != nil {
		return n, fmt.Errorf("failed to parse user data template file: %v", err)
	}
	userDataBuffer := bytes.NewBufferString("")

	userDataInput := node.UserDataTmplInput{
		GetInstanceIdCmd: "curl http://169.254.169.254/latest/meta-data/instance-id -H \"X-aws-ec2-metadata-token: $(curl -X PUT http://169.254.169.254/latest/api/token -H \"X-aws-ec2-metadata-token-ttl-seconds: 5\")\"",
	}

	err = userDataTmpl.Execute(userDataBuffer, userDataInput)
	if err != nil {
		return n, fmt.Errorf("failed to execute user data template: %v", err)
	}

	fmt.Println(userDataBuffer.String())

	return n, nil
}

func (p *AWSProvider) GetNode(nodeId string) (node.Node, error) {
	var node node.Node
	return node, nil
}
func (p *AWSProvider) ListNodes() ([]node.Node, error) {
	return nil, nil
}
func (p *AWSProvider) DeleteNode(nodeID string) error {
	return nil
}
