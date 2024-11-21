package aws

import (
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

type Node struct {
	InAWS    types.Instance
	Region   string
	AuthKey  string
	Shutdown string
}

func (i Node) GetID() string {
	return *i.InAWS.InstanceId
}

func (i Node) GetType() string {
	return string(i.InAWS.InstanceType)
}

func (i Node) GetStatus() string {
	return string(i.InAWS.State.Name)
}

func (i Node) GetRegion() string {
	return i.Region
}

func (i Node) GetAuthKey() string {
	return i.AuthKey
}

func (i Node) GetShutdown() string {
	return i.Shutdown
}

func (i Node) GetInstanceIdCmd() string {
	return "curl http://169.254.169.254/latest/meta-data/instance-id -H \"X-aws-ec2-metadata-token: $(curl -X PUT http://169.254.169.254/latest/api/token -H \"X-aws-ec2-metadata-token-ttl-seconds: 5\")\""
}

// func (i Node) GetCost() currency.Unit {
// 	return currency.USD.Amount()
// }
