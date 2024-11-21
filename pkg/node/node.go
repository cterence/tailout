package node

import (
	"golang.org/x/text/currency"
)

type Config struct {
	Region string
}

type Node struct {
	ID          string
	Config      Config
	CostPerHour currency.Unit
}

type UserDataTmplInput struct {
	GetInstanceIdCmd      string
	TailscaleAuthKey      string
	MinutesBeforeShutdown string
}

const UserDataTmpl = `echo 'net.ipv4.ip_forward = 1' | sudo tee -a /etc/sysctl.conf
echo 'net.ipv6.conf.all.forwarding = 1' | sudo tee -a /etc/sysctl.conf
sudo sysctl -p /etc/sysctl.conf

export INSTANCE_ID=$({{ .GetInstanceIdCmd }})

curl -fsSL https://tailscale.com/install.sh | sh
sudo tailscale up --auth-key={{ .TailscaleAuthKey }} --hostname=tailout-$INSTANCE_ID --advertise-exit-node --ssh
sudo echo "sudo shutdown now" | at now + {{ .MinutesBeforeShutdown }} minutes`
