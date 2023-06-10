/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
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
		// 		Long: `A longer description that spans multiple lines and likely contains
		// examples and usage of using your application. For example:

		// Cobra is a CLI library for Go that empowers applications.
		// This application is a tool to generate the needed files
		// to quickly create a Cobra application.`,

		// Uncomment the following line if your bare application
		// has an action associated with it:
		// Run: func(cmd *cobra.Command, args []string) { },
	}
)

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(InitConfig)
	// Find home directory.

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.xit.yaml)")
	rootCmd.PersistentFlags().BoolP("dry-run", "", false, "Dry run the command")

	viper.BindPFlag("dry_run", rootCmd.PersistentFlags().Lookup("dry-run"))
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
	}

	viper.AutomaticEnv()
	err := viper.ReadInConfig()
	if err != nil {
		fmt.Println("Failed to read config:", err)
		return
	}
}
