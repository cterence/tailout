package config

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

type Config struct {
	Tailscale      TailscaleConfig `mapstructure:"tailscale"`
	UI             UIConfig        `mapstructure:"ui"`
	Region         string          `mapstructure:"region"`
	Create         CreateConfig    `mapstructure:"create"`
	NonInteractive bool            `mapstructure:"non_interactive"`
	DryRun         bool            `mapstructure:"dry_run"`
	Stop           StopConfig      `mapstructure:"stop"`
}

type CreateConfig struct {
	Shutdown string `mapstructure:"shutdown"`
	Connect  bool   `mapstructure:"connect"`
}
type TailscaleConfig struct {
	BaseURL string `mapstructure:"base_url"`
	APIKey  string `mapstructure:"api_key"`
	Tailnet string `mapstructure:"tailnet"`
}

type StopConfig struct {
	All bool `mapstructure:"all"`
}

type UIConfig struct {
	Port    string `mapstructure:"port"`
	Address string `mapstructure:"address"`
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
	v.AddConfigPath(".")

	err := v.ReadInConfig()
	if err != nil {
		var configFileNotFound viper.ConfigFileNotFoundError
		if !errors.As(err, &configFileNotFound) {
			return fmt.Errorf("failed to read configuration file: %w", err)
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

	if err := v.BindPFlags(flags); err != nil {
		return fmt.Errorf("failed to bind flags: %w", err)
	}

	// Bind tailscale and command specific nested flags and remove prefix when binding
	// FIXME: This is a workaround for a limitation of Viper, found here:
	// https://github.com/spf13/viper/issues/1072
	var bindErr error
	flags.Visit(func(f *pflag.Flag) {
		if bindErr != nil {
			return
		}
		flagName := strings.ReplaceAll(f.Name, "-", "_")
		if err := v.BindPFlag(cmdName+"."+f.Name, flags.Lookup(f.Name)); err != nil {
			bindErr = fmt.Errorf("failed to bind flag %s: %w", f.Name, err)
			return
		}
		if strings.HasPrefix(flagName, "tailscale_") {
			if err := v.BindPFlag("tailscale."+strings.TrimPrefix(flagName, "tailscale_"), flags.Lookup(f.Name)); err != nil {
				bindErr = fmt.Errorf("failed to bind tailscale flag %s: %w", f.Name, err)
				return
			}
		}
	})
	if bindErr != nil {
		return bindErr
	}

	// Useful for debugging viper fully-merged configuration
	// spew.Dump(v.AllSettings())

	if err := v.Unmarshal(c); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return nil
}

// bindEnvironmentVariables inspects iface's structure and recursively binds its
// fields to environment variables. This is a workaround to a limitation of
// Viper, found here:
// https://github.com/spf13/viper/issues/188#issuecomment-399884438
func bindEnvironmentVariables(v *viper.Viper, iface interface{}, parts ...string) {
	ifv := reflect.ValueOf(iface)
	ift := reflect.TypeOf(iface)
	for i := range ift.NumField() {
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
			if err := v.BindEnv(strings.Join(append(parts, tv), ".")); err != nil {
				panic(fmt.Sprintf("failed to bind environment variable %s: %v", strings.Join(append(parts, tv), "."), err))
			}
		}
	}
}
