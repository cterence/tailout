package config

import (
	"fmt"
	"os"
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

func (c *Config) Load(flags *pflag.FlagSet, cmdName string) error {
	v := viper.New()

	// Xit looks for configuration files called config.yaml, config.json,
	// config.toml, config.hcl, etc.
	v.SetConfigName("config")

	// Xit looks for configuration files in the common configuration
	// directories.
	v.AddConfigPath("/etc/xit/")
	v.AddConfigPath("$HOME/.xit/")

	// Viper logs the configuration file it uses, if any.
	if err := v.ReadInConfig(); err == nil {
		fmt.Fprintf(os.Stderr, "Using config file: %s\n", v.ConfigFileUsed())
	}

	// Xit can be configured with environment variables that start with
	// XIT_.
	v.SetEnvPrefix("xit")
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
