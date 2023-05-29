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
	cobra.OnInitialize(initConfig)
	// Find home directory.

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.xit.yaml)")
	rootCmd.PersistentFlags().StringP("ts-api-key", "", "", "TailScale API Key")
	rootCmd.PersistentFlags().StringP("ts-auth-key", "", "", "TailScale Auth Key")
	rootCmd.PersistentFlags().StringP("ts-tailnet", "", "", "TailScale Tailnet")
	rootCmd.PersistentFlags().StringP("region", "", "", "AWS Region to create the instance into")
	rootCmd.PersistentFlags().BoolP("dry-run", "", false, "Dry run the command")

	viper.BindPFlag("ts_api_key", rootCmd.PersistentFlags().Lookup("ts-api-key"))
	viper.BindPFlag("ts_auth_key", rootCmd.PersistentFlags().Lookup("ts-auth-key"))
	viper.BindPFlag("ts_tailnet", rootCmd.PersistentFlags().Lookup("ts-tailnet"))
	viper.BindPFlag("region", rootCmd.PersistentFlags().Lookup("region"))
	viper.BindPFlag("dry_run", rootCmd.PersistentFlags().Lookup("dry-run"))

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func initConfig() {
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
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
