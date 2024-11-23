package aws

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"text/template"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2Types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/cterence/tailout/pkg/types"
)

type Provider struct {
	Client      aws.Config
	Region      ec2Types.Region
	Credentials aws.Credentials
}

func (p *Provider) Init(ctx context.Context, pc types.ProviderConfig) error {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(pc.Region))
	if err != nil {
		return fmt.Errorf("failed to load AWS SDK config, %v", err)
	}
	creds, err := cfg.Credentials.Retrieve(ctx)
	if err != nil {
		return fmt.Errorf("failed to get credentials: %v", err)
	}
	p.Client = cfg
	p.Credentials = creds
	return nil
}

func (p *Provider) CreateNode(ctx context.Context, nc types.NodeConfig, dryRun bool) (types.Node, error) {
	fmt.Println(p.Details())
	n := Node{
		AuthKey:  nc.AuthKey,
		Shutdown: nc.Shutdown,
	}
	ec2Svc := ec2.NewFromConfig(p.Client)

	// TODO: give more control on this ?
	// TODO: paginated request
	amazonLinuxAMIsOutput, err := ec2Svc.DescribeImages(ctx, &ec2.DescribeImagesInput{
		Filters: []ec2Types.Filter{
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
	if len(amazonLinuxAMIsOutput.Images) == 0 {
		return n, fmt.Errorf("failed to get AMIs: no AMI retrieved from filter")
	}

	bestAMI := amazonLinuxAMIsOutput.Images[0]

	if len(amazonLinuxAMIsOutput.Images) > 0 {
		for _, i := range amazonLinuxAMIsOutput.Images[1:] {
			bestImageTime, err := time.Parse(time.RFC3339, *bestAMI.DeprecationTime)
			if err != nil {
				return n, fmt.Errorf("failed to parse AMI deprecation time: %v", err)
			}
			currentImageTime, err := time.Parse(time.RFC3339, *bestAMI.DeprecationTime)
			if err != nil {
				return n, fmt.Errorf("failed to parse AMI deprecation time: %v", err)
			}
			if currentImageTime.After(bestImageTime) {
				bestAMI = i
			}
		}
	}

	// TODO: generic method for logging in to tailscale or headscale
	userDataTmpl, err := template.New("userData").Parse(types.UserDataTmpl)
	if err != nil {
		return n, fmt.Errorf("failed to parse user data template file: %v", err)
	}
	userDataBuffer := bytes.NewBufferString("")

	err = userDataTmpl.Execute(userDataBuffer, n)
	if err != nil {
		return n, fmt.Errorf("failed to execute user data template: %v", err)
	}

	userData := base64.StdEncoding.EncodeToString(userDataBuffer.Bytes())

	runInput := &ec2.RunInstancesInput{
		ImageId:      bestAMI.ImageId,
		InstanceType: ec2Types.InstanceType(n.GetType()),
		MinCount:     aws.Int32(1),
		MaxCount:     aws.Int32(1),
		UserData:     aws.String(userData),
		DryRun:       aws.Bool(dryRun),
		KeyName:      aws.String("stronghold"),
		InstanceMarketOptions: &ec2Types.InstanceMarketOptionsRequest{
			MarketType: ec2Types.MarketTypeSpot,
			SpotOptions: &ec2Types.SpotMarketOptions{
				InstanceInterruptionBehavior: ec2Types.InstanceInterruptionBehaviorTerminate,
			},
		},
	}

	// Run the EC2 instance
	runResult, err := ec2Svc.RunInstances(ctx, runInput)
	if err != nil {
		return n, fmt.Errorf("failed to create EC2 instance: %w", err)
	}

	if len(runResult.Instances) == 0 {
		return n, fmt.Errorf("no instances created")
	}

	n.InAWS = runResult.Instances[0]

	// Add a handler for the instance state change event
	err = ec2.NewInstanceExistsWaiter(ec2Svc).Wait(ctx, &ec2.DescribeInstancesInput{
		InstanceIds: []string{n.GetID()},
	}, time.Minute*2)
	if err != nil {
		return n, fmt.Errorf("failed to wait for instance to be created: %v", err)
	}

	fmt.Println("instance is running")

	return n, nil
}

func (p *Provider) Details() string {
	return fmt.Sprintf(`- Provider: AWS
- Account ID: %s
- Region: %s
`, p.Credentials.AccountID, p.Client.Region)
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
