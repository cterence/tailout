package types

type Node interface {
	GetID() string
	GetAuthKey() string
	GetRegion() string
	GetShutdown() string
	GetType() string
	GetStatus() string
	GetInstanceIdCmd() string
	// GetCost() currency.Unit
}

type NodeConfig struct {
	Region       string
	AuthKey      string
	Shutdown     string
	InstanceType string
}

const UserDataTmpl = `#!/bin/bash
echo 'net.ipv4.ip_forward = 1' | sudo tee -a /etc/sysctl.conf
echo 'net.ipv6.conf.all.forwarding = 1' | sudo tee -a /etc/sysctl.conf
sudo sysctl -p /etc/sysctl.conf

export INSTANCE_ID=$({{ .GetInstanceIdCmd }})

curl -fsSL https://tailscale.com/install.sh | sh
sudo tailscale up --auth-key={{ .GetAuthKey }} --hostname=tailout-$INSTANCE_ID --advertise-exit-node --ssh
sudo echo "sudo shutdown now" | at now + {{ .GetShutdown }} minutes`
