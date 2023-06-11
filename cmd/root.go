package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// rootCmd represents the base command when called without any subcommands
var (
	cfgFile string

	rootCmd = &cobra.Command{
		Use:   "xit",
		Short: "Create an instant exit node in your tailnet",
		Long: `xit is a CLI tool to create an instant exit node in your tailnet.
		
		xit will create a new exit node in your tailnet, and then connect to it. This will allow you to create a VPN tunnel to anywhere in the world.
		
		Example : xit connect xit-eu-west-3-i-048afd4880f66c596`,
	}
)

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(InitConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.xit.yaml)")
}

func InitConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".xit" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".xit")
		viper.SetEnvPrefix("xit")
	}

	viper.AutomaticEnv()
	viper.ReadInConfig()
}
