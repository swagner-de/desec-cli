package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var outputFormat string

var rootCmd = &cobra.Command{
	Use:   "desec-cli",
	Short: "CLI client for the deSEC DNS API",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVarP(&outputFormat, "output", "o", "", "Output format: table, json, yaml (default \"table\")")
}

func initConfig() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("$HOME/.config/desec-cli")
	viper.SetDefault("output", "table")
	viper.SetEnvPrefix("DESEC")
	viper.BindEnv("token")
	viper.ReadInConfig()

	if outputFormat == "" {
		outputFormat = viper.GetString("output")
	}
}

func getToken() string {
	token := viper.GetString("token")
	if token == "" {
		fmt.Fprintln(os.Stderr, "Error: no API token configured. Set DESEC_TOKEN or run \"desec-cli config init\".")
		os.Exit(1)
	}
	return token
}
