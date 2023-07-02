package config

import (
	"reflect"
	"strings"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

type Config struct {
	Region         string `mapstructure:"region"`
	NonInteractive bool   `mapstructure:"non_interactive"`
	DryRun         bool   `mapstructure:"dry_run"`

	Tailscale TailscaleConfig `mapstructure:"tailscale"`
	Create    CreateConfig    `mapstructure:"create"`
	Stop      StopConfig      `mapstructure:"stop"`
}

type CreateConfig struct {
	Connect  bool   `mapstructure:"connect"`
	Shutdown string `mapstructure:"shutdown"`
}

type TailscaleConfig struct {
	AuthKey string `mapstructure:"auth_key"`
	APIKey  string `mapstructure:"api_key"`
	Tailnet string `mapstructure:"tailnet"`
}

type StopConfig struct {
	All bool `mapstructure:"all"`
}

type Policy struct {
	ACLs                []ACL               `json:"acls,omitempty"`
	Hosts               map[string]string   `json:"hosts,omitempty"`
	Groups              map[string][]string `json:"groups,omitempty"`
	Tests               []Test              `json:"tests,omitempty"`
	TagOwners           map[string][]string `json:"tagOwners,omitempty"`
	AutoApprovers       AutoApprovers       `json:"autoApprovers,omitempty"`
	SSH                 []SSHConfiguration  `json:"ssh,omitempty"`
	DerpMap             DerpMap             `json:"derpMap,omitempty"`
	DisableIPv4         bool                `json:"disableIPv4,omitempty"`
	RandomizeClientPort bool                `json:"randomizeClientPort,omitempty"`
}

type ACL struct {
	Action string   `json:"action,omitempty"`
	Src    []string `json:"src,omitempty"`
	Dst    []string `json:"dst,omitempty"`
	Proto  string   `json:"proto,omitempty"`
}

type Test struct {
	Src    string   `json:"src,omitempty"`
	Accept []string `json:"accept,omitempty"`
	Deny   []string `json:"deny,omitempty"`
}

type AutoApprovers struct {
	Routes   map[string][]string `json:"routes,omitempty"`
	ExitNode []string            `json:"exitNode,omitempty"`
}

type SSHConfiguration struct {
	Action string   `json:"action,omitempty"`
	Src    []string `json:"src,omitempty"`
	Dst    []string `json:"dst,omitempty"`
	Users  []string `json:"users,omitempty"`
}

type DerpMap struct {
	Regions map[string]DerpRegion `json:"regions,omitempty"`
}

type DerpRegion struct {
	RegionID int    `json:"regionID,omitempty"`
	HostName string `json:"hostName,omitempty"`
}

type Node struct {
	Addresses                 []string `json:"addresses"`
	Authorized                bool     `json:"authorized"`
	BlocksIncomingConnections bool     `json:"blocksIncomingConnections"`
	ClientVersion             string   `json:"clientVersion"`
	Created                   string   `json:"created"`
	Expires                   string   `json:"expires"`
	Hostname                  string   `json:"hostname"`
	ID                        string   `json:"id"`
	IsExternal                bool     `json:"isExternal"`
	KeyExpiryDisabled         bool     `json:"keyExpiryDisabled"`
	LastSeen                  string   `json:"lastSeen"`
	MachineKey                string   `json:"NodeKey,omitempty"`
	Name                      string   `json:"name,omitempty"`
	NodeID                    string   `json:"nodeId"`
	NodeKey                   string   `json:"nodeKey"`
	OS                        string   `json:"os"`
	TailnetLockError          string   `json:"tailnetLockError,omitempty"`
	TailnetLockKey            string   `json:"tailnetLockKey,omitempty"`
	UpdateAvailable           bool     `json:"updateAvailable"`
	User                      string   `json:"user,omitempty"`
	Tags                      []string `json:"tags,omitempty"`
}

type TailscaleStatus struct {
	ControlURL             string `json:"ControlURL"`
	RouteAll               bool   `json:"RouteAll"`
	AllowSingleHosts       bool   `json:"AllowSingleHosts"`
	ExitNodeID             string `json:"ExitNodeID"`
	ExitNodeIP             string `json:"ExitNodeIP"`
	ExitNodeAllowLANAccess bool   `json:"ExitNodeAllowLANAccess"`
	CorpDNS                bool   `json:"CorpDNS"`
	RunSSH                 bool   `json:"RunSSH"`
	WantRunning            bool   `json:"WantRunning"`
	LoggedOut              bool   `json:"LoggedOut"`
	ShieldsUp              bool   `json:"ShieldsUp"`
	AdvertiseTags          string `json:"AdvertiseTags"`
	Hostname               string `json:"Hostname"`
	NotepadURLs            bool   `json:"NotepadURLs"`
	AdvertiseRoutes        string `json:"AdvertiseRoutes"`
	NoSNAT                 bool   `json:"NoSNAT"`
	NetfilterMode          int    `json:"NetfilterMode"`
	Config                 struct {
		PrivateMachineKey string `json:"PrivateMachineKey"`
		PrivateNodeKey    string `json:"PrivateNodeKey"`
		OldPrivateNodeKey string `json:"OldPrivateNodeKey"`
		Provider          string `json:"Provider"`
		LoginName         string `json:"LoginName"`
		UserProfile       struct {
			ID            int64    `json:"ID"`
			LoginName     string   `json:"LoginName"`
			DisplayName   string   `json:"DisplayName"`
			ProfilePicURL string   `json:"ProfilePicURL"`
			Roles         []string `json:"Roles"`
		} `json:"UserProfile"`
		NetworkLockKey string `json:"NetworkLockKey"`
		NodeID         string `json:"NodeID"`
	} `json:"Config"`
}

type UserNodes struct {
	User  string `json:"user"`
	Nodes []Node `json:"devices"`
}

func (c *Config) Load(flags *pflag.FlagSet, cmdName string) error {
	v := viper.New()

	// Tailout looks for configuration files called config.yaml, config.json,
	// config.toml, config.hcl, etc.
	v.SetConfigName("config")

	// Tailout looks for configuration files in the common configuration
	// directories.
	v.AddConfigPath("/etc/tailout/")
	v.AddConfigPath("$HOME/.tailout/")

	// Viper logs the configuration file it uses, if any.
	err := v.ReadInConfig()
	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return err
		}
	}

	// Tailout can be configured with environment variables that start with
	// TAILOUT_.
	v.SetEnvPrefix("tailout")
	v.AutomaticEnv()

	// Options with dashes in flag names have underscores when set inside a
	// configuration file or with environment variables.
	flags.SetNormalizeFunc(func(fs *pflag.FlagSet, name string) pflag.NormalizedName {
		name = strings.ReplaceAll(name, "-", "_")
		return pflag.NormalizedName(name)
	})

	// Nested configuration options set with environment variables use an
	// underscore as a separator.
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	bindEnvironmentVariables(v, *c)

	v.BindPFlags(flags)

	// Bind tailscale and command specific nested flags and remove prefix when binding
	// FIXME: This is a workaround for a limitation of Viper, found here:
	// https://github.com/spf13/viper/issues/1072
	flags.Visit(func(f *pflag.Flag) {
		flagName := strings.ReplaceAll(f.Name, "-", "_")
		v.BindPFlag(cmdName+"."+f.Name, flags.Lookup(f.Name))
		if strings.HasPrefix(flagName, "tailscale_") {
			v.BindPFlag("tailscale."+strings.TrimPrefix(flagName, "tailscale_"), flags.Lookup(f.Name))
		}
	})

	// Useful for debugging viper fully-merged configuration
	// spew.Dump(v.AllSettings())

	return v.Unmarshal(c)
}

// bindEnvironmentVariables inspects iface's structure and recursively binds its
// fields to environment variables. This is a workaround to a limitation of
// Viper, found here:
// https://github.com/spf13/viper/issues/188#issuecomment-399884438
func bindEnvironmentVariables(v *viper.Viper, iface interface{}, parts ...string) {
	ifv := reflect.ValueOf(iface)
	ift := reflect.TypeOf(iface)
	for i := 0; i < ift.NumField(); i++ {
		val := ifv.Field(i)
		typ := ift.Field(i)
		tv, ok := typ.Tag.Lookup("mapstructure")
		if !ok {
			continue
		}
		switch val.Kind() {
		case reflect.Struct:
			bindEnvironmentVariables(v, val.Interface(), append(parts, tv)...)
		default:
			v.BindEnv(strings.Join(append(parts, tv), "."))
		}
	}
}
