package gcp

import "google.golang.org/api/compute/v1"

type Node struct {
	InGCP    compute.Instance
	Region   string
	AuthKey  string
	Shutdown string
}

func (i Node) GetID() string {
	return i.InGCP.Name
}

func (i Node) GetType() string {
	return i.InGCP.Kind
}

func (i Node) GetStatus() string {
	return i.InGCP.Status
}

func (i Node) GetRegion() string {
	return i.InGCP.Zone
}

func (i Node) GetAuthKey() string {
	return i.AuthKey
}

func (i Node) GetShutdown() string {
	return i.Shutdown
}

func (i Node) GetInstanceIdCmd() string {
	return ""
}

// func (i Node) GetCost() currency.Unit {
// 	return currency.USD.Amount()
// }
